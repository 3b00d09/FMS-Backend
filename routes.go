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
		if c.Get("X-Forwarded-Proto") != "https" {
			redirectURL := "https://" + c.Hostname() + c.OriginalURL()
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
	app.Get("/owned-org", handlers.HandleGetOwnedOrgDetails)
	app.Get("view-org", handlers.HandleViewOrg)
	app.Get("/view-org-members", handlers.HandleViewOrgMembers)
	app.Get("/invite-user", handlers.HandleInviteUser)
	app.Put("/change-org-name", handlers.HandleChangeOrgName)
	app.Put("/update-member-role", handlers.HandleChangeMemberRole)
	app.Delete("/remove-member", handlers.HandleRemoveMember)
	app.Delete("/delete-org", handlers.HandleDeleteOrg)

	// folder-related routes
	app.Get("/view-folder-children", handlers.HandleViewFolderChildren)
	app.Post("/add-folder", handlers.HandleCreateFolder)
	app.Post("/add-file", handlers.HandleUploadFile)
	app.Delete("/delete-file", handlers.HandleDeleteFile)
	app.Delete("/delete-folder", handlers.HandleDeleteFolder)
	app.Get("/download-file", handlers.HandleDownloadFile)

	// user-related routes
	app.Get("/view-user-orgs", handlers.HandleViewUserOrgs)
	app.Get("/users", handlers.HandleSearchUsers)
	app.Get("/user-invites", handlers.HandleGetUserInvites)
	app.Get("/accept-invite", handlers.HandleAcceptInvite)
	app.Get("/decline-invite", handlers.HandleDeclineInvite)
	app.Get("notifications", handlers.HandleGetUserNotifications)
	app.Get("/read-notification", handlers.HandleMarkNotificationAsRead)
	app.Post("/change-password", handlers.HandleChangePassword)
	app.Post("/change-username", handlers.HandleChangeUsername)
	app.Delete("/delete-account", handlers.HandleDeleteAccount)
}
