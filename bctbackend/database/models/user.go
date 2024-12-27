package models

type User struct {
	UserId       Id
	RoleId       Id
	CreatedAt    Timestamp
	LastActivity *Timestamp
	Password     string
}

func NewUser(
	userId Id,
	roleId Id,
	createdAt Timestamp,
	lastActivity *Timestamp,
	password string) *User {

	return &User{
		UserId:       userId,
		RoleId:       roleId,
		CreatedAt:    createdAt,
		LastActivity: lastActivity,
		Password:     password,
	}
}
