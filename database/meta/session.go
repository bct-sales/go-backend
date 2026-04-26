package meta

var Session = SessionMetadata{
	Table:          "sessions",
	SessionID:      "session_id",
	UserID:         "user_id",
	ExpirationTime: "expiration_time",
}

type SessionMetadata struct {
	Table          string
	SessionID      string
	UserID         string
	ExpirationTime string
}
