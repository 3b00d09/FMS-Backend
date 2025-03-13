package database

import (
	"database/sql"
	"log"
)

func RunSchema(db *sql.DB) {

	const schema string = `
	CREATE TABLE IF NOT EXISTS user (
		id TEXT NOT NULL PRIMARY KEY,
		username TEXT NOT NULL,
		password TEXT NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS user_session (
		id TEXT NOT NULL PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES user(id) ON DELETE CASCADE,
		active_expires INTEGER NOT NULL
	);
	`

	_, err := db.Exec(schema)

	if err != nil {
		log.Fatalf("Error running schema: %s\n", err.Error())
	}

}
