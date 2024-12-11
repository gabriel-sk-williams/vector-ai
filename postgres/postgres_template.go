package postgres

import (
	"context"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (pgx Pgx) ListTemplates(orgId string, workspaceId string) ([]model.Template, error) {
	templates := []model.Template{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, org_id, workspace_id, name, text, user_id, author_name, created_at FROM templates WHERE org_id=$1 AND workspace_id=$2`, orgId, workspaceId)
	if err != nil {
		return []model.Template{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var template model.Template
		if err := rows.Scan(&template.ID, &template.OrgID, &template.WorkspaceID, &template.Name, &template.Text, &template.UserID, &template.AuthorName, &template.CreatedAt); err != nil {
			return []model.Template{}, err
		}
		templates = append(templates, template)
	}

	return templates, err
}

func (pgx Pgx) GetTemplate(templateId string) (model.Template, error) {
	var template model.Template
	if err := pgx.Driver.QueryRow(context.Background(), "SELECT id, org_id, workspace_id, name, text, user_id, author_name, created_at FROM templates WHERE id=$1", templateId).Scan(&template.ID, &template.OrgID, &template.WorkspaceID, &template.Name, &template.Text, &template.UserID, &template.AuthorName, &template.CreatedAt); err != nil {
		return template, err
	}
	return template, nil
}

func (pgx Pgx) CreateTemplate(orgId string, workspaceId string, templateName string, text string, userId string, authorName string, timestamp string) (model.Template, error) {

	uuid := uuid.New()
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"INSERT INTO templates (id, org_id, workspace_id, name, text, user_id, author_name, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", uuid, orgId, workspaceId, templateName, text, userId, authorName, timestamp)

	if err != nil || commandTag.RowsAffected() != 1 {
		var template model.Template
		return template, err
	}

	return pgx.GetTemplate(uuid.String())
}

func (pgx Pgx) UpdateTemplate(templateId string, templateName string, templateText string) (model.Template, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"UPDATE templates SET name=$1, text=$2 WHERE id=$3", templateName, templateText, templateId)

	if err != nil || commandTag.RowsAffected() != 1 {
		var template model.Template
		return template, err
	}

	return pgx.GetTemplate(templateId)
}

func (pgx Pgx) DeleteTemplate(templateId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM templates WHERE id=$1", templateId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
