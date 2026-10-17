# ApplyDiff — Interview Prep (dev + bisnis)

Status: plan · Owner: F Y Pahlevi · Date: 2026-10-17

## Why

Semua repo interview prep itu **library statis**. Bagus buat belajar, tapi
gak nyambung ke lamaran yang lagi jalan.

ApplyDiff udah punya JD, company, dan structured CV elu. Jadi prep-nya bisa
**digenerate dari lamaran nyata**:

- "Lo apply Senior Backend di RevenueCat. Loop-nya kemungkinan: recruiter →
  coding → system design → behavioral. Topik yang paling mungkin: billing,
  idempotency, webhooks, rate limiting."
- "Cerita STAR yang cocok buat 'conflict with a peer': ambil dari project X di
  CV elu. Jangan ngarang."

Itu pembeda yang gak ada di 4 repo yang kita pelajari. Mereka jual **konten**.
Kita jual **konten yang relevan sama lamaran elu**, lokal, tanpa langganan.

## Yang kita pelajari

| Repo | Yang diadopsi | Yang gak diambil |
|---|---|---|
| `umitkacar/awesome-interview` + `awesome-faang` | Rubrik 5 dimensi per tipe interview, peta "apa yang sebenarnya diuji", daftar failure mode, STAR compression (Action 70%) | Link list eksternal |
| `AkashSingh3031/The-Complete-FAANG-Preparation` | Taksonomi topik (DSA+CP, technical subjects, LLD, HLD, projects, behavioral) | Isi materi (berat, gak terstruktur) |
| `swadhinbiswas/FAANG-Playbook` | Template system design 9 bagian, rubric mock 16K, study plan per level, company notes | Website + LeetCode DB (bukan ranah kita) |
| `devwithsal.com` (publik) | Struktur 6 langkah ber-timebox (0–5, 5–10, 10–15, 15–25, 25–40, 40–45) | Materi berbayar (jangan disalin) |

Aturan: **adopsi struktur, tulis konten sendiri.** Jangan copy teks berbayar.

## Framework yang di-ship sebagai data

Bukan prosa. Tiap framework = langkah + timebox + rubrik + pertanyaan probe,
disimpan di DB, dipakai TUI/Web/CLI.

1. **System design** — 6 langkah: requirements → sizing → arsitektur → data
   model → deep dive 1–3 → failure/observability → tradeoff recap.
2. **Coding** — clarify → contoh → brute force → optimasi → kode → tes →
   kompleksitas.
3. **Behavioral** — STAR dengan Action 70%, wajib 2 cerita per kategori
   (failure, conflict, disagree-and-commit, ambiguity, deadline, influence).
4. **Business case** (non-dev) — clarify objective → issue tree → analisis
   angka → rekomendasi → risiko.
5. **Product case** — problem → user → metrik → opsi → prioritas (RICE) →
   tradeoff.

## Objek inti (DB-first)

Semua di SQLite. Gak ada folder per-lamaran (konsisten sama keputusan
DB-first).

- `interview_kit` — per application: loop stages, topik, prompt, pertanyaan
  buat interviewer
- `story_bank` — STAR story, kategori, bukti (link ke fakta di CV)
- `mock_session` — tipe, prompt, timebox, catatan, skor rubrik
- `rubric_score` — 5 dimensi × 1–5 per sesi
- `topic_note` — 13 topik teknis + framework bisnis
- `drill` — latihan pendek 15 menit per topik

## Fitur

1. **Generate kit dari JD** — loop yang mungkin, topik, prompt system design,
   pertanyaan behavioral yang dipetakan ke values company, pertanyaan buat
   mereka.
2. **Mock mode** — timer per langkah, catatan, rubrik 5 dimensi, skor 1–5.
3. **Story bank grounded** — cerita diambil dari fakta CV. Gate fakta yang udah
   ada dipakai ulang: angka yang gak ada di facts file **gak boleh masuk**.
4. **Topic drills** — 13 topik (load balancer, caching, DB, queue, CDN, rate
   limiting, consistent hashing, sharding, dll) + 15 menit per sesi.
5. **Business / product case** — framework di atas + studi kasus (pricing,
   churn, marketplace, unit economics).
6. **Progress + weak spot** — skor per dimensi, topik yang paling sering
   jelek.
7. **Export prep sheet** — satu artefak (PDF/Markdown) per lamaran, isinya kit
   + cerita + catatan. Ini satu-satunya output ke disk.

## CLI

```bash
applydiff interview kit <job-key>              # generate + simpan kit
applydiff interview kit <job-key> --json
applydiff interview mock <job-key> --type system-design|behavioral|coding|case|product
applydiff interview drill <topic> --minutes 15
applydiff interview story add|list <job-key>
applydiff interview score <session-id> --communication 4 --problem-solving 3 ...
applydiff interview export <job-key> --pdf     # tulis ke output/
```

## TUI

Tab **Interview**: kit (loop + topik), mock timer (timebox per langkah),
rubrik scoring pakai `j/k` + `1-5`, story bank browse. Warna Phosphor.

## Web

`/interview/[job-key]`: kit, mock runner, rubrik, story bank. Keyboard sama
seperti layar lain.

## Fase + checklist

### Fase A — Data model + konten framework
- [ ] Tabel: interview_kit, story_bank, mock_session, rubric_score, topic_note, drill
- [ ] Seed 5 framework (system design, coding, behavioral, business case, product case)
- [ ] Seed 13 topic_note + 6 story kategori
- [ ] Rubrik 5 dimensi × 3 tipe (coding, system design, behavioral) + 2 tipe bisnis

### Fase B — Kit generator
- [ ] Ambil JD + company + role dari application
- [ ] Prompt LLM: loop stages, topik, prompt, pertanyaan buat interviewer
- [ ] Semua output lewat approval diff (sama seperti CV)
- [ ] Simpan ke interview_kit

### Fase C — Mock engine
- [ ] Timer per langkah (timebox)
- [ ] Catatan per langkah
- [ ] Rubrik 1–5 per dimensi + komentar
- [ ] Simpan mock_session + rubric_score
- [ ] Resume sesi yang ketinggalan

### Fase D — Story bank grounded
- [ ] Ekstrak kandidat cerita dari facts + CV
- [ ] Gate fakta: angka tanpa sumber = ditolak
- [ ] Map ke 6 kategori, tandai yang kosong
- [ ] Editor STAR (S/T/A/R + learning)

### Fase E — Bisnis & product case
- [ ] Framework business case + product case
- [ ] 6 studi kasus awal (pricing, churn, marketplace, unit economics, growth, ops)
- [ ] Prompt mock + rubrik

### Fase F — Permukaan
- [ ] CLI subcommand `interview`
- [ ] Tab Interview di TUI
- [ ] Halaman `/interview/[job-key]` di Web
- [ ] Export prep sheet ke `output/`

## Non-goals

- Gak nyalin materi berbayar atau repo orang.
- Gak ngarang pengalaman. Cerita wajib grounded.
- Gak auto-join interview atau auto-jawab.
- Gak jadi LeetCode clone.

## Definisi selesai

- Kit bisa digenerate dari JD nyata dan di-approve lewat diff.
- Mock jalan di CLI + TUI + Web dengan skor tersimpan.
- Cerita STAR gak bisa lolos kalau angkanya gak ada di facts.
- Prep sheet keluar sebagai PDF di `output/`.
- `go test ./...` + `npm run build` pass.
