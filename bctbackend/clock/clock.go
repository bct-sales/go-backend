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

type SystemClock struct{}

func (c *SystemClock) Now() models.Timestamp {
	return models.Now()
}
