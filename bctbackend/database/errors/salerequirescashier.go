package errors

type SaleRequiresCashierError struct{}

func (e *SaleRequiresCashierError) Error() string {
	return "sale requires a cashier"
}
