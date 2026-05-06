# VOD Upload Flow — End-to-End Map & Flaw Analysis

## 1. Overview

User uploads video file via web → VOD service stores raw file in MinIO + creates DB row + enqueues transcode job → Transcode worker polls jobs, runs FFmpeg, uploads HLS, calls back VOD service to mark `ready` → Web reads VOD and displays duration.

---

## 2. End-to-End Flow

| # | Phase | File | Action |
|---|-------|------|--------|
| 1 | Web upload call | `web/lib/api/vod.ts:63-80` | `UploadVOD()` POST multipart `/vods/upload` (`disableTimeout: true`) |
| 2 | HTTP handler | `backend/vod/handlers/vod/upload_vod_private.go:13-57` | Parse multipart (32MB buffer), extract file + title/desc/visibility |
| 3 | Service entry | `backend/vod/services/vod/upload_vod.go:25-114` | Validate ext (.mp4/.mov/.avi/.mkv/.webm); generate UUID |
| 4 | Raw storage | `backend/vod/services/vod/upload_vod.go:58-69` | MinIO PUT `raw-videos/{vodId}/{filename}` |
| 5 | DB insert | `backend/vod/services/vod/upload_vod.go:77-95` | VOD row: `status=processing`, `duration=0` |
| 6 | Enqueue job | `backend/vod/services/vod/upload_vod.go:98-111` | Insert `transcode_jobs` row, `status=pending`, `max_attempts=3` |
| 7 | Worker poll | `backend/transcode/worker/worker.go:68-86` | Loop every 5s, `SELECT ... FOR UPDATE` next pending job |
| 8 | Mark processing | `backend/transcode/worker/worker.go:131` | `UPDATE transcode_jobs SET status='processing', started_at=now(), attempts++` |
| 9 | Download raw | `backend/transcode/worker/worker.go:171-175` | Pull raw file from MinIO |
| 10 | FFmpeg transcode | `backend/transcode/transcoder/file_transcoder.go:42-66` | `ffmpeg -i in -c:v libx264 -c:a aac -f hls -hls_time 10 -hls_list_size 0 -master_pl_name index.m3u8` |
| 11 | Upload HLS | `backend/transcode/worker/worker.go:193-198` | Push segments + playlist to MinIO |
| 12 | Thumbnail | `backend/transcode/worker/worker.go:202-209` | Separate ffmpeg invocation |
| 13 | Callback | `backend/transcode/gateway/livestream/http/http.go:132-177` | `PATCH /v1/internal/vods/{id}/status` with `{status, playbackUrl, thumbnailUrl}` |
| 14 | Apply update | `backend/vod/repositories/vod/update_vod_status.go:12-39` | `UPDATE vods SET status, playback_url, thumbnail_url, updated_at` |
| 15 | Mark job done | `backend/transcode/worker/worker.go:219` | `UPDATE transcode_jobs SET status='completed', completed_at=now()` |
| 16 | Cleanup | `backend/transcode/worker/worker.go:225-227` | Delete raw file from MinIO |
| 17 | Web display | `web/components/livestream/vod-card.tsx:91` | `formatSeconds(vod.duration)` → `0:00` when 0 |

---

## 3. DB Schema (`backend/vod/migrations/0001_init_vod_tables.sql`)

`vods` table:
- `duration BIGINT DEFAULT 0`
- `status VARCHAR(20) DEFAULT 'ready'` (overridden to `processing` on insert)
- `playback_url`, `original_file_url`, `thumbnail_url`, `view_count`, timestamps

`transcode_jobs` table:
- `status VARCHAR(20) DEFAULT 'pending'`
- `attempts`, `max_attempts=3`, `error_message`, `started_at`, `completed_at`

---

## 4. ROOT CAUSE — Why duration shows 0:00

Duration is **never written** anywhere in pipeline. Three missing links:

### Flaw A — FFmpeg never extracts duration
`backend/transcode/transcoder/file_transcoder.go:104-111` returns only `(masterPlaylistPath, thumbnailPath, error)`. No `ffprobe` call, no parsing of FFmpeg stderr metadata, no HLS playlist scan for `#EXTINF` sums.

### Flaw B — Callback payload omits duration
`backend/transcode/gateway/livestream/http/http.go:141-149` sends:
```json
{ "status": "ready", "playbackUrl": "...", "thumbnailUrl": "..." }
```
No `duration` field.

