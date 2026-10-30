# Search — Channels, Live Streams, Videos

## 1. Overview

One search box finds **channels (users)**, **live streams** (only those live
now) and **VODs**. Matching tolerates small typos and ignores Vietnamese
accents, and the matched part of each result is highlighted.

| Surface | Behaviour |
|---------|-----------|
| Desktop header | Dropdown grouped by type (Channels / Live now / Videos), **max 5 per type**. Enter or "See all results" opens `/search`. |
| Phone header | Search icon links to `/search` (no popup). |
| `/search` page | Tabs **All / Channels / Live / Videos**. All = 5 per type + "See all"; type tabs load more on scroll (infinite). |

Product decisions behind it:

- Phones use a full page, not a dialog: the old dialog was a cramped box with
  users only.
- VODs/lives match on **title, description and streamer name** (the user asked
  for all three).
- Matching is **accent-insensitive** ("tro choi" finds "Trò chơi"): most
  Vietnamese users type without accents on phone keyboards.
- Search text and tab live in the URL (`/search?q=jane&type=vods`) so a search
  can be shared or bookmarked. `type` is omitted for the All tab.

---

## 2. Flow

```
header input ──(300 ms debounce)──► 3 preview queries (limit 5) ──► dropdown
      │ Enter / "See all"
      ▼
/search?q=…&type=…  ──► All tab: same 3 preview queries (shared cache)
                     └► type tab: infinite query (page size 20)

GET /users/search        → user service  (fuzzy username match)
GET /livestreams/search  → livestream service ─┐
GET /vods/search         → vod service ────────┴─► user service /v1/users/search
                                                   (streamer-name ids, see §4.3)
```

The Envoy gateway already routes `GET /vods/...` and `GET /livestreams/...` as
public, so the new paths needed no gateway change. In the vod service mux,
`GET /v1/vods/search` wins over `GET /v1/vods/{vodId}` because Go's ServeMux
picks the more specific pattern (same as `/v1/vods/author`).

---

## 3. API

None of the search endpoints return `total`; `meta` carries `page` and
`page_size` only (`page: 0` is dropped by `omitempty`). The web client treats a
short page as the end of the list.

| Endpoint | Params | Returns |
|----------|--------|---------|
| `GET /v1/users/search` | `username` (≤ 100 characters), `page` (default 0, max 10000), `limit` (default **10**, 1..20) | `PublicUser[]` |
| `GET /v1/vods/search` | `q` (trimmed, 1..100 characters), `page` (max 10000), `limit` (default 20, max 20) | `VOD[]` — public + `ready` only |
| `GET /v1/livestreams/search` | same as vods | `Livestream[]` — public + live now (`ended_at IS NULL`) only |

Lengths are counted in characters (runes), not bytes. Queries that are not
valid UTF-8 or contain a NUL byte are rejected with an input error (Postgres
would otherwise fail the query with a 500). `page` is capped at 10000 so
`limit * page` cannot overflow.

`page`/`limit` on user search are new and optional; the default of 10 keeps
older callers (e.g. the new-conversation dialog) unchanged. Contract entries
are also in `API_CONTRACT_AUDIT.md` → "Search".

---

## 4. Matching & ranking (backend)

### 4.1 Normalisation

Query and text are compared as `immutable_unaccent(lower(...))`.
`unaccent()` is not `IMMUTABLE`, so it cannot be indexed directly; the
migrations add an `IMMUTABLE` wrapper `immutable_unaccent(text)` that pins the
`public.unaccent` dictionary. **Queries must use exactly the same expression as
the indexes** or the indexes are skipped.

### 4.2 VODs and live streams

A row matches if **any** of:

