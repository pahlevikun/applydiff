# ApplyDiff — Scam guard + channel (plan)

Status: plan · Owner: F Y Pahlevi · Date: 2026-10-17

## Why

Dari thread @arifin_eff (Threads) dan reply-nya, dua sinyal nyata:

1. **Scam di lamaran remote itu umum.** Warning yang muncul: verifikasi company,
   jangan mau disuruh clone repo. Itu RCE klasik yang dibungkus "technical test".
2. **Channel network masih menang.** Dari ~8 reply, 2 dapat kerja dari kenalan
   atau WhatsApp, 3 dari LinkedIn, nol dari "scan board sendiri". Board itu
   pintu masuk, bukan penentu.

ApplyDiff sudah punya scan + filter + match. Yang belum ada: **penjaga scam** dan
**pencatatan channel**. Dua-duanya murah dan nutup lubang nyata.

## 1. Scam guard

Bukan LLM-first. Aturan deterministik dulu, LLM cuma buat menjelaskan.

### Aturan keras (block, wajib punya alasan)

| Sinyal | Kenapa |
|---|---|
| Domain email gak cocok sama domain company | `hr@gmail.com` yang ngaku startup US |
| Minta bayar (training, kit, deposit) | Selalu scam |
| Minta KTP/rekening/selfie sebelum interview | Identitas dipakai buat pinjol |
| Task "clone repo ini lalu jalankan" | Eksekusi kode asing = RCE |
| Repo/task tanpa penjelasan konteks bisnis | Test asli punya konteks |
| Kontak cuma WhatsApp/Telegram tanpa domain company | Gak ada jejak |
| Salary di JD jauh di atas pasar untuk level itu | Umpan |

### Aturan lunak (tandai, tetap boleh lanjut)

- Domain company umur < 6 bulan (perlu lookup whois)
- Alamat kantor gak ada / cuma "remote worldwide"
- Deskripsi JD identik dengan posting lain (copy-paste scam farm)
- Recruiter gak punya profil LinkedIn yang cocok

### Bentuk di app

- Tiap job punya `risk` (0–100) + daftar `flags` dengan alasan.
- Di feed, job berisiko ditandai, bukan disembunyikan. User yang putuskan.
- `applydiff scan --min-risk 0` buat lihat semua.
- `applydiff show <key>` nampilin flags lengkap.

## 2. Channel & referral

Field baru di aplikasi:

- `channel`: `board` | `referral` | `inbound` | `linkedin` | `email` | `other`
- `referrer`: nama/handle (opsional)
- `notes`

Gunanya: **statistik jujur**. Setelah 20 lamaran, app bisa bilang "referral 4/6
reply, board 1/14". Itu insight yang gak dimiliki tool lain.

`applydiff stats` → reply rate per channel. TUI + Web nampilin di Applications.

## 3. Freshness + kompetisi

Dari thread @aaksanurirwan:

- Filter keras: `posted_within: 5h` (buat mode agresif)
- Filter lunak: jumlah pelamar kalau ada di source (Greenhouse kadang ada)
- Prioritas = freshness × match × (1 / pelamar)

## Fase + checklist

### Fase A — Model
- [ ] Kolom `risk`, `flags` di job; `channel`, `referrer`, `notes` di application
- [ ] Migrasi timestamp (bukan nomor urut)
- [ ] Seed daftar flag + bobot

### Fase B — Detektor
- [ ] Aturan keras: domain email, minta bayar, KTP, clone repo, kontak tanpa domain
- [ ] Aturan lunak: umur domain, alamat kosong, JD duplikat, recruiter gak jelas
- [ ] `RiskJob(jd, company) -> {score, flags}` dengan test per aturan
- [ ] LLM opsional: jelaskan kenapa berisiko (bukan yang memutuskan)

### Fase C — Permukaan
- [ ] `applydiff scan --min-risk`, `show` nampilin flags
- [ ] TUI: badge risiko di Feed, detail di Job
- [ ] Web: badge + panel alasan
- [ ] `applydiff stats` (reply rate per channel)
- [ ] Form channel/referrer di Application

### Fase D — Freshness
- [ ] Filter `posted_within`
- [ ] Skor prioritas pakai jumlah pelamar kalau tersedia
- [ ] Preset "agresif" (fresh < 5h, pelamar rendah)

## Non-goals

- Gak ngeklaim bisa mendeteksi semua scam. Ini filter, bukan jaminan.
- Gak auto-tolak. Selalu user yang putuskan.
- Gak scraping LinkedIn (ToS).

## Definisi selesai

- Job berisiko punya skor + alasan yang bisa dibaca.
- Satu test per aturan keras, semuanya lulus.
- `stats` nampilin reply rate per channel dari data nyata.
- Filter freshness jalan di scan + feed.
