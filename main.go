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

	dbClient, err := database.ConnectDatabase(dbURL, dbToken)

	if err != nil {
		log.Fatalf("Error connecting to database: %s \n", err.Error())
	}

	// close when main ends
	defer dbClient.Close()

	// run the schema when the database connection is established
	database.RunSchema(dbClient)

	app := fiber.New()
	SetupRoutes(app)

	fmt.Printf("app listening on http://localhost%s\n", port)
	app.Listen(port)

}
