package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	SessionIdByteLength      = 16
	SessionCookieName        = "bct_session_id"
	Hour                     = 60 * 60
	SessionDurationInSeconds = 24 * Hour // TODO
)

func HashPassword(password string, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	hash.Write([]byte(salt))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func GenerateUniqueSessionId() string {
	bytes := make([]byte, SessionIdByteLength)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}

	// Note: base64 leads to trouble
	return hex.EncodeToString(bytes)
}
