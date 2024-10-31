package models

type User struct {
	UserId    Id
	RoleId    Id
	Timestamp Timestamp
	Password  string
}

func NewUser(
	userId Id,
	roleId Id,
	timestamp Timestamp,
	password string) *User {

	return &User{
		UserId:    userId,
		RoleId:    roleId,
		Timestamp: timestamp,
		Password:  password,
	}
}
