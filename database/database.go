package database

import (
	"database/sql"
	"fmt"
	"log"

	// keep this import even though IDE says its unused we need it to connect to DB
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var dbClient *sql.DB

func ConnectDatabase(dbUrl string, dbToken string) {

	dbConnectionUrl := fmt.Sprintf("%s?authToken=%s", dbUrl, dbToken)

	// cant use shorthand err := because dbClient will be locally scoped
	var err error
	dbClient, err = sql.Open("libsql", dbConnectionUrl)
	if err != nil {
		fmt.Println(err.Error())
		log.Fatal("Error connecting to database")
	}

	// run the schema when the database connection is established
	RunSchema()
}
