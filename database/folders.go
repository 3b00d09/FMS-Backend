package database

import (
	"database/sql"
	"fms/ioOperations"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
)

func CreateFolder(userId string, folderName string, orgId string) error {
	folderExists, err := FolderExists(folderName, nil, orgId)

	if err != nil {
		return err
	}

	if folderExists {
		return fmt.Errorf("folder exists")
	}

	tx, err := dbClient.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	statement, err := tx.Prepare("INSERT INTO folder (org_id, uploader_id, name) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	res, err := statement.Exec(orgId, userId, folderName)

	if err != nil {
		return err
	}

	// get the folder id of the inserted row
	folderId, err := res.LastInsertId()

	// don't want the upload function to error out if we are unable to send out a notification because the file itself got uploaded
	if err != nil {
		log.Printf("error: could not read file ID: %v", err.Error())
	}

	// convert the id to a string
	payloadID := strconv.FormatInt(folderId, 10)

	// attempt to create folder at root level
	// if this operation fails and the function returns, the transaction will rollback because of defer tx.rollback
	err = ioOperations.CreateOrgFolderAtRoot(payloadID, orgId)
	if err != nil {
		log.Printf("ERROR: CREATING A FOLDER AT ROOT OF ORG FAILED. Org: %v Folder: %v \n", orgId, payloadID)
		// if the folder was not created then the transaction should not go through
		return err
	}

	// we only commit the transaction if the folder was created
	err = tx.Commit()
	if err != nil {
		return err
	}
	// send notification to all org members + org owner if applicable
	// this is a non-critical operation so neither transaction nor folder creation care about the result
	err = SendNotificationToOrgMembers(orgId, userId, "folder upload", "Uploaded a folder to", payloadID, folderName)
	if err != nil {
		log.Printf("error: could not send out notification to upload folder: %v", err.Error())
	}

	return nil
}

func CreateFolderAsChild(userId string, folderName string, orgId string, parentFolderName string) error {
	folderExists, err := FolderExists(folderName, &parentFolderName, orgId)

	if err != nil {
		return err
	}

	if folderExists {
		return fmt.Errorf("folder exists")
	}

	statement, err := dbClient.Prepare(`
		INSERT INTO folder (org_id, uploader_id, name, parent_folder_id) 
		VALUES (?, ?, ?,(SELECT id FROM folder WHERE name = ? AND org_id = ?))
		`)

	if err != nil {
		fmt.Print(err.Error())
	}

	defer statement.Close()

	res, err := statement.Exec(orgId, userId, folderName, parentFolderName, orgId)
	if err != nil {
		fmt.Print(err.Error())
	}

	// get the file id of the inserted row
	folderId, err := res.LastInsertId()

	// don't want the upload function to error out if we are unable to send out a notification because the file itself got uploaded
	if err != nil {
		log.Printf("error: could not read file ID: %v", err.Error())
	}

	// convert the id to a string
	payloadID := strconv.FormatInt(folderId, 10)

	path, err := getFolderPath(payloadID)

	if err != nil {
		log.Printf("ERROR: UNABLE TO READ FOLDER PATH FOR FOLDER ID: %v ORG ID: %v\n", payloadID, orgId)
	}

	err = ioOperations.CreateOrgFolderAsChild(path)
	if err != nil {
		log.Printf("ERROR: UNABLE TO CREATE FOLDER FOR FOLDER ID: %v ORG ID: %v \n", payloadID, orgId)
	}

	// send notification to all org members + org owner if applicable
	err = SendNotificationToOrgMembers(orgId, userId, "folder upload", "Uploaded a folder to", payloadID, folderName)
	if err != nil {
		log.Printf("error: could not send out notification to upload folder: %v", err.Error())
	}

	return nil
}

func GetRootFolderOfOrg(orgId string) []FolderData {

	var folders []FolderData

	statement, err := dbClient.Prepare(`
		SELECT 
			folder.id, 
			folder.org_id, 
			user.username, 
			folder.name, 
			folder.parent_folder_id, 
			folder.created_at,
			COALESCE(SUM(file.size), 0) AS total_size
		FROM folder 
		LEFT JOIN user ON user.id = folder.uploader_id
		LEFT JOIN file ON file.folder_id = folder.id
		WHERE folder.org_id = ? AND folder.parent_folder_id IS NULL
		GROUP BY folder.id, folder.org_id, user.username, folder.name, folder.parent_folder_id, folder.created_at
		ORDER BY folder.created_at DESC
	`)

	if err != nil {
		return folders
	}

	defer statement.Close()

	rows, err := statement.Query(orgId)
	if err != nil {
		return folders
	}
	defer rows.Close()

	for rows.Next() {
		var folder FolderData
		err := rows.Scan(
			&folder.Id,
			&folder.OrgId,
			&folder.Uploader,
			&folder.Name,
			&folder.ParentFolderId,
			&folder.CreatedAt,
			&folder.Size,
		)
		if err != nil {
			continue
		}
		folders = append(folders, folder)
	}

	if err = rows.Err(); err != nil {
		return folders
	}

	return folders

}

func GetFolderChildren(folderName string, orgId string) []FolderData {
	var folders []FolderData

	statement, err := dbClient.Prepare(`
		SELECT 
			folder.id, folder.org_id, user.username, folder.name, 
			folder.parent_folder_id, folder.created_at, COALESCE(SUM(file.size), 0) AS total_size
		FROM folder 
		LEFT JOIN user ON user.id = folder.uploader_id
		LEFT JOIN file ON file.folder_id = folder.id
		WHERE folder.parent_folder_id = (SELECT id FROM folder WHERE name = ? AND org_id = ? LIMIT 1) AND folder.org_id = ?
		GROUP BY folder.id, folder.org_id, user.username, folder.name, folder.parent_folder_id, folder.created_at
		ORDER BY folder.created_at DESC
	`)

	if err != nil {
		return folders
	}

	defer statement.Close()

	rows, err := statement.Query(folderName, orgId, orgId)
	if err != nil {
		return folders
	}
	defer rows.Close()

	for rows.Next() {
		var folder FolderData
		err := rows.Scan(
			&folder.Id,
			&folder.OrgId,
			&folder.Uploader,
			&folder.Name,
			&folder.ParentFolderId,
			&folder.CreatedAt,
			&folder.Size,
		)
		if err != nil {
			continue
		}
		folders = append(folders, folder)
	}

	if err = rows.Err(); err != nil {
		return folders
	}

	return folders
}

func DeleteFolder(folderId string, userId string, orgId string, folderName string) error {
	folderPath, err := getFolderPath(folderId)

	if err != nil {
		log.Printf("ERROR: UNABLE TO PARSE FOLDER PATH TREE. FOLDER ID:%v ORG ID:%v \n", folderId, orgId)
	}

	statement, err := dbClient.Prepare("DELETE FROM folder WHERE id = ?")
	if err != nil {
		return err
	}

	defer statement.Close()

	result, err := statement.Exec(folderId)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("something went wrong")
	}

	err = ioOperations.DeleteOrgFolder(folderPath)
	if err != nil {
		log.Printf("ERROR: UNABLE TO DELETE CHILD FOLDER IN AN ORG. FOLDER ID:%v ORG ID:%v \n", folderId, orgId)
	}

	// send notification to all org members + org owner if applicable
	err = SendNotificationToOrgMembers(orgId, userId, "folder delete", "Deleted a folder from", folderId, folderName)
	if err != nil {
		log.Printf("error: could not send out notification to delete folder: %v", err.Error())
	}

	return nil
}

