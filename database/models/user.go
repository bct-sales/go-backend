package models

type User struct {
	UserId       ID
	RoleId       RoleId
	CreatedAt    Timestamp
	LastActivity *Timestamp
	Password     string
}
