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
	ID           string `json:"id"`
	Name         string `json:"name"`
	Creator_id   string `json:"creatorId"`
	Storage_used int    `json:"storageUsed"`
	MemberCount  int    `json:"memberCount"`
}

type OrganisationMembers struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	JoinedAt string `json:"joinedAt"`
}

type FolderData struct {
	Id             *int64 `json:"id"`
	OrgId          int64  `json:"orgId"`
	Uploader       string `json:"uploader"`
	Name           string `json:"name"`
	ParentFolderId *int64 `json:"parentFolderId" `
	CreatedAt      string `json:"createdAt"`
}

type FileData struct {
	Id             *int64 `json:"id"`
	OrgId          int64  `json:"orgId"`
	Uploader       string `json:"uploader"`
	Name           string `json:"name"`
	ParentFolderId *int64 `json:"parentFolderId"`
	CreatedAt      string `json:"createdAt"`
	Type           string `json:"type"`
	Size           int64  `json:"size"`
}

type OrgInvite struct {
	Id        *int64 `json:"id"`
	OrgId     int64  `json:"orgId"`
	OrgOwner  string `json:"orgOwner"`
	OrgName   string `json:"orgName"`
	InvitedAt string `json:"invitedAt"`
}
