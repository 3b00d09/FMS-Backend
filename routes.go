package main

import (
	"fms/handlers"
	"path/filepath"

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

	app.Post("/register", handlers.HandleRegister)
	app.Post("/login", handlers.HandleLogin)

	app.Get("/auth-user", handlers.AuthRequest)

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
