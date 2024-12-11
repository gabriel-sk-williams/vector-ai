package postgres

import (
	"vector-ai/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pgx struct {
	Driver *pgxpool.Pool
}

type Controls interface {
	ListInvites() ([]model.Invite, error)
	ListInvitesByOrgId(string) ([]model.Invite, error)
	GetInvite(string) (model.Invite, error)
	CreateInvite(string, string, string) (model.Invite, error)
	DeleteInvite(string) error

	ListOrgs() ([]model.Org, error)
	ListOrgsByUserId(string) ([]model.Org, error)
	CreateOrg(string) (model.Org, error)
	GetOrg(string) (model.Org, error)
	UpdateOrgName(string, string) (model.Org, error)  // UpdateOrg
	UpdateOrgLevel(string, string) (model.Org, error) // UpdateOrg
	DeleteOrg(string) error

	CreateOrgStripeSubscriptionAssociation(string, string, bool) (model.OrgStripeSubscriptionAssociation, error)
	GetOrgStripeSubscriptionAssociationByOrgId(string) (model.OrgStripeSubscriptionAssociation, error)
	GetOrgStripeSubscriptionAssociationByStripeId(string) (model.OrgStripeSubscriptionAssociation, error)
	UpdateOrgStripeSubscriptionAssociation(bool, string) (model.OrgStripeSubscriptionAssociation, error)
	DeleteOrgStripeSubscriptionAssociation(string) error

	ListUserOrgAssociations(string) ([]model.UserOrgAssociation, error)
	ListUserOrgAssociationsByUserId(string) ([]model.UserOrgAssociation, error)
	GetUserOrgAssociation(string, string) (model.UserOrgAssociation, error)
	CreateUserOrgAssociation(string, string) (model.UserOrgAssociation, error)
	DeleteUserOrgAssociation(string, string) error

	ListUsers(string) ([]model.User, error)
	GetUser(string) (model.User, error)
	GetUserName(string) (string, error)
	CreateUser(string, string, string, string) (model.User, error)
	DeleteUser(string) error

	ListOrgRoleAssignments(string) ([]model.OrgRoleAssignment, error)
	ListOrgRoleAssignmentsByUserId(string, string) ([]model.OrgRoleAssignment, error)
	GetOrgRoleAssignmentById(string) (model.OrgRoleAssignment, error)
	GetOrgRoleAssignmentByOrgUser(string, string) (model.OrgRoleAssignment, error)
	CreateOrgRoleAssignment(string, string, string) (model.OrgRoleAssignment, error)
	UpdateOrgRoleAssignment(string, string) (model.OrgRoleAssignment, error)
	DeleteOrgRoleAssignment(string) error

	ListOrgConfigs(string) ([]model.OrgConfig, error)
	GetOrgConfig(string) (model.OrgConfig, error)
	CreateOrgConfig(string, string, string, string) (model.OrgConfig, error)
	UpdateOrgConfig(string, string) (model.OrgConfig, error)
	ClearOrgConfigs(string) error // no route
	DeleteOrgConfig(string) error

	ListWorkspaceConfigs(string) ([]model.WorkspaceConfig, error)
	GetWorkspaceConfig(string, string) (model.WorkspaceConfig, error)
	CreateWorkspaceConfig(string, string, string, int64) (model.WorkspaceConfig, error)
	UpdateWorkspaceConfig(string, string, int64) (model.WorkspaceConfig, error)
	ClearWorkspaceConfigs(string) error // no route
	DeleteWorkspaceConfig(string) error

	ListUserConfigs(string) ([]model.UserConfig, error)
	GetUserConfig(string) (model.UserConfig, error)
	CreateUserConfig(string, string, string, string) (model.UserConfig, error)
	UpdateUserConfig(string, string) (model.UserConfig, error)
	ClearUserConfigs(string) error // no route
	DeleteUserConfig(string) error

	ListWorkspaces(string) ([]model.Workspace, error)
	GetWorkspace(string) (model.Workspace, error)
	CreateWorkspace(string, string) (model.Workspace, error)
	UpdateWorkspace(string, string) (model.Workspace, error)
	DeleteWorkspace(string) error

	ListConversations(string) ([]model.Conversation, error)
	CreateConversation(string, string) (model.Conversation, error)
	GetConversation(string) (model.Conversation, error)
	DeleteConversation(string) error
	ClearMessages(string) error // ClearConversation
	UpdateConversation(string, string) (model.Conversation, error)

	ListMessages(string) ([]model.Message, error)
	CreateMessage(string, string, string, string, string, string, string) (model.Message, error)
	CreateAIMessage(string, string, string, string, string, string) (model.Message, error)
	CreateEmptyMessage(string, string, string) (model.Message, error)
	GetMessage(string) (model.Message, error)
	GetLastMessage(string) (model.Message, error)
	DeleteMessage(string) error

	ListDocuments(string) ([]model.Document, error)
	CreateDocument(uuid.UUID, string, string, string, int64, int64, int64, string) (model.Document, error)
	UpdateDocument(string, string, int64, int64, int64, string) (model.Document, error)
	GetDocument(string) (model.Document, error)
	GetTotalFileSizeAmount(string) (int64, error)
	ClearDocuments(string) error
	DeleteDocument(string) error

	ListDriveDocumentSync(string) ([]model.DriveDocumentSync, error)
	ListDriveDocumentSyncByParentId(string, string) ([]model.DriveDocumentSync, error)
	GetDriveDocumentSync(string) (model.DriveDocumentSync, error)
	CreateDriveDocumentSync(string, string, string, string, string, string) (model.DriveDocumentSync, error)
	UpdateDriveDocumentSyncLastModified(string, string) (model.DriveDocumentSync, error)
	DeleteDriveDocumentSync(string) error
	DeleteDriveDocumentSyncByDocumentId(string) error

	ListDriveFolderSync(string) ([]model.DriveFolderSync, error)
	GetDriveFolderSync(string) (model.DriveFolderSync, error)
	CreateDriveFolderSync(string, string, string, string, string, string) (model.DriveFolderSync, error)
	UpdateDriveFolderLastSynced(string, string) (model.DriveFolderSync, error)
	DeleteDriveFolderSync(string) error

	ListContextsByDocumentId(string) ([]model.Context, error)
	ListContextsByMessageId(string) ([]model.Context, error)
	CreateContext(string, string, string) (model.Context, error)
	GetContext(string) (model.Context, error)
	DeleteContext(string) error

	ListTemplates(string, string) ([]model.Template, error)
	CreateTemplate(string, string, string, string, string, string, string) (model.Template, error)
	GetTemplate(string) (model.Template, error)
	UpdateTemplate(string, string, string) (model.Template, error)
	DeleteTemplate(string) error

	ListUsageEvents(string) ([]model.UsageEvent, error)
	GetUsageEvent(string) (model.UsageEvent, error)
	CreateUsageEvent(string, int64, string) (model.UsageEvent, error)
	DeleteUsageEvent(string) error

	ListTags(string) ([]model.Tag, error)
	CreateTag(string, string, string) (model.Tag, error)
	GetTag(string) (model.Tag, error)
	UpdateTag(string, string, string) (model.Tag, error)
	DeleteTag(string) error

	ListDocumentTagAssociations(string) ([]model.DocumentTagAssociation, error)
	ListDocumentsByTagId(string) ([]model.Document, error)
	ListTagsByDocumentId(string) ([]model.Tag, error)
	CreateDocumentTagAssociation(string, string, string) (model.DocumentTagAssociation, error)
	GetDocumentTagAssociation(string, string) (model.DocumentTagAssociation, error)
	DeleteDocumentTagAssociation(string, string) error
}
