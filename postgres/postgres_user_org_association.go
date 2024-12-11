package postgres

import (
	"context"
	"vector-ai/model"
)

func (pgx Pgx) ListUserOrgAssociations(orgId string) ([]model.UserOrgAssociation, error) {
	userOrgAssociations := []model.UserOrgAssociation{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT org_id, user_id FROM user_org_associations WHERE org_id=$1`, orgId)

	if err != nil {
		return userOrgAssociations, err
	}
	defer rows.Close()

	for rows.Next() {
		var userOrgAssociation model.UserOrgAssociation
		if err := rows.Scan(&userOrgAssociation.OrgID, &userOrgAssociation.UserID); err != nil {
			return []model.UserOrgAssociation{}, err
		}
		userOrgAssociations = append(userOrgAssociations, userOrgAssociation)
	}

	return userOrgAssociations, err
}

func (pgx Pgx) ListUserOrgAssociationsByUserId(userId string) ([]model.UserOrgAssociation, error) {
	userOrgAssociations := []model.UserOrgAssociation{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT org_id, user_id FROM user_org_associations WHERE user_id=$1`, userId)

	if err != nil {
		return userOrgAssociations, err
	}
	defer rows.Close()

	for rows.Next() {
		var userOrgAssociation model.UserOrgAssociation
		if err := rows.Scan(&userOrgAssociation.OrgID, &userOrgAssociation.UserID); err != nil {
			return []model.UserOrgAssociation{}, err
		}
		userOrgAssociations = append(userOrgAssociations, userOrgAssociation)
	}

	return userOrgAssociations, err
}

func (pgx Pgx) CreateUserOrgAssociation(userId string, orgId string) (model.UserOrgAssociation, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(), "INSERT INTO user_org_associations (org_id, user_id) VALUES ($1, $2)", orgId, userId)
	if err != nil || commandTag.RowsAffected() != 1 {
		var userOrg model.UserOrgAssociation
		return userOrg, err
	}
	return pgx.GetUserOrgAssociation(orgId, userId)
}

func (pgx Pgx) GetUserOrgAssociation(orgId string, userId string) (model.UserOrgAssociation, error) {
	var userOrg model.UserOrgAssociation

	if err := pgx.Driver.QueryRow(context.Background(), "SELECT org_id, user_id FROM user_org_associations WHERE org_id=$1 AND user_id=$2", orgId, userId).Scan(&userOrg.OrgID, &userOrg.UserID); err != nil {
		return userOrg, err
	}
	return userOrg, nil
}

// Deletes a specific user-organization association.
func (pgx Pgx) DeleteUserOrgAssociation(orgId string, userId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM user_org_associations WHERE org_id=$1 AND user_id=$2", orgId, userId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
