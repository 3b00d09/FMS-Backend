package main

import (
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func SetupRoutes(app *fiber.App) {

	app.Use(cors.New(cors.Config{
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Content-Length", "Accept-Language", "Accept-Encoding", "Connection", "Access-Control-Allow-Origin"},
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowCredentials: true,
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello world!")
	})

	app.Get("/auth-test", func(c fiber.Ctx) error {
		cookie := c.Cookies("session_token")
		if len(cookie) == 0 {
			c.Status(fiber.StatusUnauthorized)
			return c.JSON(fiber.Map{
				"message": "Missing cookie",
			})
		}

		return c.JSON(fiber.Map{
			"message": "success",
			"value":   cookie,
		})
	})

	app.Get("/cookie-test", func(c fiber.Ctx) error {
		cookie := &fiber.Cookie{
			Name:   "session_token",
			Value:  "123456",
			MaxAge: int(time.Now().Add(3600 * time.Hour * 24 * 7).Unix()),
			Path:   "/",
			Domain: "localhost",
		}

		c.Cookie(cookie)
		return c.JSON(fiber.Map{
			"message": "cookie set",
		})
	})

	app.Post("/upload-test", func(c fiber.Ctx) error{
		file, err := c.FormFile("file")
		if err != nil{
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "failed to get file " + err.Error(),
			})
		}

		fileName := file.Filename
		fileExt := filepath.Ext(fileName)

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"success": "file name: " + fileName + " file ext: " + fileExt,
		})
	})
}
