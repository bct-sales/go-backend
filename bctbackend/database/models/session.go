package models

type SessionId = string

type Session struct {
	SessionId      SessionId
	UserId         Id
	ExpirationTime Timestamp
}

func NewSession(sessionId SessionId, userId Id, expirationTime Timestamp) *Session {
	return &Session{
		SessionId:      sessionId,
		UserId:         userId,
		ExpirationTime: expirationTime,
	}
}