### Flaw C — UPDATE SQL omits duration column
`backend/vod/repositories/vod/update_vod_status.go:13-15`:
```sql
UPDATE vods SET status=$1,
  playback_url=COALESCE($2, playback_url),
  thumbnail_url=COALESCE($3, thumbnail_url),
  updated_at=now()
WHERE id=$4
```
Duration column not touched. Stays at insert default `0`.

Result: `duration=0` BIGINT → frontend `formatSeconds(0)` → `"0:00"`.

---

## 5. Fix for Reported Bug — Duration shows `0:00`

Severity: HIGH (bug). Three coordinated changes:

**Step 1 — Extract duration in transcoder**
- File: `backend/transcode/transcoder/file_transcoder.go:104-111`
- Add `ffprobe -v error -show_entries format=duration -of csv=p=0 <inputPath>` invocation before or in parallel with ffmpeg.
- Parse stdout float seconds → `int64` (round or floor).
- Change return signature: `(masterPlaylistPath, thumbnailPath string, durationSec int64, err error)`.
- Fallback: if ffprobe fails, sum HLS `#EXTINF:<n>` lines from generated playlist.

**Step 2 — Forward duration through worker + gateway**
- File: `backend/transcode/worker/worker.go:185, 212`
  - Capture new return value from `TranscodeFile`.
  - Pass into `UpdateVODStatus(ctx, vodId, "ready", playbackURL, thumbnailURL, durationSec)`.
- File: `backend/transcode/gateway/livestream/http/http.go:132-177`
  - Add `Duration int64` to request DTO and method signature.
  - Include `"duration": <n>` in JSON payload at line 141-149.

**Step 3 — Persist duration in VOD service**
- File: `backend/vod/handlers/vod/update_vod_status_internal.go:14-18`
  - Add `Duration *int64 \`json:"duration"\`` to `UpdateVODStatusRequest`.
  - Pass to service call at line 42.
- File: `backend/vod/services/vod/update_vod_status.go` (and downstream)
  - Thread `duration *int64` to repo.
- File: `backend/vod/repositories/vod/update_vod_status.go:13-15`
  - Add `duration = COALESCE($N, duration)` to UPDATE clause.

**Backfill (optional)**
- One-time script: for each VOD with `status=ready` and `duration=0`, ffprobe its HLS or the still-present raw, update column. Or trigger re-transcode.

---

## 6. Other Flaws + Recommended Actions

### F1 — Worker poll uses `FOR UPDATE` not `FOR UPDATE SKIP LOCKED`
- Severity: MEDIUM (scalability)
- File: `backend/transcode/worker/worker.go:105-113`
- Problem: multiple worker replicas block on row lock instead of skipping to next pending job. Concurrency = 1 globally.
- Fix: change query to `... FOR UPDATE SKIP LOCKED`. Also consider adding `WHERE created_at < now() - interval '1s'` to avoid race with insert commit.
- Optional: tune poll interval, add LISTEN/NOTIFY for instant pickup.

### F2 — `disableTimeout: true` on web upload, no abort UI
- Severity: LOW (UX)
- File: `web/lib/api/vod.ts:75`
- Problem: stalled upload hangs forever; user has no cancel button or progress indicator.
- Fix: pass `AbortController` to fetch; expose Cancel button in upload UI; show progress via `XMLHttpRequest` upload events or chunked TUS-style upload.

### F3 — No upload size cap → disk-fill DoS
- Severity: HIGH (security)
- File: `backend/vod/handlers/vod/upload_vod_private.go:23-30`
- Problem: `r.ParseMultipartForm(32 << 20)` only sets memory buffer — files >32MB spill to disk with no upper bound. Attacker uploads multi-TB file → fills disk.
- Fix: wrap before ParseMultipartForm:
  ```go
  r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_BYTES) // e.g. 5 GiB
  ```
- Also: validate `fileHeader.Size` and reject before reading body.

### F4 — Raw file orphaned on transcode failure
- Severity: MEDIUM (storage cost)
- File: `backend/transcode/worker/worker.go:225-227` (delete only on success)
- Problem: jobs that exhaust `max_attempts` leave raw file in MinIO forever.
- Fix: in failure branch (when `attempts >= max_attempts`), also delete raw, OR move to a `failed-videos/` prefix with TTL/lifecycle policy. Add MinIO bucket lifecycle rule: auto-expire `raw-videos/` after N days regardless.

