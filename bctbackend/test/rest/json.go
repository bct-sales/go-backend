package rest

import "encoding/json"

func ToJson(x any) string {
	jsonData, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	return string(jsonData)
}

func FromJson[T any](jsonString string) *T {
	var x T
	err := json.Unmarshal([]byte(jsonString), &x)
	if err != nil {
		panic(err)
	}
	return &x
}
