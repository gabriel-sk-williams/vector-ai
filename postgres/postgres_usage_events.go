package postgres

import (
	"context"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (pgx Pgx) ListUsageEvents(orgId string) ([]model.UsageEvent, error) {
	usageEvents := []model.UsageEvent{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, org_id, total_file_size_amount, timestamp FROM usage_events WHERE org_id=$1 ORDER BY timestamp ASC
	`, orgId)
	if err != nil {
		return []model.UsageEvent{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var usageEvent model.UsageEvent
		if err := rows.Scan(&usageEvent.ID, &usageEvent.OrgID, &usageEvent.TotalFileSizeAmount, &usageEvent.Timestamp); err != nil {
			return []model.UsageEvent{}, err
		}
		usageEvents = append(usageEvents, usageEvent)
	}

	return usageEvents, err
}

func (pgx Pgx) GetUsageEvent(usageEventId string) (model.UsageEvent, error) {
	var usageEvent model.UsageEvent
	if err := pgx.Driver.QueryRow(context.Background(), `SELECT id, org_id, total_file_size_amount, timestamp FROM usage_events WHERE id=$1`, usageEventId).Scan(&usageEvent.ID, &usageEvent.OrgID, &usageEvent.TotalFileSizeAmount, &usageEvent.Timestamp); err != nil {
		return usageEvent, err
	}
	return usageEvent, nil
}

func (pgx Pgx) CreateUsageEvent(orgId string, totalFileSizeAmount int64, timestamp string) (model.UsageEvent, error) {
	uuid := uuid.New()

	commandTag, err := pgx.Driver.Exec(context.Background(),
		"INSERT INTO usage_events (id, org_id, total_file_size_amount, timestamp) VALUES ($1, $2, $3, $4)", uuid, orgId, totalFileSizeAmount, timestamp)

	if err != nil || commandTag.RowsAffected() != 1 {
		var usageEvent model.UsageEvent
		return usageEvent, err
	}

	return pgx.GetUsageEvent(uuid.String())
}

func (pgx Pgx) DeleteUsageEvent(usageEventId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM usage_events WHERE id=$1", usageEventId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