| # | Rule | Postgres | Example |
|---|------|----------|---------|
| 1 | Title similar overall | `title % q` (trigram similarity ≥ `pg_trgm.similarity_threshold`, default 0.3) | `ranked grnd` → "Ranked grind…" |
| 2 | Query close to part of the title | `q <% title` (word similarity ≥ `pg_trgm.word_similarity_threshold`, default 0.6) | `diamnd` → "…Diamond…" |
| 3 | Title contains query | `title LIKE '%q%'` (wildcards in `q` escaped) | `ia` → "Diamond" |
| 4 | Query close to part of the description | `q <% description` | `boss figth` |
| 5 | Description contains query | `description LIKE '%q%'` | `boss fight` |
| 6 | Streamer name matches | `user_id = ANY($ids)` (§4.3) | `jane` → all of streamer_jane's videos |

Ranking, highest first:

```
GREATEST(
  similarity(title, q),
  word_similarity(q, title),
  0.5 * word_similarity(q, description)   -- description counts half
) + 0.5 if matched via streamer name
then view_count DESC, created_at/started_at DESC, id   -- id keeps paging stable
```

Queries: `backend/vod/repositories/vod/search_public_vods.go`,
`backend/livestream/repositories/livestream/search_live_livestreams.go`.

### 4.3 Streamer-name lookup

`services/vod/search.go` and `services/livestream/search.go` call the user
gateway method `SearchUserIds(ctx, query, 10)` → `GET /v1/users/search`.

- Only for queries of **≥ 3 characters**: the user search tolerates several
  typos, so 1–2 letters would match almost every streamer.
- **2 s timeout.** On timeout or error the search logs one warning and
  continues with title/description rules only (no failure, no fabricated
  results). The gateway method returns wrapped errors and does not log itself.

### 4.4 Channels (user search)

Pre-existing logic, only paging added:
`username % q OR levenshtein(username, q) <= 5`, ordered by similarity, then
edit distance, then `u.id`. Usernames are **not** accent-folded, and
`levenshtein` is case-sensitive. See §8 for the known weakness.

---

## 5. Highlighting (frontend)

`web/utils/highlight.ts` → `findHighlightRanges(text, query)`, rendered by
`web/components/search/highlight.tsx` as `<mark>`. Runs per field (title,
username, description snippet), case- and accent-insensitive (`đ` → `d`).

1. Whole query as a phrase, exact.
2. Each query word, exact substring.
3. A word with no exact hit (≥ 4 letters) highlights words within a small edit
   distance: 1 edit up to 5 letters, 2 up to 9, else 3. A swap of neighbours
   counts as 1 (`jnae` → jane). Prefixes count (`jan` → jane_streams).
4. Single-letter queries only highlight word starts.

When a result matched by description only, the row shows a ~60-char snippet
around the first match (`excerptAroundMatch`).

The highlighter is a heuristic and does **not** mirror Postgres trigram
scoring. Expect occasional mismatches: a streamer-name match has nothing
highlighted in the title, and `fight` highlights "night" (one edit).

---

## 6. Frontend map

| Concern | File |
|---------|------|
| Header search (dropdown, phone icon, mirrors URL on `/search`) | `web/app/(main)/_components/header/search.tsx` |
| Header layering (`z-40`, needed so the dropdown is above page cards) + Suspense | `web/app/(main)/_components/header/header.tsx` |
| Search page + tabs | `web/app/(main)/search/page.tsx`, `_components/search-view.tsx`, `search-all-tab.tsx`, `search-type-tab.tsx`, `search-states.tsx` |
| Result rows (compact/full), row skeleton, type → row map | `web/components/search/search-result-rows.tsx` |
| Search input | `web/components/search/search-input.tsx` |
| Shared constants/helpers (`SEARCH_TYPES`, `SEARCH_TABS`, `searchHref`, `parseSearchTab`, 300 ms debounce, `RAW_INTERPOLATION`) | `web/lib/search.ts` |
| Query hooks (preview limit 5, page size 20, 30 s stale time) | `web/hooks/queries/use-search.ts` |
| URL state `?q=&type=` (debounced writes, survives tab clicks mid-typing, back/forward cancels pending writes) | `web/hooks/use-search-url.ts`, `web/hooks/use-debounced-value.ts` |
| API calls | `web/lib/api/user.ts`, `vod.ts`, `livestream.ts` |
| Strings | `common.json` → `search.*` (en-US, vi-VN) |
| Underlined tab style (shared with channel page) | `web/components/ui/tabs.tsx` → `UNDERLINE_TABS_*` |

