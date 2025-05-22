package database

import (
	"database/sql"
	"fms/auth"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// takes in username and password, attempts to create user, returns user id or error
func CreateUser(username string, password string) (string, error) {

	usernameExists, err := UsernameExists(username)

	if err != nil {
		return "", err
	}

	if usernameExists {
		return "", fmt.Errorf("username taken")
	}

	hashedPassword := auth.GenerateHashedPassword(password)
	userId := uuid.New().String()

	statement, err := dbClient.Prepare("INSERT INTO user (id, username, password) VALUES (?, ?, ?)")

	if err != nil {
		return "", err
	}

	defer statement.Close()

	_, err = statement.Exec(userId, username, hashedPassword)

	if err != nil {
		return "", err
	}

	return userId, nil
}

func CreateSession(userId string) (UserSession, error) {

	statement, err := dbClient.Prepare("INSERT INTO user_session (id, user_id, expires_at) VALUES (?, ?, ?)")

	if err != nil {
		return UserSession{}, err
	}

	defer statement.Close()

	sessionId := uuid.New().String()
	// 30 days expiry
	expiresAt := time.Now().Add(time.Hour * 24 * 30).Unix()

	_, err = statement.Exec(sessionId, userId, expiresAt)

	if err != nil {
		return UserSession{}, err
	}

	return UserSession{
		ID:        sessionId,
		UserID:    userId,
		ExpiresAt: expiresAt,
	}, nil

}

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
			return "", fmt.Errorf("invalid username")
		}
		return "", err
	}

	if !auth.CheckPasswordHash(password, user.password) {
		return "", fmt.Errorf("incorrect password")
	}

	return user.id, nil

}

func GetUser(userId string) (User, error) {
	var user User

	statement, err := dbClient.Prepare("SELECT id, username FROM user WHERE id = ?")

	if err != nil {
		return user, err
	}

	defer statement.Close()

	err = statement.QueryRow(userId).Scan(&user.ID, &user.Username)

	if err != nil {
		return User{}, err
	}

	return user, nil
}

func GetUserWithSession(sessionId string) UserWithSession {
	var userWithSession UserWithSession

	statement, err := dbClient.Prepare(`
		SELECT user.id, user.username, user_session.id, user_session.expires_at 
		FROM user_session 
		LEFT JOIN user ON user_session.user_id = user.id 
		WHERE user_session.id = ?
	`)

	if err != nil {
		return userWithSession
	}

	defer statement.Close()

	err = statement.QueryRow(sessionId).Scan(&userWithSession.User.ID, &userWithSession.User.Username, &userWithSession.Session.ID, &userWithSession.Session.ExpiresAt)

	if err != nil {
		return UserWithSession{}
	}

	return userWithSession
}

func InvalidateSession(sessionId string) {
	statement, err := dbClient.Prepare("DELETE FROM user_session WHERE user_session.id = ?")

	if err != nil {
		fmt.Println(err)
	}

	defer statement.Close()

	_, err = statement.Exec(sessionId)

	if err != nil {
		fmt.Println(err)
	}
}

func GetUsernameById(id string) (string, error) {
	var username string

	statement, err := dbClient.Prepare("SELECT username FROM user WHERE username =  LIMIT 1)")
	if err != nil {
		return username, err
	}

	defer statement.Close()

	err = statement.QueryRow(username).Scan(&username)

	if err != nil {
		return username, err
	}

	return username, nil
}

func ChangePassword(userId string, password string) error {
	statement, err := dbClient.Prepare("UPDATE user SET password = ? WHERE id = ?")
	if err != nil {
		return err
	}

	defer statement.Close()

	hashedPassword := auth.GenerateHashedPassword(password)

	result, err := statement.Exec(hashedPassword, userId)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("could not update password. please try again later or contact support")
	}

	return nil

}

func ChangeUsername(userId string, username string) error {
	statement, err := dbClient.Prepare("UPDATE user SET username = ? WHERE id = ?")
	if err != nil {
		return err
	}

	defer statement.Close()

	result, err := statement.Exec(username, userId)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("could not update username. please try again later or contact support")
	}

	return nil
}

func DeleteAccount(userId string) error {
	statement, err := dbClient.Prepare("DELETE FROM user WHERE id = ?")
	if err != nil {
		return err
	}

	defer statement.Close()

	result, err := statement.Exec(userId)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("could not delete account. please try again later or contact support")
	}

	return nil
}
