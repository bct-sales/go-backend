package queries

import (
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"bctbackend/security"
	"database/sql"
	"errors"
	"fmt"
)

func AddSession(
	db DatabaseQuerier,
	userId models.Id,
	expirationTime models.Timestamp) (r_result models.SessionId, r_err error) {

	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

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

func GetSessionById(db DatabaseQuerier, sessionId models.SessionId) (r_result *models.Session, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row := db.QueryRow(
		`
			SELECT
				user_id,
				expiration_time
			FROM
				sessions
			WHERE
				session_id = ?
		`,
		sessionId,
	)

	var userId models.Id
	var expirationTime models.Timestamp
	if err := row.Scan(&userId, &expirationTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to get session with id %s: %w", sessionId, dberr.ErrNoSuchSession)
		}
		return nil, err
	}

	session := models.Session{
		SessionID:      sessionId,
		UserID:         userId,
		ExpirationTime: expirationTime,
	}
	return &session, nil
}

type SessionData struct {
	UserId         models.Id
	RoleId         models.RoleId
	ExpirationTime models.Timestamp
}

// GetSessionData returns information about the given session.
// The function only returns valid session data if the session has not expired.
// ErrNoSuchSession is returned if no unexpired session is found with the given sessionId.
func GetSessionData(db DatabaseQuerier, sessionId models.SessionId, currentTime models.Timestamp) (r_result *SessionData, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row := db.QueryRow(
		`
			SELECT
				users.user_id, role_id, expiration_time
			FROM
				sessions
			INNER JOIN
				users ON sessions.user_id = users.user_id
			WHERE
				session_id = ? AND ? < expiration_time
		`,
		sessionId,
		currentTime,
	)

	var userId models.Id
	var roleId models.RoleId
	var expirationTime models.Timestamp
	if err := row.Scan(&userId, &roleId.Id, &expirationTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dberr.ErrNoSuchSession
		}
		return nil, err
	}

	sessionData := SessionData{
		UserId:         userId,
		RoleId:         roleId,
		ExpirationTime: expirationTime,
	}
	return &sessionData, nil
}

func GetSessions(db DatabaseQuerier) (r_result []models.Session, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := db.Query(
		`
			SELECT
				session_id,
				user_id,
				expiration_time
			FROM
				sessions
		`,
	)

	if err != nil {
		return nil, err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var sessions []models.Session

	for rows.Next() {
		var sessionId models.SessionId
		var userId models.Id
		var expirationTime models.Timestamp
		if err := rows.Scan(&sessionId, &userId, &expirationTime); err != nil {
			return nil, err
		}

		session := models.Session{
			SessionID:      sessionId,
			UserID:         userId,
			ExpirationTime: expirationTime,
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return sessions, nil
}

func DeleteSession(db DatabaseQuerier, sessionId models.SessionId) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

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
		return dberr.ErrNoSuchSession
	}

	return nil
}

func DeleteSessionWithUser(db DatabaseQuerier, userId models.Id) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	_, err := db.Exec(
		`
			DELETE FROM sessions
			WHERE user_id = ?
		`,
		userId,
	)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func DeleteExpiredSessions(db DatabaseQuerier, cutOff models.Timestamp) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query := `
		DELETE FROM sessions
		WHERE expiration_time < ?
	`
	_, err := db.Exec(query, cutOff)

	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

func DeleteAllSessions(db DatabaseQuerier) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query := `
		DELETE FROM sessions
	`
	_, err := db.Exec(query)

	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	return nil
}

func GetTables(db DatabaseQuerier) (r_result []string, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := db.Query(
		`
			SELECT
				name
			FROM
				sqlite_schema
			WHERE
				type ='table' AND name NOT LIKE 'sqlite_%';
		`,
	)

	if err != nil {
		return nil, err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var tableNames []string

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}

		tableNames = append(tableNames, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return tableNames, nil
}
