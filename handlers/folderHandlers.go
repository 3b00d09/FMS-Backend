package handlers

import (
	"fms/database"
	"fms/ioOperations"
	"fmt"
	"log"
	"mime"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func HandleCreateFolder(c fiber.Ctx) error {

	// authenticate the request
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	type addFolderStruct struct {
		Name   string `json:"name" validate:"required"`
		Org_id string `json:"org_id" validate:"required"`
	}

	var addFolderData addFolderStruct

	err = c.Bind().Body(&addFolderData)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Internal server error.",
		})
	}

	validate := validator.New()

	err = validate.Struct(addFolderData)

	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Missing form data.",
		})
	}

	// check if folder name is empty after trimming
	if strings.TrimSpace(addFolderData.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Folder name cannot be empty",
		})
	}

	// check that folder name contains only alphanumeric characters
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(addFolderData.Name) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Folder name must contain only alphanumeric characters",
		})
	}

	// check if folder name is "root" (case insensitive)
	if strings.ToLower(addFolderData.Name) == "root" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Folder name cannot be root",
		})
	}

	// check if folder name is too long
	if len(addFolderData.Name) > 13 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Folder name is too long. Max length is 13 characters",
		})
	}

	parentFolderName := c.Query("parent-folder")

	if len(parentFolderName) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Missing parent folder name.",
		})
	}

	if addFolderData.Name == parentFolderName {
		return c.SendStatus(fiber.StatusConflict)
	}

	_, role, err := database.CanViewOrg(userWithSession.User.ID, addFolderData.Org_id)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	if strings.ToLower(role) != "owner" && strings.ToLower(role) != "editor" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "You do not have permissions to carry out this operation",
		})
	}

	if parentFolderName == "root" {
		err = database.CreateFolder(userWithSession.User.ID, addFolderData.Name, addFolderData.Org_id)
		if err != nil {
			if strings.Contains(err.Error(), "exists") {
				return c.SendStatus(fiber.StatusConflict)
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	} else {
		err = database.CreateFolderAsChild(userWithSession.User.ID, addFolderData.Name, addFolderData.Org_id, parentFolderName)
		if err != nil {
			if strings.Contains(err.Error(), "exists") {
				return c.SendStatus(fiber.StatusConflict)
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": "true",
	})

}

func HandleViewFolderChildren(c fiber.Ctx) error {
	// authenticate the request
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// grab the data from the url search queries
	folderName := c.Query("folder_name")
	orgId := c.Query("org_id")

	// validate that the data exists
	if len(folderName) == 0 || len(orgId) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Missing URL data",
		})
	}

	// check if the user has permission to view this org's content
	// the middle variable is user role which is irrelevent in this context
	canView, _, err := database.CanViewOrg(userWithSession.User.ID, orgId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if !canView {
		return c.SendStatus(fiber.StatusForbidden)
	}

	// variables to hold the folders and files belonging to an org
	var folderChildren []database.FolderData
	var fileChildren []database.FileData

	// root level folders and files are differnet from others in that they don't have a foreign key to other folders
	// a distinction must be made
	if folderName == "root" {
		folderChildren = database.GetRootFolderOfOrg(orgId)
		fileChildren = database.GetRootFilesOfOrg(orgId)
	} else {
		folderChildren = database.GetFolderChildren(folderName, orgId)
		fileChildren = database.GetFolderFiles(folderName, orgId)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"folders": folderChildren,
		"files":   fileChildren,
	})
}

