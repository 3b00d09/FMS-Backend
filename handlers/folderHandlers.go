package handlers

import (
	"fms/database"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func HandleCreateFolder(c fiber.Ctx) error {

	type addFolderStruct struct {
		Name   string `json:"name" validate:"required"`
		Org_id string `json:"org_id" validate:"required"`
	}

	var addFolderData addFolderStruct

	err := c.Bind().Body(&addFolderData)
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

	parentFolderName := c.Query("parent-folder")

	if len(parentFolderName) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Missing parent folder name.",
		})
	}

	cookie := c.Cookies("session_token")
	if len(cookie) == 0 {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Missing cookie",
		})
	}

	user := database.GetUserWithSession(cookie)

	if len(user.User.ID) == 0 {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// need to do permissions later

	if parentFolderName == "root" {
		err = database.CreateFolder(user.User.ID, addFolderData.Name, addFolderData.Org_id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	} else {
		err = database.CreateFolderAsChild(user.User.ID, addFolderData.Name, addFolderData.Org_id, parentFolderName)
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
	cookie := c.Cookies("session_token")
	if len(cookie) == 0 {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Missing cookie",
		})
	}

	user := database.GetUserWithSession(cookie)

	if len(user.User.ID) == 0 {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	folderName := c.Query("folder_name")
	orgId := c.Query("org_id")

	if len(folderName) == 0 || len(orgId) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing URL data",
		})
	}

	var folderChildren []database.FolderData
	var fileChildren []database.FileData

	if folderName == "root" {
		folderChildren = database.GetRootFolderOfOrg(orgId)
		fileChildren = database.GetRootFilesOfOrg(orgId)
	} else {
		folderChildren = database.GetFolderChildren(folderName, orgId)
		fileChildren = database.GetFolderFiles(folderName, orgId)
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"folders": folderChildren,
		"files":   fileChildren,
	})
}
