.PHONY: test

cleantest:
	go clean -testcache
	go test -tags=test ./...

test:
	go test -tags=test ./...

coverage:
	go test -tags=test -coverprofile=coverage-data ./...
	go tool cover -html=coverage-data

swagger:
	swag init -g rest/main.go

rest: build swagger
	go run . server

build:
	go build
