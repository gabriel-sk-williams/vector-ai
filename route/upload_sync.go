package route

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	c "vector-ai/constants"
	"vector-ai/drive"
	"vector-ai/model"
	"vector-ai/util"
)

func (s Session) SyncDrive(folderIds []string) {
	orgId := s.orgId
	workspaceId := s.workspaceId

	cc := s.token.Claims.(*model.ClerkClaims)
	userId := cc.Subject

	options := Options{
		VectorSize:   1536, // no longer used here
		ChunkSize:    500,  // default: 200
		ChunkOverlap: 100,  // default: 4000
	}

	provider := "oauth_google"
	token, status, err := s.handler.checkToken(userId, provider)

	if err != nil { // send error via websockets
		s.handler.TR.Broadcast(model.TokenError(err.Error(), status, workspaceId, userId))
		return
	}

	config, err := drive.GetDriveConfig()
	check(err)
	service, err := drive.GetDriveService(config, token)
	check(err)
	s.handler.DRV = drive.Drv{Service: service}

	// sync folders ids -> additive and neutral but not subtractive
	if len(folderIds) > 0 {
		for _, folderId := range folderIds {

			df, err := s.handler.DRV.GetDriveFolderById(folderId)
			check(err)

			lastSynced := time.Now().Format(time.RFC3339)
			_, err = s.handler.PG.CreateDriveFolderSync(workspaceId, df.Name, df.DriveID, df.DriveParentID, drive.ServiceType, lastSynced)
			check(err)
		}
	}

	// possibly use folderIds from frontend
	folderSyncs, err := s.handler.PG.ListDriveFolderSync(workspaceId)
	check(err)

	documentSyncs, err := s.handler.PG.ListDriveDocumentSync(workspaceId)
	check(err)

	syncReport, err := s.handler.createSyncReport(folderSyncs, documentSyncs)
	check(err)

	syncedFolders := syncReport.SyncedFolders

	// folderRecords := make([]model.FolderRecord, len(syncedFolders)) // folderIds
	syncProfiles := make([]model.SyncProfile, len(syncedFolders)) // folderIds

	for i, folderReport := range syncedFolders {
		folderId := folderReport.DriveID
		// folderName := folderReport.Name
		syncReport := folderReport.SyncReport

		// check for folderId in manifest
		_, ok := s.manifest[folderId]
		if !ok {
			s.manifest[folderId] = make(map[string]model.FileRecord)
		}

		var records []model.FileRecord
		syncProfile := model.SyncProfile{FolderID: folderId}

		if syncReport.HasNew() {
			newRecords, newProfiles := syncReport.ManifestNew(workspaceId)
			records = append(records, newRecords...)
			syncProfile.New = newProfiles
		}
		if syncReport.HasUpdated() {
			updatedRecords, updatedProfiles := syncReport.ManifestUpdated(workspaceId)
			records = append(records, updatedRecords...)
			syncProfile.Updated = updatedProfiles
		}
		if syncReport.HasMissing() {
			missingRecords, missingProfiles := syncReport.ManifestMissing(workspaceId)
			records = append(records, missingRecords...)
			syncProfile.Missing = missingProfiles
		}

		for _, record := range records {
			documentId := record.ID.String()
			s.manifest[folderId][documentId] = record
		}

		syncProfiles[i] = syncProfile
	}

	// check if user is subscribed and limit to 5mb
	subscription, err := s.handler.PG.GetOrgStripeSubscriptionAssociationByOrgId(orgId)

	if !subscription.Active || err != nil {
		syncSize := util.UploadSize(syncReport)

		currentFileSizeAmount, err := s.handler.PG.GetTotalFileSizeAmount(orgId)
		check(err)

		fmt.Println(syncSize)
		fmt.Println(currentFileSizeAmount)
		sum := syncSize + currentFileSizeAmount

		stats := fmt.Sprintf(`%d + %d = %d`, syncSize, currentFileSizeAmount, sum)
		fmt.Println(stats)

		if currentFileSizeAmount+syncSize > c.NonSubscriberFileUploadLimit {
			fmt.Println("Limit Breached")
			s.handler.TR.Broadcast(model.NotAuthorized("User is limited to 5MB upload", http.StatusPaymentRequired, workspaceId, userId))
			return
		}
	}

	s.handler.TR.Broadcast(model.SendManifest("active", workspaceId, s.manifest))

	maxGoroutines := 4
	guard := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup
	for _, syncProfile := range syncProfiles {
		folderId := syncProfile.FolderID

		if syncProfile.HasNew() {
			for _, profile := range syncProfile.New {

				profile := profile
				dlp := model.DownloadProfile{
					ManifestData: profile.ManifestData,
					DriveOrigin:  profile.DriveOrigin,
				}

				vsp := model.VectorStorageProfile{
					OrgID:       orgId,
					WorkspaceID: profile.WorkspaceID,
					DocumentID:  profile.DocumentID,
				}

				guard <- struct{}{}
				wg.Add(1)

				go func(profile model.NewDriveProfile, dlp model.DownloadProfile, vsp model.VectorStorageProfile) {
					defer wg.Done()
					var evs model.EventStream
					evs, body, exportType := s.handler.downloadDriveFile(evs, dlp)
					evs, parsedDoc := s.handler.parseBody(evs, profile.ManifestData, body, exportType)
					evs, chunks := s.handler.splitEmbedUpload(evs, vsp, parsedDoc, s.embedder, options)
					evs = s.handler.syncNew(evs, profile, chunks, options)

					documentId := profile.DocumentID
					record := s.manifest[folderId][documentId]

					record.EventStream = evs
					record.OperationSuccessful = true
					s.manifest[folderId][documentId] = record

					if s.allDone() {
						s.handler.TR.Broadcast(model.SendManifest("done", workspaceId, s.manifest))
						// then delete all files
					}
					<-guard
				}(profile, dlp, vsp)
			}
		}

		if syncProfile.HasUpdated() {
			for _, profile := range syncProfile.Updated {

				profile := profile
				dlp := model.DownloadProfile{
					ManifestData: profile.ManifestData,
					DriveOrigin:  profile.DriveOrigin,
				}

				vsp := model.VectorStorageProfile{
					OrgID:       orgId,
					WorkspaceID: profile.WorkspaceID,
					DocumentID:  profile.DocumentID,
				}

				guard <- struct{}{}
				wg.Add(1)

				go func(profile model.UpdatedDriveProfile, dlp model.DownloadProfile, vsp model.VectorStorageProfile) {
					defer wg.Done()
					var evs model.EventStream
					evs = s.handler.DeleteVectors(evs, vsp)
					evs, body, exportType := s.handler.downloadDriveFile(evs, dlp)
					evs, parsedDoc := s.handler.parseBody(evs, profile.ManifestData, body, exportType)
					evs, chunks := s.handler.splitEmbedUpload(evs, vsp, parsedDoc, s.embedder, options)
					evs = s.handler.syncUpdated(evs, profile, chunks, options)

					documentId := profile.DocumentID
					record := s.manifest[folderId][documentId]

					record.EventStream = evs
					record.OperationSuccessful = true
					s.manifest[folderId][documentId] = record

					if s.allDone() {
						s.handler.TR.Broadcast(model.SendManifest("done", workspaceId, s.manifest))
						// then delete all files
					}
					<-guard
				}(profile, dlp, vsp)
			}
		}

		if syncProfile.HasMissing() {
			for _, profile := range syncProfile.Missing {
				profile := profile

				vsp := model.VectorStorageProfile{
					OrgID:       orgId,
					WorkspaceID: profile.WorkspaceID,
					DocumentID:  profile.DocumentID,
				}

				guard <- struct{}{}
				wg.Add(1)

				go func(profile model.MissingDriveProfile, vsp model.VectorStorageProfile) {
					defer wg.Done()
					var evs model.EventStream
					evs = s.handler.DeleteVectors(evs, vsp)
					evs = s.handler.syncMissing(evs, profile)

					documentId := profile.DocumentID
					record := s.manifest[folderId][documentId]

					record.EventStream = evs
					record.OperationSuccessful = true
					s.manifest[folderId][documentId] = record

					if s.allDone() {
						s.handler.TR.Broadcast(model.SendManifest("done", workspaceId, s.manifest))
						// then delete all records
					}
					<-guard
				}(profile, vsp)
			}
		}
	}
	wg.Wait()
	//ctx.Done()

	if subscription.Active {
		record, err := s.handler.createUsageEvent(orgId)
		fmt.Println(record)
		check(err)
	}

}
