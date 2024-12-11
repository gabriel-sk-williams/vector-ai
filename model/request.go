package model

import (
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"
	"goyave.dev/goyave/v4/validation"
)

var (
	WorkspaceCreateProps = validation.RuleSet{
		"name":      validation.List{"required", "string"},
		"templates": validation.List{"required", "array:string"},
	}

	DriveListProps = validation.RuleSet{
		"q":       validation.List{"required", "string"},
		"orderBy": validation.List{"required", "string"},
		"limit":   validation.List{"required", "integer"},
	}

	DriveDocumentProps = validation.RuleSet{
		"folders": validation.List{"required", "array:string"},
	}

	WorkspaceConfigProps = validation.RuleSet{
		"value": validation.List{"required", "integer"},
	}

	OrgRoleAssignmentProps = validation.RuleSet{
		"userIds": validation.List{"required", "string"},
	}
)

type ClerkClaims struct {
	jwt.StandardClaims
	TokenType string
}

type ClerkToken struct {
	oauth2.Token
	Audience  string `json:"aud,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	Id        string `json:"jti,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	NotBefore int64  `json:"nbf,omitempty"`
	Subject   string `json:"sub,omitempty"`
}

type WebSocketsMessage struct {
	ID             string         `db:"id" json:"id"`
	WorkspaceID    string         `db:"workspace_id" json:"workspaceId"`
	ConversationID string         `db:"conversation_id" json:"conversationId"`
	TemplateID     sql.NullString `db:"template_id" json:"templateId"`
	QueryText      string         `db:"query_text" json:"queryText"`
	VssText        string         `db:"vss_text" json:"vssText"`
	ForceContext   string         `db:"force_context" json:"forceContext"`
	ResponseSchema string         `db:"response_schema" json:"responseSchema"`
	AuthorType     string         `db:"author_type" json:"authorType"`
	AuthorName     string         `db:"author_name" json:"authorName"`
	Timestamp      time.Time      `db:"timestamp" json:"timestamp"`
}

type UserWebsocketEnvelope struct {
	Message      *WebSocketsMessage `json:"message"`
	DriveFolders *DriveFolders      `json:"driveFolders"`
	SyncFolders  *SyncFolders       `json:"syncFolders"`
	Token        *string            `json:"token"`
	// WebsocketFile *WebsocketFile     `json:"file"`
}

type DriveFolders struct {
	Folders []string `json:"folders"`
}

type SyncFolders struct {
	Folders []string `json:"folders"`
}

type FileHeader struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}

type VssOptions struct {
	VssDocumentLimit uint32 `json:"vssDocumentLimit"`
	VssChunkLimit    uint32 `json:"vssChunkLimit"`
}
