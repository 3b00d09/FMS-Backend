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
