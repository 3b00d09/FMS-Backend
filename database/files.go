package database

import (
	"database/sql"
	"fms/ioOperations"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
)

func UploadFileToRoot(file *multipart.FileHeader, orgId string, uploaderId string) error {
	fileExists, err := FileExists(file.Filename, nil, nil)
	if err != nil {
		return err
	}
	if fileExists {
		return fmt.Errorf("file name already exists in this location")
	}

	tx, err := dbClient.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	statement, err := tx.Prepare(`
	 	INSERT INTO file (org_id, uploader_id, name, type, size)
		VALUES (?, ?, ?, ?, ?)
	 `)

	if err != nil {
		fmt.Print(err.Error())
	}

	defer statement.Close()

	res, err := statement.Exec(orgId, uploaderId, file.Filename, filepath.Ext(file.Filename), file.Size)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("file name already exists in this location")
		} else {
			return err
		}
	}

	// get the file id of the inserted row
	fileId, err := res.LastInsertId()

	// don't want the upload function to error out if we are unable to send out a notification because the file itself got uploaded
	if err != nil {
		log.Printf("error: could not read file ID: %v", err.Error())
	}

	// convert the id to a string
	payloadID := strconv.FormatInt(fileId, 10)

	err = ioOperations.CreateOrgFileAtRoot(file, payloadID, orgId)
	if err != nil {
		log.Printf("ERROR CREATING FILE WITH ID: %v ORG ID %s, error: %s", fileId, orgId, err.Error())
		return err
	}

	tx.Commit()
	// send notification to all org members + org owner if applicable
	err = SendNotificationToOrgMembers(orgId, uploaderId, "file upload", "Uploaded a file to", payloadID, file.Filename)
	if err != nil {
		log.Printf("error: could not send out notification to file upload: %v", err.Error())
	}

	return nil

}

func UploadFileToFolder(file *multipart.FileHeader, orgId string, parentFolderName string, uploaderId string) error {
	fileExists, err := FileExists(file.Filename, &parentFolderName, &orgId)
	if err != nil {
		return err
	}
	if fileExists {
		return fmt.Errorf("file name already exists in this location")
	}

	//  get the folder ID so we can find its path
	var folderId string
	err = dbClient.QueryRow("SELECT id FROM folder WHERE name = ? AND org_id = ?", parentFolderName, orgId).Scan(&folderId)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("folder not found")
		}
		return err
	}

	// get the folder path
	folderPath, err := getFolderPath(folderId)
	if err != nil {
		return fmt.Errorf("error getting folder path: %w", err)
	}

	tx, err := dbClient.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()
	statement, err := tx.Prepare(`
	 	INSERT INTO file (org_id, uploader_id, name, type, size, folder_id)
		VALUES (?, ?, ?, ?, ?, (SELECT id FROM folder WHERE name = ? AND org_id = ?))
	 `)

	if err != nil {
		return err
	}

	defer statement.Close()

	res, err := statement.Exec(orgId, uploaderId, file.Filename, filepath.Ext(file.Filename), file.Size, parentFolderName, orgId)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("file name already exists in this location")
		} else {
			return err
		}
	}

	// get the file id of the inserted row
	fileId, err := res.LastInsertId()

	// don't want the upload function to error out if we are unable to send out a notification because the file itself got uploaded
	if err != nil {
		log.Printf("error: could not read file ID: %v", err.Error())
	}

	// convert the id to a string
	payloadID := strconv.FormatInt(fileId, 10)

	err = ioOperations.CreateOrgFileAsChild(file, payloadID, folderPath)
	if err != nil {
		log.Printf("ERROR CREATING FILE ID %s, error: %s", payloadID, err.Error())
		return err
	}

	tx.Commit()
	// send notification to all org members + org owner if applicable
	err = SendNotificationToOrgMembers(orgId, uploaderId, "file upload", "Uploaded a file to", payloadID, file.Filename)
	if err != nil {
		log.Printf("error: could not send out notification to file upload: %v", err.Error())
	}

	return nil
}

