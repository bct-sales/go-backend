package models

type SessionId string

func (sessionId SessionId) String() string {
	return string(sessionId)
}

type Session struct {
	SessionID      SessionId
	UserID         ID
	ExpirationTime Timestamp
}
