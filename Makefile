.PHONY: test

cleantest:
	go clean -testcache

test:
	go test -tags=test ./...

ttest: cleantest
	go test -count 10 -failfast -tags=test ./...

testhv:
	go test -tags=test -run ^TestHeavyLoad ./...

coverage:
	go test -tags=test -coverprofile=coverage-data ./...
	go tool cover -html=coverage-data

rest: build
	go run . -v server --debug

build:
	go build

lint:
	golangci-lint run | tee lint.txt