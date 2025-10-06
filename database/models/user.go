package models

type User struct {
	UserId       ID
	RoleId       RoleID
	CreatedAt    Timestamp
	LastActivity *Timestamp
	Password     string
}
