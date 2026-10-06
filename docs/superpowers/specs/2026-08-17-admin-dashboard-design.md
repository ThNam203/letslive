# Admin Dashboard — Phase 1 Design

**Date:** 2026-08-17
**Branch:** feat/admin-dashboard (to be created)
**Author:** ThNam203

---

## Overview

A standalone admin subsystem for platform operators: its own backend service, its own database, its own frontend app. Phase 1 is read-only: a stats overview (five platform-wide counts), plus browsable lists and detail pages for streams and VODs across all users. No moderation actions (no edit/ban/remove), no admin account management UI, no video playback in the admin panel yet.

Admin identity has zero overlap with regular user identity: separate accounts table, separate JWT secret, separate login flow. Compromising the main app's auth does not grant admin access, and vice versa. This was chosen over adding a `role` column + RBAC to the existing `user`/`auth` services specifically to keep the two trust boundaries fully independent — no changes to `backend/auth` or `backend/user`'s existing auth code paths.

---

## Decisions Made

| Question | Decision |
|---|---|
| Scope | Phase 1 = RBAC-free standalone system, read-only: stats overview + browsable streams/VODs lists and detail pages. Moderation actions, account management UI, reports/flagging, and trend charts are later phases. |
| Isolation | Fully separate service (`backend/admin`), database (`admin_db`), and frontend (`admin-web/`). Not a role bolted onto the existing user model. |
| Admin account creation | Seeded by hand via SQL. No signup flow, no admin-management UI this phase. |
| Session strategy | Single JWT on login (~24h expiry), own secret. No refresh-token table/rotation — re-login daily is acceptable for a low-traffic internal tool. |
| Stats aggregation | `admin` service calls new internal (unauthenticated, internal-network-only) stats endpoints on `user`/`livestream`/`vod`/`finance` directly via service discovery — no new aggregator/analytics service. |
| Stats content | Core counts only: total users, new signups (7d), active streams, total VODs, total revenue. No time-series/trend charts (would need time-bucketed queries — deferred). |
| Streams/VODs content depth | Metadata list + detail page (title, owner, status, visibility, view count, timestamps). No video playback embedded in the admin panel — that's deferred with the moderation-action phase, when it'd actually inform a decision. |
| Streams/VODs list scope | All records regardless of visibility (public + private) — an admin needs to see private content too. Existing per-user/public-only list methods (`GetRecommendedLivestreams`, `GetPopular`, `GetByUser`) don't cover this; new unscoped, paginated list methods are added. |
| List filtering | Paginated, newest-first, no search/filter in phase 1 — add when it's actually needed (YAGNI). |

---

## Data Model

### New database — `admin_db`

Own Postgres database inside the existing shared `main_db` container, same pattern as every other service (own `ADMIN_DB_USER`/`ADMIN_DB_PASSWORD`).

```sql
-- 0001_create_admin_accounts_table.sql
CREATE TABLE admin_accounts (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email         VARCHAR(255) NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

No FK to the main `users` table — fully separate identity space.

---

## API Design

### `admin` service — new endpoints

```
POST /v1/admin/login              # email + password -> sets ADMIN_ACCESS_TOKEN httpOnly cookie (public)
GET  /v1/admin/stats              # aggregated stats                                 (private, admin JWT)
GET  /v1/admin/streams            # paginated list, all visibility                   (private, admin JWT)
GET  /v1/admin/streams/{id}       # single stream detail                             (private, admin JWT)
GET  /v1/admin/vods               # paginated list, all visibility                   (private, admin JWT)
GET  /v1/admin/vods/{id}          # single VOD detail                                (private, admin JWT)
```

These are the Go service's internal route registrations (`sm.Handle(...)`, same as every other service). Kong's `admin` service entry sets `path: /v1` at the service level (same pattern as `finance`/`vod`/etc.), which prefixes `/v1` onto the upstream request — so the browser/`admin-web`-facing path drops it: `POST /admin/login`, `GET /admin/stats`, `GET /admin/streams`, etc. (confirmed by how `web/` already calls `finance`'s `/v1/wallet` as just `/wallet` through Kong).

`GET /v1/admin/streams` and `GET /v1/admin/vods` accept `?page=&limit=` (default `limit=20`), newest-first. Response items are enriched with the owner's `username` (see username lookup below) since the underlying records only carry `userId`.

`POST /v1/admin/login` request body:
```json
{ "email": "admin@letslive.app", "password": "..." }
```

`GET /v1/admin/stats` response:
```json
{
  "totalUsers": 12345,
  "newSignups7d": 210,
  "activeStreams": 18,
  "totalVods": 4032,
  "totalRevenue": 985000
}
```

### `user` / `livestream` / `vod` / `finance` services — new internal endpoints

```
GET /v1/internal/stats                     # service-to-service only, no JWT, internal network
```

- `user`: `{ "totalUsers": ..., "newSignups7d": ... }`
- `livestream`: `{ "activeStreams": ... }`
- `vod`: `{ "totalVods": ... }`
- `finance`: `{ "totalRevenue": ... }`

```
# livestream service
GET /v1/internal/livestreams?page=&limit=  # all visibility, newest-first
GET /v1/internal/livestreams/{id}          # single record