func GetRootFilesOfOrg(orgId string) []FileData {
	var files []FileData
	statement, err := dbClient.Prepare(`
		SELECT file.id, file.folder_id, file.org_id, user.username, file.name, file.type, file.size, file.uploaded_at  
		FROM file 
		LEFT JOIN user ON user.id = file.uploader_id
		WHERE org_id = ? AND folder_id IS NULL
		ORDER BY uploaded_at DESC`)
	if err != nil {
		fmt.Print(err.Error())
		return files
	}

	defer statement.Close()

	rows, err := statement.Query(orgId)

	if err != nil {
		fmt.Print(err.Error())
		return files
	}

	for rows.Next() {
		var file FileData
		err := rows.Scan(
			&file.Id,
			&file.ParentFolderId,
			&file.OrgId,
			&file.Uploader,
			&file.Name,
			&file.Type,
			&file.Size,
			&file.CreatedAt,
		)
		if err != nil {
			continue
		}
		files = append(files, file)
	}

	return files

}

func GetFolderFiles(folderName string, orgId string) []FileData {
	var files []FileData
	statement, err := dbClient.Prepare(`
		SELECT file.id, file.folder_id, file.org_id, user.username, file.name, file.type, file.size, file.uploaded_at  
		FROM file 
		LEFT JOIN user ON user.id = file.uploader_id
		WHERE org_id = ? AND folder_id = (SELECT id FROM folder WHERE name = ? AND org_id = ?)
		ORDER BY uploaded_at DESC`)
	if err != nil {
		fmt.Print(err.Error())
		return files
	}

	defer statement.Close()

	rows, err := statement.Query(orgId, folderName, orgId)

	if err != nil {
		fmt.Print(err.Error())
		return files
	}

	for rows.Next() {
		var file FileData
		err := rows.Scan(
			&file.Id,
			&file.ParentFolderId,
			&file.OrgId,
			&file.Uploader,
			&file.Name,
			&file.Type,
			&file.Size,
			&file.CreatedAt,
		)
		if err != nil {
			continue
		}
		files = append(files, file)
	}

	return files
}

func FileExists(fileName string, folderName *string, orgId *string) (bool, error) {
	if folderName == nil {
		statement, err := dbClient.Prepare("SELECT COUNT(id) FROM file WHERE name = ? AND folder_id IS NULL")
		if err != nil {
			return true, err
		}
		defer statement.Close()
		result := statement.QueryRow(fileName)
		var count int
		err = result.Scan(&count)
		if err != nil {
			return true, err
		}
		if count == 0 {
			return false, nil
		} else {
			return true, nil
		}

	} else {
		statement, err := dbClient.Prepare("SELECT COUNT(id) FROM file WHERE name = ? AND folder_id = (SELECT id FROM folder WHERE name = ? AND org_id = ?)")
		if err != nil {
			return true, err
		}
		defer statement.Close()
		result := statement.QueryRow(fileName, folderName, orgId)
		var count int
		err = result.Scan(&count)
		if err != nil {
			return true, err
		}
		if count == 0 {
			return false, nil
		} else {
			return true, nil
		}
	}
}

func DeleteFile(fileId string, orgId string, userId string, fileName string) error {
	path, err := GetFilePath(fileId)
	if err != nil {
		fmt.Println(err)
	}

	statement, err := dbClient.Prepare("DELETE FROM file WHERE id = ?")
	if err != nil {
		return err
	}

	defer statement.Close()

	result, err := statement.Exec(fileId)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("something went wrong")
	}

	err = ioOperations.DeleteOrgChildFile(path)
	if err != nil {
		fmt.Printf("ERROR REMOVING FILE WITH ID: %v\n", fileId)
	}

	// send notification to all org members + org owner if applicable
	err = SendNotificationToOrgMembers(orgId, userId, "file delete", "Delete a file from", fileId, fileName)
	if err != nil {
		log.Printf("error: could not send out notification to file delete: %v", err.Error())
	}
	return nil
}

// helper function to get the full path of a folder by recursively finding its parent folders
func GetFilePath(fileId string) (string, error) {
	var orgId string
	var folderId sql.NullString

	statement, err := dbClient.Prepare("SELECT org_id, folder_id FROM file WHERE id = ?")
	if err != nil {
		return "", err
	}

	defer statement.Close()

	err = statement.QueryRow(fileId).Scan(&orgId, &folderId)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("file not found")
		}
		return "", err
	}

	var parentPath string
	// base case: if the folder has no parent (root level folder)
	if !folderId.Valid || folderId.String == "" {
		parentPath = filepath.Join("appdata", fmt.Sprintf("org-%s", orgId))
	} else {
		parentPath, err = getFolderPath(folderId.String)
		if err != nil {
			return "", fmt.Errorf("error GETTING FOLDER PATH FOR FILE ID: %v: error: %s", fileId, err.Error())
		}
	}

	// return the full path
	return filepath.Join(parentPath, fmt.Sprintf("file-%s", fileId)), nil
}
