package postgres

import (
	"context"
	"vector-ai/model"
)

func (pgx Pgx) ListDocumentTagAssociations(workspaceId string) ([]model.DocumentTagAssociation, error) {
	documentTagAssociations := []model.DocumentTagAssociation{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT workspace_id, document_id, tag_id FROM document_tag_associations WHERE workspace_id=$1
	`, workspaceId)

	if err != nil {
		return documentTagAssociations, err
	}
	defer rows.Close()

	for rows.Next() {
		var documentTagAssociation model.DocumentTagAssociation
		if err := rows.Scan(&documentTagAssociation.WorkspaceID, &documentTagAssociation.DocumentID, &documentTagAssociation.TagID); err != nil {
			return []model.DocumentTagAssociation{}, err
		}
		documentTagAssociations = append(documentTagAssociations, documentTagAssociation)
	}

	return documentTagAssociations, err
}

func (pgx Pgx) ListDocumentsByTagId(tagId string) ([]model.Document, error) {
	documents := []model.Document{}

	rows, err := pgx.Driver.Query(context.Background(), `
	SELECT id, workspace_id, name, mime_type, size, vectors, chunk_size, timestamp FROM documents
	WHERE id=ANY 
	(SELECT document_id FROM document_tag_associations WHERE tag_id=$1)`, tagId)

	if err != nil {
		return documents, err
	}
	defer rows.Close()

	for rows.Next() {
		var file model.Document
		if err := rows.Scan(&file.ID, &file.WorkspaceID, &file.Name, &file.MIMEType, &file.Size, &file.Vectors, &file.ChunkSize, &file.Timestamp); err != nil {
			return []model.Document{}, err
		}
		documents = append(documents, file)
	}

	return documents, err
}

func (pgx Pgx) ListTagsByDocumentId(documentId string) ([]model.Tag, error) {
	tags := []model.Tag{}

	rows, err := pgx.Driver.Query(context.Background(), `
	SELECT id, org_id, name, color FROM tags
	WHERE id=ANY 
	(SELECT tag_id FROM document_tag_associations WHERE document_id=$1)`, documentId)

	if err != nil {
		return tags, err
	}
	defer rows.Close()

	for rows.Next() {
		var file model.Tag
		if err := rows.Scan(&file.ID, &file.OrgID, &file.Name, &file.Color); err != nil {
			return []model.Tag{}, err
		}
		tags = append(tags, file)
	}

	return tags, err
}

func (pgx Pgx) CreateDocumentTagAssociation(workspaceId string, documentId string, tagId string) (model.DocumentTagAssociation, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(), `INSERT INTO document_tag_associations (workspace_id, document_id, tag_id) VALUES ($1, $2, $3)`, workspaceId, documentId, tagId)
	if err != nil || commandTag.RowsAffected() != 1 {
		var documentTag model.DocumentTagAssociation
		return documentTag, err
	}
	return pgx.GetDocumentTagAssociation(documentId, tagId)
}

func (pgx Pgx) GetDocumentTagAssociation(documentId string, tagId string) (model.DocumentTagAssociation, error) {
	var documentTag model.DocumentTagAssociation

	if err := pgx.Driver.QueryRow(context.Background(), `SELECT workspace_id, document_id, tag_id FROM document_tag_associations WHERE document_id=$1 AND tag_id=$2`, documentId, tagId).Scan(&documentTag.WorkspaceID, &documentTag.DocumentID, &documentTag.TagID); err != nil {
		return documentTag, err
	}
	return documentTag, nil
}

// Deletes a specific user-organization association.
func (pgx Pgx) DeleteDocumentTagAssociation(documentId string, tagId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), `DELETE FROM document_tag_associations WHERE document_id=$1 AND tag_id=$2`, documentId, tagId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
