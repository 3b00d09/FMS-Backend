package database

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
)

func UploadFileToRoot(file *multipart.FileHeader, orgId string, uploaderId string) {
	statement, err := dbClient.Prepare(`
	 	INSERT INTO file (org_id, uploader_id, name, type, size)
		VALUES (?, ?, ?, ?, ?)
	 `)

	if err != nil {
		fmt.Print(err.Error())
	}

	defer statement.Close()

	_, err = statement.Exec(orgId, uploaderId, file.Filename, filepath.Ext(file.Filename), file.Size)
	if err != nil {
		fmt.Print(err.Error())
	}
}

func UploadFileToFolder(file *multipart.FileHeader, orgId string, parentFolderName string, uploaderId string) {
	statement, err := dbClient.Prepare(`
	 	INSERT INTO file (org_id, uploader_id, name, type, size, folder_id)
		VALUES (?, ?, ?, ?, ?, (SELECT id FROM folder WHERE name = ? AND org_id = ?))
	 `)

	if err != nil {
		fmt.Print(err.Error())
	}

	defer statement.Close()

	_, err = statement.Exec(orgId, uploaderId, file.Filename, filepath.Ext(file.Filename), file.Size, parentFolderName, orgId)
	if err != nil {
		fmt.Print(err.Error())
	}
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
