.PHONY: test

cleantest:
	go clean -testcache
	go test -tags=test ./...

test:
	BCT_FONT_DIR=`realpath ..` BCT_FONT_FILE=arial.ttf BCT_FONT_FAMILY=Arial go test -tags=test ./...

coverage:
	go test -tags=test -coverprofile=coverage-data ./...
	go tool cover -html=coverage-data

swagger:
	swag init -g server/server.go

rest: swagger build
	go run . -v server --debug

build:
	go build

lint:
	golangci-lint run