package route

import (
	"time"
	"vector-ai/model"
	"vector-ai/util"
)

func (h Handler) createSyncReport(folderSyncs []model.DriveFolderSync, documentSyncs []model.DriveDocumentSync) (model.DriveSyncReport, error) {

	// ultimately returns modified DriveDocument (because it doesn't exist in syncEntries)
	checkForNewItems := func(syncEntries []model.DriveDocumentSync, currentItems []model.DriveDocument) []model.NewDriveItem {
		newItems := []model.NewDriveItem{}
		for _, item := range currentItems {
			currentId := item.DriveID
			if !util.HasSyncEntry(syncEntries, currentId) {
				ndi := model.NewDriveItem{DriveDocument: item}
				newItems = append(newItems, ndi)
			}
		}

		return newItems
	}

	checkForUpdatedItems := func(syncEntries []model.DriveDocumentSync, currentItems []model.DriveDocument) []model.UpdatedDriveItem {
		updatedItems := []model.UpdatedDriveItem{}
		for _, syncEntry := range syncEntries {

			// get original file size
			document, err := h.PG.GetDocument(syncEntry.DocumentID)
			check(err)

			currentItem, exists := util.GetItemByDriveId(currentItems, syncEntry.DriveID) // get exact item
			if exists {                                                                   // compare
				if syncEntry.LastModified.Format(time.RFC3339) != currentItem.LastModified.Format(time.RFC3339) {
					udi := model.UpdatedDriveItem{
						SyncID:           syncEntry.ID,
						DocumentID:       syncEntry.DocumentID,
						OriginalFileSize: document.Size,
						DriveDocument:    currentItem,
					}
					updatedItems = append(updatedItems, udi)
				}
			}
		}

		return updatedItems
	}

	checkForMissingItems := func(syncEntries []model.DriveDocumentSync, currentItems []model.DriveDocument) []model.MissingDriveItem {
		missingItems := []model.MissingDriveItem{}
		for _, syncEntry := range syncEntries {
			syncId := syncEntry.DriveID
			if !util.HasDriveItem(currentItems, syncId) {
				document, err := h.PG.GetDocument(syncEntry.DocumentID)
				check(err) // 'couldn't get document'

				mdi := model.MissingDriveItem{
					SyncID:     syncEntry.ID,
					DocumentID: syncEntry.DocumentID,
					CoreDocumentProps: model.CoreDocumentProps{
						Name:     document.Name,
						Size:     int64(document.Size),
						MimeType: document.MIMEType,
					},
				}
				missingItems = append(missingItems, mdi)
			}
		}
		return missingItems
	}

	syncedFolders := []model.FolderReport{}

	folderIds := util.MapFolderIds(folderSyncs)

	for _, folderId := range folderIds {

		driveDocuments, err := h.DRV.ListChildren(folderId)
		if err != nil {
			return model.DriveSyncReport{}, err
		}
		// driveCheck(err)

		driveSyncs := util.ReduceByParent(documentSyncs, folderId)

		newDocs := checkForNewItems(driveSyncs, driveDocuments)
		updatedDocs := checkForUpdatedItems(driveSyncs, driveDocuments)
		missingDocs := checkForMissingItems(driveSyncs, driveDocuments)

		folder, err := h.DRV.GetDriveDataById(folderId)
		if err != nil {
			return model.DriveSyncReport{}, err
		}
		// driveCheck(err)

		fr := model.FolderReport{
			DriveID: folder.Id,
			Name:    folder.Name,
			SyncReport: model.SyncReport{
				New:     newDocs,
				Updated: updatedDocs,
				Missing: missingDocs,
			},
		}

		if !fr.SyncReport.IsEmpty() {
			syncedFolders = append(syncedFolders, fr)
		}
	}

	return model.DriveSyncReport{SyncedFolders: syncedFolders}, nil
}
