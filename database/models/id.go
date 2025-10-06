package models

import (
	"strconv"
)

type ID int64

func ParseId(str string) (ID, error) {
	identifier, err := strconv.ParseInt(str, 10, 64)

	if err != nil {
		return 0, err
	}

	return ID(identifier), nil
}

func (id ID) String() string {
	return strconv.FormatInt(id.Int64(), 10)
}

func (id ID) Int64() int64 {
	return int64(id)
}
