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

	CREATE TABLE IF NOT EXISTS organisation (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, 
		name TEXT NOT NULL UNIQUE,
		creator_id TEXT NOT NULL REFERENCES user(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS folder(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		org_id INTEGER NOT NULL REFERENCES organisation(id) ON DELETE CASCADE,
		uploader_id TEXT NOT NULL REFERENCES user(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		parent_folder_id INTEGER REFERENCES folder(id) ON DELETE CASCADE DEFAULT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS file(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		folder_id INTEGER NOT NULL REFERENCES folder(id) ON DELETE CASCADE,
		uploader_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		size INTEGER NOT NULL,
		uploaded_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)


	`
	_, err := dbClient.Exec(schema)

	// properly handle later, dont want fatals in the app
	if err != nil {
		log.Fatalf("Error running schema: %s\n", err.Error())
	}

}
