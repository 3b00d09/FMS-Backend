package main

import (
	"fms/database"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

const port string = ":3000"

func main() {

	// NEED TO REDO SO THE FRONTEND KNOWS THE DATABASE IS DEAD WITHOUT KILLING ENTIRE SERVER
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// lookup env instead of loadenv to be EXTRA sure :)
	dbURL, exists := os.LookupEnv("DATABASE_URL")
	if !exists {
		log.Fatal("ENV Error: DATABASE_URL not found")
	}

	dbToken, exists := os.LookupEnv("DATABASE_TOKEN")
	if !exists {
		log.Fatal("ENV Error: DATABASE_TOKEN not found")
	}

	database.ConnectDatabase(dbURL, dbToken)

	app := fiber.New()
	SetupRoutes(app)

	fmt.Printf("app listening on http://localhost%s\n", port)
	app.Listen(port)

}