func HandleUploadFile(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to get file: " + err.Error(),
		})
	}

	orgId := c.FormValue("orgId")
	parentFolderName := c.FormValue("parentFolderName")

	if file == nil || orgId == "" || parentFolderName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required form data",
		})
	}

	// map that holds allowed file types
	allowedFileTypes := map[string]bool{
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"image/jpeg": true,
		"image/png":  true,
	}
	// get file extension and check if it's allowed
	fileExt := filepath.Ext(file.Filename)
	contentType := ""

	// map common extensions to MIME types
	switch strings.ToLower(fileExt) {
	case ".pdf":
		contentType = "application/pdf"
	case ".doc":
		contentType = "application/msword"
	case ".docx":
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	}

	if !allowedFileTypes[contentType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File type not supported. Please upload PDF, DOC, DOCX, JPG, or PNG files",
		})
	}

	// file size validation
	// 10 (mb) * 1024 * 1024
	maxFileSize := int64(10 * 1024 * 1024)
	if file.Size > maxFileSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Upload limit exceeded. Maximum file size is 10MB",
		})
	}

	// file name validation
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	if invalidChars.MatchString(file.Filename) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File name contains invalid characters",
		})
	}

	_, role, err := database.CanViewOrg(userWithSession.User.ID, orgId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	if strings.ToLower(role) != "owner" && strings.ToLower(role) != "editor" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "You do not have permissions to carry out this operation",
		})
	}

	if parentFolderName == "root" {
		err := database.UploadFileToRoot(file, orgId, userWithSession.User.ID)
		if err != nil {
			if strings.Contains(err.Error(), "exists") {
				return c.SendStatus(fiber.StatusConflict)
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	} else {
		err := database.UploadFileToFolder(file, orgId, parentFolderName, userWithSession.User.ID)
		if err != nil {
			if strings.Contains(err.Error(), "exists") {
				return c.SendStatus(fiber.StatusConflict)
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleDeleteFile(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	orgId := c.Query("org-id")
	fileId := c.Query("file-id")
	fileName := c.Query("file-name")

	if len(orgId) == 0 || len(fileId) == 0 || len(fileName) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required form data",
		})
	}

	_, role, err := database.CanViewOrg(userWithSession.User.ID, orgId)

	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	if strings.ToLower(role) != "owner" && strings.ToLower(role) != "editor" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "You do not have permissions to carry out this operation",
		})
	}

	err = database.DeleteFile(fileId, orgId, userWithSession.User.ID, fileName)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleDeleteFolder(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	orgId := c.Query("org-id")
	folderId := c.Query("folder-id")
	folderName := c.Query("folder-name")

	if len(orgId) == 0 || len(folderId) == 0 || len(folderName) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required form data",
		})
	}

	_, role, err := database.CanViewOrg(userWithSession.User.ID, orgId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	if strings.ToLower(role) != "owner" && strings.ToLower(role) != "editor" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "You do not have permissions to carry out this operation",
		})
	}

	err = database.DeleteFolder(folderId, userWithSession.User.ID, orgId, folderName)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleDownloadFile(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	orgId := c.Query("org-id")
	fileId := c.Query("file-id")
	fileType := c.Query("file-type")
	if len(orgId) == 0 || len(fileId) == 0 || len(fileType) == 0 {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// if the file name was not able to be decoded then we fallback to "download"
	// some files have special but not illegal characters that can cause issues if not encoded then decoded
	fileName, err := url.QueryUnescape(c.Query("file-name"))
	if err != nil {
		log.Printf("Error decoding filename: %v", err)
		fileName = "download"
	}

	// verify that the user has permission to download this file
	_, _, err = database.CanViewOrg(userWithSession.User.ID, orgId)

	if err != nil {
		return c.SendStatus(fiber.StatusForbidden)
	}

	// get filepath from this function that walks the database table and collects folder-ids until it hits null which is root level
	filePath, err := database.GetFilePath(fileId)
	if err != nil {
		fmt.Println(err.Error())
	}

	// check that the file actually exists on disk
	err = ioOperations.FileExists(filePath)

	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	// set the response headers to tell the browser to initiate a download operation
	encodedFilename := mime.QEncoding.Encode("utf-8", fileName)
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, encodedFilename, url.PathEscape(fileName)))
	// parse the mime type of the file based on the type
	c.Set("Content-Type", getMimeType(fileType))
	return c.SendFile(filePath)

}

func getMimeType(fileType string) string {
	// Remove the dot if present
	// client does this already but you can never be too safe
	fileType = strings.TrimPrefix(fileType, ".")

	switch strings.ToLower(fileType) {
	case "pdf":
		return "application/pdf"
	case "txt":
		return "text/plain"
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "svg":
		return "image/svg+xml"
	case "doc", "docx":
		return "application/msword"
	case "xls", "xlsx":
		return "application/vnd.ms-excel"
	default:
		// default binary mime type
		return "application/octet-stream"
	}
}
