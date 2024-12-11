package model

type HTTPResponse struct {
	Message string `json:"message"`
}

type ConsumableContext struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	DocumentID  string `json:"documentId"`
	Text        string `json:"text"`
}

type AdminResponse struct {
	IsAdmin bool `json:"isAdmin"`
}

type DriveError struct {
	Status       int    `json:"status"`
	ErrorMessage string `json:"errorMessage"`
}

type DriveList struct {
	ErrorStatus bool
	Files       []SharedDrive
}

type GoogleDriveFileList struct {
	ErrorStatus bool
	Files       []GoogleDriveFile
}

type SharedDrive struct {
	DriveID string `json:"driveId"`
	Name    string `json:"name"`
}

type WorkspaceResponse struct {
	Workspace
	LastMessage       Message `json:"lastMessage"`
	NumberOfDocuments int64   `json:"numberOfDocuments"`
	DocumentsSize     int64   `json:"documentsSize"`
}
