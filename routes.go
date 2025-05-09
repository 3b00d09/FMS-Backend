package main

import (
	"fms/database"
	"fms/handlers"
	"fmt"
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

	// even though cloudflare seems to handle redirects, can never be too safe?
	// middleware to force https
	app.Use(func(c fiber.Ctx) error {
		// Cloudflare sets this header for you
		if c.Get("X-Forwarded-Proto") != "https" {
			redirectURL := "https://" + c.Hostname() + c.OriginalURL()
			// Build a 301 redirect and return its error
			return c.
				Redirect().
				Status(fiber.StatusMovedPermanently).
				To(redirectURL)
		}
		return c.Next()
	})

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello world!")
	})

	app.Post("/register", handlers.HandleRegister)
	app.Post("/login", handlers.HandleLogin)
	app.Get("/logout", handlers.HandleLogout)

	app.Get("/auth-user", handlers.AuthRequest)

	app.Post("/add-org", handlers.HandleAddOrg)
	app.Get("/owned-org", handlers.HandleGetOwnedOrg)

	app.Post("/add-folder", handlers.HandleCreateFolder)

	app.Get("/view-folder-children", handlers.HandleViewFolderChildren)

	app.Post("/upload-test", func(c fiber.Ctx) error {
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

		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "failed to parse form: " + err.Error(),
			})
		}

		files := form.File["files[]"]
		orgId := form.Value["orgId"]
		parentFolderName := form.Value["parentFolderName"]
		if len(files) == 0 || len(orgId) == 0 || len(parentFolderName) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing Form Data.",
			})
		}

		if parentFolderName[0] == "root" {
			database.UploadFileToRoot(files[0], orgId[0], user.User.ID)
		} else {
			database.UploadFileToFolder(files[0], orgId[0], parentFolderName[0], user.User.ID)
		}

		fileInfo := []map[string]string{}
		for _, file := range files {
			fileName := file.Filename
			fileExtension := filepath.Ext(fileName)
			fileSize := file.Size

			fileInfo = append(fileInfo, map[string]string{
				"name": fileName,
				"type": fileExtension,
				"size": fmt.Sprint(fileSize),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data":   fileInfo,
			"folder": parentFolderName,
			"org":    orgId,
		})
	})
}
