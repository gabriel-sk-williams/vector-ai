package postgres

import (
	"context"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (pgx Pgx) ListDriveFolderSync(workspaceId string) ([]model.DriveFolderSync, error) {
	driveSyncs := []model.DriveFolderSync{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, workspace_id, name, drive_id, drive_parent_id, drive_service_type, last_synced FROM drive_folder_sync WHERE workspace_id=$1`, workspaceId)
	if err != nil {
		return []model.DriveFolderSync{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var sync model.DriveFolderSync
		if err := rows.Scan(&sync.ID, &sync.WorkspaceID, &sync.Name, &sync.DriveID, &sync.DriveParentID, &sync.DriveServiceType, &sync.LastSynced); err != nil {
			return []model.DriveFolderSync{}, err
		}
		driveSyncs = append(driveSyncs, sync)
	}

	return driveSyncs, err
}

func (pgx Pgx) GetDriveFolderSync(syncId string) (model.DriveFolderSync, error) {
	var sync model.DriveFolderSync
	if err := pgx.Driver.QueryRow(context.Background(), "SELECT id, workspace_id, name, drive_id, drive_parent_id, drive_service_type, last_synced FROM drive_folder_sync WHERE id=$1", syncId).Scan(&sync.ID, &sync.WorkspaceID, &sync.Name, &sync.DriveID, &sync.DriveParentID, &sync.DriveServiceType, &sync.LastSynced); err != nil {
		return sync, err
	}
	return sync, nil
}

func (pgx Pgx) CreateDriveFolderSync(workspaceId string, name string, driveID string, driveParentId string, driveServiceType string, lastSynced string) (model.DriveFolderSync, error) {

	uuid := uuid.New()
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"INSERT INTO drive_folder_sync (id, workspace_id, name, drive_id, drive_parent_id, drive_service_type, last_synced) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING", uuid, workspaceId, name, driveID, driveParentId, driveServiceType, lastSynced)

	if err != nil || commandTag.RowsAffected() != 1 {
		var sync model.DriveFolderSync
		return sync, err
	}

	return pgx.GetDriveFolderSync(uuid.String())
}

// update with new last_synced
func (pgx Pgx) UpdateDriveFolderLastSynced(syncId string, lastSynced string) (model.DriveFolderSync, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"UPDATE drive_folder_sync SET last_synced=$1 WHERE id=$2", lastSynced, syncId)

	if err != nil || commandTag.RowsAffected() != 1 {
		var sync model.DriveFolderSync
		return sync, err
	}

	return pgx.GetDriveFolderSync(syncId)
}

func (pgx Pgx) UpdateDriveFolderName(syncId string, name string) (model.DriveFolderSync, error) {
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"UPDATE drive_folder_sync SET name=$1 WHERE id=$2", name, syncId)

	if err != nil || commandTag.RowsAffected() != 1 {
		var sync model.DriveFolderSync
		return sync, err
	}

	return pgx.GetDriveFolderSync(syncId)
}

func (pgx Pgx) DeleteDriveFolderSync(syncId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM drive_folder_sync WHERE id=$1", syncId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
