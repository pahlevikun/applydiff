# ApplyDiff — CLI + TUI + Web (proper UI/UX)

Status: plan · Owner: F Y Pahlevi · Date: 2026-10-17

## Why

Core Go udah jalan (scan, match, cv, tailor, diff, filter, application, store,
export). Sekarang cuma kebuka lewat HTTP API + Nuxt web. Kita mau tiga pintu
masuk ke **satu core yang sama**:

1. **CLI** — scriptable, non-interactive, buat otomasi (cron, pipe).
2. **TUI** — interactive di terminal, keyboard-driven, "alive" seperti
   claude-code / codex / cursor-cli. Ini rumah alami buat review diff per-hunk.
3. **Web** — UI proper buat yang mau lihat penuh.

Satu core, satu SQLite, tiga muka. **Gak ada logika bisnis di UI.**

## Palette — LOCKED: PHOSPHOR

Amber CRT di warm black. Nuansa terminal, bukan hijau klise.

```
--bg      #12100c   --surface  #1a1712   --surface2 #221e17
--line    #2f2a21   --line2    #3d372c
--fg      #f0e6d2   --muted    #a99b83   --dim      #6f6555
--accent  #ffb454   --accent-dim #5c3f12
--add     #9ccb5a   --del      #e5644e
paper #fbfaf7  ink #1c1f25  paper-line #e6e3db
```

Rules (biar gak jadi slop):

- Accent cuma buat **satu** hal per layar (CTA/aktif). Bukan dekorasi.
- `add`/`del` cuma buat diff. Jangan dipakai buat status umum.
- Gak ada gradient. Gak ada glow berlebih. Border 1px, radius 10px / 7px.
- Font: IBM Plex Sans (UI) + IBM Plex Mono (angka, kode, diff).
- Motion minimal: fade 120ms, gak ada bounce.

## Arsitektur

```
             ┌──────────────── internal/ (core, no UI) ────────────────┐
             │ job source scan match cv tailor diff filter extract    │
             │ application store export llm                           │
             └───────▲───────────────────────▲────────────────────────┘
                     │ in-process            │ in-process
        cmd/applydiff (CLI subcommands)   cmd/applydiff tui (Bubble Tea)
                     │
                     │ net/http
                     ▼
              cmd/applydiff serve  ──►  web/ (Nuxt 3)
```

- CLI + TUI: panggil core **langsung** (in-process). Cepat, gak perlu server.
- Web: lewat HTTP API yang sama (`/api/...`).
- Satu sumber kebenaran: semua aksi nulis ke SQLite lewat `internal/store`.
- Audit log: setiap approve/apply tercatat (actor, waktu, hunk).

## Surface 1 — CLI

Binary: `applydiff`. Non-interactive. Semua command punya `--json`.

| Command | Fungsi |
|---|---|
| `applydiff scan [--source S] [--preset P]` | tarik job, dedupe, liveness probe |
| `applydiff jobs [--filter ...] [--json]` | list job + skor filter |
| `applydiff show <job-key>` | detail JD + match reasons |
| `applydiff match <job-key>` | hitung skor match vs CV |
| `applydiff tailor <job-key>` | generate CV variant + cover letter (LLM) |
| `applydiff diff <job-key> [--color=auto]` | tampilin diff per-hunk (unified) |
| `applydiff approve <job-key> --hunk N[,N]` | approve hunk tertentu |
| `applydiff approve <job-key> --all` | approve semua hunk (gate tetap wajib eksplisit) |
| `applydiff reject <job-key> --hunk N` | tolak hunk |
| `applydiff build <job-key>` | compile PDF LaTeX (xelatex) |
| `applydiff apply <job-key> [--mode manual\|assisted\|auto]` | tandai / kirim |
| `applydiff export <file>` / `import <file>` | SQLite portability |
| `applydiff serve [--port 8787]` | start HTTP API + web |
| `applydiff tui` | buka TUI |

Aturan CLI:

- Exit code: `0` ok · `2` gak ada match · `3` butuh approval · `4` error LLM ·
  `5` error build LaTeX. Biar bisa dipakai di script.
- `--json` output stabil buat pipe.
- `--dry-run` di semua command yang nulis.
- Gak ada auto-submit. `apply` default `manual`.

## Surface 2 — TUI

Framework: **Bubble Tea** + Lip Gloss + Bubbles (charmbracelet). Satu binary.

Layar:

1. **Feed** — list job, kolom: judul, company, region, umur posting, skor.
   Filter preset di kiri, vim keys.
2. **Job** — JD lengkap + match reasons (kenapa cocok/gak).
3. **Diff review** ← HERO. Persis `git add -p` tapi buat CV:
   per-hunk, `j`/`k` pindah hunk, `space` toggle, `a` accept, `x` reject,
   `e` edit manual, `b` build PDF setelah semua keputusannya.
