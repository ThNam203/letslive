# Stories — Detailed Plan

Instagram-style short-lived posts (photo or ≤60s video) visible to followers for 24h with view-tracking, reactions, and DM replies. Design mirrors the existing VOD pipeline + follow graph + eventbus patterns already in this repo so it slots in cleanly rather than inventing new infrastructure.

---

## 1. Scope and phasing

Feature is large enough that it should ship in stages. Each phase is independently shippable.

| Phase | Scope |
|---|---|
| **P0 — MVP** | Post photo story with audience picker (public / followers / specific users / only me); view stories from people you follow; 24h expiry; self-delete; auth-gated media access |
| **P1** | Video stories (transcoded); "who viewed mine" list; push notification when followed creator posts |
| **P2** | Emoji reactions; reply-to-story (creates/extends DM thread); mute a creator's stories |
| **P3** | Highlights (pin past stories to profile); polls/quizzes/question stickers; "close friends" preset lists (reusable audiences) |
| **P4** | Story-archive view for self; cross-post to VOD |

The rest of this plan covers **P0 and P1 in detail**, with P2+ sketched in "Out of scope for MVP."

---

## 2. Architecture

```
                             ┌─────────────────────────────────────────┐
                             │ web / mobile clients                    │
                             │   - bubble strip on home feed           │
                             │   - fullscreen player (tap next/prev)   │
                             │   - capture/upload sheet                │
                             └───────────────┬─────────────────────────┘
                                             │
                                    ┌────────▼────────┐
                                    │      Kong       │ /stories/* (JWT)
                                    └────────┬────────┘
                                             │
                    ┌────────────────────────▼────────────────────────┐
                    │ story service (Go, :7784)                       │
                    │  - POST /stories (multipart) → MinIO + PG row   │
                    │  - GET /stories/feed → followed creators' active│
                    │  - GET /stories/mine → my stories + viewers     │
                    │  - POST /stories/:id/view → mark seen           │
                    │  - DELETE /stories/:id → owner only             │
                    └───┬───────────────┬─────────────┬───────────────┘
                        │               │             │
        raw media ──────▼──┐    pg: stories,   publish ▼
                           │    story_views    kafka letslive.story
                  ┌────────▼──┐               ┌────────────────────┐
                  │   MinIO   │               │ user service       │
                  │ stories/  │               │  (notification     │
                  │  {id}/... │               │   consumer)        │
                  └──────┬────┘               └────────────────────┘
                         │
                         ▼ (video only)
                  ┌──────────┐     writes processed HLS
                  │ transcode│ ──► back to MinIO, PATCHes
                  │  service │     story.playback_url
                  └──────────┘
```

**Service placement:** new `story` service (Go), alongside vod/livestream — not bolted onto an existing service. Stories have distinct access patterns (feed-centric, expiry, view-tracking) that would bloat VOD if co-located. Reuses VOD's MinIO + transcode-job pattern rather than duplicating it.

**Key reuses (no new infra):**
- MinIO for media (already routed through Kong at `/files/`)
- Transcode service for video → HLS (VOD pattern)
- User service's `GetFollowedUserIds()` for feed query
- User service's notification table + handler for push notifications
- Kafka eventbus (`letslive.story` topic) for cross-service events
- Auth: Kong JWT plugin + `ACCESS_TOKEN` cookie (identical to `/vods/upload`)

---

## 3. Data model

New Postgres DB: `story_db`, owned by story service.

```sql
-- migration 0001_stories.sql
CREATE TABLE stories (
    id              UUID PRIMARY KEY,
    author_id       UUID NOT NULL,              -- no FK: cross-service
    media_type      VARCHAR(10) NOT NULL,       -- 'image' | 'video'
    raw_url         TEXT NOT NULL,              -- MinIO object path (not client-facing)
    hls_path        TEXT,                       -- MinIO folder for HLS (video only)
    thumbnail_path  TEXT,                       -- MinIO object path for thumbnail
    duration_ms     INTEGER,                    -- video only
    status          VARCHAR(20) NOT NULL,       -- 'processing' | 'ready' | 'failed'
    caption         VARCHAR(200),               -- optional
    visibility      VARCHAR(20) NOT NULL,       -- 'public' | 'followers' | 'specific' | 'self'
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL        -- created_at + 24h, enforced at insert
);
CREATE INDEX ON stories (author_id, created_at DESC);
CREATE INDEX ON stories (expires_at);          -- for purge job

-- Allowlist for visibility='specific'. Empty for other visibility values.
-- Also used as "included list" UI state regardless of whether the author
-- switches visibility later (preserves intent).
CREATE TABLE story_audiences (
    story_id    UUID NOT NULL,
    user_id     UUID NOT NULL,
    PRIMARY KEY (story_id, user_id)
);
CREATE INDEX ON story_audiences (user_id);     -- "stories I'm specifically allowed in"

CREATE TABLE story_views (
    story_id    UUID NOT NULL,
    viewer_id   UUID NOT NULL,
    viewed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (story_id, viewer_id)
);
CREATE INDEX ON story_views (viewer_id);       -- "which of creator X's stories have I seen"
```

