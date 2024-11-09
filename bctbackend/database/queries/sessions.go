package queries

import (
	models "bctbackend/database/models"
	"bctbackend/security"
	"database/sql"
	"log"
)

func AddSession(
	db *sql.DB,
	userId models.Id) (string, error) {

	sessionId := security.GenerateUniqueSessionId()

	_, err := db.Exec(
		`
			INSERT INTO sessions (session_id, user_id)
			VALUES (?, ?)
		`,
		sessionId,
		userId,
	)

	if err != nil {
		return "", err
	}

	return sessionId, nil
}

func GetSessionById(
	db *sql.DB,
	sessionId string) (*models.Session, error) {

	row := db.QueryRow(
		`
			SELECT session_id, user_id
			FROM sessions
			WHERE session_id = ?
		`,
		sessionId,
	)

	var session models.Session

	err := row.Scan(
		&session.SessionId,
		&session.UserId,
	)

	if err == sql.ErrNoRows {
		sessions, err := GetSessions(db)

		if err != nil {
			panic("oh no")
		}

		for _, s := range sessions {
			log.Printf("session: %v\n", s)
		}

		log.Printf("Session count: %v\n", len(sessions))

		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func GetSessions(db *sql.DB) ([]models.Session, error) {
	rows, err := db.Query(
		`
			SELECT session_id, user_id
			FROM sessions
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var sessions []models.Session

	for rows.Next() {
		var session models.Session

		err := rows.Scan(&session.SessionId, &session.UserId)

		if err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}