4. **Sources** — status tiap board (ok/gagal, jumlah, terakhir).
5. **Applications** — pipeline (draft → ready → applied), status email.

Keymap global: `?` help · `tab`/`shift+tab` ganti pane · `/` filter ·
`g`/`G` atas/bawah · `q` keluar · `ctrl+c` quit paksa.

Feel: warna Phosphor, border tipis, status bar bawah (mode + hint),
spinner buat kerja LLM, gak ada frame yang berkedip.

## Surface 3 — Web

Nuxt 3. Wireframe s01–s06 jadi halaman nyata.

- Terapin token Phosphor ke `web/assets/theme.css` (atau Tailwind config).
- Halaman: Jobs · Match · Diff (approve/reject per-hunk) · My CV · Interview · Sources.
- Nav 5 item konsisten.
- Keyboard: `/` fokus search, `j/k` navigasi list, `space` toggle hunk di Diff.
- Diff view = komponen utama, sama persis logikanya dengan TUI + CLI.
- Gak ada auto-submit. Gak ada loading spinner generik. Skeleton disesuaikan bentuk.

## Data model — DB-first

**SQLite adalah satu-satunya sumber kebenaran.** Semua state hidup di DB:
jobs, applications, CV master + varian per lamaran, hunks + keputusan, status,
catatan, riwayat.

Disk cuma dipakai buat **artefak hasil build**:

```
output/
  2026-10-17-revenuecat-senior-backend.pdf
  2026-10-17-anovium-fullstack.pdf
```

Aturan:

- Gak ada direktori per-lamaran. Gak ada YAML/JSON sebagai sumber state.
- PDF (dan `.tex` sementara) ditulis ke `output/`. Nama file =
  `<tanggal>-<company>-<role>`.
- "Apa yang gue kirim" dibaca dari DB (diff + keputusan hunk), bukan dari folder.
- Pindah mesin = copy file `.db` (atau `applydiff export`). Backup = copy `.db`.
- DB rusak = gak ada recovery dari disk. Jadi `export` wajib gampang.

## Fase + checklist

### Fase A — Token & desain
- [ ] `web/assets/theme.css` pakai token Phosphor
- [ ] Tailwind/setara di-config pakai token itu (bukan warna hardcode)
- [ ] Ikon bespoke (bukan lucide default)
- [ ] Font IBM Plex Sans/Mono self-host di web
- [ ] Wireframe s01–s06 di-regenerate pakai token final

### Fase B — CLI
- [ ] Refactor `cmd/applydiff` ke subcommand (cobra)
- [ ] `scan`, `jobs`, `show`, `match`, `tailor`
- [ ] `diff` (output unified, `--color`)
- [ ] `approve` / `reject` per-hunk + `--all`
- [ ] `build`, `apply`, `serve`, `tui`
- [ ] `export` / `import` (pindah dari flag lama)
- [ ] Exit code + `--json` + `--dry-run`
- [ ] Test: `go test ./cmd/...`

### Fase C — TUI
- [ ] Setup Bubble Tea (model, update, view)
- [ ] Layar Feed + filter preset
- [ ] Layar Job + match reasons
- [ ] Layar Diff review (hero) + keymap
- [ ] Layar Sources + Applications
- [ ] Status bar + help overlay
- [ ] Snapshot test render (tanpa TTY)

### Fase D — Web proper
- [ ] Terapin token + komponen (button, chip, card, list)
- [ ] Halaman Jobs + filter panel
- [ ] Halaman Match
- [ ] Halaman Diff + approve/reject per-hunk
- [ ] Halaman My CV (structured editor)
- [ ] Halaman Interview + Sources
- [ ] Keyboard shortcuts
- [ ] Responsive (min 1280, turun ke 1024)

### Fase E — Packaging & distribusi
- [ ] Embed hasil build Nuxt ke binary (`go:embed`), `applydiff serve` jalan tanpa Node
- [ ] Satu binary: CLI + TUI + serve
- [ ] `applydiff build` tulis PDF ke `output/` (nama `<tanggal>-<company>-<role>`)
- [ ] Docker image (binary + TeX Live) → user cukup punya Docker
- [ ] Fallback: kalau `xelatex` gak ada, pakai container `texlive/texlive`
- [ ] Release binary (linux/darwin/windows) + checksum
- [ ] Makefile: `make cli`, `make tui`, `make web`, `make pdf`, `make docker`
- [ ] README: quickstart 3 muka + jalur zero-install

## Definisi selesai

- Tiga muka jalan dari core yang sama, gak ada logika dobel.
- Diff per-hunk konsisten di CLI, TUI, Web.
- Gak ada auto-submit di default.
- `go build ./...`, `go test ./...`, `npm run build` pass.
- Palette Phosphor kepakai konsisten, gak ada warna lepas.
