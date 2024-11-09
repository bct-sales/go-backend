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
	SessionDurationInSeconds = 60 * 60
)

func HashPassword(password string, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	hash.Write([]byte(salt))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func GenerateUniqueSessionId() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}
