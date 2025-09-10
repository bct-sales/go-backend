package clock

import "bctbackend/database/models"

// A ManualClock allows to have a clock where the current time can be advanced manually.
// Meant for testing and debugging.
type ManualClock struct {
	currentTime models.Timestamp
	tickers     []*ManualTicker
}

// NewManualClock creates a new manual clock which is set to the given initial time.
func NewManualClock(initialTime models.Timestamp) *ManualClock {
	return &ManualClock{currentTime: initialTime, tickers: nil}
}

// NewTicker creates a new ticker associated with the clock.
func (c *ManualClock) NewTicker(duration int, callback func()) Ticker {
	ticker := &ManualTicker{
		nextTick: c.currentTime + models.Timestamp(duration),
		interval: models.Timestamp(duration),
		active:   true,
		callback: callback,
	}

	c.tickers = append(c.tickers, ticker)
	return ticker
}

// Now returns the time the clock is set at.
func (c *ManualClock) Now() models.Timestamp {
	return c.currentTime
}

type ManualTicker struct {
	nextTick models.Timestamp // time left to wait until next tick.
	interval models.Timestamp // interval at which this ticker invokes its callback.
	active   bool             // whether the timer is active or not
	callback func()           // callback to be called at every tick
}

// Stop deactivates the timer.
func (t *ManualTicker) Stop() {
	t.active = false
}

// Advance advances the manual clock's time.
// Tickers will call their associated callbacks as needed, in order.
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

// PruneInactiveTickers removes stopped tickers from the list of tickers.
func (c *ManualClock) PruneInactiveTickers() {
	activeTimers := []*ManualTicker{}

	for _, ticker := range c.tickers {
		if ticker.active {
			activeTimers = append(activeTimers, ticker)
		}
	}

	c.tickers = activeTimers
}

// findEarliestTicker returns the ticker that will be the first to tick if time is advanced.
// Returns nil if there is no such ticker.
func (c *ManualClock) findEarliestTicker() *ManualTicker {
	var earliestTicker *ManualTicker = nil

	for _, ticker := range c.tickers {
		if ticker.active && (earliestTicker == nil || ticker.nextTick < earliestTicker.nextTick) {
			earliestTicker = ticker
		}
	}

	return earliestTicker
}

// StopAllTickers deactivates all tickers associated with the clock.
func (c *ManualClock) StopAllTickers() {
	c.tickers = nil
}
