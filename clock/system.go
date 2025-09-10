package clock

import (
	"bctbackend/database/models"
	"time"
)

// A SystemClock gets its time from the OS.
type SystemClock struct {
	tickers []*SystemTicker
}

// Now returns the current system time.
func (c *SystemClock) Now() models.Timestamp {
	return models.Now()
}

// NewSystemClock creates a new system clock.
func NewSystemClock() *SystemClock {
	return &SystemClock{tickers: nil}
}

type SystemTicker struct {
	stopChannel chan int // channel to signal the ticker to stop
}

// NewTicker creates a new ticker that will call the given callback function at regular times.
func (clock *SystemClock) NewTicker(durationInSeconds int, callback func()) Ticker {
	ticker := time.NewTicker(time.Duration(durationInSeconds) * time.Second)
	stopChannel := make(chan int)

	go func() {
		for {
			select {
			case <-ticker.C:
				callback()
			case <-stopChannel:
				ticker.Stop()
				return
			}
		}
	}()

	systemTicker := SystemTicker{stopChannel: stopChannel}

	return &systemTicker
}

// Stop deactivates the ticker.
// If the callback is running at the moment, it will finish its work.
func (ticker *SystemTicker) Stop() {
	if ticker.stopChannel != nil {
		ticker.stopChannel <- 0
		ticker.stopChannel = nil // Prevent further use
	}
}

// StopAllTickers deactivates all times associated with the clock.
// All callbacks will be allowed to finish their work.
func (clock *SystemClock) StopAllTickers() {
	for _, ticker := range clock.tickers {
		ticker.Stop()
	}
	clock.tickers = nil // Clear the list of tickers
}