# vod service
GET /v1/internal/vods?page=&limit=         # all visibility, newest-first
GET /v1/internal/vods/{id}                 # single record

# user service — batch username lookup, used to enrich stream/VOD list items
POST /v1/internal/users/batch              # body: { "ids": ["uuid1", "uuid2", ...] } -> [{ "id": ..., "username": ... }]
```

Batch lookup is POST+body rather than GET+query — an id list isn't a resource filter, it's an RPC-style batch read, and a query string has no safe bound on how many ids it can carry. Also matches the existing internal-endpoint convention (`vod`'s `create_vod_internal.go` is POST+JSON body). No auth on any of these — trust boundary is the internal Docker/Consul network, not reachable via Kong.

---

## Data Flow

Browser-facing paths below drop the `/v1` prefix (Kong adds it back for the upstream call — see note above).

```
Browser (admin-web)
  → POST /admin/login (via Kong, plain proxy, no auth)
  admin service [/v1/admin/login]
    → verify email/password against admin_accounts (bcrypt)
    → sign JWT (ADMIN_JWT_SECRET, ~24h expiry)
    → Set-Cookie: ADMIN_ACCESS_TOKEN (httpOnly, Secure, SameSite=None, MaxAge=24h) — matches how the main app's login never returns the token in the JSON body either
  ← 200 empty body

Browser (admin-web)
  → GET /admin/stats (via Kong, plain proxy) -> RequireAdminAuth middleware verifies JWT with ADMIN_JWT_SECRET
  admin service [/v1/admin/stats]
    → discovery.Registry.ServiceAddress("user") → GET http://<addr>/v1/internal/stats
    → discovery.Registry.ServiceAddress("livestream") → GET http://<addr>/v1/internal/stats
    → discovery.Registry.ServiceAddress("vod") → GET http://<addr>/v1/internal/stats
    → discovery.Registry.ServiceAddress("finance") → GET http://<addr>/v1/internal/stats
    → merge into one response (each call independent — one failing does not fail the others)
  ← { totalUsers, newSignups7d, activeStreams, totalVods, totalRevenue }

Browser (admin-web)
  → GET /admin/streams?page=1 (via Kong, plain proxy) -> RequireAdminAuth middleware verifies JWT with ADMIN_JWT_SECRET
  admin service [/v1/admin/streams]
    → livestream gateway → GET /v1/internal/livestreams?page=1
    → user gateway → POST /v1/internal/users/batch { ids: [userId1, userId2, ...] } (dedup'd from the page's results)
    → merge username onto each stream item
  ← [{ id, title, ownerUsername, visibility, status, viewCount, startedAt, endedAt }, ...]

# GET /admin/vods follows the same shape, via the vod + user gateways.
```

---

## Backend Code Structure

### New service `backend/admin`

Mirrors the existing Go service skeleton used by `vod`/`finance`/etc.

```
cmd/                             # main.go, service bootstrap
config/                          # config.go (DB, JWT secret, service discovery)
domains/
  admin_account.go                # AdminAccount struct + AdminAccountRepository interface
  stats.go                        # Stats struct (aggregated response)
