.PHONY: build build-pi build-admin build-admin-pi test cover cover-html clean help

## build: compile binary for current platform
build:
	go build -o kurokku ./cmd/kurokku

## build-pi: cross-compile for Raspberry Pi (linux/arm64)
build-pi:
	GOOS=linux GOARCH=arm64 go build -o kurokku-pi ./cmd/kurokku

## build-admin: compile admin binary for current platform
build-admin:
	go build -o kurokku-admin ./cmd/kurokku-admin

## build-admin-pi: cross-compile admin for Raspberry Pi (linux/arm64)
build-admin-pi:
	GOOS=linux GOARCH=arm64 go build -o kurokku-admin-pi ./cmd/kurokku-admin

## test: run all unit tests
test:
	go test ./...

## cover: run tests with coverage summary
cover:
	go test -cover ./...

## cover-html: run tests and open HTML coverage report
cover-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## clean: remove build artifacts and coverage output
clean:
	rm -f kurokku kurokku-pi kurokku-admin kurokku-admin-pi coverage.out

## help: show this help message
help:
	@echo "Usage: make <target>"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'

.DEFAULT_GOAL := help
