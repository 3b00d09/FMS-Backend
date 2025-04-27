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

func CreateOrg(userId string, orgName string) error {
	statement, err := dbClient.Prepare("INSERT INTO organisation (name, creator_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(orgName, userId)

	if err != nil {
		return err
	}

	return nil

}

func GetUserOrg(userId string) Organisation {
	var organisation Organisation

	statement, err := dbClient.Prepare("SELECT * FROM organisation WHERE creator_id = ?")
	if err != nil {
		fmt.Println(err.Error())
		return organisation
	}
	defer statement.Close()

	err = statement.QueryRow(userId).Scan(&organisation.ID, &organisation.Name, &organisation.Creator_id)

	if err != nil {
		return Organisation{}
	}

	return organisation
}

func CreateFolder(userId string, folderName string, orgId string) error {
	statement, err := dbClient.Prepare("INSERT INTO folder (org_id, uploader_id, name) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(orgId, userId, folderName)

	if err != nil {
		return err
	}

	return nil
}
