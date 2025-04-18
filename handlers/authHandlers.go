package handlers

import (
	"fms/database"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func HandleRegister(c fiber.Ctx) error {
	var registerData database.UserCredentials;

	err := c.Bind().Body(&registerData)

	if err != nil{
		fmt.Println(err.Error())
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Internal server error.",
		})
	}

	validate := validator.New()

	err = validate.Struct(registerData);

	if err != nil{
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Missing form data.",
		})
	}

	fmt.Println(registerData.Password, registerData.Username)

	// attemps to create a user and return a session ID if successful
	userId, err := database.CreateUser(registerData.Username, registerData.Password)

	if(err != nil){
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	session, err := database.CreateSession(userId)

		if(err != nil){
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"session": session,
	})
}

func HandleLogin(c fiber.Ctx) error{
	var loginData database.UserCredentials;

	err := c.Bind().Body(&loginData)

	if err != nil{
		fmt.Println(err.Error())
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Internal server error.",
		})
	}

	validate := validator.New()

	err = validate.Struct(loginData);

	if err != nil{
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Missing form data.",
		})
	}
	
	userId, err := database.UserExists(loginData.Username, loginData.Password)

	if(err != nil){
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	session, err := database.CreateSession(userId)

	if err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"session": session,
	})

}


func AuthRequest(c fiber.Ctx) error{
	cookie := c.Cookies("session_token")
	if len(cookie) == 0 {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "Missing cookie",
		})
	}

	session, err := database.GetSession(cookie);

	if err != nil{
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if (session.ExpiresAt < time.Now().Unix()){
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message":"cookie expired",
		})
	}

	user, err := database.GetUser(session.UserID)

	if err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"user": user,
	})

}