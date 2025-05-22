package handlers

import (
	"fms/database"
	"fmt"
	"strings"

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
		fmt.Println(err.Error())
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if hasExceededLimit {
		return c.SendStatus(fiber.StatusConflict)
	}

	err = database.AcceptOrgInvite(userWithSession.User.ID, orgId, userWithSession.User.Username)
	if err != nil {
		fmt.Println(err.Error())
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

	return c.SendStatus(fiber.StatusOK)
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

func HandleChangePassword(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	currPassword := c.FormValue("current-password")
	newPassword := c.FormValue("new-password")
	confirmNewPassword := c.FormValue("confirm-new-password")

	if newPassword != confirmNewPassword {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Passwords don't match",
		})
	}

	// dont care about the return value of this function other than error
	// if there is no error user exists
	_, err = database.UserExists(userWithSession.User.Username, currPassword)

	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	err = database.ChangePassword(userWithSession.User.ID, newPassword)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusOK)

	// if len(registerData.Password) < passwordLengthMin || len(registerData.Password) > passwordLengthMax {
	// 	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
	// 		"error": "Password length must be between 6 and 12 characters",
	// 	})
	// }
}

func HandleChangeUsername(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	username := c.FormValue("username")
	username = strings.TrimSpace(username)

	if len(username) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "username cannot be empty",
		})
	}

	// validate length of input
	if len(username) < usernameLengthMin || len(username) > usernameLengthMax {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Username length must be between 6 and 12 characters",
		})
	}

	exists, err := database.UsernameExists(username)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	if exists {
		return c.SendStatus(fiber.StatusConflict)
	}
	err = database.ChangeUsername(userWithSession.User.ID, username)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusOK)

}

func HandleDeleteAccount(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	err = database.DeleteAccount(userWithSession.User.ID)

	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusOK)
}
