package errors

import "fmt"

type AuthenticationError struct {
	Reason error
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication error: %v", e.Reason)
}

func (e *AuthenticationError) Unwrap() error {
	return e.Reason
}
