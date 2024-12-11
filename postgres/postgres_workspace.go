package postgres

import (
	"context"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (pgx Pgx) ListWorkspaces(orgId string) ([]model.Workspace, error) {
	workspaces := []model.Workspace{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, name, org_id FROM workspaces WHERE org_id=$1`, orgId)
	if err != nil {
		return []model.Workspace{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var workspace model.Workspace
		if err := rows.Scan(&workspace.ID, &workspace.Name, &workspace.OrgID); err != nil {
			return []model.Workspace{}, err
		}
		workspaces = append(workspaces, workspace)
	}

	return workspaces, err
}

func (pgx Pgx) GetWorkspace(workspaceId string) (model.Workspace, error) {
	var workspace model.Workspace
	if err := pgx.Driver.QueryRow(context.Background(), "SELECT id, name, org_id FROM workspaces WHERE id=$1", workspaceId).Scan(&workspace.ID, &workspace.Name, &workspace.OrgID); err != nil {
		return workspace, err
	}
	return workspace, nil
}

// workspace
func (pgx Pgx) CreateWorkspace(orgId string, workspaceId string) (model.Workspace, error) {
	uuid := uuid.New()
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"INSERT INTO workspaces (id, org_id, name) VALUES ($1, $2, $3)", uuid, orgId, workspaceId)

	if err != nil || commandTag.RowsAffected() != 1 {
		var workspace model.Workspace
		return workspace, err
	}

	return pgx.GetWorkspace(uuid.String())
}

// updates an existing workspace by ID.
func (pgx Pgx) UpdateWorkspace(workspaceId string, workspaceName string) (model.Workspace, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"UPDATE workspaces SET name=$1 WHERE id=$2", workspaceName, workspaceId)

	if err != nil || commandTag.RowsAffected() != 1 {
		var workspace model.Workspace
		return workspace, err
	}

	return pgx.GetWorkspace(workspaceId)
}

// deletes a workspace by ID.
func (pgx Pgx) DeleteWorkspace(workspaceId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM workspaces WHERE id=$1", workspaceId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
