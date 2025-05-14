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

	statement, err := dbClient.Prepare("SELECT id, name, creator_id FROM organisation WHERE id = ?")
	if err != nil {
		fmt.Println(err.Error())
		return organisation
	}
	defer statement.Close()

	err = statement.QueryRow(orgId, orgId).Scan(&organisation.ID, &organisation.Name, &organisation.Creator_id)

	if err != nil {
		return Organisation{}
	}

	return organisation

}

func GetUserOrg(userId string) Organisation {
	var organisation Organisation

	statement, err := dbClient.Prepare("SELECT id, name, creator_id FROM organisation WHERE creator_id = ?")
	if err != nil {
		fmt.Println(err.Error())
		return organisation
	}
	defer statement.Close()

	err = statement.QueryRow(userId).Scan(&organisation.ID, &organisation.Name, &organisation.Creator_id)

	if err != nil {
		return Organisation{}
	}

	statement2, err := dbClient.Prepare("SELECT SUM(size) FROM file WHERE org_id = ?")
	if err != nil {
		return organisation
	}

	defer statement2.Close()

	err = statement2.QueryRow(organisation.ID).Scan(&organisation.Storage_used)

	if err != nil {
		return organisation
	}

	return organisation
}

func InviteUserToOrg(username string, ownerId string) error {

	statement, err := dbClient.Prepare(`
		INSERT INTO org_invites (org_id, user_id) 
		SELECT organisation.id, user.id 
		FROM organisation, user  
		WHERE organisation.creator_id = ? AND user.username = ?
	`)

	if err != nil {
		return err
	}

	defer statement.Close()

	result, err := statement.Exec(ownerId, username)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("Operation failed. Please try again later.")
	}

	return nil
}
