package postgres

import (
	"context"
	"vector-ai/model"

	"github.com/google/uuid"
)

// admin only
func (pgx Pgx) ListOrgs() ([]model.Org, error) {
	orgs := []model.Org{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, name, level FROM org`)

	if err != nil {
		return orgs, err
	}
	defer rows.Close()

	for rows.Next() {
		var org model.Org
		if err := rows.Scan(&org.ID, &org.Name, &org.Level); err != nil {
			return []model.Org{}, err
		}
		orgs = append(orgs, org)
	}

	return orgs, err
}

func (pgx Pgx) ListOrgsByUserId(userId string) ([]model.Org, error) {
	orgs := []model.Org{}

	rows, err := pgx.Driver.Query(context.Background(), "SELECT id, name, level FROM org WHERE id=ANY (SELECT org_id FROM user_org_associations WHERE user_id=$1)", userId)

	if err != nil {
		return orgs, err
	}
	defer rows.Close()

	for rows.Next() {
		var org model.Org
		if err := rows.Scan(&org.ID, &org.Name, &org.Level); err != nil {
			return []model.Org{}, err
		}
		orgs = append(orgs, org)
	}

	return orgs, err
}

//
// edit below
//

// GetOrg retrieves a specific org by ID.
func (pgx Pgx) GetOrg(orgId string) (model.Org, error) {
	var org model.Org

	if err := pgx.Driver.QueryRow(context.Background(), "SELECT id, name, level FROM org WHERE id=$1", orgId).Scan(&org.ID, &org.Name, &org.Level); err != nil {
		return org, err
	}
	return org, nil
}

// CreateOrg creates a new org.
func (pgx Pgx) CreateOrg(orgName string) (model.Org, error) {
	uuid := uuid.New()
	commandTag, err := pgx.Driver.Exec(context.Background(), "INSERT INTO org (id, name, level) VALUES ($1, $2, $3)", uuid, orgName, 0) // all orgs start at level 0
	if err != nil || commandTag.RowsAffected() != 1 {
		var org model.Org
		return org, err
	}
	return pgx.GetOrg(uuid.String())
}

// UpdateOrg updates an existing org name by ID.
func (pgx Pgx) UpdateOrgName(orgId string, orgName string) (model.Org, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(), "UPDATE org SET name=$1 WHERE id=$2", orgName, orgId)
	if err != nil || commandTag.RowsAffected() != 1 {
		var org model.Org
		return org, err
	}
	return pgx.GetOrg(orgId)
}

// UpdateOrg updates an existing org level by ID.
func (pgx Pgx) UpdateOrgLevel(orgId string, orgLevel string) (model.Org, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(), "UPDATE org SET level=$1 WHERE id=$2", orgLevel, orgId)
	if err != nil || commandTag.RowsAffected() != 1 {
		var org model.Org
		return org, err
	}
	return pgx.GetOrg(orgId)
}

// DeleteOrg deletes an org by ID.
func (pgx Pgx) DeleteOrg(orgId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM org WHERE id=$1", orgId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
