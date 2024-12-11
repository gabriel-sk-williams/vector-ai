package model

import (
	"encoding/json"
	"log"
)

type Envelope struct {
	Data           []byte `json:"data"`
	UpdateType     string `json:"updateType"`
	WorkspaceID    string `json:"workspaceId"`
	ConversationID string `json:"conversationId"`
}

type StatusUpdate struct {
	Text string `json:"text"`
}

type ProgressUpdate struct {
	DocumentID string      `json:"documentId"`
	Event      UploadEvent `json:"uploadEvent"`
	Progress   int         `json:"progress"`
}

type AuthError struct {
	UserID       string `json:"userId"`
	Status       int    `json:"status"`
	ErrorMessage string `json:"errorMessage"`
}

// happening in SyncDrive and UploadDrive
func TokenError(errorMessage string, status int, workspaceId string, userId string) Envelope {
	update, err := json.Marshal(AuthError{UserID: userId, Status: status, ErrorMessage: errorMessage})
	check(err)

	return Envelope{Data: update, UpdateType: "AuthError", WorkspaceID: workspaceId, ConversationID: "nil"}
}

func NotAuthorized(errorMessage string, status int, workspaceId string, userId string) Envelope {
	update, err := json.Marshal(AuthError{UserID: userId, Status: status, ErrorMessage: errorMessage})
	check(err)

	return Envelope{Data: update, UpdateType: "AuthError", WorkspaceID: workspaceId, ConversationID: "nil"}
}

func UserResponse(msg Message, workspaceId string, conversationId string) Envelope {
	data, err := json.Marshal(msg)
	check(err)

	return Envelope{Data: data, UpdateType: "UserResponse", WorkspaceID: workspaceId, ConversationID: conversationId}
}

func AiResponse(msg Message, workspaceId string, conversationId string) Envelope {
	data, err := json.Marshal(msg)
	check(err)

	return Envelope{Data: data, UpdateType: "AIResponse", WorkspaceID: workspaceId, ConversationID: conversationId}
}

func AiStreamChunk(chunk []byte, workspaceId string, conversationId string) Envelope {
	return Envelope{Data: chunk, UpdateType: "AIStreamChunk", WorkspaceID: workspaceId, ConversationID: conversationId}
}

func QueryStatus(text string, workspaceId string, conversationId string) Envelope {
	update, err := json.Marshal(StatusUpdate{Text: text})
	check(err)

	return Envelope{Data: update, UpdateType: "QueryStatus", WorkspaceID: workspaceId, ConversationID: conversationId}
}

func VssResponse(context ContextHolder, workspaceId string, conversationId string) Envelope {
	vssJson, err := json.Marshal(context)
	check(err)

	return Envelope{Data: vssJson, UpdateType: "VssResponse", WorkspaceID: workspaceId, ConversationID: conversationId}
}

// upload event
func UploadStatus(event UploadEvent, workspaceId string, documentId string, progress int) Envelope {
	update, err := json.Marshal(ProgressUpdate{DocumentID: documentId, Event: event, Progress: progress})
	check(err)

	return Envelope{Data: update, UpdateType: "UploadStatus", WorkspaceID: workspaceId, ConversationID: "nil"}
}

// upload event
func ErrorStatus(event UploadEvent, workspaceId string, documentId string, progress int) Envelope {
	update, err := json.Marshal(ProgressUpdate{DocumentID: documentId, Event: event, Progress: progress})
	check(err)

	return Envelope{Data: update, UpdateType: "ErrorStatus", WorkspaceID: workspaceId, ConversationID: "nil"}
}

func UploadManifest(stage string, workspaceId string, records []FolderRecord) Envelope {
	manifest, err := json.Marshal(Manifest{Stage: stage, FolderRecords: records})
	check(err)

	return Envelope{Data: manifest, UpdateType: "UploadManifest", WorkspaceID: workspaceId, ConversationID: "nil"}
}

func SendManifest(stage string, workspaceId string, records map[string]map[string]FileRecord) Envelope {
	manifest, err := json.Marshal(MapManifest{Stage: stage, FolderRecords: records})
	check(err)

	return Envelope{Data: manifest, UpdateType: "UploadManifest", WorkspaceID: workspaceId, ConversationID: "nil"}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
