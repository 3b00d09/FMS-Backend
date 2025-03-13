package database

import (
	"database/sql"
	"fmt"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func ConnectDatabase(dbUrl string, dbToken string) (*sql.DB, error) {

	dbConnectionUrl := fmt.Sprintf("%s?authToken=%s", dbUrl, dbToken)

	return sql.Open("libsql", dbConnectionUrl)
}
