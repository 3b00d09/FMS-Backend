package database

import (
	"database/sql"
	"fms/auth"
	"fmt"
)

func UsernameExists(username string) (bool, error) {
	var exists bool

	statement, err := dbClient.Prepare("SELECT EXISTS (SELECT id FROM user WHERE username = ? LIMIT 1)")
	if err != nil {
		return false, err
	}

	defer statement.Close()

	err = statement.QueryRow(username).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func UserExists(username string, password string) (string, error) {
	statement, err := dbClient.Prepare("SELECT * FROM user WHERE username = ?")
	if err != nil {
		return "", err
	}
	defer statement.Close()

	var user struct {
		id       string
		username string
		password []byte
	}

	err = statement.QueryRow(username).Scan(&user.id, &user.username, &user.password)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("User does not exist")
		}
		return "", err
	}

	if !auth.CheckPasswordHash(password, user.password) {
		return "", fmt.Errorf("Incorrect password.")
	}

	return user.id, nil

}
