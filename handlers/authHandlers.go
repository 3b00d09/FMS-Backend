package handlers

import (
	"fms/database"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

var passwordLengthMin = 0
var passwordLengthMax = 12

var usernameLengthMin = 0
var usernameLengthMax = 12

func HandleRegister(c fiber.Ctx) error {
	// variable that will hold the form data submitted by the user
	var registerData database.UserCredentials

	// attempt to read request body
	err := c.Bind().Body(&registerData)

	// if the server is unable to read the body, it returns a HTTP code 401
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// validator will ensure that even if the body is successfully read, the data in it matches what the server expects
	validate := validator.New()

	err = validate.Struct(registerData)

	// if the body doesn't contain the data the server expects, the server returns HTTP code 401
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// validate length of input
	if len(registerData.Username) < usernameLengthMin || len(registerData.Username) > usernameLengthMax {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Username length must be between 6 and 12 characters",
		})
	}

	if len(registerData.Password) < passwordLengthMin || len(registerData.Password) > passwordLengthMax {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Password length must be between 6 and 12 characters",
		})
	}

	// attemps to create a user and return a session ID if successful
	userId, err := database.CreateUser(registerData.Username, registerData.Password)

	// user creation failed
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	session, err := database.CreateSession(userId)

	// session creation failed
	// the failed at field helps the client figure out if the user creation and session creation both failed or just one of them failed
	// this is important as we don't want the user to register again with the same credentials in case a user was created successfully but session failed to create
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error":    err.Error(),
			"failedAt": "session",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"session": session,
	})
}

func HandleLogin(c fiber.Ctx) error {
	var loginData database.UserCredentials

	// attempt to read request body
	err := c.Bind().Body(&loginData)

	// if the server is unable to read the body, it returns a HTTP code 401
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// validator will ensure that even if the body is successfully read, the data in it matches what the server expects
	validate := validator.New()

	err = validate.Struct(loginData)

	// if the body doesn't contain the data the server expects, the server returns HTTP code 401
	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// validate length of input
	if len(loginData.Username) < usernameLengthMin || len(loginData.Username) > usernameLengthMax {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Username length must be between 6 and 12 characters",
		})
	}

	if len(loginData.Password) < passwordLengthMin || len(loginData.Password) > passwordLengthMax {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Password length must be between 6 and 12 characters",
		})
	}

	// attempt to match the submitted credentials against the database
	userId, err := database.UserExists(loginData.Username, loginData.Password)

	// if credentials don't match
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// attempt to create a session for the user after successfully matching credentials
	session, err := database.CreateSession(userId)

	// if session creation was unsuccessful
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// return session details so the client can create a browser cookie since cookies can't be created directly here
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"session": session,
	})

}

// function to be called on every request
// attemps to parse session_token cookie and passes into a function that authenticates the cookie
// if the cookie is valid we return the user if not we return nil
func AuthRequest(c fiber.Ctx) error {
	userWithSession, err := database.AuthenticateCookie(c.Cookies("session_token"))

	if err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"error": nil,
		"user":  userWithSession.User,
	})

}

func HandleLogout(c fiber.Ctx) error {
	// attempt to read in the session cookie
	cookie := c.Cookies("session_token")
	if len(cookie) == 0 {
		// if the cookie doesn't exist this endpoint shouldn't be computed
		c.Status(fiber.StatusUnauthorized)
	}

	// delete session from the database
	database.InvalidateSession(cookie)

	return c.SendStatus(fiber.StatusOK)
}
