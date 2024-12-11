package postgres

import (
	bg "context"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (pgx Pgx) ListContextsByDocumentId(documentId string) ([]model.Context, error) {
	contexts := []model.Context{}

	rows, err := pgx.Driver.Query(bg.Background(), `SELECT id, document_id, message_id, point_id FROM context WHERE document_id=$1`, documentId)
	if err != nil {
		return []model.Context{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var context model.Context
		if err := rows.Scan(&context.ID, &context.DocumentID, &context.MessageID, &context.PointID); err != nil {
			return []model.Context{}, err
		}
		contexts = append(contexts, context)
	}

	return contexts, err
}

func (pgx Pgx) ListContextsByMessageId(messageId string) ([]model.Context, error) {
	contexts := []model.Context{}

	rows, err := pgx.Driver.Query(bg.Background(), `SELECT id, document_id, message_id, point_id FROM context WHERE message_id=$1`, messageId)
	if err != nil {
		return []model.Context{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var context model.Context
		if err := rows.Scan(&context.ID, &context.DocumentID, &context.MessageID, &context.PointID); err != nil {
			return []model.Context{}, err
		}
		contexts = append(contexts, context)
	}

	return contexts, err
}

func (pgx Pgx) GetContext(contextId string) (model.Context, error) {
	var context model.Context
	if err := pgx.Driver.QueryRow(bg.Background(), "SELECT id, document_id, message_id, point_id FROM context WHERE id=$1", contextId).Scan(&context.ID, &context.DocumentID, &context.MessageID, &context.PointID); err != nil {
		return context, err
	}
	return context, nil
}

func (pgx Pgx) CreateContext(documentId string, messageId string, pointId string) (model.Context, error) {
	uuid := uuid.New()
	commandTag, err := pgx.Driver.Exec(bg.Background(),
		"INSERT INTO context (id, document_id, message_id, point_id) VALUES ($1, $2, $3, $4)", uuid, documentId, messageId, pointId)

	if err != nil || commandTag.RowsAffected() != 1 {
		var context model.Context
		return context, err
	}

	return pgx.GetContext(uuid.String())
}

func (pgx Pgx) DeleteContext(contextId string) error {
	commandTag, err := pgx.Driver.Exec(bg.Background(), "DELETE FROM context WHERE id=$1", contextId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
