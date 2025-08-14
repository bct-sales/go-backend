package clock

import (
	"bctbackend/database/models"
	"time"
)

type SystemClock struct {
	tickers []*SystemTicker
}

func (c *SystemClock) Now() models.Timestamp {
	return models.Now()
}

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

type SystemTicker struct {
	stopChannel chan int // Channel to signal the ticker to stop
}

func (clock *SystemClock) NewTicker(duration int, callback func()) Ticker {
	ticker := time.NewTicker(time.Duration(duration) * time.Second)
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

func (ticker *SystemTicker) Stop() {
	if ticker.stopChannel != nil {
		ticker.stopChannel <- 0
		ticker.stopChannel = nil // Prevent further use
	}
}

func (clock *SystemClock) StopAllTickers() {
	for _, ticker := range clock.tickers {
		ticker.Stop()
	}
	clock.tickers = nil // Clear the list of tickers
}
