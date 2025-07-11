package rest

import (
	"bctbackend/database/models"
	"fmt"
	"time"
)

type DateTime struct {
	Year   int `json:"year"`
	Month  int `json:"month"`
	Day    int `json:"day"`
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
	Second int `json:"second"`
}

func (t DateTime) String() string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", t.Year, t.Month, t.Day, t.Hour, t.Minute, t.Second)
}

func ConvertTimestampToDateTime(unixTimestamp models.Timestamp) DateTime {
	t := time.Unix(unixTimestamp.Int64(), 0)

	return DateTime{
		Year:   t.Year(),
		Month:  int(t.Month()),
		Day:    t.Day(),
		Hour:   t.Hour(),
		Minute: t.Minute(),
		Second: t.Second(),
	}
}
