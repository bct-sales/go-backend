package models

type User struct {
	UserID       ID
	RoleID       RoleID
	CreatedAt    Timestamp
	LastActivity *Timestamp
	Password     string
}
