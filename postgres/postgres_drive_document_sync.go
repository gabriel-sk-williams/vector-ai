package postgres

import (
	"context"
	"fmt"
	"vector-ai/model"

	"github.com/google/uuid"
)

func (pgx Pgx) ListDriveDocumentSync(workspaceId string) ([]model.DriveDocumentSync, error) {
	driveSyncs := []model.DriveDocumentSync{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, workspace_id, document_id, drive_id, drive_parent_id, drive_service_type, last_modified FROM drive_document_sync WHERE workspace_id=$1`, workspaceId)
	if err != nil {
		return []model.DriveDocumentSync{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var sync model.DriveDocumentSync
		if err := rows.Scan(&sync.ID, &sync.WorkspaceID, &sync.DocumentID, &sync.DriveID, &sync.DriveParentID, &sync.DriveServiceType, &sync.LastModified); err != nil {
			return []model.DriveDocumentSync{}, err
		}
		driveSyncs = append(driveSyncs, sync)
	}

	return driveSyncs, err
}

func (pgx Pgx) ListDriveDocumentSyncByParentId(workspaceId string, driveParentId string) ([]model.DriveDocumentSync, error) {
	driveSyncs := []model.DriveDocumentSync{}

	rows, err := pgx.Driver.Query(context.Background(), `SELECT id, workspace_id, document_id, drive_id, drive_parent_id, drive_service_type, last_modified FROM drive_document_sync WHERE workspace_id=$1 AND drive_parent_id=$2`, workspaceId, driveParentId)
	if err != nil {
		return []model.DriveDocumentSync{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var sync model.DriveDocumentSync
		if err := rows.Scan(&sync.ID, &sync.WorkspaceID, &sync.DocumentID, &sync.DriveID, &sync.DriveParentID, &sync.DriveServiceType, &sync.LastModified); err != nil {
			return []model.DriveDocumentSync{}, err
		}
		driveSyncs = append(driveSyncs, sync)
	}

	return driveSyncs, err
}

func (pgx Pgx) GetDriveDocumentSync(syncId string) (model.DriveDocumentSync, error) {
	var sync model.DriveDocumentSync
	if err := pgx.Driver.QueryRow(context.Background(), `SELECT id, workspace_id, document_id, drive_id, drive_parent_id, drive_service_type, last_modified FROM drive_document_sync WHERE id=$1`, syncId).Scan(&sync.ID, &sync.WorkspaceID, &sync.DocumentID, &sync.DriveID, &sync.DriveParentID, &sync.DriveServiceType, &sync.LastModified); err != nil {
		return sync, err
	}
	return sync, nil
}

func (pgx Pgx) CreateDriveDocumentSync(workspaceId string, documentId string, driveServiceId string, driveParentId string, driveServiceType string, lastModified string) (model.DriveDocumentSync, error) {

	uuid := uuid.New()
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"INSERT INTO drive_document_sync (id, workspace_id, document_id, drive_id, drive_parent_id, drive_service_type, last_modified) VALUES ($1, $2, $3, $4, $5, $6, $7)", uuid, workspaceId, documentId, driveServiceId, driveParentId, driveServiceType, lastModified)

	if err != nil || commandTag.RowsAffected() != 1 {
		var sync model.DriveDocumentSync
		return sync, err
	}

	return pgx.GetDriveDocumentSync(uuid.String())
}

// update with new last_modified
func (pgx Pgx) UpdateDriveDocumentSyncLastModified(syncId string, lastModified string) (model.DriveDocumentSync, error) {
	fmt.Println("uddslm:", syncId, lastModified)
	commandTag, err := pgx.Driver.Exec(context.Background(),
		"UPDATE drive_document_sync SET last_modified=$1 WHERE id=$2", lastModified, syncId)

	if err != nil || commandTag.RowsAffected() != 1 {
		var sync model.DriveDocumentSync
		return sync, err
	}

	return pgx.GetDriveDocumentSync(syncId)
}

func (pgx Pgx) DeleteDriveDocumentSync(syncId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM drive_document_sync WHERE id=$1", syncId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}

func (pgx Pgx) DeleteDriveDocumentSyncByDocumentId(documentId string) error {
	commandTag, err := pgx.Driver.Exec(context.Background(), "DELETE FROM drive_document_sync WHERE document_id=$1", documentId)
	if err != nil || commandTag.RowsAffected() != 1 {
		return err
	}
	return nil
}
