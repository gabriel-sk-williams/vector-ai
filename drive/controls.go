package drive

import (
	"fmt"
	"net/http"
	"vector-ai/model"

	"google.golang.org/api/drive/v3"
)

const (
	ServiceType = "google"
)

// Google Drive
// OAuth2 scopes used by this API.
const (
	// See, edit, create, and delete all of your Google Drive files
	DriveScope = "https://www.googleapis.com/auth/drive"

	// See, create, and delete its own configuration data in your Google
	// Drive
	DriveAppdataScope = "https://www.googleapis.com/auth/drive.appdata"

	// View your Google Drive apps
	DriveAppsReadonlyScope = "https://www.googleapis.com/auth/drive.apps.readonly"

	// See, edit, create, and delete only the specific Google Drive files
	// you use with this app
	DriveFileScope = "https://www.googleapis.com/auth/drive.file"

	// View and manage metadata of files in your Google Drive
	DriveMetadataScope = "https://www.googleapis.com/auth/drive.metadata"

	// See information about your Google Drive files
	DriveMetadataReadonlyScope = "https://www.googleapis.com/auth/drive.metadata.readonly"

	// View the photos, videos and albums in your Google Photos
	DrivePhotosReadonlyScope = "https://www.googleapis.com/auth/drive.photos.readonly"

	// See and download all your Google Drive files
	DriveReadonlyScope = "https://www.googleapis.com/auth/drive.readonly"

	// Modify your Google Apps Script scripts' behavior
	DriveScriptsScope = "https://www.googleapis.com/auth/drive.scripts"

	FolderMimeType = "application/vnd.google-apps.folder"
)

// Google Docs
const (
	// See, edit, create, and delete all your Google Docs documents
	DocumentsScope = "https://www.googleapis.com/auth/documents"

	// See all your Google Docs documents
	DocumentsReadonlyScope = "https://www.googleapis.com/auth/documents.readonly"
)

type Drv struct {
	Service *drive.Service
}

type Controls interface {
	ListDocs(string, string, int64) ([]model.GoogleDriveFile, error)
	ListSharedDrives() ([]model.SharedDrive, error)
	ListChildren(string) ([]model.DriveDocument, error)
	GetDriveFolderById(string) (model.DriveFolder, error)
	GetDriveDataById(string) (*drive.File, error)
	ExportDriveFile(string, string) (*http.Response, error)
	DownloadDriveFile(string) (*http.Response, error)
}

func check(err error) {
	if err != nil {
		fmt.Println("Google Drive error:", err)
	}
}