**Visibility semantics:** a viewer `V` may see a story authored by `A` iff one of:
- `V = A` (author always sees own), OR
- `visibility = 'public'`, OR
- `visibility = 'followers'` AND `V` follows `A`, OR
- `visibility = 'specific'` AND `(storyId, V) ∈ story_audiences`.

`visibility = 'self'` restricts to the author only — useful for drafts, archive staging, and "post just to see what it looks like." Enforced identically for metadata queries and media fetches.

**Why store URLs as MinIO paths, not public URLs:** with four visibility modes the client must go through an auth-gated media endpoint on the story service (section 5 + 11). Raw MinIO paths are internal; the API serializer rewrites them to `/v1/stories/:id/media` URLs at response time.

**Expiry strategy:** rows are not deleted at 24h. Queries filter by `expires_at > NOW()`. A nightly purge job (cron in story service) hard-deletes rows + MinIO objects where `expires_at < NOW() - 7d` to reclaim storage but keep a week-long archival window for P3 highlights/archive. Delete cascade handled in-app (stories → story_audiences + story_views), not via FK, so views table outlives the story row briefly during delete — fine.

**Why `expires_at` stored (not computed):** lets you change the lifetime per-story later (e.g., "close friends" stories could be shorter) without migrating code that computes.

---

## 4. REST API

All under `/v1` on story service; Kong route `Story_Routes` at path `/stories` with JWT plugin (protected routes) + a public read route.

| Method | Path | Auth | Body / Query | Purpose |
|---|---|---|---|---|
| `POST` | `/v1/stories` | required | multipart: `file`, `caption?`, `visibility`, `audience_user_ids?`, `duration_ms?` | Upload photo or video story. `visibility` ∈ `public`/`followers`/`specific`/`self`. `audience_user_ids` is a JSON array (≤ 200 entries), required when `visibility='specific'`, ignored otherwise. Returns story in `processing` (video) or `ready` (image) state. |
| `GET` | `/v1/stories/feed` | required | — | Returns `{ author: UserStub, stories: Story[], hasUnseen: bool }[]` for each author the caller follows **and is allowed to see** (visibility filter applied server-side). Sorted: unseen authors first, then by most-recent story. |
| `GET` | `/v1/stories/user/:userId` | optional | — | Active stories for one creator, filtered to those the caller may see. Anonymous caller sees only `public` stories and is never recorded as a viewer. |
| `GET` | `/v1/stories/mine` | required | — | My own active stories with view counts + per-story viewer lists + current audience list (for `specific`). |
| `GET` | `/v1/stories/:id/media` | required* | — | **Auth-gated media proxy.** Streams image bytes or redirects to HLS playlist. Enforces visibility (403 if not authorized). `*`public stories accept anonymous callers. |
| `GET` | `/v1/stories/:id/playlist.m3u8` | required* | — | Video stories: returns HLS playlist with segment URLs rewritten to include a short-lived token bound to `(storyId, viewerId)`. |
| `GET` | `/v1/stories/:id/segment/:name` | token | — | HLS segment endpoint. Validates the token from playlist; streams segment from MinIO. |
| `POST` | `/v1/stories/:id/view` | required | — | Idempotent: record caller as viewer. 204 on success (insert-ignore on conflict). Rejected with 403 if caller can't see the story. |
| `DELETE` | `/v1/stories/:id` | required | — | Owner-only; soft-deletes the story immediately (sets `expires_at = NOW()`). |
| `PATCH` | `/v1/stories/:id` *(internal)* | — | `{ hls_path, thumbnail_path, status, duration_ms? }` | Called by transcode service to mark video ready. Not exposed through Kong. |

