package models

type SessionID string

func (sessionId SessionID) String() string {
	return string(sessionId)
}

type Session struct {
	SessionID      SessionID
	UserID         ID
	ExpirationTime Timestamp
}
