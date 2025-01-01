package errors

type WrongPasswordError struct{}

func (e *WrongPasswordError) Error() string {
	return "wrong password"
}
