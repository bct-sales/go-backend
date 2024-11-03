coverage:
	go test -coverprofile=coverage-data ./...
	go tool cover -html=coverage-data

swagger:
	swag init -g rest/main.go

rest: swagger
	go run . server