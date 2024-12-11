package postgres

import (
	"context"
	"fmt"
	"vector-ai/model"
)

func (pgx Pgx) ListUsers(orgId string) ([]model.User, error) {
	users := []model.User{}

	// Get all userIds associated with orgId from user_org_associations
	// Get all Users from userIds
	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, first_name, last_name, email from users 
	WHERE id IN (
	SELECT user_id from user_org_associations WHERE org_id=$1
		)`, orgId)

	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email); err != nil {
			return []model.User{}, err
		}
		users = append(users, user)
	}

	return users, err
}

// GetOrg retrieves a specific org by ID.
func (pgx Pgx) GetUser(userId string) (model.User, error) {
	var user model.User

	if err := pgx.Driver.QueryRow(context.Background(), "SELECT id, first_name, last_name, email FROM users WHERE id=$1", userId).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email); err != nil {
		return user, err
	}
	return user, nil
}

func (pgx Pgx) GetUserName(userId string) (string, error) {
	var user model.User
	var userName string

	if err := pgx.Driver.QueryRow(context.Background(), "SELECT first_name, last_name FROM users WHERE id=$1", userId).Scan(&user.FirstName, &user.LastName); err != nil {
		return userName, err
	}

	userName = fmt.Sprintf(`%s %s`, user.FirstName, user.LastName)
	return userName, nil
}

func (pgx Pgx) CreateUser(userId string, firstName string, lastName string, email string) (model.User, error) {

	commandTag, err := pgx.Driver.Exec(context.Background(),
		`
		INSERT INTO users (id, first_name, last_name, email) VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE
		SET first_name=$2, last_name=$3, email=$4 WHERE users.id=$1
		`, userId, firstName, lastName, email)

	if err != nil || commandTag.RowsAffected() != 1 {
		var user model.User
		return user, err
	}

	return pgx.GetUser(userId)
}

// DeleteOrg deletes an org by ID.
func (pgx Pgx) DeleteUser(userId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), `DELETE FROM users WHERE id=$1`, userId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
