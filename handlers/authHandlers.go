package handlers

import (
	"fms/database"
	"fmt"
	"time"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/gofiber/fiber/v3"
)

func HandleRegister(c fiber.Ctx) error {
	var registerData database.UserCredentials

	err := c.Bind().Body(&registerData)

	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Internal server error.",
		})
	}

	validate := validator.New()

	err = validate.Struct(registerData)

	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Missing form data.",
		})
	}

	// attemps to create a user and return a session ID if successful
	userId, err := database.CreateUser(registerData.Username, registerData.Password)

	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	session, err := database.CreateSession(userId)

	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"session": session,
	})
}

func HandleLogin(c fiber.Ctx) error {
	var loginData database.UserCredentials

	err := c.Bind().Body(&loginData)

	if err != nil {
		fmt.Println(err.Error())
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "Internal server error.",
		})
	}

	validate := validator.New()
	// build an English translator
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ := uni.GetTranslator("en")

	// register the built‐in English translations for tags like "required", "email", etc.
	en_translations.RegisterDefaultTranslations(validate, trans)

	err = validate.Struct(loginData)

	if err != nil {
		errs := err.(validator.ValidationErrors)

		// this gives you a map[fieldName]translatederror
		translated := errs.Translate(trans)

		return c.
			Status(fiber.StatusBadRequest).
			JSON(fiber.Map{
				"error":  "Missing form data",
				"errors": translated,
			})
	}

	userId, err := database.UserExists(loginData.Username, loginData.Password)

	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	session, err := database.CreateSession(userId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"session": session,
	})

}

// function to be called on every request
// reads in cookies and looks for session token, if it exists then validate token, otherwise return http unauthorised
func AuthRequest(c fiber.Ctx) error {
	cookie := c.Cookies("session_token")
	if len(cookie) == 0 {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"error": "Missing cookie",
		})
	}

	userWithSession := database.GetUserWithSession(cookie)

	// if the token exists but the value is invalid we won't get a user
	if userWithSession.User.ID == "" {
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized request",
		})
	}

	// validate session lifetime
	if userWithSession.Session.ExpiresAt < time.Now().Unix() {
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "cookie expired",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"user": userWithSession.User,
	})

}

func HandleLogout(c fiber.Ctx) error {
	cookie := c.Cookies("session_token")
	if len(cookie) == 0 {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"error": "Missing cookie",
		})
	}

	err := database.InvalidateSession(cookie)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": "true",
	})
}
