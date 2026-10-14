.PHONY: help dev test scan build

help:
	@echo "Targets: dev, test, scan, build"

dev:
	@echo "Run backend: go run ./cmd/applydiff"
	@echo "Run web: cd web && npm run dev"

test:
	go test ./...

scan:
	go run ./cmd/scan

build:
	go build ./...
