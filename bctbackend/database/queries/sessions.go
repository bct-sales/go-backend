package queries

import (
	models "bctbackend/database/models"
	"bctbackend/security"
	"database/sql"
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
	sessionId models.SessionId) (*models.Session, error) {

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
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &session, nil
}

type SessionData struct {
	UserId models.Id
	RoleId models.Id
}

func GetSessionData(db *sql.DB, sessionId models.SessionId) (*SessionData, error) {
	row := db.QueryRow(
		`
			SELECT users.user_id, role_id
			FROM sessions INNER JOIN users ON sessions.user_id = users.user_id
			WHERE session_id = ?
		`,
		sessionId,
	)

	var sessionData SessionData
	err := row.Scan(
		&sessionData.UserId,
		&sessionData.RoleId,
	)

	if err == sql.ErrNoRows {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &sessionData, nil
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