### F5 — No `failed` status path → VODs stuck in `processing` forever
- Severity: HIGH (data integrity / UX)
- Files: `backend/transcode/worker/worker.go` (failure branch), `backend/vod/services/vod/update_vod_status.go`, `web/types/vod.ts` (already has `"failed"`).
- Problem: when `attempts >= max_attempts`, only `transcode_jobs.status` updates. VOD row `status='processing'` never transitions. Frontend can't show error.
- Fix:
  1. In worker failure handler, after marking job failed, call `UpdateVODStatus(ctx, vodId, "failed", "", "", 0)`.
  2. Ensure `update_vod_status` repo allows setting `failed`.
  3. Frontend: render "Processing failed" state in `vod-card.tsx` when `status === "failed"`.

### F6 — `playback_url`/`thumbnail_url` use `COALESCE`, can't be cleared
- Severity: LOW (consistency)
- File: `backend/vod/repositories/vod/update_vod_status.go:13-15`
- Problem: passing nil leaves old value; no way to reset to NULL. Inconsistent with `status` which is direct assignment.
- Fix: decide policy — either always overwrite (drop COALESCE), or document COALESCE as intentional. Recommend overwrite for `failed` transitions to clear stale URLs.

### F7 — Two ffmpeg invocations (transcode + thumbnail)
- Severity: LOW (perf)
- File: `backend/transcode/transcoder/file_transcoder.go:107-109`
- Problem: separate ffmpeg call for thumbnail = duplicate decode cost.
- Fix: combine into one ffmpeg with multiple outputs:
  ```
  ffmpeg -i in -map 0 -c:v libx264 -c:a aac -f hls ... \
                -map 0:v -ss 3 -vframes 1 -update 1 thumb.jpg
  ```
- Or extract thumbnail from a generated HLS segment (no re-decode of source).

### F8 — Migration `status` default mismatches service
- Severity: LOW (correctness)
- File: `backend/vod/migrations/0001_init_vod_tables.sql`
- Problem: column default is `'ready'` but service inserts `'processing'`. Direct INSERT bypassing service yields `ready` VOD with NULL `playback_url` → broken player.
- Fix: change column default to `'processing'`. New migration:
  ```sql
  ALTER TABLE vods ALTER COLUMN status SET DEFAULT 'processing';
  ```

### F9 — No transcode idempotency check
- Severity: MEDIUM (correctness / waste)
- File: `backend/transcode/worker/worker.go:150-231`
- Problem: if worker crashes after `UpdateVODStatus` but before `transcode_jobs.status='completed'`, retry re-transcodes a `ready` VOD. Wastes work; could overwrite playable HLS mid-watch.
- Fix:
  1. At start of `doTranscode`, fetch current VOD status. If `ready`, mark job completed and skip.
  2. Wrap `UpdateVODStatus` + `transcode_jobs UPDATE completed` in single transaction (or use 2-phase: write status only via job-completion-tied trigger).
  3. Use an idempotency key (e.g. `transcode_jobs.id`) on output paths so retries are deterministic.

### F10 — Extension-only file validation
- Severity: HIGH (security)
- File: `backend/vod/services/vod/upload_vod.go:35-44`
- Problem: only checks filename suffix. Renamed `.exe` → `.mp4` accepted. Mismatch with stored mimeType.
- Fix:
  1. Read first 512 bytes, call `http.DetectContentType` (Go stdlib) before storing.
  2. Reject if not in `[video/mp4, video/quicktime, video/x-matroska, video/webm, video/x-msvideo]`.
  3. Reset reader (`io.MultiReader` of buffered head + rest) before MinIO upload.
  4. Also: have transcode worker bail early if ffprobe reports no video stream.

---

## 7. Suggested Prioritization

| Priority | Item | Why |
|----------|------|-----|
| P0 | §5 duration fix (A/B/C) | Reported user bug |
| P0 | F3 MaxBytesReader | Disk-fill DoS |
| P0 | F10 magic-byte MIME check | Arbitrary file upload |
| P1 | F5 failed-status path | UX dead-end on errors |
| P1 | F1 SKIP LOCKED | Blocks horizontal scaling |
| P1 | F9 idempotency | Crash-recovery correctness |
| P2 | F4 orphan cleanup | Storage cost over time |
| P2 | F8 default mismatch | Hardening |
| P3 | F2, F6, F7 | Polish |
