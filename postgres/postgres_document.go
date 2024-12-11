package postgres

import (
	"context"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (pgx Pgx) ListDocuments(workspaceId string) ([]model.Document, error) {
	files := []model.Document{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, workspace_id, name, mime_type, size, vectors, chunk_size, timestamp FROM documents WHERE workspace_id=$1`, workspaceId)
	if err != nil {
		return []model.Document{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var file model.Document
		if err := rows.Scan(&file.ID, &file.WorkspaceID, &file.Name, &file.MIMEType, &file.Size, &file.Vectors, &file.ChunkSize, &file.Timestamp); err != nil {
			return []model.Document{}, err
		}
		files = append(files, file)
	}

	return files, err
}

func (pgx Pgx) GetDocument(fileId string) (model.Document, error) {
	var file model.Document
	if err := pgx.Driver.QueryRow(context.Background(), `SELECT id, workspace_id, name, mime_type, size, vectors, chunk_size, timestamp FROM documents WHERE id=$1`, fileId).Scan(&file.ID, &file.WorkspaceID, &file.Name, &file.MIMEType, &file.Size, &file.Vectors, &file.ChunkSize, &file.Timestamp); err != nil {
		return file, err
	}
	return file, nil
}

func (pgx Pgx) GetTotalFileSizeAmount(orgId string) (int64, error) {
	var totalFileSizeAmount int64

	// Get all userIds associated with orgId from user_org_associations
	// Get all Users from userIds
	if err := pgx.Driver.QueryRow(context.Background(),
		`SELECT COALESCE(SUM(size),0) FROM documents WHERE workspace_id IN 
		(SELECT id from workspaces WHERE org_id=$1)
	`, orgId).Scan(&totalFileSizeAmount); err != nil {
		return totalFileSizeAmount, err
	}

	// COALESCE is a function that will return the first non NULL value from the list.
	// sum(size) was return null when user has zero documents

	return totalFileSizeAmount, nil
}

// no longer in use with randomly generated uuid
func (pgx Pgx) CreateDocument(uuid uuid.UUID, workspaceId string, fileName string, mimeType string, fileSize int64, vectors int64, chunkSize int64, timestamp string) (model.Document, error) {
	// uuid := uuid.New()
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"INSERT INTO documents (id, workspace_id, name, mime_type, size, vectors, chunk_size, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", uuid, workspaceId, fileName, mimeType, fileSize, vectors, chunkSize, timestamp)

	if err != nil || commandTag.RowsAffected() != 1 {
		var file model.Document
		return file, err
	}

	return pgx.GetDocument(uuid.String())
}

func (pgx Pgx) UpdateDocument(documentId string, fileName string, fileSize int64, vectors int64, chunkSize int64, timestamp string) (model.Document, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"UPDATE documents SET name=$1, size=$2, vectors=$3, chunk_size=$4, timestamp=$5 WHERE id=$6", fileName, fileSize, vectors, chunkSize, timestamp, documentId)

	if err != nil || commandTag.RowsAffected() != 1 {
		var document model.Document
		return document, err
	}

	return pgx.GetDocument(documentId)
}

func (pgx Pgx) ClearDocuments(workspaceId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM documents WHERE workspace_id=$1", workspaceId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}

func (pgx Pgx) DeleteDocument(documentId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM documents WHERE id=$1", documentId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