**Validation:**
- Max 10 active stories per author (24h window) — same shape as chat-command cap.
- `caption` ≤ 200 chars.
- `visibility` is required; `specific` requires `audience_user_ids.length ≥ 1` and ≤ 200; IDs must be valid UUIDs. Self-ID in the audience is silently dropped (author is always in audience).
- `file`: image ≤ 10MB (`.jpg`/`.png`/`.webp`); video ≤ 50MB, ≤ 60s, (`.mp4`/`.mov`/`.webm`).
- `duration_ms` for video; server verifies against probed file (ffprobe in transcode).

**Response shape (Story):**
```json
{
  "id": "...", "authorId": "...",
  "mediaType": "image" | "video",
  "mediaUrl": "/v1/stories/.../media",               // image
  "playlistUrl": "/v1/stories/.../playlist.m3u8",    // video
  "thumbnailUrl": "/v1/stories/.../media?thumb=1",
  "durationMs": 1234, "caption": "...", "status": "ready",
  "visibility": "specific",
  "audienceUserIds": ["..."],   // only in /mine; omitted in /feed and /user/:id
  "createdAt": "...", "expiresAt": "..."
}
```

---

## 5. Media pipeline

**Images** (happy path, synchronous):
1. Client `POST /v1/stories` multipart with visibility + optional audience.
2. Story service validates mime + size + visibility payload, generates storyId, uploads to MinIO at `stories/{storyId}/raw.{ext}`.
3. Server generates thumbnail inline (resize to 320w) — simple image lib in Go, no job needed — writes to `stories/{storyId}/thumb.jpg`.
4. Insert DB row with `status = 'ready'`, `visibility`, and (if specific) rows into `story_audiences`.
5. Publish `story.created` to Kafka (payload includes `visibility` and `audienceUserIds` so the notification consumer can fan out to the right set).
6. Return 201 with the Story; `mediaUrl`/`thumbnailUrl` are story-service URLs, not raw MinIO paths.