migrations/
  0001_create_admin_accounts_table.sql
repositories/admin_account/
  admin_account.go                # postgres repo + constructor
  get_by_email.go
gateway/
  user/http/http.go               # GET /v1/internal/stats, POST /v1/internal/users/batch via discovery.Registry
  livestream/http/http.go         # GET /v1/internal/stats, /livestreams, /livestreams/{id}
  vod/http/http.go                # GET /v1/internal/stats, /vods, /vods/{id}
  finance/http/http.go            # GET /v1/internal/stats
services/
  auth.go                         # AuthService{repo} — verify password, issue JWT
  stats.go                        # StatsService{userGW, livestreamGW, vodGW, financeGW}
  streams.go                      # StreamsService{livestreamGW, userGW} — list/get + username enrichment
  vods.go                         # VODsService{vodGW, userGW} — list/get + username enrichment
handlers/
  auth/
    auth.go                       # AuthHandler embeds BaseHandler
    login_public.go                # POST /v1/admin/login
  stats/
    stats.go                       # StatsHandler embeds BaseHandler
    get_stats_private.go           # GET /v1/admin/stats
  streams/
    streams.go                     # StreamsHandler embeds BaseHandler
    list_streams_private.go        # GET /v1/admin/streams
    get_stream_private.go          # GET /v1/admin/streams/{id}
  vods/
    vods.go                        # VODsHandler embeds BaseHandler
    list_vods_private.go           # GET /v1/admin/vods
    get_vod_private.go             # GET /v1/admin/vods/{id}
dto/
  auth.go                          # LoginRequestDTO, LoginResponseDTO
  stats.go                         # StatsResponseDTO
  stream.go                        # StreamListItemDTO, StreamDetailDTO
  vod.go                           # VODListItemDTO, VODDetailDTO
types/
  jwt_claims.go                    # AdminClaims{ AdminId string }
utils/
  password.go                      # bcrypt hash/verify (same lib as backend/auth, separate code)
```

### `user` / `livestream` / `vod` / `finance` services — new files each

```
# user
handlers/user/get_stats_internal.go        # GET /v1/internal/stats
handlers/user/get_usernames_internal.go    # POST /v1/internal/users/batch
repositories/user/get_stats.go
repositories/user/get_by_ids.go

# livestream
handlers/livestream/get_stats_internal.go
handlers/livestream/list_livestreams_internal.go
handlers/livestream/get_livestream_internal.go
repositories/livestream/get_stats.go
repositories/livestream/get_all_paginated.go   # new: unscoped, all visibility

# vod
handlers/vod/get_stats_internal.go
handlers/vod/list_vods_internal.go
handlers/vod/get_vod_internal.go
repositories/vod/get_stats.go
repositories/vod/get_all_paginated.go          # new: unscoped, all visibility

# finance
handlers/finance/get_stats_internal.go
repositories/finance/get_stats.go
```

---

## Frontend — `admin-web/` (new top-level directory)

Separate Next.js app, own `package.json` and `Dockerfile`, sibling to `web/` and `backend/`. No i18n/locale routing — internal tool, single language.

```
admin-web/
  app/
    login/
      page.tsx                    # email/password form -> POST /admin/login (Kong), sets cookie
    (dashboard)/
      layout.tsx                  # RequireAdminAuth wrapper (redirects to /login if no token) + nav
      page.tsx                    # stats overview — 5 <StatCard />
      streams/
        page.tsx                  # paginated table: title, owner, status, visibility, viewers, started
        [id]/page.tsx              # stream detail — full metadata, no playback
      vods/
        page.tsx                  # paginated table: title, owner, status, visibility, views, duration, created
        [id]/page.tsx              # VOD detail — full metadata, no playback
      _components/
        stat-card.tsx              # loading / error / success states, independent per card
        data-table.tsx              # shared paginated table (used by streams + vods list pages)
  lib/
    api/
      admin.ts                    # Login(), GetStats(), ListStreams(), GetStream(), ListVods(), GetVod()
  hooks/
    use-admin-auth.ts              # token cookie read/write, logout
