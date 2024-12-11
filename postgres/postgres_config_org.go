package postgres

import (
	"context"
	"vector-ai/model"

	"github.com/google/uuid"
)

// List all org configs associated with an id
func (pgx Pgx) ListOrgConfigs(orgId string) ([]model.OrgConfig, error) {
	orgConfigs := []model.OrgConfig{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, configuration_id, org_id, property, value from org_config 
	WHERE org_id=$1`, orgId)

	if err != nil {
		return orgConfigs, err
	}
	defer rows.Close()

	for rows.Next() {
		var config model.OrgConfig
		if err := rows.Scan(&config.ID, &config.ConfigurationID, &config.OrgID, &config.Property, &config.Value); err != nil {
			return []model.OrgConfig{}, err
		}
		orgConfigs = append(orgConfigs, config)
	}

	return orgConfigs, err
}

// Retrieves a specific config by ID.
func (pgx Pgx) GetOrgConfig(configId string) (model.OrgConfig, error) {
	var config model.OrgConfig

	if err := pgx.Driver.QueryRow(context.Background(), `SELECT id, configuration_id, org_id, property, value FROM org_config WHERE id=$1`, configId).Scan(&config.ID, &config.ConfigurationID, &config.OrgID, &config.Property, &config.Value); err != nil {
		return config, err
	}
	return config, nil
}

func (pgx Pgx) CreateOrgConfig(configurationId string, orgId string, property string, value string) (model.OrgConfig, error) {
	uuid := uuid.New()

	commandTag, err := pgx.Driver.Exec(context.Background(),
		`INSERT INTO org_config (id, configuration_id, org_id, property, value) VALUES ($1, $2, $3, $4, $5)
		`, uuid, configurationId, orgId, property, value)

	if err != nil || commandTag.RowsAffected() != 1 {
		var orgConfig model.OrgConfig
		return orgConfig, err
	}

	return pgx.GetOrgConfig(uuid.String())
}

// Update an existing org config value by ID.
func (pgx Pgx) UpdateOrgConfig(configId string, value string) (model.OrgConfig, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"UPDATE org_config SET value=$1 WHERE id=$2", value, configId)

	if err != nil || commandTag.RowsAffected() != 1 {
		var config model.OrgConfig
		return config, err
	}

	return pgx.GetOrgConfig(configId)
}

// Deletes a orgConfig by ID.
func (pgx Pgx) DeleteOrgConfig(configId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), `DELETE FROM org_config WHERE id=$1`, configId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}

// Deletes all configs associated with an org.
func (pgx Pgx) ClearOrgConfigs(orgId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), `DELETE FROM org_config WHERE org_id=$1`, orgId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
