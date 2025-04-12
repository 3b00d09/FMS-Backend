package database

type User struct {
	ID       string `db:"id"`
	Username string `db:"username"`
}

// validate required tag is important for the library that makes sure the request body has all fields
type UserCredentials struct {
	//ID       string `json:"id" db:"id"`
	Username string `json:"username" db:"username" validate:"required"`
	Password string `json:"password" db:"password" validate:"required"`
	//FirstName string `json:"firstName" db:"first_name" validate:"required"`
	//LastName  string `json:"lastName" db:"last_name" validate:"required"`
	//Email     string `json:"email" db:"email" validate:"required,email"`
}

type UserSession struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	ExpiresAt int64  `db:"active_expires"`
}

type UserWithSession struct {
	User    User
	Session UserSession
}

type Organisation struct {
	ID         string `db:"id"`
	Name       string `db:"name"`
	Creator_id string `db:"creator_id"`
}