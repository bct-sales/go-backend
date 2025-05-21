package models

import (
	"strconv"
	"time"
)

type Id int64
type MoneyInCents = int64
type Timestamp = int64

func NewId(id int64) Id {
	return Id(id)
}

func NewMoneyInCents(moneyInCents int64) MoneyInCents {
	return MoneyInCents(moneyInCents)
}

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

	return NewId(id), nil
}

func IdToInt64(id Id) int64 {
	return int64(id)
}

func IdToString(id Id) string {
	return strconv.FormatInt(int64(id), 10)
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
