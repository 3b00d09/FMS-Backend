package handlers

import (
	"fms/database"

	"github.com/gofiber/fiber/v3"
)

func HandleSearchUsers(c fiber.Ctx) error {
	// authenticate the request
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	searchInput := c.Query("username")

	if len(searchInput) == 0 {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	result, err := database.SearchUsers(searchInput, userWithSession.User.ID)

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
	// authenticate the request
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	username := c.Query("username")
	orgId := c.Query("org-id")

	if len(username) == 0 {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// no need to permission check on this function because it passes in the requesting user's id so the invite will automatically go to their org
	err = database.InviteUserToOrg(username, userWithSession.User.ID, orgId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleGetUserInvites(c fiber.Ctx) error {
	// authenticate the request
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	invites, err := database.GetUserInvites(userWithSession.User.ID)

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
	// authenticate the request
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	orgId := c.Query("org_id")

	if len(orgId) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Missing Org Id",
		})
	}

	hasExceededLimit, err := database.HasExceededLimit(userWithSession.User.ID)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if hasExceededLimit {
		return c.SendStatus(fiber.StatusConflict)
	}

	err = database.AcceptOrgInvite(userWithSession.User.ID, orgId, userWithSession.User.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusAccepted)
}

func HandleDeclineInvite(c fiber.Ctx) error {
	// authenticate the request
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	orgId := c.Query("org_id")

	if len(orgId) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Missing Org Id",
		})
	}

	err = database.DeclineOrgInvite(userWithSession.User.ID, orgId, userWithSession.User.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusAccepted)
}

func HandleGetUserNotifications(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	notifications, err := database.GetUserNotifications(userWithSession.User.ID)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"notifications": notifications,
	})
}

func HandleMarkNotificationAsRead(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	notifId := c.Query("id")
	// this endpoint optionally takes an "all" flag that lets the server know whether all notifs should be marked as read or just one
	clearAll := c.Query("all")

	if len(notifId) == 0 && len(clearAll) == 0 {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	if len(clearAll) > 0 {
		err = database.MarkAllAsRead(userWithSession.User.ID)

		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.SendStatus(fiber.StatusOK)

	} else {
		err = database.MarkAsRead(notifId, userWithSession.User.ID)

		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.SendStatus(fiber.StatusOK)
	}
}
