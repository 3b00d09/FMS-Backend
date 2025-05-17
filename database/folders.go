package database

import "fmt"

func CreateFolder(userId string, folderName string, orgId string) error {
	folderExists, err := FolderExists(folderName, nil, orgId)

	if err != nil {
		return err
	}

	if folderExists {
		return fmt.Errorf("folder exists")
	}
	statement, err := dbClient.Prepare("INSERT INTO folder (org_id, uploader_id, name) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(orgId, userId, folderName)

	if err != nil {
		return err
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

	_, err = statement.Exec(orgId, userId, folderName, parentFolderName, orgId)
	if err != nil {
		fmt.Print(err.Error())
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

func DeleteFolder(folderId string) error {
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
