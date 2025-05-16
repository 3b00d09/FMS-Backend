package main

import (
	"fms/handlers"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func SetupRoutes(app *fiber.App) {

	// configuring the app
	app.Use(cors.New(cors.Config{
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Content-Length", "Accept-Language", "Accept-Encoding", "Connection", "Access-Control-Allow-Origin"},
		AllowOrigins:     []string{"http://localhost:5173", "https://fmsatiya.live"},
		AllowMethods:     []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowCredentials: true,
	}))

	// even though cloudflare seems to handle redirects, can never be too safe
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

	// auth routes
	app.Post("/register", handlers.HandleRegister)
	app.Post("/login", handlers.HandleLogin)
	app.Get("/logout", handlers.HandleLogout)
	app.Get("/auth-user", handlers.AuthRequest)

	// org-related routes
	app.Post("/add-org", handlers.HandleAddOrg)
	// currently unused
	app.Get("/owned-org", handlers.HandleGetOwnedOrg)
	app.Get("view-org", handlers.HandleViewOrg)
	app.Get("/view-org-members", handlers.HandleViewOrgMembers)
	app.Get("/invite-user", handlers.HandleInviteUser)

	// folder-related routes
	app.Get("/view-folder-children", handlers.HandleViewFolderChildren)
	app.Post("/add-folder", handlers.HandleCreateFolder)
	app.Post("/add-file", handlers.HandleUploadFile)

	// user-related routes
	app.Get("/view-user-orgs", handlers.HandleViewUserOrgs)
	app.Get("/users", handlers.HandleSearchUsers)
	app.Get("/user-invites", handlers.HandleGetUserInvites)
	app.Get("/accept-invite", handlers.HandleAcceptInvite)
	app.Get("/decline-invite", handlers.HandleDeclineInvite)
}
