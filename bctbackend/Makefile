.PHONY: test

cleantest:
	go clean -testcache
	go test -tags=test ./...

test:
	go test -tags=test ./...

coverage:
	go test -coverprofile=coverage-data ./...
	go tool cover -html=coverage-data

swagger:
	swag init -g rest/main.go

rest: swagger
	go run . server