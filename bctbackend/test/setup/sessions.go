package setup

import (
	models "bctbackend/database/models"
	queries "bctbackend/database/queries"
	"database/sql"
)

func Session(db *sql.DB, userId models.Id) string {
	return AddSessionToDatabaseWithExpiration(db, userId, 3600)
}

func AddSessionToDatabaseWithExpiration(db *sql.DB, userId models.Id, secondsBeforeExpiration int64) string {
	expirationTime := models.Now() + secondsBeforeExpiration
	sessionId, err := queries.AddSession(db, userId, expirationTime)

	if err != nil {
		panic(err)
	}

	return sessionId
}
