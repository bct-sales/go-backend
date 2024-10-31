package models

type Id = int64
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
