package security

import (
	"bctbackend/database/models"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	SessionIDByteLength      = 16
	SessionCookieName        = "bct_session_id"
	Second                   = 1
	Minute                   = 60 * Second
	Hour                     = 60 * Minute
	SessionDurationInSeconds = 24 * Hour
)

func HashPassword(password string, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	hash.Write([]byte(salt))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func GenerateUniqueSessionID() models.SessionID {
	bytes := make([]byte, SessionIDByteLength)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}

	// Note: base64 leads to trouble
	return models.SessionID(hex.EncodeToString(bytes))
}
