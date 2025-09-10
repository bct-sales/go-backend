package clock

import "bctbackend/database/models"

type ManualClock struct {
	currentTime models.Timestamp
	tickers     []*ManualTicker
}

func NewManualClock(initialTime models.Timestamp) *ManualClock {
	return &ManualClock{currentTime: initialTime, tickers: nil}
}

func (c *ManualClock) NewTicker(duration int, callback func()) Ticker {
func (c *ManualClock) NewTicker(durationInSeconds int, callback func()) Ticker {
	ticker := &ManualTicker{
		nextTick: c.currentTime + models.Timestamp(durationInSeconds),
		interval: models.Timestamp(durationInSeconds),
		active:   true,
		callback: callback,
	}

	c.tickers = append(c.tickers, ticker)
	return ticker
}

func (c *ManualClock) Now() models.Timestamp {
	return c.currentTime
}

type ManualTicker struct {
	nextTick models.Timestamp
	interval models.Timestamp
	active   bool
	callback func()
}

func (t *ManualTicker) Stop() {
	t.active = false
}

func (c *ManualClock) Advance(duration models.Timestamp) {
	c.PruneInactiveTickers()

	target := c.currentTime + duration

	for {
		ticker := c.findEarliestTicker()
		if ticker == nil || ticker.nextTick > target {
			break
		}
		c.currentTime = ticker.nextTick
		ticker.callback()
		ticker.nextTick += ticker.interval
		c.PruneInactiveTickers()
	}

	c.currentTime = target
}

func (c *ManualClock) PruneInactiveTickers() {
	activeTimers := []*ManualTicker{}

	for _, ticker := range c.tickers {
		if ticker.active {
			activeTimers = append(activeTimers, ticker)
		}
	}

	c.tickers = activeTimers
}

func (c *ManualClock) findEarliestTicker() *ManualTicker {
	var earliestTicker *ManualTicker = nil

	for _, ticker := range c.tickers {
		if ticker.active && (earliestTicker == nil || ticker.nextTick < earliestTicker.nextTick) {
			earliestTicker = ticker
		}
	}

	return earliestTicker
}

func (c *ManualClock) StopAllTickers() {
	c.tickers = nil
}
