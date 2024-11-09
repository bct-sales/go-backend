package models

type SessionId = string

type Session struct {
	SessionId SessionId
	UserId    Id
}

func NewSession(sessionId SessionId, userId Id) *Session {
	return &Session{
		SessionId: sessionId,
		UserId:    userId,
	}
}
