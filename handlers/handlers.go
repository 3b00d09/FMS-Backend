package handlers

import (
	"fms/database"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func HandleAddOrg(c fiber.Ctx) error {
	type addOrgStruct struct {
		Name       string
		Creator_id string
	}

	var addOrgData addOrgStruct

	err := c.Bind().Body(&addOrgData)
	if err != nil {
		fmt.Println(err.Error())
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Internal server error.",
		})
	}

	validate := validator.New()

	err = validate.Struct(addOrgData)

	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Missing form data.",
		})
	}

	err = database.CreateOrg(addOrgData.Creator_id, addOrgData.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Success",
	})
}

func HandleGetOwnedOrg(c fiber.Ctx) error {
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

	org := database.GetUserOrg(user.User.ID)

	return c.JSON(fiber.Map{
		"data": org,
	})

}

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

	fmt.Println(user.User.ID, addFolderData.Name, addFolderData.Org_id)
	err = database.CreateFolder(user.User.ID, addFolderData.Name, addFolderData.Org_id)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": "true",
	})

}

func HandleGetRootFolder(c fiber.Ctx) error {
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

	// this should eventually verify that the user is a member of the org first
	var rootFolder []database.FolderData = database.GetRootFolderOfOrg(c.Query("org_id"))

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"data": rootFolder,
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

	var folderChildren = database.GetFolderChildren(c.Query("folder_name"), c.Query("org_id"))
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"data": folderChildren,
	})
}
