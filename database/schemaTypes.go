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

type JoinedOrganisation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CreatorName string `json:"creatorName"`
	Role        string `json:"role"`
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
	Size           int64  `json:"size"`
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

type Notification struct {
	ID            int    `json:"id"`
	OrgName       string `json:"orgName"`
	ActorUsername string `json:"actorName"`
	Message       string `json:"message"`
	NotifType     string `json:"notifType"`
	Payload_name  string `json:"payloadName"`
	IsRead        bool   `json:"isRead"`
	CreatedAt     string `json:"createdAt"`
}
