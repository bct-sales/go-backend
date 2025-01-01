package errors

type SaleMissingItemsError struct{}

func (e *SaleMissingItemsError) Error() string {
	return "sale must have at least one item"
}
