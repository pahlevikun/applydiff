# ApplyDiff desktop (Tauri v2)

Scaffold for the desktop shell described in the standalone distribution plan.

## Layout

- `src-tauri/` — Tauri v2 app (webview + menu stubs)
- Go binary sidecar: `applydiff serve --desktop --port 8787` prints `LISTENING http://127.0.0.1:8787`
- Tectonic sidecar: second `externalBin` for PDF export (not vendored in this scaffold)

## Local dev

```bash
# Terminal 1
go run ./cmd/applydiff serve --desktop --port 8787

# Terminal 2 — after `npm create tauri-app` or copying src-tauri template
cd desktop/src-tauri && cargo tauri dev
```

## CI / signing

Production bundles use [tauri-action](https://github.com/tauri-apps/tauri-action) with secrets:

- Apple: `APPLE_CERTIFICATE`, `APPLE_CERTIFICATE_PASSWORD`, `APPLE_SIGNING_IDENTITY`, notarization API key
- Windows: `TAURI_SIGNING_PRIVATE_KEY`
- Updater: `TAURI_PRIVATE_KEY`, `TAURI_KEY_PASSWORD`

This repo only documents the pipeline; secrets are not configured here.

## Menu stubs

- **Install CLI on PATH** — symlink or copy sidecar `applydiff` into `/usr/local/bin` (platform-specific; TODO in Tauri Rust menu handler)
- **Connect Cursor / Claude** — runs `applydiff mcp install cursor|claude`
