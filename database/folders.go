package database

func CreateFolder(userId string, folderName string, orgId string) error {
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

func GetRootFolderOfOrg(orgId string) []FolderData {

	var folders []FolderData

	statement, err := dbClient.Prepare(`
    	SELECT 
        folder.id, folder.org_id, user.username, folder.name, 
        folder.parent_folder_id, folder.created_at
    	FROM folder 
    	LEFT JOIN user ON user.id = folder.uploader_id
    	WHERE org_id = ? AND parent_folder_id IS NULL
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
