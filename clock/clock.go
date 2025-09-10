package clock

import "bctbackend/database/models"

type Clock interface {
	Now() models.Timestamp
	NewTicker(durationInSeconds int, callback func()) Ticker
	StopAllTickers()
}

type Ticker interface {
	Stop()
}
