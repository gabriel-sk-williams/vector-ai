package postgres

import (
	"context"
	"vector-ai/model"

	"github.com/google/uuid"
)

// List all user configs associated with an id
func (pgx Pgx) ListUserConfigs(userId string) ([]model.UserConfig, error) {
	userConfigs := []model.UserConfig{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, configuration_id, user_id, property, value from user_config 
	WHERE user_id=$1`, userId)

	if err != nil {
		return userConfigs, err
	}
	defer rows.Close()

	for rows.Next() {
		var config model.UserConfig
		if err := rows.Scan(&config.ID, &config.ConfigurationID, &config.UserID, &config.Property, &config.Value); err != nil {
			return []model.UserConfig{}, err
		}
		userConfigs = append(userConfigs, config)
	}

	return userConfigs, err
}

// Retrieves a specific config by ID.
func (pgx Pgx) GetUserConfig(configId string) (model.UserConfig, error) {
	var config model.UserConfig

	if err := pgx.Driver.QueryRow(context.Background(), `SELECT id, configuration_id, user_id, property, value FROM user_config WHERE id=$1`, configId).Scan(&config.ID, &config.ConfigurationID, &config.UserID, &config.Property, &config.Value); err != nil {
		return config, err
	}
	return config, nil
}

func (pgx Pgx) CreateUserConfig(configurationId string, userId string, property string, value string) (model.UserConfig, error) {
	uuid := uuid.New()

	commandTag, err := pgx.Driver.Exec(context.Background(),
		`INSERT INTO user_config (id, configuration_id, user_id, property, value) VALUES ($1, $2, $3, $4, $5)
		`, uuid, configurationId, userId, property, value)

	if err != nil || commandTag.RowsAffected() != 1 {
		var userConfig model.UserConfig
		return userConfig, err
	}

	return pgx.GetUserConfig(uuid.String())
}

// Update an existing user config value by ID.
func (pgx Pgx) UpdateUserConfig(configId string, value string) (model.UserConfig, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"UPDATE user_config SET value=$1 WHERE id=$2", value, configId)

	if err != nil || commandTag.RowsAffected() != 1 {
		var config model.UserConfig
		return config, err
	}

	return pgx.GetUserConfig(configId)
}

// Deletes a userConfig by ID.
func (pgx Pgx) DeleteUserConfig(configId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), `DELETE FROM user_config WHERE id=$1`, configId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}

// Deletes all configs associated with a userID.
func (pgx Pgx) ClearUserConfigs(userId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), `DELETE FROM user_config WHERE user_id=$1`, userId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
