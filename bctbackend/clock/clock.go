package clock

import "bctbackend/database/models"

type Clock interface {
	Now() models.Timestamp
	NewTicker(duration int, callback func()) Ticker
	StopAllTickers()
}

type Ticker interface {
	Stop()
}
