package queries

import (
	"context"
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"bctbackend/database/queries/internal/isql"
	"bctbackend/security"
	"database/sql"
	"errors"
	"fmt"
)

func AddSession(
	ctx context.Context,
	db DatabaseQuerier,
	userID models.ID,
	expirationTime models.Timestamp) (r_result models.SessionID, r_err error) {

	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExists(ctx, db, userID); err != nil {
		return "", fmt.Errorf("failed to add session: %w", err)
	}

	sessionID := security.GenerateUniqueSessionID()

	err := isql.New(db).AddSession(ctx, isql.AddSessionParams{
		SessionID:      sessionID,
		UserID:         userID,
		ExpirationTime: int64(expirationTime),
	})

	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func GetSessionByID(ctx context.Context, db DatabaseQuerier, sessionID models.SessionID) (r_result *models.Session, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row, err := isql.New(db).GetSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to get session with id %s: %w", sessionID, dberr.ErrNoSuchSession)
		}
		return nil, err
	}

	session := models.Session{
		SessionID:      sessionID,
		UserID:         row.UserID,
		ExpirationTime: models.Timestamp(row.ExpirationTime),
	}
	return &session, nil
}

type SessionData struct {
	UserID         models.ID
	RoleID         models.RoleID
	ExpirationTime models.Timestamp
}

// GetSessionData returns information about the given session.
// The function only returns valid session data if the session has not expired.
// ErrNoSuchSession is returned if no unexpired session is found with the given sessionID.
func GetSessionData(ctx context.Context, db DatabaseQuerier, sessionID models.SessionID, currentTime models.Timestamp) (r_result *SessionData, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row, err := isql.New(db).GetSessionData(ctx, isql.GetSessionDataParams{
		SessionID:      sessionID,
		ExpirationTime: int64(currentTime),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dberr.ErrNoSuchSession
		}
		return nil, err
	}

	sessionData := SessionData{
		UserID:         row.UserID,
		RoleID:         models.RoleID{ID: models.ID(row.RoleID)},
		ExpirationTime: models.Timestamp(row.ExpirationTime),
	}
	return &sessionData, nil
}

func GetSessions(ctx context.Context, db DatabaseQuerier) (r_result []models.Session, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := isql.New(db).GetSessions(ctx)
	if err != nil {
		return nil, err
	}

	var sessions []models.Session

	for i := range rows {
		session := models.Session{
			SessionID:      rows[i].SessionID,
			UserID:         rows[i].UserID,
			ExpirationTime: models.Timestamp(rows[i].ExpirationTime),
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func DeleteSession(ctx context.Context, db DatabaseQuerier, sessionID models.SessionID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rowsAffected, err := isql.New(db).DeleteSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if rowsAffected == 0 {
		return dberr.ErrNoSuchSession
	}

	return nil
}

func DeleteSessionWithUser(ctx context.Context, db DatabaseQuerier, userID models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	err := isql.New(db).DeleteSessionWithUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func DeleteExpiredSessions(ctx context.Context, db DatabaseQuerier, cutOff models.Timestamp) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	err := isql.New(db).DeleteExpiredSessions(ctx, int64(cutOff))
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

func DeleteAllSessions(ctx context.Context, db DatabaseQuerier) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	err := isql.New(db).DeleteAllSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	return nil
}

// NOTE: GetTables is kept as raw SQL; the sqlite_schema table cannot be
// declared in sqlc/schemas, so sqlc cannot generate this query.
func GetTables(ctx context.Context, db DatabaseQuerier) (r_result []string, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := db.QueryContext(ctx,
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
