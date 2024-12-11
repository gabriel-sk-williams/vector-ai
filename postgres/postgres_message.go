package postgres

import (
	"context"
	"time"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (pgx Pgx) ListMessages(conversationId string) ([]model.Message, error) {
	messages := []model.Message{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, workspace_id, conversation_id, template_id, text, author_type, author_name, timestamp FROM messages WHERE conversation_id=$1 ORDER BY timestamp ASC
	`, conversationId)
	if err != nil {
		return []model.Message{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var message model.Message
		if err := rows.Scan(&message.ID, &message.WorkspaceID, &message.ConversationID, &message.TemplateID, &message.Text, &message.AuthorType, &message.AuthorName, &message.Timestamp); err != nil {
			return []model.Message{}, err
		}
		messages = append(messages, message)
	}

	return messages, err
}

func (pgx Pgx) GetMessage(messageId string) (model.Message, error) {
	var message model.Message
	if err := pgx.Driver.QueryRow(context.Background(), `SELECT id, workspace_id, conversation_id, template_id, text, author_type, author_name, timestamp FROM messages WHERE id=$1`, messageId).Scan(&message.ID, &message.WorkspaceID, &message.ConversationID, &message.TemplateID, &message.Text, &message.AuthorType, &message.AuthorName, &message.Timestamp); err != nil {
		return message, err
	}
	return message, nil
}

func (pgx Pgx) GetLastMessage(workspaceId string) (model.Message, error) {
	var message model.Message
	if err := pgx.Driver.QueryRow(context.Background(), `SELECT id, workspace_id, conversation_id, template_id, text, author_type, author_name, timestamp FROM messages WHERE workspace_id=$1 ORDER BY timestamp DESC LIMIT 1`, workspaceId).Scan(&message.ID, &message.WorkspaceID, &message.ConversationID, &message.TemplateID, &message.Text, &message.AuthorType, &message.AuthorName, &message.Timestamp); err != nil {

		return model.Message{
			AuthorName: "SYSTEM",
			Text:       "No messages found in this workspace.",
			Timestamp:  time.Now()}, err
		// return message, err
	}
	return message, nil
}

func (pgx Pgx) CreateMessage(workspaceId string, conversationId string, templateId string, text string, authorType string, authorName string, timestamp string) (model.Message, error) {
	uuid := uuid.New()

	commandTag, err := pgx.Driver.Exec(context.Background(),
		"INSERT INTO messages (id, workspace_id, conversation_id, template_id, text, author_type, author_name, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", uuid, workspaceId, conversationId, templateId, text, authorType, authorName, timestamp)

	if err != nil || commandTag.RowsAffected() != 1 {
		var message model.Message
		return message, err
	}

	return pgx.GetMessage(uuid.String())
}

func (pgx Pgx) CreateAIMessage(workspaceId string, conversationId string, text string, authorType string, authorName string, timestamp string) (model.Message, error) {
	uuid := uuid.New()

	commandTag, err := pgx.Driver.Exec(context.Background(),
		"INSERT INTO messages (id, workspace_id, conversation_id, template_id, text, author_type, author_name, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", uuid, workspaceId, conversationId, nil, text, authorType, authorName, timestamp)

	if err != nil || commandTag.RowsAffected() != 1 {
		var message model.Message
		return message, err
	}

	return pgx.GetMessage(uuid.String())
}

func (pgx Pgx) CreateEmptyMessage(workspaceId string, conversationId string, timestamp string) (model.Message, error) {
	uuid := uuid.New()

	commandTag, err := pgx.Driver.Exec(context.Background(),
		"INSERT INTO messages (id, workspace_id, conversation_id, template_id, text, author_type, author_name, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", uuid, workspaceId, conversationId, nil, "", "SYSTEM", "SYSTEM", timestamp)

	if err != nil || commandTag.RowsAffected() != 1 {
		var message model.Message
		return message, err
	}

	return pgx.GetMessage(uuid.String())
}

func (pgx Pgx) ClearMessages(conversationId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM messages WHERE conversation_id=$1", conversationId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}

func (pgx Pgx) DeleteMessage(messageId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM messages WHERE id=$1", messageId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
