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
