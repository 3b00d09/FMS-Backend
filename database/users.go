package database

import (
	"fmt"
	"time"
)

func AuthenticateCookie(cookie string) (*UserWithSession, error) {
	if len(cookie) == 0 {
		return nil, fmt.Errorf("cookie value is missing")
	}

	userWithSession := GetUserWithSession(cookie)

	// if the token exists but the value is invalid we won't get a user
	if userWithSession.User.ID == "" {
		return nil, fmt.Errorf("cookie value is invalid")
	}

	// validate session lifetime
	if userWithSession.Session.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("cookie expired")
	}

	return &userWithSession, nil
}

func SearchUsers(username string, userId string) ([]string, error) {
	var users []string
	// this function searches for users who are not equal to the user who is searching
	// and are not members of the user who is searching's organisation
	// and have not been invited already by the organisation that the user created
	statement, err := dbClient.Prepare(`
		SELECT u.username 
		FROM user u
		WHERE u.username LIKE ? COLLATE NOCASE 
		AND u.id != ?
		AND u.id NOT IN (
			SELECT om.user_id 
			FROM org_members om
			JOIN organisation o ON om.org_id = o.id
			WHERE o.creator_id = ?
		)
		AND u.id NOT IN (
			SELECT oi.user_id
			FROM org_invites oi
			JOIN organisation o ON oi.org_id = o.id
			WHERE o.creator_id = ?
		)
	`)
	if err != nil {
		return users, err
	}

	defer statement.Close()

	queryString := fmt.Sprint(username, "%")
	rows, err := statement.Query(queryString, userId, userId, userId)

	if err != nil {
		return users, err
	}

	for rows.Next() {
		var tmpUsername string
		err := rows.Scan(&tmpUsername)
		if err != nil {
			continue
		}
		users = append(users, tmpUsername)
	}

	return users, nil
}

func GetUserInvites(userId string) ([]OrgInvite, error) {
	statement, err := dbClient.Prepare(`
		SELECT 
		o.name as org_name,
		u.username as creator_username,
		i.invited_at,
		i.id,
		o.id as org_id
		FROM organisation o
		JOIN user u ON o.creator_id = u.id
		JOIN org_invites i ON o.id = i.org_id
		WHERE i.user_id = ? AND i.status = 'pending';

	`)
	if err != nil {
		return []OrgInvite{}, err
	}

	defer statement.Close()

	rows, err := statement.Query(userId)

	if err != nil {
		return []OrgInvite{}, err
	}

	var invites []OrgInvite

	for rows.Next() {
		var invite OrgInvite
		err := rows.Scan(&invite.OrgName, &invite.OrgOwner, &invite.InvitedAt, &invite.Id, &invite.OrgId)
		if err != nil {
			fmt.Print(err.Error())
			continue
		}

		invites = append(invites, invite)
	}

	return invites, nil
}

func AcceptOrgInvite(userId string, orgId string) error {
	statement, err := dbClient.Prepare("UPDATE org_invites SET status = 'accepted' WHERE org_id = ? AND user_id = ?")
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

	err = AddMemberToOrg(userId, orgId)

	if err != nil {
		return err
	}

	return nil
}

func DeclineOrgInvite(userId string, orgId string) error {
	statement, err := dbClient.Prepare("UPDATE org_invites SET status = 'declined' WHERE org_id = ? AND user_id = ?")
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
