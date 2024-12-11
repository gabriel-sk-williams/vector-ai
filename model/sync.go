package model

import "github.com/google/uuid"

type SyncTree struct {
	FolderSyncs   []DriveFolderSync   `json:"folderSyncs"`
	DocumentSyncs []DriveDocumentSync `json:"documentSyncs"`
	Files         []Document          `json:"files"`
	ParentageMap  map[string]string   `json:"parentageMap"`
}

type DriveSyncReport struct {
	SyncedFolders []FolderReport `json:"syncedFolders"`
}

type FolderReport struct {
	DriveID    string     `json:"driveId"`
	Name       string     `json:"name"`
	SyncReport SyncReport `json:"syncReport"`
}

type SyncReport struct {
	New     []NewDriveItem     `json:"new"`
	Updated []UpdatedDriveItem `json:"updated"`
	Missing []MissingDriveItem `json:"missing"`
}

type SyncProfile struct {
	FolderID string
	New      []NewDriveProfile
	Updated  []UpdatedDriveProfile
	Missing  []MissingDriveProfile
}

func (dsr DriveSyncReport) Entries() (int, int, int, int) {
	var total int
	var new int
	var updated int
	var missing int
	for _, folderReport := range dsr.SyncedFolders {
		sr := folderReport.SyncReport
		total += len(sr.New) + len(sr.Updated) + len(sr.Missing)
		new += len(sr.New)
		updated += len(sr.Updated)
		missing += len(sr.Missing)
	}
	return total, new, updated, missing
}

func (s SyncReport) Entries() int {
	return len(s.New) + len(s.Updated) + len(s.Missing)
}

func (s SyncReport) ManifestNew(workspaceId string) ([]FileRecord, []NewDriveProfile) {
	fileRecords := make([]FileRecord, len(s.New))
	newDriveProfiles := make([]NewDriveProfile, len(s.New))

	for i, item := range s.New {
		uuid := uuid.New()

		mData := ManifestData{
			ID:                uuid,
			DocumentID:        uuid.String(),
			WorkspaceID:       workspaceId,
			CoreDocumentProps: item.CoreDocumentProps,
		}

		fileRecords[i] = FileRecord{
			ManifestData:        mData,
			EventStream:         EventStream{},
			OperationSuccessful: false,
		}

		newDriveProfiles[i] = NewDriveProfile{
			ManifestData: mData,
			DriveOrigin:  item.DriveOrigin,
			LastModified: item.LastModified,
		}
	}

	return fileRecords, newDriveProfiles
}

func (s SyncReport) ManifestUpdated(workspaceId string) ([]FileRecord, []UpdatedDriveProfile) {
	fileRecords := make([]FileRecord, len(s.Updated))
	updatedDriveProfiles := make([]UpdatedDriveProfile, len(s.Updated))

	for i, item := range s.Updated {

		uuid, err := uuid.Parse(item.DocumentID)
		check(err)

		mData := ManifestData{
			ID:                uuid,
			DocumentID:        uuid.String(),
			WorkspaceID:       workspaceId,
			CoreDocumentProps: item.CoreDocumentProps,
		}

		fileRecords[i] = FileRecord{
			ManifestData:        mData,
			EventStream:         EventStream{},
			OperationSuccessful: false,
		}

		updatedDriveProfiles[i] = UpdatedDriveProfile{
			ManifestData: mData,
			DriveOrigin:  item.DriveOrigin,
			SyncID:       item.SyncID,
			LastModified: item.LastModified,
		}
	}

	return fileRecords, updatedDriveProfiles
}

func (s SyncReport) ManifestMissing(workspaceId string) ([]FileRecord, []MissingDriveProfile) {
	fileRecords := make([]FileRecord, len(s.Missing))
	missingDriveProfiles := make([]MissingDriveProfile, len(s.Missing))

	for i, item := range s.Missing {

		uuid, err := uuid.Parse(item.DocumentID)
		check(err)

		mData := ManifestData{
			ID:                uuid,
			DocumentID:        uuid.String(),
			WorkspaceID:       workspaceId,
			CoreDocumentProps: item.CoreDocumentProps,
		}

		fileRecords[i] = FileRecord{
			ManifestData:        mData,
			EventStream:         EventStream{},
			OperationSuccessful: false,
		}

		missingDriveProfiles[i] = MissingDriveProfile{
			ManifestData: mData,
			SyncID:       item.SyncID,
		}
	}

	return fileRecords, missingDriveProfiles
}

func (s SyncReport) IsEmpty() bool {
	return len(s.New) == 0 && len(s.Updated) == 0 && len(s.Missing) == 0
}

func (s SyncReport) HasNew() bool {
	return len(s.New) > 0
}

func (s SyncReport) HasUpdated() bool {
	return len(s.Updated) > 0
}

func (s SyncReport) HasMissing() bool {
	return len(s.Missing) > 0
}

func (s SyncProfile) IsEmpty() bool {
	return len(s.New) == 0 && len(s.Updated) == 0 && len(s.Missing) == 0
}

func (s SyncProfile) HasNew() bool {
	return len(s.New) > 0
}

func (s SyncProfile) HasUpdated() bool {
	return len(s.Updated) > 0
}

func (s SyncProfile) HasMissing() bool {
	return len(s.Missing) > 0
}
