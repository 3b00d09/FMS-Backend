package handlers

import (
	"fms/database"

	"github.com/gofiber/fiber/v3"
)

func HandleSearchUsers(c fiber.Ctx) error {
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

	searchInput := c.Query("username")

	if len(searchInput) == 0 {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	result, err := database.SearchUsers(searchInput, user.User.ID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"users": result,
	})

}

func HandleInviteUser(c fiber.Ctx) error {
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

	username := c.Query("username")

	if len(username) == 0 {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	err := database.InviteUserToOrg(username, user.User.ID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusAccepted)
}

func HandleGetUserInvites(c fiber.Ctx) error {
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

	invites, err := database.GetUserInvites(user.User.ID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"invites": invites,
	})
}

func HandleAcceptInvite(c fiber.Ctx) error {
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

	orgId := c.Query("org_id")

	if len(orgId) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Missing Org Id",
		})
	}

	err := database.AcceptOrgInvite(user.User.ID, orgId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusAccepted)
}

func HandleDeclineInvite(c fiber.Ctx) error {
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

	orgId := c.Query("org_id")

	if len(orgId) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Missing Org Id",
		})
	}

	err := database.DeclineOrgInvite(user.User.ID, orgId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusAccepted)
}
