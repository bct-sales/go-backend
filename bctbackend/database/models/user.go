package models

type User struct {
	UserId       Id
	RoleId       RoleId
	CreatedAt    Timestamp
	LastActivity *Timestamp
	Password     string
}
