//go:build test

package helpers

import (
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"database/sql"
)

type AddSessionData struct {
	secondsBeforeExpiration int64
}

func WithExpiration(secondsBeforeExpiration int64) func(*AddSessionData) {
	return func(data *AddSessionData) {
		data.secondsBeforeExpiration = secondsBeforeExpiration
	}
}

func AddSessionToDatabase(db *sql.DB, userId models.Id, options ...func(*AddSessionData)) string {
	data := &AddSessionData{
		secondsBeforeExpiration: 3600,
	}

	for _, option := range options {
		option(data)
	}

	expirationTime := models.Now() + data.secondsBeforeExpiration
	sessionId, err := queries.AddSession(db, userId, expirationTime)

	if err != nil {
		panic(err)
	}

	return sessionId
}
