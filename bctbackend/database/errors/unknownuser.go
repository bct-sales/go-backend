package errors

type UnknownUserError struct{}

func (e *UnknownUserError) Error() string {
	return "unknown user"
}
