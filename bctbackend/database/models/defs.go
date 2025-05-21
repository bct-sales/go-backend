package models

import (
	"strconv"
)

type Id int64
type MoneyInCents int64

func ParseId(string string) (Id, error) {
	id, err := strconv.ParseInt(string, 10, 64)

	if err != nil {
		return 0, err
	}

	return Id(id), nil
}

func (id Id) String() string {
	return strconv.FormatInt(id.Int64(), 10)
}

func (id Id) Int64() int64 {
	return int64(id)
}

func (m MoneyInCents) String() string {
	return strconv.FormatInt(int64(m), 10)
}

func IsValidPrice(price MoneyInCents) bool {
	return price > 0
}