```

---

## Infra Wiring

Mirrors the pattern from recent commits (`58cc58b` — force-recreate Kong on deploy for route changes; `612a740` — wire Stripe secrets into deploy `.env`):

- `docker-compose.yaml` + `docker-compose-dev.yaml`: new `admin` and `admin-web` service blocks; `admin_db` env vars (`ADMIN_DB_USER`, `ADMIN_DB_PASSWORD`) added to the shared `main_db` block's init.
- `configs/kong.yml`: new `admin` service entry (`host: admin.service.consul`, `path: /v1`), plain proxy routes for `/admin/*` — **no** Kong `jwt` plugin attached (precedent: `MinIO`/`Transcode` routes are also plugin-less). Reason: Kong's existing `jwt` plugin config here uses a single shared consumer/secret with no ACL — it validates *a* signature exists, not *which* secret signed it, so a second consumer wouldn't actually stop a main-app token from passing on an admin route. Instead, the `admin` service does its own full signature verification (`jwt.ParseWithClaims` with `ADMIN_JWT_SECRET`, not the `ParseUnverified` other services use) in a `RequireAdminAuth` middleware — a token signed with the main app's `ACCESS_TOKEN_SECRET` fails this outright, giving real isolation without new Kong plugin/ACL config.
- `.github/workflows/build-and-publish-images.yml`: add build/publish steps for the two new images.
- `.github/workflows/deploy.yml`: wire `ADMIN_DB_USER`, `ADMIN_DB_PASSWORD`, `ADMIN_JWT_SECRET` into the deploy `.env`.

---

## Error Handling

- Each of the 4 internal stats calls is independent: `StatsService` collects results and errors per-source. `GET /v1/admin/stats` always returns 200 with all 5 fields; a field whose source call failed is `null` instead of a number, so one service being down doesn't fail the whole request.
- `admin-web` stat cards render per-card error state with retry — no mock/fallback data (hard rule: never fall back to fake data on error).
- Admin login failure returns a generic "invalid credentials" for both unknown-email and wrong-password cases (no user enumeration).
- Internal endpoints (stats, streams/vods list+detail, username batch lookup) are unauthenticated by design (internal-network-only) — not reachable through Kong, no public route configured for them.
- `GET /v1/admin/streams`/`/vods`: if the username-lookup call to `user` fails, items are still returned with `ownerUsername: null` rather than failing the whole list — the list itself isn't blocked by an enrichment failure.
- `GET /v1/admin/streams/{id}`/`/vods/{id}` for a nonexistent id returns 404, matching the existing per-service `GetById` convention.

---

## Testing

- `backend/admin`: unit tests for login (bad password, unknown email, bcrypt verify), JWT middleware (missing/expired/malformed token), `StatsService` aggregation (one gateway call failing doesn't fail the others), `StreamsService`/`VODsService` list+detail (including username-lookup failure -> `null` owner, not a hard failure) and 404 on unknown id.
- `user`: unit tests for `get_stats_internal` and the batch username lookup (unknown ids silently omitted, not errored).
- `livestream`/`vod`: unit tests for `get_stats_internal` and the new unscoped `get_all_paginated` repository query (confirms private-visibility records are included, unlike the existing public-only queries).
- `finance`: unit test for `get_stats_internal`.
- `admin-web`: component tests for `StatCard` (loading/error/success), the shared paginated `data-table` (loading/error/empty/pagination), login form validation, `RequireAdminAuth` redirect behavior.

---

## Out of Scope (Phase 1)

- Ban/suspend/remove actions, on users, streams, or VODs. Note: `UserStatus.DISABLED` already exists on the `user` service's user model and could back a future suspend action without a new migration.
- Editing stream/VOD metadata from the admin panel.
- Video playback (live stream preview or VOD player) embedded in the admin panel.
- Search/filter on the streams/VODs lists (plain pagination only, newest-first).
- Admin account management UI (creating/revoking admin accounts stays SQL-only).
- Reports/flagging queue.
- Trend charts / time-series stats.
- Refresh-token rotation for admin sessions (single long-lived token is sufficient for now).
