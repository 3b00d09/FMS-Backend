package database

import (
	"database/sql"
	"fmt"
	"strings"
)

func CreateOrg(userId string, orgName string) error {

	// first the server must verify that the user does not already have an org created
	statement, err := dbClient.Prepare("SELECT id FROM organisation WHERE creator_id = ?")
	if err != nil {
		return err
	}

	defer statement.Close()

	var orgId string

	row := statement.QueryRow(userId)

	err = row.Scan(&orgId)

	// sql returns an error if no rows are found
	// server catches this error to prevent an early exit from the function
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	// if an orgid exists, an org exists, therefore the server returns an error
	if len(orgId) != 0 {
		return fmt.Errorf("org limit exceeded. Users are only allowed to create one Org")
	}

	// org does not exist and the server can create one for the user
	statement, err = dbClient.Prepare("INSERT INTO organisation (name, creator_id) VALUES (?, ?)")

	if err != nil {
		return err
	}

	_, err = statement.Exec(orgName, userId)

	if err != nil {
		// if org name is taken
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("organization with this name already exists")
		}
		return err
	}

	return nil

}

func GetOrgById(orgId string) *Organisation {
	var organisation Organisation

	statement, err := dbClient.Prepare(`
        SELECT 
            o.id,
            o.name,
            o.creator_id,
            COALESCE(SUM(f.size), 0),
            (SELECT COUNT(*) FROM org_members WHERE org_id = o.id)
        FROM organisation o
        LEFT JOIN file f ON o.id = f.org_id
        WHERE o.id = ?
        GROUP BY o.id, o.name, o.creator_id;
    `)
	if err != nil {
		return nil
	}
	defer statement.Close()

	err = statement.QueryRow(orgId).Scan(
		&organisation.ID,
		&organisation.Name,
		&organisation.Creator_id,
		&organisation.Storage_used,
		&organisation.MemberCount,
	)

	if err != nil {
		return nil
	}

	return &organisation

}

func CanViewOrg(userId string, orgId string) (bool, string, error) {
	statement, err := dbClient.Prepare("SELECT o.creator_id, m.role FROM organisation o LEFT JOIN org_members m ON m.org_id = o.id AND m.user_id = ? WHERE o.id = ?")

	if err != nil {
		return false, "", err
	}

	var memberRole sql.NullString
	var creatorId string

	defer statement.Close()

	err = statement.QueryRow(userId, orgId).Scan(&creatorId, &memberRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil
		}
		return false, "", err
	}

	if creatorId == userId {
		return true, "owner", nil
	}

	if memberRole.Valid {
		return true, memberRole.String, nil
	}

	return false, "", nil

}

func GetUserOrg(userId string) *Organisation {
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
		return nil
	}
	defer statement.Close()

	err = statement.QueryRow(userId).Scan(&organisation.ID, &organisation.Name, &organisation.Creator_id, &organisation.Storage_used, &organisation.MemberCount)

	if err != nil {
		return nil
	}

	return &organisation
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

func GetOrgMembers(orgId string) []*OrganisationMembers {
	statement, err := dbClient.Prepare(`
		SELECT user.username, org_members.role, org_members.joined_at 
		FROM org_members 
		LEFT JOIN user ON user.id = org_members.user_id
		WHERE org_members.org_id = ?
	`)

	if err != nil {
		return nil
	}

	var organisationMembers []*OrganisationMembers

	defer statement.Close()

	rows, err := statement.Query(orgId)

	if err != nil {
		return nil
	}

	for rows.Next() {
		var member OrganisationMembers
		err := rows.Scan(&member.Username, &member.Role, &member.JoinedAt)
		if err != nil {
			continue
		}
		organisationMembers = append(organisationMembers, &member)
	}

	err = rows.Err()

	if err != nil {
		return nil
	}

	return organisationMembers

}

func GetJoinedOrgs(userId string) []*JoinedOrganisation {
	var organisations []*JoinedOrganisation
	statement, err := dbClient.Prepare(`
		SELECT organisation.id, organisation.name, user.username, org_members.role
		FROM organisation
		JOIN org_members ON org_members.org_id = organisation.id
		JOIN user ON user.id = organisation.creator_id
		WHERE org_members.user_id = ?
  	`)

	if err != nil {
		return nil
	}

	defer statement.Close()

	rows, err := statement.Query(userId)

	if err != nil {
		return nil
	}

	for rows.Next() {
		var org JoinedOrganisation
		err := rows.Scan(&org.ID, &org.Name, &org.CreatorName, &org.Role)
		if err != nil {
			continue
		}
		organisations = append(organisations, &org)
	}

	return organisations
}

func ChangeOrgName(orgId string, orgName string) error {
	statement, err := dbClient.Prepare("UPDATE organisation SET name = ? WHERE id = ?")

	if err != nil {
		return err
	}

	defer statement.Close()

	result, err := statement.Exec(orgName, orgId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("unable to update name. please try again later")
	}

	return nil
}
