package security

import (
	"crypto/sha256"
	"fmt"
)

func HashPassword(password string, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	hash.Write([]byte(salt))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
