//go:build test

package helpers

import (
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"context"
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

func AddSessionToDatabase(ctx context.Context, db *sql.DB, userID models.ID, currentTime models.Timestamp, options ...func(*AddSessionData)) models.SessionID {
	data := &AddSessionData{
		secondsBeforeExpiration: 3600,
	}

	for _, option := range options {
		option(data)
	}

	expirationTime := currentTime + models.Timestamp(data.secondsBeforeExpiration)
	sessionID, err := queries.AddSession(ctx, db, userID, expirationTime)

	if err != nil {
		panic(err)
	}

	return sessionID
}