The preview queries are shared between the dropdown and the All tab (same query
keys), so opening `/search` from the dropdown does not refetch.

Notes for changes here:

- Strings that interpolate the query pass `RAW_INTERPOLATION`: i18next escapes
  HTML by default and React escapes again (`rock & roll` would show `&amp;`).
- Header layout per breakpoint is chosen with CSS (`md:hidden` / `hidden
  md:flex`), not `useMediaQuery`, so phones don't flash the desktop input
  before hydration.
- The dropdown closes on outside pointer-down, Escape, or keyboard focus
  leaving it. Blur without a `relatedTarget` is ignored on purpose (Safari does
  not focus clicked links, so closing on it would swallow the click).

---

## 7. Mock mode (`npm run dev:mock`)

`web/mocks/search.ts` + handlers in `mocks/handlers/{user,vod,livestream}.ts`.
Mocks decide matches with the **frontend highlighter** (kept on purpose so
typo matching can be demoed without a backend) and rank exact
(accent-insensitive) hits above typo hits, title > streamer > description
within each. Close to the backend, not identical — verify ranking on a real DB.
The `/vods/search` mock must stay above `/vods/:vodId` (MSW matches in order).

---

## 8. Operations, gaps and follow-ups

**Migrations** — `backend/vod/migrations/0004_add_vod_search.sql`,
`backend/livestream/migrations/0006_add_livestream_search.sql`:
`CREATE EXTENSION pg_trgm`, `CREATE EXTENSION unaccent`, the
`immutable_unaccent` wrapper, GIN `gin_trgm_ops` indexes on normalised title
and description; livestream 0006 also adds `idx_livestream_user_id` (without
it the streamer-match branch of the OR forces a sequential scan). Down drops
indexes, function and both extensions (no
`CASCADE`, so it fails loudly if something else starts using them).

| # | Status | Item |
|---|--------|------|
| G1 | **Unverified** | The SQL and migrations were never run (no Postgres in the dev environment when written). Before relying on it: run the migrations and `EXPLAIN` a search on a dev DB to confirm the trigram indexes and the `user_id` index are used. |
| G2 | **Unverified** | `CREATE EXTENSION` needs privileges. Both extensions are "trusted" on Postgres 13+ (DB owner can install), but the production DB user's privileges were not checked. |
| G3 | Known weakness | User search `levenshtein(...) <= 5` is far too loose for short names (`xyz` matches every username of ≤ 5 letters) and also feeds streamer-name matching (§4.3). Suggested fix: length-scaled limit (1 edit up to 5 letters, 2 above) on `lower(username)`. |
| G4 | Known gap | Usernames are not accent-insensitive. |
| G5 | Pre-existing risk | `vods.title` and `livestreams.view_count` are nullable in SQL but non-pointer in Go; a NULL row fails the row scan for search **and** the popular endpoints. |
| G6 | Untested path | Infinite scroll beyond page 1 was not exercised in the UI (mock data has < 20 hits per query). Same sentinel pattern as the home feed. |
| G7 | By decision | No unit tests on either side yet (team wants to review how tests are written first). `utils/highlight.ts` is the first candidate. |
| G8 | Idea | No result count per tab: the endpoints don't return `total` (a `count(*)` over fuzzy matches is expensive). |
| G9 | Idea | Keyboard navigation (↑/↓) in the desktop dropdown (would make it a proper ARIA combobox); recent searches. |
| G10 | Follow-up | Each live/VOD row fetches its streamer with `usePublicUser` (cached per user, but up to 10 requests per dropdown and 20 per page). Better: have the search endpoints return `username`/`profilePicture`, or batch-fetch distinct user ids. The streamer line reserves its height so rows don't jump meanwhile. |
