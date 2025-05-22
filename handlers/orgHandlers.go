package handlers

import (
	"fms/database"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func HandleAddOrg(c fiber.Ctx) error {

	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// variable to hold the data submitted
	type addOrgStruct struct {
		Name string
	}

	var addOrgData addOrgStruct

	// attempt to parse request body
	err = c.Bind().Body(&addOrgData)
	// if the server is unable to read the body, it returns a HTTP code 401
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// validate that even if the body has data, it matches what the server expects
	validate := validator.New()

	err = validate.Struct(addOrgData)

	// if the data doesn't match what the server expects
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// attempt to create org in the database
	// the create org func checks for the constraint that ensures only 1 org can be created by a user
	_, err = database.CreateOrg(userWithSession.User.ID, addOrgData.Name)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleChangeOrgName(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	orgId := c.Query("org_id")
	orgName := c.Query("org_name")

	if len(orgId) == 0 || len(orgName) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Missing URL params.",
		})
	}

	canView, role, err := database.CanViewOrg(userWithSession.User.ID, orgId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if strings.ToLower(role) != "owner" || !canView {
		return c.SendStatus(fiber.StatusForbidden)
	}

	err = database.ChangeOrgName(orgId, orgName, userWithSession.User.ID)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)

}

func HandleGetOwnedOrgDetails(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	orgId := c.Query("org_id")

	if len(orgId) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Missing URL params.",
		})
	}

	canView, role, err := database.CanViewOrg(userWithSession.User.ID, orgId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if strings.ToLower(role) != "owner" || !canView {
		return c.SendStatus(fiber.StatusForbidden)
	}

	org := database.GetOrgById(orgId)
	members := database.GetOrgMembers(orgId)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"org":     org,
		"members": members,
	})

}

func HandleViewOrg(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	orgId := c.Query("org_id")

	if len(orgId) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Missing URL params.",
		})
	}

	canView, role, err := database.CanViewOrg(userWithSession.User.ID, orgId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if !canView {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized request. Access Denied.",
		})
	}

	org := database.GetOrgById(orgId)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"org":  org,
		"role": role,
	})

}

func HandleViewOrgMembers(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	ownedOrg := database.GetUserOrg(userWithSession.User.ID)

	if len(ownedOrg.ID) == 0 {
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"org": nil,
		})
	}

	members := database.GetOrgMembers(ownedOrg.ID)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"members": members,
	})

}

func HandleViewUserOrgs(c fiber.Ctx) error {
	// authenticate the request
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// fetch the user's created org and joined orgs
	// data that doesn't exist will return nil (null)

	ownedOrg := database.GetUserOrg(userWithSession.User.ID)
	joinedOrgs := database.GetJoinedOrgs(userWithSession.User.ID)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"joinedOrgs": joinedOrgs,
		"ownedOrg":   ownedOrg,
	})
}

func HandleChangeMemberRole(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	memberUsername := c.Query("username")
	newRole := c.Query("role")

	if len(newRole) == 0 || (newRole != "Editor" && newRole != "Viewer") {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	if len(memberUsername) == 0 {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	err = database.ChangeOrgMemberRole(userWithSession.User.ID, memberUsername, newRole)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleRemoveMember(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	memberUsername := c.Query("username")

	if len(memberUsername) == 0 {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	err = database.RemoveOrgMember(userWithSession.User.ID, memberUsername)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func HandleDeleteOrg(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	orgId := database.GetUserOrg(userWithSession.User.ID)

	if orgId == nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	err = database.DeleteOrg(orgId.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)

}
