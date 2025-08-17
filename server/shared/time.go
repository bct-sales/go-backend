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
	unixTime := time.Unix(unixTimestamp.Int64(), 0)

	return DateTime{
		Year:   unixTime.Year(),
		Month:  int(unixTime.Month()),
		Day:    unixTime.Day(),
		Hour:   unixTime.Hour(),
		Minute: unixTime.Minute(),
		Second: unixTime.Second(),
	}
}
