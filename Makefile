.PHONY: help cli tui web pdf docker release test build dev dev-web serve plugins

BIN := bin/applydiff
GO := go
PORT ?= 8080

help:
	@echo "Targets: cli, tui, web, pdf, docker, release, release-snapshot, test, build"
	@echo "  serve     — build CLI and run API + embedded UI (PORT=$(PORT))"
	@echo "  dev       — API on :$(PORT) + Nuxt dev (live web, proxies /api)"
	@echo "  dev-web   — Nuxt dev only (run 'make serve' in another terminal)"
	@echo "  plugins"

plugins:
	@chmod +x scripts/generate-plugins.sh integrations/hooks/*.sh
	@./scripts/generate-plugins.sh

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

# CI publishes on git tag via .github/workflows/release.yml + GoReleaser.
release:
	./scripts/release.sh

release-snapshot:
	goreleaser release --snapshot --clean

test:
	$(GO) test ./...

build: cli web

serve: cli
	./$(BIN) serve --port $(PORT)

dev-web:
	cd web && npm run dev

# One terminal: API + hot-reload web (needs: cd web && npm ci once).
dev: cli
	@echo "API  http://127.0.0.1:$(PORT)"
	@echo "Web  Nuxt dev (open the URL it prints; /api → :$(PORT))"
	@trap 'kill 0' INT TERM EXIT; \
	./$(BIN) serve --port $(PORT) & \
	cd web && npm run dev
