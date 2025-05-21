package models

import (
	"strconv"
	"time"
)

type Timestamp int64

func Now() Timestamp {
	return Timestamp(time.Now().Unix())
}

func (ts Timestamp) String() string {
	return strconv.FormatInt(ts.Int64(), 10)
}

func (ts Timestamp) FormattedDateTime() string {
	return time.Unix(ts.Int64(), 0).String()
}

func (ts Timestamp) Int64() int64 {
	return int64(ts)
}
