.PHONY: help cli tui web pdf docker release test

BIN := bin/applydiff
GO := go

help:
	@echo "Targets: cli, tui, web, pdf, docker, release, test"

cli:
	$(GO) build -o $(BIN) ./cmd/applydiff

tui: cli

web:
	cd web && npm ci && npm run generate
	rm -rf internal/webembed/public
	cp -a web/.output/public internal/webembed/public

pdf: cli
	@echo "Usage: ./$(BIN) build <job-key>  (requires xelatex or Docker)"

docker:
	docker build -t applydiff:latest .

release:
	./scripts/release.sh

test:
	$(GO) test ./...