**Videos** (asynchronous, reuses VOD pattern):
1. Same upload to MinIO at `stories/{storyId}/raw.{ext}`; same visibility insert.
2. Insert DB row with `status = 'processing'`.
3. Enqueue transcode job in `transcode_jobs` table (existing infra). Job type `story_hls`.
4. Transcode service picks it up, generates HLS playlist + segments + thumbnail (first frame), writes to `stories/{storyId}/hls/...` and `stories/{storyId}/thumb.jpg`.
5. Transcode calls story service `PATCH /v1/stories/:id` internal endpoint (service-to-service; auth by Consul-discovered network — same trust model as today).
6. Story service sets `status = 'ready'`, `hls_path`, `thumbnail_path`, `duration_ms`, publishes `story.created` to Kafka **only now** (delayed publish — the audience shouldn't be notified of a broken story).

**Why publish on ready, not on upload:** avoids race where a follower opens the bubble before HLS exists.

**Media access path (all four visibility modes):**
- Client calls `GET /v1/stories/:id/media` (image) or `GET /v1/stories/:id/playlist.m3u8` (video).
- Story service: load story → check `can_view(caller, story)` (section 3 semantics) → if denied, 403.
- **Image:** presign a short-lived MinIO URL (5 min TTL) and redirect (302) to it. MinIO already exposes presigned GETs; no new infra.
- **Video playlist:** fetch the stored HLS index from MinIO, rewrite each segment URL to `/v1/stories/:id/segment/<name>?t=<token>`. Token is a signed blob `{sid, vid, exp}` (HS256 with the service's shared secret; re-uses `ACCESS_TOKEN_SECRET` or a dedicated one). TTL 10 minutes — long enough for a 60s video to play through a slow connection but short enough that leaked URLs don't live forever.
- **Video segment:** validate token → 302 to presigned MinIO URL for the segment.

**Why proxy instead of permanent presigned URLs at upload time:** presigned URLs are bearer-equivalent once leaked. With the proxy, we re-check `can_view` every time a new playlist is fetched, which means revoking audience (e.g., unfollow, delete-from-audience) takes effect within one playlist-refresh window instead of being immortal. Proxy overhead for image/video bytes is one MinIO-presign call per access — cheap.

---

## 6. Realtime + notifications

**P0: no realtime.** Clients poll `/feed` when the user opens the app or pulls to refresh.

**P1: push notifications.** Event flow:
```
story service → Kafka letslive.story (type=story.created,
                    payload={storyId, authorId, visibility, audienceUserIds[]})
             → user service notification-consumer
             → resolve target set by visibility:
                 public     → followers of authorId (GetFollowers)
                 followers  → followers of authorId
                 specific   → audienceUserIds from payload
                 self       → [] (no-op, skip)
             → for each target user, INSERT INTO notifications
               (type='story_posted', reference_id=storyId, ...)
             → client polls /user/me/notifications (existing flow) or receives via DM-WS push
```

`public` and `followers` both notify followers — the difference is *who can view*, not *who gets notified*. A public story still shouldn't spam non-followers with notifications.

The "followers-of-X" lookup needs a new user-service method: currently we have `GetFollowedUserIds(userId)` (people X follows), not `GetFollowers(userId)` (people following X). Add repo method + an **internal-only** handler reachable from story service's notification consumer.

**Trade-off:** for a creator with 100k followers, fanning out 100k notification rows is heavy. Mitigation: a `notification_preferences` check (users opt in per-creator), or a fan-out-on-read model. For MVP with small follower counts, direct fan-out is fine. Flag this as a follow-up.

**Live view counts** (real-time "three people are watching now") — P3 feature. Would use the existing chat presence service + a new `/story-ws` endpoint, but not worth the complexity for MVP.

---

## 7. Frontend — Web

### Route structure
- `web/app/[lng]/(main)/page.tsx` (home) — add `<StoryBubbleStrip />` at top.
- `web/app/[lng]/(main)/stories/[userId]/page.tsx` — fullscreen viewer. Route is navigated into, not a modal, so deep-linking + back-button work.
- `web/app/[lng]/(main)/settings/stories/page.tsx` — "My stories and viewers" (P1).

### New files
```
web/types/story.ts                                  Story, StoryFeedEntry, StoryViewer types
web/lib/api/story.ts                                CreateStory, GetStoryFeed, GetUserStories,
                                                    GetMyStories, MarkStoryViewed, DeleteStory
web/components/story/bubble-strip.tsx               horizontal scroll of avatar bubbles
web/components/story/bubble.tsx                     one avatar with ring (unseen=gradient, seen=gray)
web/components/story/viewer.tsx                     fullscreen; tap zones, progress bars, hold-to-pause
web/components/story/viewer-progress.tsx            top progress bars (one per story in the set)
web/components/story/composer.tsx                   upload dialog: file picker + caption input
web/components/story/viewer-list.tsx                P1: who-viewed-mine list
web/lib/i18n/locales/{en,vi}/stories.json           new namespace
web/app/[lng]/(main)/stories/[userId]/page.tsx      viewer route
web/app/[lng]/(main)/settings/stories/page.tsx      P1
```

### Composer entry point
Floating "+ Story" button on home (desktop) or bottom-nav action (mobile web). Opens `<StoryComposer />` modal. After upload returns, optimistically inserts bubble at the front of the strip as "processing" (grayed) until feed refresh reveals `status=ready`.

### Composer — visibility picker
Inside the composer, below the caption, a segmented control with four options:

| Option | Meaning | Secondary UI |
|---|---|---|
| 🌐 **Public** | Anyone can view | — |
| 👥 **Followers** (default) | Only accounts that follow me | — |
| 🧑‍🤝‍🧑 **Specific users** | A chosen allowlist | Opens a searchable user-picker (reuses existing user-search endpoint); chips for selected users; ≤ 200 allowed |
| 🔒 **Only me** | Just me — useful for drafts / testing look | Small note: "Won't notify anyone" |

Default = **Followers**. Selection persists per-session (localStorage), not per-user (to avoid accidentally posting to the wrong audience after switching).

When visibility is `specific`, the selected userIds go into the multipart form as a JSON-encoded `audience_user_ids` field. The viewer-bar of the viewer UI (section below) shows a small icon indicating audience scope so the author can tell at a glance.

### Viewer UX (matches Instagram)
- Top progress bars: one slot per story in current author's set; active bar fills over `duration_ms` (video) or 5s (image).
- Tap left/right third: prev/next story.
- Tap center: open reply input (P2).
- Long-press: pause progress.
- Swipe down/Esc: dismiss.
- On story-visible-for-≥500ms: fire `POST /v1/stories/:id/view`.
- At end of author's set: advance to next author's bubble. At end of feed: close.

### Media rendering
- Image: `<img src={playback_url} />` with object-fit cover.
- Video: `<video src={playback_url} autoplay muted playsInline />` — if HLS, wrap with `hls.js` (already a dep for VOD player — reuse).

---

## 8. Frontend — Mobile (Flutter)

### Route
- `AppRoutes.storyViewer(userId)` added to `app_router.dart`.
- Home screen (`features/home/presentation/home_screen.dart`) gains `StoryBubbleStrip` widget at top.

### New files
```
mobile/lib/models/story.dart                                 Story, StoryFeedEntry, StoryViewer
mobile/lib/features/stories/data/story_repository.dart       mirrors web story.ts API surface
mobile/lib/features/stories/presentation/
    story_bubble_strip.dart
    story_bubble.dart
    story_viewer_screen.dart                                 GestureDetector zones + AnimationController
    story_viewer_progress.dart
    story_composer_sheet.dart
mobile/lib/features/stories/settings/story_viewers_screen.dart   P1
mobile/lib/providers.dart                                    + storyRepositoryProvider
mobile/lib/core/network/api_endpoints.dart                   + stories* endpoints
mobile/lib/core/router/app_router.dart                       + storyViewer route
mobile/l10n/app_{en,vi}.arb                                  stories* + settingsNavStories keys
```

### Video playback
Reuse existing `video_player` + HLS adapter from VOD player feature. For images, `CachedNetworkImage`.

### Upload
`image_picker` (already a dep) provides camera + gallery. Dio multipart upload through existing `ApiClient.upload(...)`.

### Composer — visibility picker (mobile)
Bottom-sheet over the composer with a `FSelect`-style list of four options (public / followers / specific / self), mirroring the web UX. Selecting **Specific users** pushes a user-search screen (reuses the existing search UX from DM "new conversation"). Selected userIds are held in the composer's Riverpod notifier and POSTed as a JSON-encoded `audience_user_ids` multipart field alongside the file.

---

## 9. Kafka event definitions

Add to `backend/shared/pkg/eventbus/events/story.go`:

```go
const (
    EventStoryCreated = "story.created"
    EventStoryDeleted = "story.deleted"
    EventStoryViewed  = "story.viewed"  // P3 — currently not emitted; views stay within story DB
)

type StoryCreatedPayload struct {
    StoryID   string    `json:"storyId"`
    AuthorID  string    `json:"authorId"`
    MediaType string    `json:"mediaType"`
    CreatedAt time.Time `json:"createdAt"`
}
```

Topic: `letslive.story` (add to `topics.go`, 3 partitions, key by `authorId` so one author's events stay in-order).

---

## 10. i18n

New namespace `stories` (web) / `stories*` keys (mobile), e.g.:

```
composer.title, composer.caption_placeholder, composer.post, composer.uploading,
viewer.replied, viewer.seen_by_count, viewer.delete_confirm,
bubble.live_now_badge, bubble.your_story, bubble.add,
errors.file_too_large, errors.video_too_long, errors.unsupported_format,
notification.new_story_title, notification.new_story_body
```

Both `en` and `vi` from day one — matches existing convention.

---

## 11. Security and abuse

### Access control — visibility enforcement

A single `can_view(callerId, story)` predicate gates every metadata and media endpoint. Implemented once in the story service; reused from every handler.

```
can_view(caller, story) =
    caller == story.authorId                              -- author always
 OR story.visibility == 'public'                          -- anyone
 OR (story.visibility == 'followers'
     AND caller != null
     AND follow_repo.Exists(caller, story.authorId))      -- follower check
 OR (story.visibility == 'specific'
     AND caller != null
     AND audience_repo.Contains(story.id, caller))
 -- 'self' falls through to false unless caller == author
```

The follower check is a synchronous gRPC/HTTP call into the user service (or a cached denormalized follower-set — P1 optimization; don't bother for MVP). Rate-limit story-service → user-service with a short TTL (30s) cache per `(viewer, author)` since the feed page hammers this predicate once per story.

### Media gate

All media (image bytes, HLS playlist, HLS segments, thumbnails) flow through the story service, not direct MinIO URLs. Guarantees that a leaked media URL stops working when:
- the author changes visibility,
- the author removes a user from the audience,
- the user unfollows the author,
- the story expires or is deleted.

Segment tokens are signed with a 10-minute expiry to bound replay.

### Other

- **Rate limit:** 10 stories/24h per author at service level (same as chat-command `MAX_PER_OWNER_PER_SCOPE`).
- **Audience size cap:** ≤ 200 userIds per `specific` story. (Anything larger is really "followers" with extra steps.)
- **File validation:** mime sniff server-side, not just extension.
- **Size caps:** enforced at multipart parser before MinIO write.
- **Auth on view-record:** `POST /:id/view` requires auth AND `can_view`; anonymous callers of a public story fetch media but never land in the viewer list.
- **Delete path:** owner-only guard exactly like chat-command delete (`story.authorId !== userId → 403`).
- **No report/block yet** — deliberate: ship P0 without moderation hooks, add in P3 once volume warrants it.

---

## 12. Deliberately out of scope for MVP

| Item | Why deferred |
|---|---|
| Close-friends list | Requires new relationship model + UI; P3 |
| Story reactions (heart/fire/etc.) | New table + UI; P2 |
| Reply-to-story → DM | Requires DM thread integration; P2 |
| Highlights (pin to profile) | Requires "move to permanent storage" + profile UI; P3 |
| Polls/quizzes/question stickers | Large UX surface; P3 |
| Cross-post to VOD | Needs a pipeline hand-off contract; P4 |
| "Close friends" preset audience list | Reusable named audience (save-a-group); composer accepts only ad-hoc per-story lists at MVP; P3 |
| Edit visibility/audience after posting | Authors must delete + repost. Keeps media-gate semantics simple; P2 |
| Live viewer count ("3 watching now") | Needs WebSocket + presence per-story; P3 |
| Moderation/reporting | Same stance as chat-command: small scale = acceptable |
| Mute a creator's stories | One-line preference table; P2 |

---

## 13. Testing checklist

Manual test matrix once P0 is done:

- [ ] Upload image story → appears in my bubble → visible to a follower account within poll interval
- [ ] Upload video story → status transitions `processing → ready` → playback works on web and mobile
- [ ] Open viewer → auto-advances through all my stories → closes at end
- [ ] Tap right/left → next/prev navigation
- [ ] Viewer records `POST /view` only once per (story, viewer)
- [ ] Anonymous user can view a public account's story but view count doesn't increment
- [ ] Delete my story → disappears for other viewers on next poll
- [ ] Story older than 24h stops appearing in feeds
- [ ] 10-stories-per-24h cap rejects the 11th
- [ ] Switch `lng` cookie → all story UI re-renders translated
- [ ] `docker compose up` clean bring-up after new migration runs

---

## 14. Open questions (decide before implementation)

These are the call-out items. Pick each before P0 starts.

1. **Video max duration.** Instagram is 60s per slide. Keep 60s? Shorter for storage?
2. ~~Who can view~~ — **resolved**: four-mode audience picker (public / followers / specific / self) with auth-gated media proxy. See sections 3, 5, 11.
3. **Fan-out strategy for notifications.** Direct-insert per follower (simple, won't scale past a few thousand followers) vs fan-out-on-read. Going with direct-insert for MVP okay?
4. **Storage cleanup interval.** I've sketched "hard-delete rows + MinIO objects at 7d post-expiry." Sound good, or shorter (24h = no archive at all) / longer (30d to enable memories)?
5. **Mobile capture UX.** In-app camera (like Instagram — complex, requires camera plugin work) vs just `image_picker` sheet (trivial). I assumed the latter.
6. **Default visibility.** Followers assumed. Should we instead default to the *last* visibility the user selected (per-session) to reduce accidents? I've sketched localStorage persistence in §7; is that acceptable?
7. **Public stories and non-follower notifications.** A public story notifies only the author's followers (section 6). Is that right, or should public-only broadcast more widely (e.g., recommendation feed, trending)?

---

## 15. Rough file-count + scope

- Backend: ~17 files (new service boilerplate + handlers/services/domains/migrations/kafka + media-gate handler + audience repo + signed-token utility).
- Web: ~14 new files + 3 edits (home, layout, settings nav). Adds user-picker reuse for audience selection.
- Mobile: ~12 new files + 4 edits (router, providers, l10n, api_endpoints). Adds audience-search screen.
- Kong + docker-compose: ~2 edits.
- User service: +1 repo method (`GetFollowers`) + internal handler.
- i18n: 2 new JSONs + 2 `.arb` diffs (with visibility labels × 4).

**Estimate for P0:** ~2.5 weeks of focused work for one engineer. The media-gate + HLS segment token work is the main addition over the initial plan; video-ready → publish-notification chain still touches four services. P1 adds roughly 3–4 days.
