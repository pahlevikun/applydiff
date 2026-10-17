.PHONY: help cli tui web pdf test build dev

help:
	@echo "Targets: cli, tui, web, pdf, test, build, dev"

cli:
	go build -o bin/applydiff ./cmd/applydiff

tui: cli

web:
	cd web && npm ci && npm run build

pdf: cli
	@echo "Build PDF via: bin/applydiff build <job-key>"

test:
	go test ./...

build: cli web

dev:
	@echo "API: go run ./cmd/applydiff serve --port 8787"
	@echo "Web: cd web && npm run dev"
