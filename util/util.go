package util

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	"vector-ai/model"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

func ConvertVector(v []float64) []float32 {
	v32 := make([]float32, len(v))
	for i, f := range v {
		v32[i] = float32(f)
	}
	return v32
}

func GetContext() (context.Context, context.CancelFunc) {
	apiKey := os.Getenv("QDRANT_API_KEY")
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	md := metadata.New(map[string]string{"api-key": apiKey})
	ctx = metadata.NewOutgoingContext(ctx, md)
	//defer cancel()
	return ctx, cancel
}

func GetContextWithDuration(duration int) (context.Context, context.CancelFunc) {
	apiKey := os.Getenv("QDRANT_API_KEY")

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration)*time.Second)
	md := metadata.New(map[string]string{"api-key": apiKey})
	ctx = metadata.NewOutgoingContext(ctx, md)
	//defer cancel()
	return ctx, cancel
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func GetTemplateIDsByOrgType(orgType string) []string {
	if orgType == "LAW" {
		return []string{
			"ef5b2907-d67f-44a4-a210-8cfeaffec4c8",
			"4c931174-2f28-49a2-be41-c9306f8f767a",
			"4cedf7d6-73f3-4789-a33e-e00d4bb7522d", // default
		}
	} else if orgType == "PHD" {
		return []string{
			"4cedf7d6-73f3-4789-a33e-e00d4bb7522d", // default
		}
	} else if orgType == "WRITING" {
		return []string{
			"4cedf7d6-73f3-4789-a33e-e00d4bb7522d", // default
		}
	} else {
		return []string{
			"4cedf7d6-73f3-4789-a33e-e00d4bb7522d", // default
		}
	}
}

func ReduceFileSize(documents []model.Document) int64 {
	var fileSize int64

	for _, document := range documents {
		fileSize += document.Size
	}

	return fileSize
}

func GetAdjacentIndeces(index int64, adjacentRange int64) []int64 {

	indeces := []int64{}
	start := index - adjacentRange
	end := index + adjacentRange

	// for value := range [start..end] {} <- possible with Go 1.22

	for i := start; i <= end; i++ {
		indeces = append(indeces, i)
	}

	return indeces
}

func MapFolderIds(folderSyncs []model.DriveFolderSync) []string {
	folderIds := []string{}
	for _, folder := range folderSyncs {
		folderIds = append(folderIds, folder.DriveID)
	}
	return folderIds
}

func HasSyncEntry(entries []model.DriveDocumentSync, searchId string) bool {
	for _, entry := range entries {
		if entry.DriveID == searchId {
			return true
		}
	}
	return false
}

func HasDriveItem(items []model.DriveDocument, searchId string) bool {
	for _, item := range items {
		if item.DriveID == searchId {
			return true
		}
	}
	return false
}

func HasId(origins []model.DriveOrigin, searchId string) bool {
	for _, item := range origins {
		if item.DriveID == searchId {
			return true
		}
	}
	return false
}

func GetItemByDriveId(items []model.DriveDocument, searchId string) (model.DriveDocument, bool) {
	for _, item := range items {
		if item.DriveID == searchId {
			return item, true
		}
	}
	return model.DriveDocument{}, false
}

func ReduceByParent(allChildren []model.DriveDocumentSync, folderId string) []model.DriveDocumentSync {
	reducedChildren := []model.DriveDocumentSync{}
	for _, child := range allChildren {
		if child.DriveParentID == folderId {
			reducedChildren = append(reducedChildren, child)
		}
	}
	return reducedChildren
}

func GetParent(parents []string) string {
	if len(parents) == 0 {
		return "root"
	} else {
		return parents[0]
	}
}

func ConvertBytesToMB(tfsa int64) int64 {
	return tfsa / 1000000
}

func GetFileRecord(folderRecords []model.FolderRecord, folderId string, id uuid.UUID) (*model.FileRecord, bool) {
	for i, fr := range folderRecords {
		if fr.FolderID == folderId {
			for j, record := range fr.FileRecords {
				if record.ManifestData.ID == id {
					return &folderRecords[i].FileRecords[j], true
				}
			}
		} else {
			continue
		}
	}
	return &model.FileRecord{}, false
}

func CreateParentageMap(folderSyncs []model.DriveFolderSync, documentSyncs []model.DriveDocumentSync) map[string]string {

	pm := map[string]string{}

	// map each folder to parent
	for _, folderSync := range folderSyncs {
		pm[folderSync.DriveID] = folderSync.DriveParentID
	}

	// map each document to parent
	for _, documentSync := range documentSyncs {
		pm[documentSync.DocumentID] = documentSync.DriveParentID
	}

	return pm
}

func MarshalVssOptions(configs []model.WorkspaceConfig) model.VssOptions {

	configMap := make(map[string]uint32)
	for _, config := range configs {
		configMap[config.Property] = uint32(config.Value)
	}

	jsonbody, err := json.Marshal(configMap)
	check(err)

	options := model.VssOptions{}
	if err := json.Unmarshal(jsonbody, &options); err != nil {
		// do error check
		fmt.Println(err)
		//return
	}

	return options
}

func IsAdmin(orgs []model.Org) bool {
	isAdmin := false
	for _, org := range orgs {
		if org.Level == 99 || org.ID == "org_2Tm1YSHW6dbroQKWsIQKws6" {
			isAdmin = true
		}
	}

	return isAdmin
}

func UploadSize(report model.DriveSyncReport) int64 {

	var add int64
	var subtract int64

	for _, folder := range report.SyncedFolders {
		for _, new := range folder.SyncReport.New {
			add += new.Size
			fmt.Println("new", new.Size)
		}

		for _, updated := range folder.SyncReport.Updated {
			add += updated.Size
			subtract += updated.OriginalFileSize

			fmt.Println("updated", updated.OriginalFileSize, updated.Size)
		}

		for _, missing := range folder.SyncReport.Missing {
			subtract += missing.Size
			fmt.Println("missing", missing.Size)
		}
	}

	return add - subtract
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
