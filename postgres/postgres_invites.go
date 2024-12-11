package postgres

import (
	"context"
	b64 "encoding/base64"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (pgx Pgx) ListInvites() ([]model.Invite, error) {
	invites := []model.Invite{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, org_id, email, token, expiration FROM invites`)

	if err != nil {
		return invites, err
	}
	defer rows.Close()

	for rows.Next() {
		var invite model.Invite
		if err := rows.Scan(&invite.ID, &invite.OrgID, &invite.Email, &invite.Token, &invite.Expiration); err != nil {
			return []model.Invite{}, err
		}
		invites = append(invites, invite)
	}

	return invites, err
}

func (pgx Pgx) ListInvitesByOrgId(orgId string) ([]model.Invite, error) {
	invites := []model.Invite{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, org_id, email, token, expiration FROM invites WHERE org_id=$1`, orgId)

	if err != nil {
		return invites, err
	}
	defer rows.Close()

	for rows.Next() {
		var invite model.Invite
		if err := rows.Scan(&invite.ID, &invite.OrgID, &invite.Email, &invite.Token, &invite.Expiration); err != nil {
			return []model.Invite{}, err
		}
		invites = append(invites, invite)
	}

	return invites, err
}

// GetOrg retrieves a specific org by ID.
func (pgx Pgx) GetInvite(token string) (model.Invite, error) {
	var invite model.Invite

	if err := pgx.Driver.QueryRow(context.Background(), "SELECT id, org_id, email, token, expiration FROM invites WHERE token=$1", token).Scan(&invite.ID, &invite.OrgID, &invite.Email, &invite.Token, &invite.Expiration); err != nil {
		return invite, err
	}
	return invite, nil
}

func (pgx Pgx) CreateInvite(orgId string, email string, expiration string) (model.Invite, error) {
	uuid := uuid.New()

	data := email + expiration
	token := b64.StdEncoding.EncodeToString([]byte(data))

	commandTag, err := pgx.Driver.Exec(context.Background(),
		`INSERT INTO invites (id, org_id, email, token, expiration) 
		VALUES ($1, $2, $3, $4, $5)`, uuid, orgId, email, token, expiration)

	if err != nil || commandTag.RowsAffected() != 1 {
		var invite model.Invite
		return invite, err
	}

	return pgx.GetInvite(token)
}

// DeleteOrg deletes an org by ID.
func (pgx Pgx) DeleteInvite(inviteId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM invites WHERE id=$1", inviteId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
