# Skill Drill API

REST API backend untuk aplikasi Micro-LMS gamifikasi menggunakan Go, Gin, GORM, PostgreSQL, LLM, dan Firebase Cloud Messaging.

## Persyaratan

- Go 1.25 atau lebih baru
- PostgreSQL 14 atau lebih baru
- Docker dan Docker Compose (opsional)
- API key DeepSeek atau provider LLM yang kompatibel
- Firebase service account JSON untuk notifikasi FCM (opsional saat development)

## Menjalankan Secara Lokal

1. Salin template konfigurasi:

   ```bash
   cp .env.example .env
   ```

2. Isi `DB_PASSWORD` dengan password PostgreSQL Anda dan isi `LLM_API_KEY` bila ingin menggunakan generator AI.

3. Pastikan database `skill_drill` sudah tersedia, lalu jalankan:

   ```bash
   go run ./cmd/api
   ```

## Menjalankan dengan Docker

```bash
docker compose up --build
```

API tersedia di `http://localhost:8080`. PostgreSQL tersedia di port `5432`. Compose menggunakan `DB_HOST=db` di dalam container API dan me-mount `firebase.json` sebagai file read-only.

## Konfigurasi Environment

| Variable                    | Contoh                                      | Keterangan                      |
| --------------------------- | ------------------------------------------- | ------------------------------- |
| `PORT`                      | `8080`                                      | Port HTTP API                   |
| `DB_HOST`                   | `localhost`                                 | Host PostgreSQL                 |
| `DB_USER`                   | `postgres`                                  | User PostgreSQL                 |
| `DB_PASSWORD`               | `password`                                  | Password PostgreSQL             |
| `DB_NAME`                   | `skill_drill`                               | Nama database                   |
| `DB_PORT`                   | `5432`                                      | Port PostgreSQL                 |
| `LLM_API_URL`               | `https://api.deepseek.com/chat/completions` | URL API LLM                     |
| `LLM_API_KEY`               | `your-api-key`                              | API key LLM                     |
| `LLM_MODEL`                 | `deepseek-chat`                             | Model LLM                       |
| `LOG_FILE`                  | `app.log`                                   | Lokasi file log aplikasi        |
| `FIREBASE_CREDENTIALS_FILE` | `./firebase.json`                           | Lokasi service account Firebase |
| `TIMEZONE`                  | `UTC`                                       | Zona waktu streak dan cron      |

Jangan commit `.env` atau `firebase.json` ke repository. Keduanya sudah masuk `.gitignore`.

## API Endpoints

Base URL default: `http://localhost:8080`

### `GET /api/health`

Memeriksa apakah API aktif.

```bash
curl http://localhost:8080/api/health
```

Response `200 OK`:

```json
{
  "status": "ok"
}
```

### `POST /api/questions/generate`

Membuat pertanyaan pilihan ganda menggunakan LLM. Request harus memakai `multipart/form-data`. File PDF bersifat opsional.

Form fields:

| Field           | Tipe    | Wajib | Keterangan                      |
| --------------- | ------- | ----- | ------------------------------- |
| `topic_id`      | integer | Ya    | ID topic yang sudah ada         |
| `prompt_theme`  | string  | Ya    | Tema pertanyaan                 |
| `num_questions` | integer | Ya    | Jumlah pertanyaan, 1 sampai 100 |
| `file_pdf`      | file    | Tidak | PDF sebagai konteks tambahan    |

Tanpa PDF:

```bash
curl -X POST http://localhost:8080/api/questions/generate \
	-F topic_id=1 \
	-F prompt_theme="Go concurrency" \
	-F num_questions=5
```

Dengan PDF:

```bash
curl -X POST http://localhost:8080/api/questions/generate \
	-F topic_id=1 \
	-F prompt_theme="AWS networking" \
	-F num_questions=5 \
	-F file_pdf=@./materi.pdf
```

Response `201 Created`:

```json
{
  "questions": [
    {
      "id": 1,
      "topic_id": 1,
      "question_text": "Apa kegunaan goroutine?",
      "option_a": "...",
      "option_b": "...",
      "option_c": "...",
      "option_d": "...",
      "correct_answer": "A",
      "explanation": "...",
      "created_at": "2026-08-27T13:30:00Z"
    }
  ]
}
```

LLM harus mengembalikan JSON array murni. Jumlah item yang berbeda dari `num_questions` akan ditolak dan tidak disimpan.

### `POST /api/study/submit`

Mengirim jawaban untuk sebuah pertanyaan dan memperbarui streak user.

Request:

```bash
curl -X POST http://localhost:8080/api/study/submit \
	-H 'Authorization: Bearer <token>' \
	-H 'Content-Type: application/json' \
	-d '{
		"question_id": 1,
		"answer": "A"
	}'
```

| Field         | Tipe    | Wajib | Keterangan                    |
| ------------- | ------- | ----- | ----------------------------- |
| `question_id` | integer | Ya    | ID question                   |
| `answer`      | string  | Ya    | Hanya `A`, `B`, `C`, atau `D` |

Response `200 OK`:

```json
{
  "correct": true,
  "current_streak": 4,
  "highest_streak": 7
}
```

Aturan streak:

- Submit pertama: `current_streak` menjadi `1`.
- Submit pada hari berikutnya: streak bertambah `1`.
- Jeda lebih dari satu hari: streak di-reset menjadi `0`, sesuai aturan aplikasi.
- `highest_streak` diperbarui jika `current_streak` lebih tinggi.

## Database

Saat startup, GORM menjalankan auto-migration untuk tabel:

- `users`
- `topics`
- `questions`
- `study_sessions`

Buat minimal satu topic dan satu user sebelum mencoba endpoint generator atau submit. API saat ini belum menyediakan endpoint CRUD untuk user dan topic.

## Notifikasi FCM

Worker berjalan otomatis setiap 30 menit pada rentang `18:00` sampai `23:30` sesuai `TIMEZONE`. Worker mencari user yang belum belajar hari ini dan mengirim notifikasi secara paralel melalui Firebase HTTP v1.

Service account Firebase dibaca dari `FIREBASE_CREDENTIALS_FILE`. Jika file tidak ditemukan, API tetap dapat berjalan tetapi notifikasi dinonaktifkan.

## Pengujian

```bash
go test ./...
go vet ./...
docker compose config -q
```
