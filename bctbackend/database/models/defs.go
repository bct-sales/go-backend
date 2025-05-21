package models

import (
	"strconv"
	"time"
)

type Id int64
type MoneyInCents int64
type Timestamp = int64

func NewTimestamp(timestamp int64) Timestamp {
	return Timestamp(timestamp)
}

func Now() Timestamp {
	return time.Now().Unix()
}

func ParseId(string string) (Id, error) {
	id, err := strconv.ParseInt(string, 10, 64)

	if err != nil {
		return 0, err
	}

	return Id(id), nil
}

func (id Id) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func (id Id) Int64() int64 {
	return int64(id)
}

func TimestampToString(timestamp Timestamp) string {
	return time.Unix(timestamp, 0).String()
}

func MoneyInCentsToString(moneyInCents MoneyInCents) string {
	return strconv.FormatInt(int64(moneyInCents), 10)
}

func IsValidPrice(price MoneyInCents) bool {
	return price > 0
}
