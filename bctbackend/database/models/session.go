package models

type SessionId = string

type Session struct {
	SessionID      SessionId
	UserID         Id
	ExpirationTime Timestamp
}
