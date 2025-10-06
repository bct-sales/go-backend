.PHONY: test

cleantest:
	go clean -testcache

test:
	BCT_FONT_DIR=`realpath .` BCT_FONT_FILE=arial.ttf BCT_FONT_FAMILY=Arial go test -tags=test ./...

ttest: cleantest
	BCT_FONT_DIR=`realpath .` BCT_FONT_FILE=arial.ttf BCT_FONT_FAMILY=Arial go test -count 10 -failfast -tags=test ./...

testhv:
	go test -tags=test -run ^TestHeavyLoad ./...

coverage:
	go test -tags=test -coverprofile=coverage-data ./...
	go tool cover -html=coverage-data

rest: build
	go run . -v server --debug

build: swagger
	go build

lint:
	golangci-lint run | tee lint.txt