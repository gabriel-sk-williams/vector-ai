package postgres

import (
	"context"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (pgx Pgx) ListTags(orgId string) ([]model.Tag, error) {
	tags := []model.Tag{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, org_id, name, color, FROM tags WHERE org_id=$1`, orgId)
	if err != nil {
		return []model.Tag{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.ID, &tag.OrgID, &tag.Name, &tag.Color); err != nil {
			return []model.Tag{}, err
		}
		tags = append(tags, tag)
	}

	return tags, err
}

func (pgx Pgx) CreateTag(orgId string, tagName string, tagColor string) (model.Tag, error) {
	uuid := uuid.New()

	commandTag, err := pgx.Driver.Exec(context.Background(),
		`INSERT INTO tags (id, org_id, name, color) VALUES ($1, $2, $3, $4)`, uuid, orgId, tagName, tagColor)

	if err != nil || commandTag.RowsAffected() != 1 {
		var tag model.Tag
		return tag, err
	}

	return pgx.GetTag(uuid.String())
}

func (pgx Pgx) GetTag(tagId string) (model.Tag, error) {
	var tag model.Tag
	if err := pgx.Driver.QueryRow(context.Background(), `SELECT id, org_id, name, color FROM tags WHERE id=$1`, tagId).Scan(&tag.ID, &tag.OrgID, &tag.Name, &tag.Color); err != nil {
		return tag, err
	}
	return tag, nil
}

func (pgx Pgx) UpdateTag(tagId string, tagName string, tagColor string) (model.Tag, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"UPDATE tags SET name=$1, color=$2 WHERE id=$3", tagName, tagColor, tagId)

	if err != nil || commandTag.RowsAffected() != 1 {
		var tag model.Tag
		return tag, err
	}

	return pgx.GetTag(tagId)
}

func (pgx Pgx) DeleteTag(tagId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM tags WHERE id=$1", tagId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
