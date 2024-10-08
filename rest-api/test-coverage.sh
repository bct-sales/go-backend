#!/usr/bin/env bash

go test -coverprofile=coverage-data ./...
go tool cover -html=coverage-data
