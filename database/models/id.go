package models

import (
	"strconv"
)

type Id int64

func ParseId(str string) (Id, error) {
	identifier, err := strconv.ParseInt(str, 10, 64)

	if err != nil {
		return 0, err
	}

	return Id(identifier), nil
}

func (id Id) String() string {
	return strconv.FormatInt(id.Int64(), 10)
}

func (id Id) Int64() int64 {
	return int64(id)
}
