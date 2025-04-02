package database

import (
	"log"
)

func RunSchema() {

	const schema string = `
	CREATE TABLE IF NOT EXISTS user (
		id TEXT NOT NULL PRIMARY KEY,
		username TEXT NOT NULL,
		password TEXT NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS user_session (
		id TEXT NOT NULL PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES user(id) ON DELETE CASCADE,
		expires_at INTEGER NOT NULL

	);
	`

	_, err := dbClient.Exec(schema)

	// properly handle later, dont want fatals in the app
	if err != nil {
		log.Fatalf("Error running schema: %s\n", err.Error())
	}

}
