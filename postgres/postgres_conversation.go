package postgres

import (
	"context"
	"fmt"
	"time"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (pgx Pgx) ListConversations(workspaceId string) ([]model.Conversation, error) {
	conversations := []model.Conversation{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, workspace_id, name FROM conversations WHERE workspace_id=$1
	ORDER BY (
		SELECT MAX(messages.timestamp) FROM messages WHERE messages.conversation_id = conversations.id
	) DESC
	`, workspaceId)
	if err != nil {
		return []model.Conversation{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var conversation model.Conversation
		if err := rows.Scan(&conversation.ID, &conversation.WorkspaceID, &conversation.Name); err != nil {
			return []model.Conversation{}, err
		}
		conversations = append(conversations, conversation)
	}

	return conversations, err
}

func (pgx Pgx) GetConversation(conversationId string) (model.Conversation, error) {
	var conversation model.Conversation
	if err := pgx.Driver.QueryRow(context.Background(), `SELECT id, workspace_id, name FROM conversations WHERE id=$1`, conversationId).Scan(&conversation.ID, &conversation.WorkspaceID, &conversation.Name); err != nil {
		return conversation, err
	}
	return conversation, nil
}

// TODO: allow IP address as uuid
func (pgx Pgx) CreateConversation(workspaceId string, conversationName string) (model.Conversation, error) {
	uuid := uuid.New()
	commandTag, err := pgx.Driver.Exec(context.Background(),
		`INSERT INTO conversations (id, workspace_id, name) VALUES ($1, $2, $3)`, uuid, workspaceId, conversationName)

	if err != nil || commandTag.RowsAffected() != 1 {
		var conversation model.Conversation
		return conversation, err
	}

	timestamp := time.Now().Format(time.RFC3339)

	//Create message
	_, err = pgx.CreateEmptyMessage(workspaceId, uuid.String(), timestamp)
	if err != nil {
		fmt.Println("Failed to create empty message in conversation", uuid.String(), err)
	}

	return pgx.GetConversation(uuid.String())
}

// Updates an existing conversation name by ID.
func (pgx Pgx) UpdateConversation(conversationId string, conversationName string) (model.Conversation, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(), `Update conversation SET name=$1 WHERE id=$2`, conversationName, conversationId)
	if err != nil || commandTag.RowsAffected() != 1 {
		var conversation model.Conversation
		return conversation, err
	}
	return pgx.GetConversation(conversationId)
}

// delete conversation and all associated messages
func (pgx Pgx) DeleteConversation(conversationId string) error {

	// delete conversation
	commandTag, err := pgx.Driver.Exec(context.Background(), `DELETE FROM conversations WHERE id=$1`, conversationId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}

	return nil
}
