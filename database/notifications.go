package database

import "fmt"

// type is a perserved keyword so its prefixed with an underscore
func SendNotificationToOrgMembers(orgId string, actorId string, _type string, message string, payloadId string, payloadName string) error {

	// a lot going on here
	// with recipients is a temporary table to hold all the users and the creator of an organisation, this table is referenced when inserting notifications for users
	// using a select in an insert will copy over the user ids from the recipients table as well as insert the data passed in as args into this function
	// for each userId in the recipients table, the statement will insert (uid, .... args)
	// see https://www.geeksforgeeks.org/sqlite-insert-into-select/ for insert with select
	// the actor is excluded from this operation as the actor should not recieve a notification
	statement, err := dbClient.Prepare(`
		WITH recipients AS (
		SELECT user_id AS uid
			FROM org_members
		WHERE org_id = ?
		UNION
		SELECT creator_id AS uid
			FROM organisation
		WHERE id = ?
		)

		INSERT INTO notification
		(user_id, org_id, actor_id, type, message, payload_id, payload_name)
		SELECT
		uid,      
		?,        
		?,        
		?,        
		?,
		?,
		?         
		FROM recipients
		WHERE uid != ?
	`)

	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(orgId, orgId, orgId, actorId, _type, message, payloadId, payloadName, actorId)

	if err != nil {
		return err
	}

	return nil
}

func GetUserNotifications(userId string) ([]Notification, error) {
	var notifications []Notification
	statement, err := dbClient.Prepare(`
		SELECT
			n.id,
			u.username AS actor_username,
			o.name AS org_name,
			n.message,
			n.payload_name,
			n.type,
			n.is_read,
			n.created_at
		FROM notification AS n
		JOIN "user" AS u ON u.id = n.actor_id
		JOIN organisation AS o ON o.id = n.org_id
		WHERE n.user_id = ?
		ORDER BY n.created_at DESC;
	`)

	if err != nil {
		fmt.Println(err.Error())
		return notifications, err
	}

	defer statement.Close()

	rows, err := statement.Query(userId)

	if err != nil {
		fmt.Println(err.Error())
		return notifications, err
	}

	for rows.Next() {
		var notif Notification
		err := rows.Scan(&notif.ID, &notif.ActorUsername, &notif.OrgName, &notif.Message, &notif.Payload_name, &notif.NotifType, &notif.IsRead, &notif.CreatedAt)
		if err != nil {
			fmt.Println(err.Error())
		}
		notifications = append(notifications, notif)
	}

	return notifications, nil
}

func MarkAsRead(id string, userId string) error {
	statement, err := dbClient.Prepare("UPDATE notification SET is_read = 1 WHERE id = ? AND user_id = ?")

	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(id, userId)

	if err != nil {
		return err
	}

	return nil
}

func MarkAllAsRead(userId string) error {
	statement, err := dbClient.Prepare("UPDATE notification SET is_read = 1 WHERE user_id = ?")

	if err != nil {
		return err
	}

	defer statement.Close()

	_, err = statement.Exec(userId)

	if err != nil {
		return err
	}

	return nil
}
