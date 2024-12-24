package models

type User struct {
	UserId    Id
	RoleId    Id
	CreatedAt Timestamp
	Password  string
}

func NewUser(
	userId Id,
	roleId Id,
	createdAt Timestamp,
	password string) *User {

	return &User{
		UserId:    userId,
		RoleId:    roleId,
		CreatedAt: createdAt,
		Password:  password,
	}
}
