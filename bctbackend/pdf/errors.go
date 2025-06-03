package pdf

type PdfError struct {
	Message string
	Wrapped error
}

func (e *PdfError) Error() string {
	if e.Wrapped != nil {
		return e.Message + ": " + e.Wrapped.Error()
	}
	return e.Message
}

func (e *PdfError) Unwrap() error {
	return e.Wrapped
}
