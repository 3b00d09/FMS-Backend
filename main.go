package main

import (
	"fms/database"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

const port string = ":8443"

func main() {

	// server should not run if ENV does not exist
	// instead we get logging and we can track what went wrong and fix it
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file" + err.Error())
	}

	// lookup env instead of loadenv to be EXTRA sure
	dbURL, exists := os.LookupEnv("DATABASE_URL")
	if !exists {
		log.Fatal("ENV Error: DATABASE_URL not found")
	}

	dbToken, exists := os.LookupEnv("DATABASE_TOKEN")
	if !exists {
		log.Fatal("ENV Error: DATABASE_TOKEN not found")
	}

	database.ConnectDatabase(dbURL, dbToken)

	// create a fiber app
	// body limit automatically rejects requests that exceed the defined limit
	// the response is HTTP 413
	// format is mb * 1024 * 10
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024,
	})

	// setup the endpoints for the app
	SetupRoutes(app)

	fmt.Printf("app listening on http://localhost%s\n", port)

	// necessary stuff for cloudflared tunnel
	if err := app.Listen(port, fiber.ListenConfig{
		CertFile:    ".ssl.cert",
		CertKeyFile: ".ssl.key",
	}); err != nil {
		log.Fatal(err)
	}

}
