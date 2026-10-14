.PHONY: help dev test scan build

help:
	@echo "Targets:"
	@echo "  dev   - run API server (go run ./cmd/applydiff)"
	@echo "  test  - go test ./..."
	@echo "  scan  - go run ./cmd/scan"
	@echo "  build - go build ./..."

dev:
	go run ./cmd/applydiff

test:
	go test ./...

scan:
	go run ./cmd/scan

build:
	go build ./...
