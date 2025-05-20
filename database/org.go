package database

import (
	"database/sql"
	"fms/ioOperations"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func CreateOrg(userId string, orgName string) (int64, error) {

	// do a case insensitive lookup for the org name to see if its taken or not (Org and ORG go through the unique constraint)
	statement, err := dbClient.Prepare("SELECT EXISTS(SELECT name FROM organisation WHERE name LIKE ? COLLATE NOCASE)")
	if err != nil {
		return 0, err
	}

	var exists bool

	err = statement.QueryRow(orgName).Scan(&exists)
	if err != nil {
		return 0, err
	}

	if exists {
		return 0, fmt.Errorf("organisation with this name already exists")
	}

	// verify that the user does not already have an org created
	statement, err = dbClient.Prepare("SELECT id FROM organisation WHERE creator_id = ?")
	if err != nil {
		return 0, err
	}

	defer statement.Close()

	var orgId string

	row := statement.QueryRow(userId)

	err = row.Scan(&orgId)

	// sql returns an error if no rows are found
	// server catches this error to prevent an early exit from the function
	if err != nil {
		if err != sql.ErrNoRows {
			return 0, err
		}
	}

	// if an orgid exists, an org exists, therefore the server returns an error
	if len(orgId) != 0 {
		return 0, fmt.Errorf("org limit exceeded. Users are only allowed to create one Org")
	}

	// THIS MUST BE A TRANSACTION SO IF FOLDER CREATION FAILS WE REVERT THE ORG CREATION
	// org does not exist and the server can create one for the user
	tx, err := dbClient.Begin()
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()
	statement, err = tx.Prepare("INSERT INTO organisation (name, creator_id) VALUES (?, ?)")

	if err != nil {
		return 0, err
	}

	result, err := statement.Exec(orgName, userId)

	if err != nil {
		// if org name is taken
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, fmt.Errorf("organization with this name already exists")
		}
		return 0, err
	}

	rowId, err := result.LastInsertId()

	if err != nil {
		return 0, fmt.Errorf("unknown error occured")
	}

	// attempt to create a directory for the org
	err = ioOperations.CreateOrgDir(strconv.FormatInt(rowId, 10))

	if err != nil {
		return 0, nil
	}

	// commit the tx if the directory was created successfully
	// the tx will rollback by itself because we have defer rolleback so if at any time the function returns before we commit, the tx is rolled back
	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return rowId, nil

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

func InviteUserToOrg(username string, ownerId string, orgId string) error {

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

	rowId, err := result.LastInsertId()

	// don't want the upload function to error out if we are unable to send out a notification because the operation itself worked
	if err != nil {
		log.Printf("error: could not read file ID: %v", err.Error())
	}

	// convert the id to a string
	payloadID := strconv.FormatInt(rowId, 10)

	// send notification to all org members + org owner if applicable
	err = SendNotificationToOrgMembers(orgId, ownerId, "invite", "Has been invited to join", payloadID, username)
	if err != nil {
		log.Printf("error: could not send out notification to join org: %v", err.Error())
	}

	return nil

}

func AddMemberToOrg(userId string, orgId string, username string) error {
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
	// get the file id of the inserted row
	rowId, err := result.LastInsertId()

	// don't want the upload function to error out if we are unable to send out a notification because the operation itself worked
	if err != nil {
		log.Printf("error: could not read file ID: %v", err.Error())
	}

	// convert the id to a string
	payloadID := strconv.FormatInt(rowId, 10)

	// send notification to all org members + org owner if applicable
	err = SendNotificationToOrgMembers(orgId, userId, "join org", "Is now a member of", payloadID, username)
	if err != nil {
		log.Printf("error: could not send out notification to join org: %v", err.Error())
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

func ChangeOrgName(orgId string, orgName string, userId string) error {
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

	// send notification to all org members + org owner if applicable
	err = SendNotificationToOrgMembers(orgId, userId, "org name", "Changed Org Name To", orgId, orgName)
	if err != nil {
		log.Printf("error: could not send out notification for org name change: %v", err.Error())
	}

	return nil
}

func ChangeOrgMemberRole(userId string, memberUsername string, newRole string) error {
	statement, err := dbClient.Prepare("UPDATE org_members SET role = ? WHERE user_id = (SELECT id FROM user WHERE username = ?) AND org_id = (SELECT id FROM organisation WHERE creator_id = ?)")

	if err != nil {
		return err
	}

	defer statement.Close()

	result, err := statement.Exec(newRole, memberUsername, userId)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("something went wrong. please try again later")
	}

	return nil
}

func RemoveOrgMember(userId string, memberUsername string) error {
	statement, err := dbClient.Prepare("DELETE FROM org_members WHERE user_id = (SELECT id FROM user WHERE username = ?) AND org_id = (SELECT id FROM organisation WHERE creator_id = ?)")

	if err != nil {
		return err
	}

	defer statement.Close()

	result, err := statement.Exec(memberUsername, userId)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("something went wrong. please try again later")
	}

	return nil
}

func DeleteOrg(orgId string) error {

	statement, err := dbClient.Prepare("DELETE FROM organisation WHERE id = ?")

	if err != nil {
		return err
	}

	defer statement.Close()

	result, err := statement.Exec(orgId)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("organisation not found or already deleted")
	}

	// attempt to remove the org's directory
	err = ioOperations.DeleteOrgDir(orgId)
	if err != nil {
		// don't have to abort the entire function, the database operation can still go through even if folder delete was a fail
		// folder cleanup can happen but database clean up is not ideal as it directly interacts with the user and can be misleading
		log.Printf("ERROR: FAILED TO OS DELETE ORG WITH ID %v %s \n", orgId, err.Error())
	}

	return nil
}
