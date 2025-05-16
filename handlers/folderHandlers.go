package handlers

import (
	"fms/database"
	"fmt"
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
		fmt.Println(err.Error())
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

	_, role, err := database.CanViewOrg(userWithSession.User.ID, addFolderData.Org_id)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	if strings.ToLower(role) != "owner" && strings.ToLower(role) != "editor" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "You do not have permissions to carry out this operation",
		})
	}

	parentFolderName := c.Query("parent-folder")

	if len(parentFolderName) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Missing parent folder name.",
		})
	}

	if parentFolderName == "root" {
		err = database.CreateFolder(userWithSession.User.ID, addFolderData.Name, addFolderData.Org_id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	} else {
		err = database.CreateFolderAsChild(userWithSession.User.ID, addFolderData.Name, addFolderData.Org_id, parentFolderName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
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

	_, role, err := database.CanViewOrg(userWithSession.User.ID, orgId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	if strings.ToLower(role) != "owner" && strings.ToLower(role) != "editor" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "You do not have permissions to carry out this operation",
		})
	}

	if parentFolderName == "root" {
		database.UploadFileToRoot(file, orgId, userWithSession.User.ID)
	} else {
		database.UploadFileToFolder(file, orgId, parentFolderName, userWithSession.User.ID)
	}

	return c.SendStatus(fiber.StatusAccepted)
}
