package clock

import "bctbackend/database/models"

type Clock interface {
	Now() models.Timestamp
	NewTicker(duration int, callback func()) Ticker
}

type Ticker interface {
	Stop()
}
