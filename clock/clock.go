package clock

import "bctbackend/database/models"

// Clock allows to query the current time and to create tickers.
type Clock interface {
	// Now returns the current time
	Now() models.Timestamp

	// NewTicker creates a new ticker.
	// It will cause the given callback to be called at regular intervals.
	NewTicker(durationInSeconds int, callback func()) Ticker

	// StopAllTickers stops all tickers created by this clock.
	StopAllTickers()
}

type Ticker interface {
	// Stops this ticker.
	Stop()
}
