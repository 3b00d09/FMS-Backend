package database

import "fmt"

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

func GetOrgById(orgId string) Organisation {
	var organisation Organisation

	statement, err := dbClient.Prepare("SELECT * FROM organisation WHERE id = ?")
	if err != nil {
		fmt.Println(err.Error())
		return organisation
	}
	defer statement.Close()

	err = statement.QueryRow(orgId).Scan(&organisation.ID, &organisation.Name, &organisation.Creator_id)

	if err != nil {
		return Organisation{}
	}

	return organisation

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
