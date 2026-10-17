# Tectonic PDF spike (2026-10-17)

Disposable validation that Tectonic can replace the Docker/xelatex fallback for ApplyDiff exports.

## Scope

Templates exercised (via `go test ./internal/export -run TestTectonicCompileTemplates` when `tectonic` is on PATH):

- CV (`internal/cv/export` TeX)
- Cover letter and proof page (`internal/export`)
- Interview prep PDF (`internal/interview/prep_pdf.go`) with bundled DejaVu Sans under `internal/export/fonts/`

## Platforms

| OS | Tectonic source | First compile | Offline rebuild |
|----|-----------------|---------------|-----------------|
| macOS (arm64) | Homebrew `tectonic` 0.17 (~20 MB bottle) or `Paths.Bin` | First compile of all four templates ~4.5 min (bundle download); subsequent compiles ~3 s in `go test` | `~/Library/Caches/TectonicProject.Tectonic` (~tens of MB after first run) |
| Linux (amd64) | distro package or release binary | Same | Same |
| Windows | Scoop `tectonic` | Same | `%LOCALAPPDATA%\Tectonic` |

## Size notes

- Tectonic release binary: ~10–15 MB per platform (not vendored in git; install via brew/scoop or future `Paths.Bin` downloader).
- Bundled DejaVuSans.ttf: ~740 KB in repo for interview prep fontspec `Path=`.

## Decision

Spike **passed** for template compilation when Tectonic is available. `export.ResolveEngine` prefers `xelatex` on PATH, then `tectonic` in `Paths.Bin`, then `tectonic` on PATH. Docker fallback removed.

## Follow-ups

- Optional: download verified Tectonic release into `Paths.Bin` on first PDF build (checksum table per OS/arch).
- Desktop app ships Tectonic as a Tauri sidecar (Phase 6).
