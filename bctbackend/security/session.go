package security

import (
	"bctbackend/database/models"
	"time"
)

const (
	SessionDurationInSeconds = 60 * 60 * 6
)

type SessionId = string

type Session struct {
	sessionId  SessionId
	userId     models.Id
	roleId     models.Id
	validUntil time.Time
}

var sessionIdTable = make(map[SessionId]*Session)

func registerSessionId(session *Session) {
	sessionIdTable[session.sessionId] = session
}

func CreateSession(userId models.Id, roleId models.Id, durationInSections int) SessionId {
	sessionId := GenerateUniqueSessionId()
	validUntil := time.Now().Add(time.Second * time.Duration(durationInSections))

	session := Session{
		sessionId:  sessionId,
		userId:     userId,
		roleId:     roleId,
		validUntil: validUntil,
	}

	registerSessionId(&session)

	return sessionId
}

func GetSession(sessionId SessionId) *Session {
	return sessionIdTable[sessionId]
}
