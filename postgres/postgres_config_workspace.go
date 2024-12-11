package postgres

import (
	"context"
	"vector-ai/model"

	"github.com/google/uuid"
)

// List all workspace configs associated with an id
func (pgx Pgx) ListWorkspaceConfigs(workspaceId string) ([]model.WorkspaceConfig, error) {
	workspaceConfigs := []model.WorkspaceConfig{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, configuration_id, workspace_id, property, value from workspace_config 
	WHERE workspace_id=$1`, workspaceId)

	if err != nil {
		return workspaceConfigs, err
	}
	defer rows.Close()

	for rows.Next() {
		var config model.WorkspaceConfig
		if err := rows.Scan(&config.ID, &config.ConfigurationID, &config.WorkspaceID, &config.Property, &config.Value); err != nil {
			return []model.WorkspaceConfig{}, err
		}
		workspaceConfigs = append(workspaceConfigs, config)
	}

	return workspaceConfigs, err
}

// Retrieves a specific config by ID.
func (pgx Pgx) GetWorkspaceConfig(workspaceId string, property string) (model.WorkspaceConfig, error) {
	var config model.WorkspaceConfig

	if err := pgx.Driver.QueryRow(context.Background(), `SELECT id, configuration_id, workspace_id, property, value FROM workspace_config WHERE workspace_id=$1 AND property=$2`, workspaceId, property).Scan(&config.ID, &config.ConfigurationID, &config.WorkspaceID, &config.Property, &config.Value); err != nil {
		return config, err
	}
	return config, nil
}

func (pgx Pgx) CreateWorkspaceConfig(configurationId string, workspaceId string, property string, value int64) (model.WorkspaceConfig, error) {
	uuid := uuid.New()

	commandTag, err := pgx.Driver.Exec(context.Background(),
		`INSERT INTO workspace_config (id, configuration_id, workspace_id, property, value) VALUES ($1, $2, $3, $4, $5)
		`, uuid, configurationId, workspaceId, property, value)

	if err != nil || commandTag.RowsAffected() != 1 {
		var workspaceConfig model.WorkspaceConfig
		return workspaceConfig, err
	}

	return pgx.GetWorkspaceConfig(workspaceId, property)
}

// Update an existing workspace config value by ID.
func (pgx Pgx) UpdateWorkspaceConfig(workspaceId string, property string, value int64) (model.WorkspaceConfig, error) {

	commandTag, err := pgx.Driver.Exec(context.Background(),
		"UPDATE workspace_config SET value=$1 WHERE workspace_id=$2 AND property=$3", value, workspaceId, property)

	if err != nil || commandTag.RowsAffected() != 1 {
		var config model.WorkspaceConfig
		return config, err
	}

	return pgx.GetWorkspaceConfig(workspaceId, property)
}

// Deletes a WorkspaceConfig by ID.
func (pgx Pgx) DeleteWorkspaceConfig(configId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), `DELETE FROM workspace_config WHERE id=$1`, configId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}

// Deletes all configs associated with a workspaceID.
func (pgx Pgx) ClearWorkspaceConfigs(workspaceId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), `DELETE FROM workspace_config WHERE workspace_id=$1`, workspaceId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
