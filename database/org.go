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

	// using coalesce here on size so if the org is empty size is 0 not null
	statement, err := dbClient.Prepare(`
		SELECT 
		o.id,
		o.name,
		o.creator_id,
		COALESCE(SUM(f.size), 0),
		(SELECT COUNT(*) FROM org_members WHERE org_id = o.id)
		FROM organisation o
		LEFT JOIN file f ON o.id = f.org_id
		WHERE o.creator_id = ?
		GROUP BY o.id, o.name, o.creator_id;
	`)
	if err != nil {
		return organisation
	}
	defer statement.Close()

	err = statement.QueryRow(userId).Scan(&organisation.ID, &organisation.Name, &organisation.Creator_id, &organisation.Storage_used, &organisation.MemberCount)

	if err != nil {
		fmt.Println(err.Error())
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
		return fmt.Errorf("operation failed. Please try again later")
	}

	return nil
}

func AddMemberToOrg(userId string, orgId string) error {
	statement, err := dbClient.Prepare("INSERT INTO org_members (org_id, user_id, role) VALUES (?, ?, 'Editor')")
	if err != nil {
		return err
	}

	defer statement.Close()

	result, err := statement.Exec(orgId, userId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("operation failed. Please try again later")
	}
	return nil
}

func GetOrgMembers(orgId string) ([]OrganisationMembers, error) {
	var organisationMembers []OrganisationMembers
	statement, err := dbClient.Prepare(`
		SELECT user.username, org_members.role, org_members.joined_at 
		FROM org_members 
		LEFT JOIN user ON user.id = org_members.user_id
		WHERE org_members.org_id = ?
	`)

	if err != nil {
		return organisationMembers, err
	}

	defer statement.Close()

	rows, err := statement.Query(orgId)

	if err != nil {
		return organisationMembers, err
	}

	for rows.Next() {
		var member OrganisationMembers
		err := rows.Scan(&member.Username, &member.Role, &member.JoinedAt)
		if err != nil {
			continue
		}
		organisationMembers = append(organisationMembers, member)
	}

	return organisationMembers, nil

}
