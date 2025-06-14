package path

import (
	"bctbackend/database/models"
	"fmt"
)

type cashierSalesPath struct{}

func CashierSales() *cashierSalesPath {
	return &cashierSalesPath{}
}

func (path *cashierSalesPath) WithRawCashierId(id string) string {
	return fmt.Sprintf("/api/v1/cashiers/%s/sales", id)
}

func (path *cashierSalesPath) WithSellerId(cashierId models.Id) string {
	return path.WithRawCashierId(cashierId.String())
}
