package database

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// validate required tag is important for the library that makes sure the request body has all fields
type UserCredentials struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserSession struct {
	ID        string
	UserID    string
	ExpiresAt int64
}

type UserWithSession struct {
	User    User
	Session UserSession
}

type Organisation struct {
	ID         string
	Name       string
	Creator_id string
}

type FolderData struct {
	Id             *int64 `json:"id"`
	OrgId          int64  `json:"org_id"`
	Uploader       string `json:"uploader"`
	Name           string `json:"name"`
	ParentFolderId *int64 `json:"parent_folder_id" `
	CreatedAt      string `json:"created_at"`
}

type FileData struct {
	Id             *int64 `json:"id"`
	OrgId          int64  `json:"org_id"`
	Uploader       string `json:"uploader"`
	Name           string `json:"name"`
	ParentFolderId *int64 `json:"parent_folder_id"`
	CreatedAt      string `json:"created_at"`
	Type           string `json:"type"`
	Size           int64  `json:"size"`
}
