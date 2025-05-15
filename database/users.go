package database

import "fmt"

func SearchUsers(username string, userId string) ([]string, error) {
	var users []string
	statement, err := dbClient.Prepare("SELECT username FROM user WHERE username LIKE ? COLLATE NOCASE AND id != ?")
	if err != nil {
		return users, err
	}

	defer statement.Close()

	queryString := fmt.Sprint(username, "%")
	rows, err := statement.Query(queryString, userId)

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
