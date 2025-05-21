package models

import (
	"strconv"
	"time"
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

func TimestampToString(timestamp Timestamp) string {
	return time.Unix(timestamp.Int64(), 0).String()
}

func IsValidPrice(price MoneyInCents) bool {
	return price > 0
}
