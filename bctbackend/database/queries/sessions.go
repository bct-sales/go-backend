package queries

import (
	"bctbackend/database"
	models "bctbackend/database/models"
	"bctbackend/security"
	"database/sql"
	"errors"
	"fmt"
)

func AddSession(
	db *sql.DB,
	userId models.Id,
	expirationTime models.Timestamp) (string, error) {

	if err := EnsureUserExists(db, userId); err != nil {
		return "", fmt.Errorf("failed to add session: %w", err)
	}

	sessionId := security.GenerateUniqueSessionId()

	_, err := db.Exec(
		`
			INSERT INTO sessions (session_id, user_id, expiration_time)
			VALUES (?, ?, ?)
		`,
		sessionId,
		userId,
		expirationTime,
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
			SELECT user_id, expiration_time
			FROM sessions
			WHERE session_id = ?
		`,
		sessionId,
	)

	var userId models.Id
	var expirationTime models.Timestamp
	if err := row.Scan(&userId, &expirationTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to get session with id %s: %w", sessionId, database.ErrNoSuchSession)
		}
		return nil, err
	}

	session := models.Session{
		SessionId:      sessionId,
		UserId:         userId,
		ExpirationTime: expirationTime,
	}
	return &session, nil
}

type SessionData struct {
	UserId models.Id
	RoleId models.Id
}

func GetSessionData(db *sql.DB, sessionId models.SessionId) (*SessionData, error) {
	now := models.Now()
	row := db.QueryRow(
		`
			SELECT users.user_id, role_id
			FROM sessions INNER JOIN users ON sessions.user_id = users.user_id
			WHERE session_id = ? AND ? < expiration_time
		`,
		sessionId,
		now,
	)

	var userId models.Id
	var roleId models.Id
	if err := row.Scan(&userId, &roleId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, database.ErrNoSuchSession
		}
		return nil, err
	}

	sessionData := SessionData{
		UserId: userId,
		RoleId: roleId,
	}
	return &sessionData, nil
}

func GetSessions(db *sql.DB) (r_result []models.Session, r_err error) {
	rows, err := db.Query(
		`
			SELECT session_id, user_id, expiration_time
			FROM sessions
		`,
	)

	if err != nil {
		return nil, err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var sessions []models.Session

	for rows.Next() {
		var session models.Session

		err := rows.Scan(&session.SessionId, &session.UserId, &session.ExpirationTime)

		if err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

func DeleteSession(db *sql.DB, sessionId models.SessionId) error {
	result, err := db.Exec(
		`
			DELETE FROM sessions
			WHERE session_id = ?
		`,
		sessionId,
	)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	if rowsAffected == 0 {
		return database.ErrNoSuchSession
	}

	return nil
}

func DeleteExpiredSessions(db *sql.DB, cutOff models.Timestamp) error {
	_, err := db.Exec(
		`
			DELETE FROM sessions
			WHERE expiration_time < ?
		`,
		cutOff,
	)

	return err
}
