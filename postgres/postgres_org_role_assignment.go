package postgres

import (
	"context"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (pgx Pgx) ListOrgRoleAssignments(orgId string) ([]model.OrgRoleAssignment, error) {
	orgAssignments := []model.OrgRoleAssignment{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, org_id, user_id, role FROM org_role_assigments WHERE org_id=$1`, orgId)
	if err != nil {
		return []model.OrgRoleAssignment{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var orgAssignment model.OrgRoleAssignment
		if err := rows.Scan(&orgAssignment.ID, &orgAssignment.OrgID, &orgAssignment.UserID, &orgAssignment.Role); err != nil {
			return []model.OrgRoleAssignment{}, err
		}
		orgAssignments = append(orgAssignments, orgAssignment)
	}

	return orgAssignments, err
}

func (pgx Pgx) ListOrgRoleAssignmentsByUserId(orgId string, userId string) ([]model.OrgRoleAssignment, error) {
	orgAssignments := []model.OrgRoleAssignment{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, org_id, user_id, role FROM org_role_assigments WHERE org_id=$1 AND userId=$2`, orgId, userId)
	if err != nil {
		return []model.OrgRoleAssignment{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var orgAssignment model.OrgRoleAssignment
		if err := rows.Scan(&orgAssignment.ID, &orgAssignment.OrgID, &orgAssignment.UserID, &orgAssignment.Role); err != nil {
			return []model.OrgRoleAssignment{}, err
		}
		orgAssignments = append(orgAssignments, orgAssignment)
	}

	return orgAssignments, err
}

func (pgx Pgx) GetOrgRoleAssignmentById(orgRoleAssigmentId string) (model.OrgRoleAssignment, error) {
	var orgAssignment model.OrgRoleAssignment

	if err := pgx.Driver.QueryRow(context.Background(), "SELECT id, org_id, user_id, role FROM org_role_assigments WHERE id=$1", orgRoleAssigmentId).Scan(&orgAssignment.ID, &orgAssignment.OrgID, &orgAssignment.UserID, &orgAssignment.Role); err != nil {
		return orgAssignment, err
	}
	return orgAssignment, nil
}

func (pgx Pgx) GetOrgRoleAssignmentByOrgUser(orgId string, userId string) (model.OrgRoleAssignment, error) {
	var orgAssignment model.OrgRoleAssignment

	if err := pgx.Driver.QueryRow(context.Background(), "SELECT id, org_id, user_id, role FROM org_role_assigments WHERE org_id=$1 AND user_id=$2", orgId, userId).Scan(&orgAssignment.ID, &orgAssignment.OrgID, &orgAssignment.UserID, &orgAssignment.Role); err != nil {
		return orgAssignment, err
	}
	return orgAssignment, nil
}

func (pgx Pgx) CreateOrgRoleAssignment(orgId string, userId string, role string) (model.OrgRoleAssignment, error) {
	uuid := uuid.New()
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"INSERT INTO org_role_assignments (id, org_id, user_id, role) VALUES ($1, $2, $3, $4)", uuid, orgId, userId, role)

	if err != nil || commandTag.RowsAffected() != 1 {
		var orgAssignment model.OrgRoleAssignment
		return orgAssignment, err
	}

	return pgx.GetOrgRoleAssignmentById(uuid.String())
}

func (pgx Pgx) UpdateOrgRoleAssignment(orgRoleAssignmentId string, role string) (model.OrgRoleAssignment, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"UPDATE org_role_assigments SET role=$1 WHERE id=$2", orgRoleAssignmentId, role)

	if err != nil || commandTag.RowsAffected() != 1 {
		var orgAssignment model.OrgRoleAssignment
		return orgAssignment, err
	}

	return pgx.GetOrgRoleAssignmentById(orgRoleAssignmentId)
}

func (pgx Pgx) DeleteOrgRoleAssignment(orgRoleAssignmentId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM org_role_assigments WHERE id=$1", orgRoleAssignmentId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
