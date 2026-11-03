package session

import (
	"testing"

	"bctbackend/database/models"
)

func TestParseExpirationRejectsMalformedInput(t *testing.T) {
	for _, value := range []string{"", "1", "h", "-1h", "1w", "hours"} {
		t.Run(value, func(t *testing.T) {
			command := &AddSessionCommand{expiration: value}
			if _, err := command.parseExpiration(); err == nil {
				t.Fatalf("parseExpiration(%q) accepted malformed input", value)
			}
		})
	}
}

func TestDefaultExpirationIsOneHour(t *testing.T) {
	value := NewAddSessionCommand().Flags().Lookup("expiration").DefValue
	before := models.Now()
	command := &AddSessionCommand{expiration: value}
	got, err := command.parseExpiration()
	if err != nil {
		t.Fatalf("default expiration %q: %v", value, err)
	}
	if got < before+3600 || got > models.Now()+3600 {
		t.Fatalf("default expiration = %v, want one hour from now", got)
	}
}
