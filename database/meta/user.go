package meta

var User = UserMetadata{
	Table:        "users",
	UserID:       "user_id",
	RoleID:       "role_id",
	CreatedAt:    "created_at",
	LastActivity: "last_activity",
	Password:     "password",
}

type UserMetadata struct {
	Table        string
	UserID       string
	RoleID       string
	CreatedAt    string
	LastActivity string
	Password     string
}
