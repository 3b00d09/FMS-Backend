package database

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
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
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("file name already exists in this location")
		} else {
			return err
		}
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
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("file name already exists in this location")
		} else {
			return err
		}
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

func DeleteFile(fileId string) error {
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
	return nil
}
