# Project Overview: Skill Drill API

**Objective:** Build a REST API backend using Golang for a gamified _Micro-LMS_ application (similar to Duolingo). This app trains users by having them answer questions on various topics (Civil Service Exam/CPNS, Web, Multiplatform, DevOps/MLOps, ML/DL/Gen AI, Interview Prep, etc.) consistently every day.
**Core Tech Stack:** Golang (Go 1.21+), PostgreSQL, Firebase Cloud Messaging (FCM), LLM API (DeepSeek/Gemini), and Docker.

## 1. Architecture & Directory Structure

Please use the standard Golang layout (Clean Architecture / Standard Go Project Layout):

- `/cmd/api`: Application entry point (`main.go`).
- `/internal/api`: Handlers/Controllers for HTTP requests (Use `Gin` or `Echo` framework).
- `/internal/services`: Business logic.
- `/internal/repository`: Database interactions (Use `GORM` or `sqlx`).
- `/internal/worker`: Cron Job & Goroutines logic for background tasks (FCM Spam Notifications).
- `/pkg`: Supporting utilities (LLM Client, PDF Extractor, FCM Client).

## 2. Database Schema (PostgreSQL)

Create entities/models for the following tables:

- `users`: id, username, email, fcm_token, current_streak, highest_streak, last_study_date, created_at.
- `topics`: id, name (e.g., "CPNS", "AWS", "Go"), description, created_at.
- `questions`: id, topic_id, question_text, option_a, option_b, option_c, option_d, correct_answer, explanation, created_at.
- `study_sessions`: id, user_id, topic_id, score, is_completed, created_at.

## 3. Core Features & Business Logic (STRICTLY REQUIRED)

### A. AI Question Generator (Admin Flow - REVISED)

Admins can generate questions using 2 modes (PDF is optional). Please create the service logic:
**Endpoint:** `POST /admin/generate-questions`
**Payload:** `topic_id` (int), `prompt_theme` (string), `num_questions` (int), `file_pdf` (file/multipart - _optional_).

**Service Logic:**

1. **If `file_pdf` IS NOT EMPTY (Context-Based):**
   - Extract text from the PDF using a Go library (e.g., `ledongthuc/pdf`).
   - Construct System Prompt: _"Based on the following text, generate {num_questions} multiple-choice questions about {prompt_theme}..."_
2. **If `file_pdf` IS EMPTY (Zero-Shot / Topic-Based):**
   - Construct System Prompt: _"You are an expert question creator. Generate {num_questions} random multiple-choice questions with medium to hard difficulty about {prompt_theme}..."_
3. **LLM Execution & Parsing:**
   - Send the prompt to the LLM API (provide an interface for Gemini/DeepSeek).
   - **CRITICAL:** Force the LLM to respond in a pure JSON Array format.
   - Parse the JSON response into the Golang `Question` struct.
   - Perform a _Bulk Insert_ into the PostgreSQL database.

### B. Daily Streak & Gamification (User Flow)

- Every time a user submits an answer via the `POST /study/submit` endpoint, compare today's date with the user's `last_study_date`.
- If the dates are different = increment `current_streak` by +1, update `last_study_date` to today.
- If the gap between `last_study_date` and today is > 1 day = reset `current_streak` to 0.

### C. Spam Notification Engine (Worker/Cron)

- Create a Cron Job (`robfig/cron`) that runs every 30 minutes (from 18:00 to 23:30).
- **Query Logic:** `SELECT fcm_token FROM users WHERE last_study_date != CURRENT_DATE OR last_study_date IS NULL`.
- Use **Goroutines** to broadcast to Firebase FCM so that sending thousands of notifications runs in parallel and quickly without _blocking_.
- The FCM Payload should contain an aggressive message (e.g., "Your streak is about to be lost! Open OmniPrep now!").

## 4. Expected Deliverables from the Agent

1. Setup `go.mod` and initialize dependencies.
2. Complete code for database entities and their automatic migrations.
3. The `internal/services/llm_service.go` code handling the 2 question generation modes (with/without PDF) along with examples of JSON-strict system prompts.
4. The `internal/worker/notifier.go` code implementing the Cron and Goroutines for FCM.
5. Setup `Dockerfile` and `docker-compose.yml` to run the backend app and PostgreSQL database locally.

Please generate this step-by-step. Start with project initialization, database setup, and models!
