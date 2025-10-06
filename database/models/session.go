package models

type SessionID string

func (sessionID SessionID) String() string {
	return string(sessionID)
}

type Session struct {
	SessionID      SessionID
	UserID         ID
	ExpirationTime Timestamp
}
