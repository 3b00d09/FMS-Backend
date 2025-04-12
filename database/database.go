package database

import (
	"database/sql"
	"fms/auth"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var dbClient *sql.DB

func ConnectDatabase(dbUrl string, dbToken string) {

	dbConnectionUrl := fmt.Sprintf("%s?authToken=%s", dbUrl, dbToken)

	// cant use shorthand err := because dbClient will be locally scoped
	var err error
	dbClient, err = sql.Open("libsql", dbConnectionUrl)
	if err != nil{
		log.Fatal("Error connecting to database")
	}

	// run the schema when the database connection is established
	RunSchema()
}

// takes in username and password, attempts to create user, returns user id or error 
func CreateUser(username string, password string) (string, error){

	usernameExists, err := UsernameExists(username)

	if err != nil{
		return "", err
	}

	if usernameExists{
		return "", fmt.Errorf("Username taken")
	}

	hashedPassword := auth.GenerateHashedPassword(password)
	userId := uuid.New().String()

	statement, err := dbClient.Prepare("INSERT INTO user (id, username, password) VALUES (?, ?, ?)")

	if err != nil{
		return "", err
	}

	defer statement.Close()

	_, err = statement.Exec(userId, username, hashedPassword)

	if err != nil{
		return "", err
	}

	return userId, nil
}


func CreateSession(userId string) (UserSession, error){

	statement, err := dbClient.Prepare("INSERT INTO user_session (id, user_id, expires_at) VALUES (?, ?, ?)")

	if err != nil{
		return UserSession{}, err
	}
	
	defer statement.Close()

	sessionId := uuid.New().String()
	expiresAt := time.Now().Add(3600 * time.Hour * 24 * 7).Unix()

	_, err = statement.Exec(sessionId, userId, expiresAt )

	if err != nil{
		return UserSession{}, err
	}

	return UserSession{
		ID: sessionId,
		UserID: userId,
		ExpiresAt: expiresAt,
	}, nil

}

func GetSession(sessionId string)(UserSession, error){
	var session UserSession

	statement, err := dbClient.Prepare("SELECT * FROM user_session WHERE id = ?")
	if err != nil{
		return session, err
	}
	defer statement.Close()

	err = statement.QueryRow(sessionId).Scan(&session.ID, &session.UserID, &session.ExpiresAt)

	if err != nil{
		return session, err
	}

	return session, nil
	
}

func GetUser(userId string) (User, error){
	var user User

	statement, err := dbClient.Prepare("SELECT id, username FROM user WHERE id = ?")

	if err != nil{
		return user, err
	}

	defer statement.Close()

	err = statement.QueryRow(userId).Scan(&user.ID, &user.Username)

	if err != nil{
		return User{}, err
	}

	return user, nil
}

func GetUserWithSession(sessionId string) UserWithSession{
	var userWithSession UserWithSession

	statement, err := dbClient.Prepare(`
		SELECT user.id, user.username, user_session.id, user_session.expires_at 
		FROM user_session 
		LEFT JOIN user ON user_session.user_id = user.id 
		WHERE user_session.id = ?
	`)

	if err != nil{
		fmt.Println(err.Error())
		return userWithSession
	}

	defer statement.Close()

	err = statement.QueryRow(sessionId).Scan(&userWithSession.User.ID, &userWithSession.User.Username, &userWithSession.Session.ID, &userWithSession.Session.ExpiresAt)

	if err != nil{
		fmt.Println(err.Error())
		return UserWithSession{}
	}

	fmt.Println(userWithSession)
	return userWithSession
}

func CreateOrg(userId string, orgName string) error{
	statement, err := dbClient.Prepare("INSERT INTO organisation (name, creator_id) VALUES (?, ?)")
	if err != nil{
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(orgName, userId)

	if err != nil{
		return err
	}

	return nil
	
}

func GetUserOrg(userId string) Organisation{
	var organisation Organisation

	statement, err := dbClient.Prepare("SELECT * FROM organisation WHERE creator_id = ?")
	if err != nil{
		fmt.Println(err.Error())
		return organisation
	}
	defer statement.Close()

	err = statement.QueryRow(userId).Scan(&organisation.ID, &organisation.Name, &organisation.Creator_id)

	if err != nil{
		return Organisation{}
	}

	fmt.Println(organisation)
	return organisation
}