package models

type User struct {
	UserId       Id
	RoleId       Id
	CreatedAt    Timestamp
	LastActivity *Timestamp
	Password     string
}
