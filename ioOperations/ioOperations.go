package ioOperations

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// this file uses mkdirall instead of mkdir because mkdir all creates the necessary parent folders
// can be handy if the system runs into inconsistences between database and storage data
// the second argument to mkdirall is the chmod octal value of permissions
// owner rwx, group rx, public rx
// https://chmod-calculator.com/
// go uses new octal notation and 0o755 is the same as 755 in the chmod calculator

// when an org is created, this method creates a fodler for that org at appdata root level
func CreateOrgDir(orgId string) error {
	// build the path for the org's folder
	path := filepath.Join("appdata", fmt.Sprint("org-", orgId))
	err := os.MkdirAll(path, 0o755)
	if err != nil {
		return err
	}
	return nil
}

// delete the org folder and all data inside it
func DeleteOrgDir(orgId string) error {
	// build the path for the org's folder
	path := filepath.Join("appdata", fmt.Sprint("org-", orgId))
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}

// creates a folder for an org at root level
func CreateOrgFolderAtRoot(folderId string, orgId string) error {
	path := filepath.Join("appdata", fmt.Sprint("org-", orgId), fmt.Sprint("folder-", folderId))
	err := os.MkdirAll(path, 0o755)
	if err != nil {
		return err
	}
	return nil
}

// creates a folder inside another folder in an org
// this function takes a path that is already built by the calling function
// the calling function recursively walks the folder table and builds a path via all parent_folder_id fields
func CreateOrgFolderAsChild(path string) error {
	err := os.MkdirAll(path, 0o755)
	if err != nil {
		return err
	}
	return nil
}

// doesn't care about root or not root level folders, the calling function will build the path and pass it in
// the path built is based off parent_folder_id fields in the database
func DeleteOrgFolder(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}

// creates a file for an org at root level
func CreateOrgFileAtRoot(file *multipart.FileHeader, fileId string, orgId string) error {
	orgPath := fmt.Sprintf("appdata/org-%s", orgId)
	filePath := filepath.Join(orgPath, fmt.Sprintf("file-%s", fileId))

	err := os.MkdirAll(orgPath, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create organisation directory in create file at root: %s", err.Error())
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	destination, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create destination path in create file at root: %s", err.Error())
	}
	defer destination.Close()

	_, err = io.Copy(destination, src)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %s", err.Error())
	}
	return nil
}

// creates a file for an org at as a child of a folder
func CreateOrgFileAsChild(file *multipart.FileHeader, fileId string, folderPath string) error {
	filePath := filepath.Join(folderPath, fmt.Sprintf("file-%s", fileId))

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	destination, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create destination path in create file at root: %s", err.Error())
	}
	defer destination.Close()

	_, err = io.Copy(destination, src)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %s", err.Error())
	}
	return nil
}

// doesn't care about root or not root level files, the calling function will build the path and pass it in
// the path built is based off folder_id fields in the database
func DeleteOrgChildFile(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}

func FileExists(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file does not exist: %s", err.Error())
	}
	return nil
}
