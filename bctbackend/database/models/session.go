package models

type SessionId = string

type Session struct {
	SessionID      SessionId
	UserId         Id
	ExpirationTime Timestamp
}
