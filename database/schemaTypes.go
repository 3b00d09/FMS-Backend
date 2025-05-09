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

type FolderData struct {
	Id             *int64 `json:"id" db:"id"`
	OrgId          int64  `json:"org_id" db:"org_id"`
	Uploader       string `json:"uploader" db:"uploader"`
	Name           string `json:"name" db:"name"`
	ParentFolderId *int64 `json:"parent_folder_id" db:"parent_folder_id"`
	CreatedAt      string `json:"created_at" db:"created_at"`
}

type FileData struct {
	Id             *int64 `json:"id" db:"id"`
	OrgId          int64  `json:"org_id" db:"org_id"`
	Uploader       string `json:"uploader" db:"uploader_id"`
	Name           string `json:"name" db:"name"`
	ParentFolderId *int64 `json:"parent_folder_id" db:"parent_folder_id"`
	CreatedAt      string `json:"created_at" db:"created_at"`
	Type           string `json:"type" db:"type"`
	Size           int64  `json:"size" db:"size"`
}
