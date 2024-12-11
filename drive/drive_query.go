package drive

import (
	"fmt"
	"net/http"
	"time"
	"vector-ai/model"
	"vector-ai/util"

	"google.golang.org/api/drive/v3"
)

// curl http://localhost:5000/org/org_2Tm1YSHW6dbroQKWsIQKws6/workspace/0d9a1a4d-8ffd-48da-b6e8-2222f913d3ba/gdoc
func (drv Drv) ListDocs(query string, orderBy string, limit int64) ([]model.GoogleDriveFile, error) {

	// var temp *drive.Service
	//.Q("mimeType='" + FolderMimeType + "'")
	// all documents/folders: "mimeType contains 'folder' or mimeType contains 'document' or mimeType contains 'pdf'"
	// all orderby: "folder desc,modifiedTime desc,name desc"
	// only folders "mimeType='application/vnd.google-apps.folder'"
	r, err := drv.Service.Files.List().Q(query).IncludeItemsFromAllDrives(true).SupportsAllDrives(true).OrderBy(orderBy).PageSize(limit).Fields("nextPageToken, files(id, name, kind, mimeType, size, thumbnailLink, iconLink, fullFileExtension, parents)").Do()

	gdfs := []model.GoogleDriveFile{}
	//fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, file := range r.Files {
			// fmt.Printf("%s (%s)\n", file.Name, file.FullFileExtension)
			// fmt.Printf("%s (%s)\n", file.Kind, file.MimeType) // drive#file (application/vnd.google-apps.folder)
			parent := util.GetParent(file.Parents)
			gdf := model.GoogleDriveFile{
				DriveOrigin: model.DriveOrigin{
					DriveID:          file.Id,
					DriveParentID:    parent,
					DriveServiceType: ServiceType,
				},
				CoreDocumentProps: model.CoreDocumentProps{
					Name:     file.Name,
					Size:     file.Size,
					MimeType: file.MimeType,
				},
				ExtendedDriveProps: model.ExtendedDriveProps{
					Kind:          file.Kind,
					ThumbnailLink: file.ThumbnailLink,
					FileExtension: file.FullFileExtension,
					IconLink:      file.IconLink,
				},
			}

			gdfs = append(gdfs, gdf)
		}
	}

	return gdfs, err
}

func (drv Drv) ListSharedDrives() ([]model.SharedDrive, error) {
	drives := []model.SharedDrive{}
	// var err *googleapi.Error

	r, err := drv.Service.Drives.List().Do()
	if err != nil {
		return drives, err
	}

	if len(r.Drives) == 0 {
		fmt.Println("No shared drives found")
	} else {
		for _, drive := range r.Drives {
			drive := model.SharedDrive{
				DriveID: drive.Id,
				Name:    drive.Name,
			}
			drives = append(drives, drive)
		}
	}
	return drives, err
}

func (drv Drv) ListChildren(folderId string) ([]model.DriveDocument, error) {

	query := fmt.Sprintf("'%s' in parents and (mimeType = 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' or mimeType = 'application/epub+zip' or mimeType = 'text/plain' or mimeType = 'application/pdf' or mimeType = 'application/vnd.google-apps.document') and trashed=false", folderId)

	r, err := drv.Service.Files.List().Q(query).OrderBy("name").PageSize(1000).IncludeItemsFromAllDrives(true).SupportsAllDrives(true).Fields("files(id, name, mimeType, size, iconLink, modifiedTime)").Do()
	check(err)

	docs := []model.DriveDocument{}
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, file := range r.Files {

			lastModified, err := time.Parse(time.RFC3339, file.ModifiedTime)
			check(err)

			document := model.DriveDocument{
				DriveOrigin: model.DriveOrigin{
					DriveID:          file.Id,
					DriveParentID:    folderId,
					DriveServiceType: ServiceType,
				},
				CoreDocumentProps: model.CoreDocumentProps{
					Name:     file.Name,
					Size:     file.Size,
					MimeType: file.MimeType,
				},
				LastModified: lastModified,
			}
			docs = append(docs, document)
		}
	}

	return docs, nil
}

// currently being used for folders
func (drv Drv) GetDriveDataById(fileId string) (*drive.File, error) {
	// "mimeType='application/vnd.google-apps.folder' and name='$folderName' and '### folder ID of FolderA ###' in parents"
	return drv.Service.Files.Get(fileId).SupportsAllDrives(true).Fields("id, name, mimeType, size, parents").Do()
}

func (drv Drv) GetDriveFolderById(folderId string) (model.DriveFolder, error) {

	folder, err := drv.GetDriveDataById(folderId)
	parent := util.GetParent(folder.Parents)

	dfm := model.DriveFolder{
		DriveOrigin: model.DriveOrigin{
			DriveID:          folder.Id,
			DriveParentID:    parent,
			DriveServiceType: ServiceType,
		},
		CoreDocumentProps: model.CoreDocumentProps{
			Name:     folder.Name,
			Size:     folder.Size,
			MimeType: folder.MimeType,
		},
	}

	return dfm, err
}

func (drv Drv) ExportDriveFile(driveId string, exportType string) (*http.Response, error) {
	return drv.Service.Files.Export(driveId, exportType).Download()
}

func (drv Drv) DownloadDriveFile(driveId string) (*http.Response, error) {
	return drv.Service.Files.Get(driveId).SupportsAllDrives(true).Download()
}
