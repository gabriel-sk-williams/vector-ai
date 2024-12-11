package model

import (
	"database/sql"
	"time"
)

type Invite struct {
	ID         string    `db:"id" json:"id"`
	OrgID      string    `db:"org_id" json:"orgId"`
	Email      string    `db:"email" json:"email"`
	Token      string    `db:"token" json:"token"`
	Expiration time.Time `db:"expiration" json:"expiration"`
}

// Org represents the organization structure.
type Org struct {
	ID    string `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Level int16  `db:"level" json:"level"`
}

type OrgStripeSubscriptionAssociation struct {
	StripeSubscriptionID string `db:"stripe_subscription_id" json:"stripeSubscriptionId"`
	OrgID                string `db:"org_id" json:"orgId"`
	Active               bool   `db:"active" json:"active"`
}

// UserOrgAssociation represents the user-organization association.
type UserOrgAssociation struct {
	OrgID  string `db:"org_id" json:"orgId"`
	UserID string `db:"user_id" json:"userId"`
}

type OrgRoleAssignment struct {
	ID     string `db:"id" json:"id"`
	OrgID  string `db:"org_id" json:"orgId"`
	UserID string `db:"user_id" json:"userId"`
	Role   string `db:"role" json:"role"`
}

// OrgConfig represents the organization configuration.
type OrgConfig struct {
	ID              string `db:"id" json:"id"`
	ConfigurationID string `db:"configuration_id" json:"configurationId"`
	OrgID           string `db:"org_id" json:"orgId"`
	Property        string `db:"property" json:"property"`
	Value           string `db:"value" json:"value"`
}

// UserConfig represents the user configuration.
type UserConfig struct {
	ID              string `db:"id" json:"id"`
	ConfigurationID string `db:"configuration_id" json:"configurationId"`
	UserID          string `db:"user_id" json:"userId"`
	Property        string `db:"property" json:"property"`
	Value           int64  `db:"value" json:"value"`
}

// WorkspaceConfig represents the workspace configuration.
type WorkspaceConfig struct {
	ID              string `db:"id" json:"id"`
	ConfigurationID string `db:"configuration_id" json:"configurationId"`
	WorkspaceID     string `db:"workspace_id" json:"workspaceId"`
	Property        string `db:"property" json:"property"`
	Value           int64  `db:"value" json:"value"`
}

// Configuration represents a generic configuration entity.
type Configuration struct {
	ID              string `db:"id" json:"id"`
	Property        string `db:"property" json:"property"`
	OrgConfig       bool   `db:"org_config" json:"orgConfig"`
	UserConfig      bool   `db:"user_config" json:"userConfig"`
	WorkspaceConfig bool   `db:"workspace_config" json:"workspaceConfig"`
}

// Workspace represents the workspaces associated with an organization.
type Workspace struct {
	ID    string `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	OrgID string `db:"org_id" json:"orgId"`
}

// Conversation represents a series of messages
type Conversation struct {
	ID          string `db:"id" json:"id"`
	WorkspaceID string `db:"workspace_id" json:"workspaceId"`
	Name        string `db:"name" json:"name"`
}

type Message struct {
	ID             string         `db:"id" json:"id"`
	WorkspaceID    string         `db:"workspace_id" json:"workspaceId"`
	ConversationID string         `db:"conversation_id" json:"conversationId"`
	TemplateID     sql.NullString `db:"template_id" json:"templateId"`
	Text           string         `db:"text" json:"text"`
	AuthorType     string         `db:"author_type" json:"authorType"`
	AuthorName     string         `db:"author_name" json:"authorName"`
	Timestamp      time.Time      `db:"timestamp" json:"timestamp"`
}

type Document struct {
	ID          string    `db:"id" json:"id"`
	WorkspaceID string    `db:"workspace_id" json:"workspaceId"`
	Name        string    `db:"name" json:"name"`
	MIMEType    string    `db:"mime_type" json:"mimeType"`
	Size        int64     `db:"size" json:"size"`
	Vectors     int64     `db:"vectors" json:"vectors,omitempty"`
	ChunkSize   int64     `db:"chunk_size" json:"chunkSize,omitempty"`
	Timestamp   time.Time `db:"timestamp" json:"timestamp,omitempty"`
}

type DriveDocumentSync struct {
	ID          string `db:"id" json:"id"`
	WorkspaceID string `db:"workspace_id" json:"workspaceId"`
	DocumentID  string `db:"document_id" json:"documentId"`
	DriveOrigin
	LastModified time.Time `db:"last_modified" json:"lastModified"`
}

type DriveFolderSync struct {
	ID          string `db:"id" json:"id"`
	WorkspaceID string `db:"workspace_id" json:"workspaceId"`
	Name        string `db:"name" json:"name"`
	DriveOrigin
	LastSynced time.Time `db:"last_synced" json:"lastSynced"`
}

type Template struct {
	ID          string    `db:"id" json:"id"`
	OrgID       string    `db:"org_id" json:"orgId"`
	WorkspaceID string    `db:"workspace_id" json:"workspaceId"`
	Name        string    `db:"name" json:"name"`
	Text        string    `db:"text" json:"text"`
	UserID      string    `db:"user_id" json:"userId"`
	AuthorName  string    `db:"author_name" json:"authorName"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
}

type Context struct {
	ID         string `db:"id" json:"id"`
	DocumentID string `db:"document_id" json:"documentId"`
	MessageID  string `db:"message_id" json:"messageId"`
	PointID    string `db:"point_id" json:"pointId"`
}

type User struct {
	ID        string `db:"id" json:"id"`
	FirstName string `db:"first_name" json:"firstName"`
	LastName  string `db:"last_name" json:"lastName"`
	Email     string `db:"email" json:"email"`
}

type UsageEvent struct {
	ID                  string    `db:"id" json:"id"`
	OrgID               string    `db:"org_id" json:"orgId"`
	TotalFileSizeAmount int64     `db:"total_file_size_amount" json:"totalFileSizeAmount"`
	Timestamp           time.Time `db:"timestamp" json:"timestamp"`
}

type Tag struct {
	ID    string `db:"id" json:"id"`
	OrgID string `db:"org_id" json:"orgId"`
	Name  string `db:"name" json:"name"`
	Color string `db:"color" json:"color"`
}

type DocumentTagAssociation struct {
	WorkspaceID string `db:"workspace_id" json:"workspaceId"`
	DocumentID  string `db:"document_id" json:"documentId"`
	TagID       string `db:"tag_id" json:"tagId"`
}

// not in use yet
type TagParentAssociation struct {
	ParentTagID string `db:"parent_tag_id" json:"parentTagId"`
	ChildTagID  string `db:"child_tag_id" json:"childTagId"`
}
