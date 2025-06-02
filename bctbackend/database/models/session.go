package models

type SessionId = string

type Session struct {
	SessionId      SessionId
	UserId         Id
	ExpirationTime Timestamp
}
