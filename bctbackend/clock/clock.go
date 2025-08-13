package clock

import "bctbackend/database/models"

type Clock interface {
	Now() models.Timestamp
}

type ManualClock struct {
	CurrentTime models.Timestamp
}

func (c *ManualClock) Now() models.Timestamp {
	return c.CurrentTime
}

func NewManualClock(initialTime models.Timestamp) *ManualClock {
	return &ManualClock{CurrentTime: initialTime}
}

type SystemClock struct{}

func (c *SystemClock) Now() models.Timestamp {
	return models.Now()
}

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}