func FolderExists(folderName string, parentFolderName *string, orgId string) (bool, error) {
	if parentFolderName == nil {
		statement, err := dbClient.Prepare("SELECT COUNT(id) FROM folder WHERE name = ? AND parent_folder_id IS NULL AND org_id = ?")
		if err != nil {
			return true, err
		}
		defer statement.Close()

		result := statement.QueryRow(folderName, orgId)
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
		statement, err := dbClient.Prepare("SELECT COUNT(id) FROM folder WHERE name = ? AND parent_folder_id = (SELECT id FROM folder WHERE name = ? AND org_id = ?) AND org_id = ?")
		if err != nil {
			return true, err
		}
		defer statement.Close()

		result := statement.QueryRow(folderName, parentFolderName, orgId, orgId)
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

// helper function to get the full path of a folder by recursively finding its parent folders
func getFolderPath(folderId string) (string, error) {
	var orgId string
	var parentFolderId sql.NullString

	statement, err := dbClient.Prepare("SELECT org_id, parent_folder_id FROM folder WHERE id = ?")
	if err != nil {
		return "", err
	}

	defer statement.Close()

	err = statement.QueryRow(folderId).Scan(&orgId, &parentFolderId)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("folder not found")
		}
		return "", err
	}

	// base case: if the folder has no parent (root level folder)
	if !parentFolderId.Valid || parentFolderId.String == "" {

		return filepath.Join("appdata", fmt.Sprintf("org-%s", orgId), fmt.Sprintf("folder-%s", folderId)), nil
	}

	// recursive case: Get the parent folder's path and append this folder to it
	parentPath, err := getFolderPath(parentFolderId.String)
	if err != nil {
		return "", err
	}

	// return the full path
	return filepath.Join(parentPath, fmt.Sprintf("folder-%s", folderId)), nil
}
