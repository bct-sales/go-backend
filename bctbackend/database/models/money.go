package models

import (
	"strconv"
)

type MoneyInCents int64

func (m MoneyInCents) String() string {
	return strconv.FormatInt(int64(m), 10)
}

func (m MoneyInCents) DecimalNotation() string {
	return strconv.FormatFloat(float64(m)/100.0, 'f', 2, 64)
}

func IsValidPrice(price MoneyInCents) bool {
	return price > 0
}
