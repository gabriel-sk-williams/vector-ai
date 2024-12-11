package model

import (
	"time"

	"github.com/google/uuid"
)

type MapManifest struct {
	Stage         string                           `json:"stage"` // init, done
	FolderRecords map[string]map[string]FileRecord `json:"folderRecords"`
}

// Upload Manifest
type Manifest struct {
	Stage         string         `json:"stage"` // init, done
	FolderRecords []FolderRecord `json:"folderRecords"`
}

type FolderRecord struct {
	FolderID    string       `json:"folderId,omitempty"`
	FileRecords []FileRecord `json:"fileRecords"`
}

type FileRecord struct { // TODO
	ManifestData
	EventStream
	OperationSuccessful bool `json:"operationSuccessful"`
}

type ManifestData struct {
	ID          uuid.UUID `json:"id"`
	DocumentID  string
	WorkspaceID string
	CoreDocumentProps
}

type EventStream struct {
	Events []UploadEvent `json:"eventStream"`
}

type UploadEvent struct {
	Operation string `json:"operation"` // Reading, Splitting, Embedding etc
	Action    string `json:"action"`    // Starting, Completed, Failed
	Detail    string `json:"detail"`
}

type FolderProfile struct {
	FolderID     string
	FileProfiles []NewDriveProfile
}

// Internal Drive synchronization
type NewLocalProfile struct {
	ManifestData
	Data []byte
}

type DownloadProfile struct {
	ManifestData
	DriveOrigin
}

type VectorStorageProfile struct {
	OrgID       string
	WorkspaceID string
	DocumentID  string
}

type NewDriveProfile struct {
	ManifestData
	DriveOrigin
	LastModified time.Time
}

type UpdatedDriveProfile struct {
	ManifestData
	DriveOrigin
	SyncID       string
	LastModified time.Time
}

type MissingDriveProfile struct {
	ManifestData
	SyncID string
}

// Drive synchronization
type GoogleDriveFile struct {
	DriveOrigin
	CoreDocumentProps
	ExtendedDriveProps
}

type DriveOrigin struct { // origin location identity
	DriveID          string `db:"drive_id" json:"driveId"`
	DriveParentID    string `db:"drive_parent_id" json:"driveParentId"`
	DriveServiceType string `db:"drive_service_type" json:"driveServiceType"`
	// SharedDriveID string	`json:"sharedDriveId"`
}

type CoreDocumentProps struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}

type ExtendedDriveProps struct {
	Kind          string `json:"kind"`
	ThumbnailLink string `json:"thumbnailLink"`
	FileExtension string `json:"fileExtension"`
	IconLink      string `json:"iconLink"`
}

type NewDriveItem struct {
	DriveDocument
}

type UpdatedDriveItem struct {
	SyncID           string `json:"-"` // for postgres update
	DocumentID       string `json:"-"` // for postgres update
	OriginalFileSize int64  `json:"-"` // TODO: check performance
	DriveDocument
}

type MissingDriveItem struct {
	SyncID            string `json:"-"` // for postgres delete
	DocumentID        string `json:"-"` // for postgres delete
	CoreDocumentProps        // can't have DriveOrigin bc it is missing
}

type DriveDocument struct {
	DriveOrigin
	CoreDocumentProps
	LastModified time.Time `db:"last_modified" json:"lastModified"`
}

type DriveFolder struct {
	DriveOrigin
	CoreDocumentProps
}
