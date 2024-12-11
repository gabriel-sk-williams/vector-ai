package route

import (
	"fmt"
	"net/http"
	c "vector-ai/constants"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (s Session) UploadFile(header model.CoreDocumentProps, data []byte) {

	fmt.Printf("Received file: %s, size: %d bytes mimeType: %s \n", header.Name, len(data), header.MimeType)

	orgId := s.orgId
	workspaceId := s.workspaceId
	cc := s.token.Claims.(*model.ClerkClaims)
	userId := cc.Subject
	folderId := "default"

	options := Options{
		VectorSize:   1536,
		ChunkSize:    500, // default: 200
		ChunkOverlap: 100, // default: 4000
	}

	uuid := uuid.New()
	documentId := uuid.String()

	cdp := model.CoreDocumentProps{
		Name:     header.Name,
		Size:     header.Size,
		MimeType: header.MimeType,
	}

	mrec := model.ManifestData{
		ID:                uuid,
		DocumentID:        uuid.String(),
		WorkspaceID:       workspaceId,
		CoreDocumentProps: cdp,
	}

	record := model.FileRecord{
		ManifestData:        mrec,
		EventStream:         model.EventStream{},
		OperationSuccessful: false,
	}

	nlp := model.NewLocalProfile{
		ManifestData: mrec,
		Data:         data,
	}

	vsp := model.VectorStorageProfile{
		OrgID:       orgId,
		DocumentID:  documentId,
		WorkspaceID: workspaceId,
	}

	_, ok := s.manifest[folderId]

	if ok { // add record to manifest
		s.manifest[folderId][documentId] = record
	} else { // create new []fileRecord
		s.manifest[folderId] = make(map[string]model.FileRecord)
		s.manifest[folderId][documentId] = record
	}

	// check if user is subscribed and limit to 5mb
	subscription, err := s.handler.PG.GetOrgStripeSubscriptionAssociationByOrgId(orgId)

	if !subscription.Active || err != nil {

		currentFileSizeAmount, err := s.handler.PG.GetTotalFileSizeAmount(orgId)
		check(err)

		fmt.Println(nlp.Size)
		fmt.Println(currentFileSizeAmount)
		sum := nlp.Size + currentFileSizeAmount

		stats := fmt.Sprintf(`%d + %d = %d`, nlp.Size, currentFileSizeAmount, sum)
		fmt.Println(stats)

		if currentFileSizeAmount+nlp.Size > c.NonSubscriberFileUploadLimit {
			fmt.Println("Limit Breached")
			s.handler.TR.Broadcast(model.NotAuthorized("User is limited to 5MB upload", http.StatusPaymentRequired, workspaceId, userId))
			return
		}
	}

	s.handler.TR.Broadcast(model.SendManifest("active", workspaceId, s.manifest))

	//maxGoroutines := 4
	//guard := make(chan struct{}, maxGoroutines)
	//var wg sync.WaitGroup

	//guard <- struct{}{}
	//wg.Add(1)

	go func(profile model.NewLocalProfile, vsp model.VectorStorageProfile) {
		// defer wg.Done()
		var evs model.EventStream
		evs, parsedDoc := s.handler.parseLocalUpload(evs, profile)
		evs, chunks := s.handler.splitEmbedUpload(evs, vsp, parsedDoc, s.embedder, options)
		evs = s.handler.saveLocalDocument(evs, profile, chunks, options)

		record.EventStream = evs
		record.OperationSuccessful = true
		s.manifest[folderId][documentId] = record

		if s.allDone() {
			s.handler.TR.Broadcast(model.SendManifest("done", workspaceId, s.manifest))

			// create usage event
			if subscription.Active {
				record, err := s.handler.createUsageEvent(orgId)
				fmt.Println(record)
				check(err)
			}
		}
		//<-guard
	}(nlp, vsp)

	//wg.Wait()
}

func (s Session) allDone() bool {
	for _, fileMap := range s.manifest {
		for _, record := range fileMap {
			if !record.OperationSuccessful {
				return false
			}
		}
	}
	return true
}
