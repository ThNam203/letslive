## Channel Badges & Channel Points — Detailed Plan

Twitch-style channel loyalty primitives. Every channel (streamer) gets its own economy:

- **Channel points** — an integer scoreboard earned by a *viewer* in a *specific* channel. Watching, chatting, following, donating all feed points. Viewers can redeem them for channel-scoped rewards.
- **Channel badges** — little icons rendered next to the viewer's username in that channel's chat (and only that channel's chat). Streamers define the badges. Badges are unlocked by crossing thresholds defined by the streamer.

Both systems sit on a shared **rule engine**: rules are data rows, not code. A rule says *"when event X happens, grant Y points and/or check for badge Z unlock."* New event sources plug in by emitting Kafka events — no service changes needed to add a new earn path (watch-time, comments, donations, follows, subs, poll-voting, whatever).

Design mirrors the existing chat-command / stories / finance patterns so this slots in rather than inventing new infrastructure.

---

## 1. Scope and phasing

Feature is large. Each phase is independently shippable.

| Phase | Scope |
|---|---|
| **P0 — MVP** | New `loyalty` service. Channel points balance per (channel, viewer). Two built-in earn rules: watch-time (per minute), chat-message-sent (per message). Streamer settings page to tune rates. Chat renders a points chip next to own username. |
| **P1** | Channel badges: streamer-defined badges with threshold triggers (lifetime points, watch-time, follow-duration, donation-lifetime). Badges render next to username for *all* viewers in that channel's chat. |
| **P2** | Redemptions: streamer defines channel-point rewards (highlight-my-message, unlock sub-only emote, custom text prompt). Queue + approve/deny UI for the streamer. |
| **P3** | More earn sources: donation-made (finance), follow-started (user), subscription-started (future), predict-voted, poll-voted. Multiplier rules (2× for subscribers, 1.5× during raids). Rule conditions (only while channel is live, only first N per stream, etc.). |
| **P4** | Leaderboards (top channel-point holders per channel), badge carousel on the channel profile page, import/export rule preset, streamer-to-streamer badge sharing. |

This plan covers **P0 + P1 in detail**, with P2+ sketched in "Out of scope for MVP" and "Future".

---

## 2. Architecture

```
     ┌──────────────────────────────────────────────────────────────────────┐
     │ Event producers (already exist)                                      │
     │                                                                      │
     │  chat service    livestream svc   finance svc   user svc             │
     │  (messages)      (viewer heart-   (donations)   (follows)            │
     │                   beat, 15s)                                         │
     └───┬──────────────────┬──────────────────┬─────────────────┬──────────┘
         │                  │                  │                 │
         ▼                  ▼                  ▼                 ▼
    letslive.chat     letslive.livestream  letslive.finance  letslive.user
         │                  │                  │                 │
         └──────────────────┴─────────┬────────┴─────────────────┘
                                      │ Kafka
                                      ▼
                        ┌──────────────────────────────┐
                        │ loyalty service (Go, :7785)  │
                        │                              │
                        │  - rule evaluator            │
                        │  - points ledger writer      │
                        │  - badge-unlock checker      │
                        │                              │
                        │  REST:                       │
                        │   GET  /loyalty/:chan/me     │
                        │   GET  /loyalty/:chan/top    │
                        │   GET  /loyalty/:chan/badges │
                        │   POST /loyalty/:chan/rules  │  (streamer only)
                        │   POST /loyalty/:chan/badges │  (streamer only)
                        └───┬─────────────────┬────────┘
                            │                 │
                 ┌──────────▼────┐   ┌────────▼────────────┐
                 │ loyalty_db    │   │ publishes           │
                 │  (Postgres)   │   │ letslive.loyalty    │
                 │               │   │  badge.unlocked     │
                 │ accounts      │   │  points.awarded     │
                 │ ledger        │   │                     │
                 │ rules         │   └────────┬────────────┘
                 │ badges        │            │
                 │ user_badges   │            ▼
                 └───────────────┘    chat service caches
                                      (channel, user) →
                                      {points, badges}
                                      for renderer
```

**Service placement:** new `loyalty` service (Go), alongside `finance`. Reuses finance's double-entry ledger pattern because that's *exactly* what channel points are: an append-only, per-account balance with strict auditability. A viewer who loses points because of a bug is a worse experience than a broken badge — treat points like currency.

**Why a new service instead of folding into finance:**
- Finance holds real value (SPARK, FLARE — backed by payments). Channel points are per-channel scoreboards that have no exchange rate and can't be transferred cross-channel. Mixing them into `accounts` would pollute finance's invariants.
- Loyalty has heavy write amplification from watch-time heartbeats. Finance is synchronous, audit-critical, and low-throughput per transaction. Different tuning profiles.
- Rules engine + badge storage + per-channel config are loyalty-specific surface area that has nothing to do with payments.

**Reuses:**
- Kafka eventbus (new `letslive.loyalty` topic + consumer on existing `letslive.*` topics)
- Kong JWT plugin (same cookie-based `ACCESS_TOKEN` as finance/stories)
- Postgres + goose migrations (same layout as finance)
- Chat service's existing "enrich outgoing message" path (section 7)
- Follow-existence check: user service `GetFollowedUserIds`

---

## 3. The rule engine (core design)

The brief was: *"design system should be flexible as possible (like how long you watch, how many comments, how much did you donate, etc.)"*. So rules are rows, not code.

### Event types (all come off Kafka)

A rule fires when a matching event arrives. Event types the evaluator understands at P0 + P1:

| Event type | Source topic / event | Key fields in payload |
|---|---|---|
| `watch_seconds` | `letslive.livestream` / `viewer.heartbeat` (new) | channelId, viewerId, seconds (≤ 60) |
| `chat_message_sent` | `letslive.chat` / `chat.message_sent` (new) | channelId, viewerId, text length |
| `follow_started` | `letslive.user` / `user.follow_started` | channelId (= followeeId), viewerId |
| `donation_made` | `letslive.finance` / `finance.donation_sent` | channelId (= receiverId), viewerId (= senderId), amount, currency |
| `subscription_started` *(future)* | `letslive.user` | channelId, viewerId, tier |

**Principle:** the loyalty service *never* polls. It subscribes to topics already produced by the owning service. To add a new earn path you (1) make the owning service publish the event (one-liner with the existing eventbus util), (2) add a handler in the rule evaluator to map payload → `(channelId, viewerId, magnitude)`. No loyalty schema change.

Two events don't exist yet and need to be added to their owning service as a prerequisite:
- Chat service → publish `chat.message_sent` on every non-command, non-DM chat line. Payload: `{channelId, viewerId, textLen, sentAt}`. Throttle at producer: max 1/sec per (channel, viewer) to neutralize spam. See section 9.
- Livestream service → publish `viewer.heartbeat` every 15s for each active WebSocket viewer session. Payload: `{channelId, viewerId, seconds=15}`. Emitted by the livestream service's existing presence tracker.

### Rule shape

```
rule {
  id, channelId,
  eventType,           // one of the above
  unit,                // 'per_magnitude' | 'per_event' | 'first_in_window'
  magnitudePerPoint,   // e.g. "1 point per 60 watch_seconds" → magnitudePerPoint=60
  pointsPerUnit,       // e.g. award 2 points per unit → pointsPerUnit=2
  cooldownSec,         // min seconds between awards for same (channel, viewer)
  dailyCap,            // max points from this rule / (channel, viewer) / UTC day
  conditions,          // JSONB: { liveOnly: true, minTextLen: 2, subscriberMult: 2.0, ... }
  isActive,
  createdAt, updatedAt
}
```

Examples of how rules express the brief's four hints:

| "How long you watch" | `eventType=watch_seconds, unit=per_magnitude, magnitudePerPoint=60, pointsPerUnit=1` |
| "How many comments" | `eventType=chat_message_sent, unit=per_event, pointsPerUnit=5, cooldownSec=10, dailyCap=500` |
| "Follow the channel" | `eventType=follow_started, unit=per_event, pointsPerUnit=100` (one-time; cooldownSec=infinite via idempotency key) |
| "How much did you donate" | `eventType=donation_made, unit=per_magnitude, magnitudePerPoint=100 (cents), pointsPerUnit=10` (i.e. $1 = 10 pts) |

**Idempotency.** Each incoming event carries a stable event-id (Kafka key + offset, or the source's own UUID). The evaluator writes `(rule_id, event_id)` into a `rule_applied_events` dedupe table before awarding. Duplicate events (Kafka rebalance, replay) are no-ops.

**Default ruleset on channel creation.** When a user becomes a streamer for the first time (first livestream, or first time they visit the loyalty settings page), the service seeds **three default rules** they can turn off or re-tune: watch (1/min), chat (5/msg, 10s cooldown), follow (100 once). Same pattern chat-commands uses for built-ins — it's on-ramp-friendly.

### Badge-unlock trigger

A badge can be unlocked in two ways:

1. **Threshold on a lifetime aggregate** (the common case). Examples: 100 lifetime points, 10 hours watch, $50 lifetime donated. Modeled as `{ metric, gte }` where `metric` is a precomputed per-(channel, viewer) counter we maintain (section 4).
2. **Event-driven** (rare but useful). Example: "awarded to anyone who followed on debut stream day" — a one-shot event rule with an attached `grants_badge_id`.

Both flow through the same unlock-evaluator: on every ledger write, re-check *only* the badges whose trigger metric changed. Keeps evaluation O(rules-touching-this-metric), not O(all-badges).

---

## 4. Data model

New Postgres DB: `loyalty_db`, owned by loyalty service.

```sql
-- migration 0001_loyalty.sql

CREATE TYPE loyalty_rule_unit_enum AS ENUM ('per_magnitude', 'per_event', 'first_in_window');
CREATE TYPE loyalty_event_type_enum AS ENUM (
    'watch_seconds', 'chat_message_sent', 'follow_started',
    'donation_made', 'subscription_started'
);

-- Per-channel config (seeded on first stream).
CREATE TABLE channel_settings (
    channel_id           UUID PRIMARY KEY,     -- = streamer userId
    points_name          TEXT NOT NULL DEFAULT 'Points',   -- "Bits", "Stars", whatever the streamer wants
    points_icon_url      TEXT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Rules: the flexible engine.
CREATE TABLE loyalty_rules (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id           UUID NOT NULL,
    event_type           loyalty_event_type_enum NOT NULL,
    unit                 loyalty_rule_unit_enum NOT NULL,
    magnitude_per_point  BIGINT NOT NULL DEFAULT 1,
    points_per_unit      INTEGER NOT NULL,
    cooldown_sec         INTEGER NOT NULL DEFAULT 0,
    daily_cap            INTEGER NULL,
    conditions           JSONB NOT NULL DEFAULT '{}',
    is_active            BOOLEAN NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON loyalty_rules (channel_id, event_type) WHERE is_active;

-- Ledger: append-only log of every points award.
-- Mirrors finance.ledger_entries — same "no updates, no deletes" triggers.
CREATE TABLE points_ledger (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id           UUID NOT NULL,
    viewer_id            UUID NOT NULL,
    rule_id              UUID NULL REFERENCES loyalty_rules(id),
    delta                INTEGER NOT NULL,      -- positive = earn; negative = redemption
    reason               TEXT NOT NULL,         -- 'rule:watch', 'redeem:highlight', 'admin_adjust'
    metadata             JSONB NOT NULL DEFAULT '{}',
    event_dedupe_key     TEXT NULL,             -- composite of source + event id; see rule_applied_events
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON points_ledger (channel_id, viewer_id, created_at DESC);
CREATE INDEX ON points_ledger (channel_id, created_at DESC);  -- channel leaderboard

-- Fast-path balance cache. Updated in the same transaction as the ledger insert.
CREATE TABLE points_balance (
    channel_id           UUID NOT NULL,
    viewer_id            UUID NOT NULL,
    balance              BIGINT NOT NULL DEFAULT 0,
    lifetime_earned      BIGINT NOT NULL DEFAULT 0,   -- monotonic, used by badge-unlock thresholds
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (channel_id, viewer_id)
);

-- Per-(channel, viewer) counters maintained by the evaluator for threshold badges
-- that can't cheaply read off the ledger (e.g. "10 hours watched" ≠ "10 * 60 points earned
-- from watch" once multipliers exist).
CREATE TABLE viewer_channel_metrics (
    channel_id           UUID NOT NULL,
    viewer_id            UUID NOT NULL,
    watch_seconds_total  BIGINT NOT NULL DEFAULT 0,
    chat_messages_total  INTEGER NOT NULL DEFAULT 0,
    donation_cents_total BIGINT NOT NULL DEFAULT 0,
    followed_since       TIMESTAMPTZ NULL,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (channel_id, viewer_id)
);

-- Cooldown / dailycap book-keeping (compact). One row per (channel, viewer, rule).
CREATE TABLE rule_cooldowns (
    rule_id              UUID NOT NULL,
    viewer_id            UUID NOT NULL,
    last_awarded_at      TIMESTAMPTZ NOT NULL,
    today_points         INTEGER NOT NULL DEFAULT 0,
    today_bucket_utc     DATE NOT NULL,
    PRIMARY KEY (rule_id, viewer_id)
);

-- Idempotency: we never apply the same source event twice.
CREATE TABLE rule_applied_events (
    rule_id              UUID NOT NULL,
    event_dedupe_key     TEXT NOT NULL,
    applied_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (rule_id, event_dedupe_key)
);

-- Badges (P1).
CREATE TYPE loyalty_badge_trigger_metric_enum AS ENUM (
    'lifetime_points', 'watch_seconds_total', 'chat_messages_total',
    'donation_cents_total', 'follow_age_days'
);

CREATE TABLE channel_badges (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id           UUID NOT NULL,
    name                 VARCHAR(32) NOT NULL,
    image_url            TEXT NOT NULL,                -- uploaded to MinIO, routed like other assets
    tier                 INTEGER NOT NULL,             -- sorts badges within the same metric
    trigger_metric       loyalty_badge_trigger_metric_enum NOT NULL,
    trigger_gte          BIGINT NOT NULL,              -- unlocks when metric >= this
    is_active            BOOLEAN NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX ON channel_badges (channel_id, name);
CREATE INDEX ON channel_badges (channel_id, trigger_metric, trigger_gte);  -- unlock scan

-- Once unlocked, a badge stays unlocked even if the metric drops (spent-points case).
CREATE TABLE viewer_channel_badges (
    channel_id           UUID NOT NULL,
    viewer_id            UUID NOT NULL,
    badge_id             UUID NOT NULL REFERENCES channel_badges(id) ON DELETE CASCADE,
    unlocked_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (channel_id, viewer_id, badge_id)
);
CREATE INDEX ON viewer_channel_badges (channel_id, viewer_id);   -- "what badges does V have in C"
```

**Caps:**
- 20 active rules per channel.
- 30 badges per channel.
- Rule `points_per_unit` ≤ 100; `daily_cap` ≤ 100k; points integer fits comfortably in BIGINT.

**Why a ledger *and* a balance cache:** reads are dominated by chat-render lookups (`GET balance for (channel, viewer)`) which must be p99 < 5ms. Writes re-aggregate would be O(ledger), hence the materialized `points_balance`. Updated in the same transaction so the two are never out of sync. This is finance's exact trick.

**Why `lifetime_earned` separate from `balance`:** redemptions (P2) subtract from `balance` but must not un-unlock badges. Badges threshold on `lifetime_earned`, which only grows.

---

## 5. REST API

All under `/v1` on loyalty service; Kong route `Loyalty_Routes` at `/loyalty` with JWT plugin (protected routes) + public read routes for the chat's badge lookups.

| Method | Path | Auth | Body / Query | Purpose |
|---|---|---|---|---|
| `GET` | `/v1/loyalty/:channelId/me` | required | — | My balance, lifetime earned, badges held in this channel. |
| `GET` | `/v1/loyalty/:channelId/top` | optional | `limit` (≤ 100) | Leaderboard. Returns viewer stubs + balance. Public. |
| `GET` | `/v1/loyalty/:channelId/badges` | optional | — | All badges defined in this channel (catalog). Public. |
| `GET` | `/v1/loyalty/:channelId/viewer/:viewerId/badges` | optional | — | Badges held by viewer in channel. Public (chat renderer uses this). |
| `GET` | `/v1/loyalty/:channelId/settings` | optional | — | points_name, icon, active rules, badges. Public (chat panel reads display name). |
| `POST` | `/v1/loyalty/:channelId/settings` | streamer | `{ pointsName, pointsIconUrl? }` | Channel owner sets display name + icon. |
| `POST` | `/v1/loyalty/:channelId/rules` | streamer | rule body | Create rule. Cap 20. |
| `PATCH` | `/v1/loyalty/rules/:ruleId` | streamer | partial | Update rule; channelId immutable. |
| `DELETE` | `/v1/loyalty/rules/:ruleId` | streamer | — | Soft-delete (set `is_active=false`). |
| `POST` | `/v1/loyalty/:channelId/badges` | streamer | multipart `image`, body | Create badge (image uploaded to MinIO). Cap 30. |
| `PATCH` | `/v1/loyalty/badges/:badgeId` | streamer | partial | Update badge. |
| `DELETE` | `/v1/loyalty/badges/:badgeId` | streamer | — | Delete badge (cascade revokes unlocked copies). |
| `POST` | `/v1/loyalty/:channelId/adjust` *(internal)* | streamer or admin | `{ viewerId, delta, reason }` | Admin grant/revoke points. Shows up in ledger with `reason='admin_adjust'`. |
| `POST` | `/v1/loyalty/:channelId/bulk-lookup` | — | `{ viewerIds: [] }` (≤100) | Chat's render path: batch `(channel, viewer) → {balance, badges[]}`. Cached 30s in chat. |

**Streamer auth check.** `:channelId === currentUserId` for any write route. 403 otherwise. Same pattern as chat-command ownership.

**Response — `GET /me`:**
```json
{
  "channelId": "...",
  "pointsName": "Sparks",
  "balance": 1450,
  "lifetimeEarned": 2800,
  "badges": [
    { "id": "...", "name": "regular", "imageUrl": "...", "tier": 1, "unlockedAt": "..." }
  ],
  "nextBadge": {
    "id": "...", "name": "veteran", "tier": 2, "metric": "lifetime_points",
    "currentValue": 2800, "requiredValue": 5000
  }
}
```
`nextBadge` is a UX nicety — shows what the viewer is progressing toward.

---

## 6. Event-to-points pipeline

### Consumer layout

One Kafka consumer group `loyalty-evaluator` with N subscribers. Each handler maps a source event → the rule-engine input.

```
letslive.chat          → handleChatMessageSent
letslive.livestream    → handleViewerHeartbeat
letslive.finance       → handleDonationSent
letslive.user          → handleFollowStarted, handleSubscriptionStarted(future)
```

### Evaluator path (single event)

```
on event E:
  normalize E → (channelId, viewerId, eventType, magnitude, dedupeKey)
  rules = SELECT * FROM loyalty_rules
          WHERE channel_id=$C AND event_type=$T AND is_active;

  for each r in rules:
    BEGIN TX (per-rule, serializable)
      INSERT INTO rule_applied_events (r.id, dedupeKey)
        ON CONFLICT DO NOTHING RETURNING *;   -- idempotent gate
      if no row inserted: break (already applied)

      cooldown = SELECT * FROM rule_cooldowns WHERE rule_id=r.id AND viewer_id=V;
      if cooldown and now - cooldown.last_awarded_at < r.cooldown_sec: break

      pts = calculate_points(r, magnitude, E.conditions)    -- applies conditions/multipliers
      if r.daily_cap and cooldown.today_points + pts > r.daily_cap:
          pts = clamp(r.daily_cap - cooldown.today_points, 0)
      if pts == 0: break

      INSERT INTO points_ledger (channel, viewer, rule, delta=pts, reason='rule:'+event_type, ...);
      UPSERT points_balance (channel, viewer) SET balance += pts, lifetime_earned += pts;
      UPSERT viewer_channel_metrics SET <metric for this event> += magnitude;
      UPSERT rule_cooldowns SET last_awarded_at=now, today_points += pts;

      candidates = SELECT * FROM channel_badges
                   WHERE channel=$C AND trigger_metric IN (metrics affected by this event)
                     AND is_active;
      for each b in candidates:
        if viewer's metric ≥ b.trigger_gte AND not already unlocked:
          INSERT INTO viewer_channel_badges (channel, viewer, badge, unlocked_at=now);
          emit loyalty.badge_unlocked → letslive.loyalty Kafka;
    COMMIT
```

**Throughput target.** Watch-heartbeats dominate: 10k concurrent viewers × 1 heartbeat / 15s = ~670/s steady state. Evaluator per-event work is ~3 indexed upserts + 1 bounded badge scan. Comfortably handled by a single replica; horizontal scaling by `channelId` partition key.

**Failure handling.** Evaluator is at-least-once. Idempotency table prevents double-grants. If a crash happens between inserting into `rule_applied_events` and the ledger commit, the whole TX rolls back (same TX) — the dedupe row is rolled back too, so retry re-applies cleanly.

### Outgoing events

```go
const (
    EventPointsAwarded   = "loyalty.points_awarded"
    EventBadgeUnlocked   = "loyalty.badge_unlocked"
    EventPointsRedeemed  = "loyalty.points_redeemed"  // P2
)

type PointsAwardedEvent struct {
    ChannelId  uuid.UUID `json:"channelId"`
    ViewerId   uuid.UUID `json:"viewerId"`
    RuleId     uuid.UUID `json:"ruleId"`
    Delta      int32     `json:"delta"`
    NewBalance int64     `json:"newBalance"`
    EventType  string    `json:"eventType"`
}

type BadgeUnlockedEvent struct {
    ChannelId  uuid.UUID `json:"channelId"`
    ViewerId   uuid.UUID `json:"viewerId"`
    BadgeId    uuid.UUID `json:"badgeId"`
    BadgeName  string    `json:"badgeName"`
}
```

Topic: `letslive.loyalty` (add to `topics.go`, 3 partitions, key by `channelId` so per-channel ordering holds).

The notification service (user svc) consumes `badge.unlocked` for push notifications (P1): "You just unlocked the Regular badge in coolguy's channel!".

---

## 7. Chat integration (the "badge next to username" path)

This is the most visible piece. It has to be fast — chat renders on every keystroke a remote viewer types.

### Enrichment path

```
sender types → chatServer.ts receives WS message
             ↓
       loyaltyCache.get(channelId, viewerId)
       (30s TTL local Map; falls back to REST /bulk-lookup)
             ↓
       augment outgoing message JSON:
         { ..., loyalty: { balance: 1450, badges: [{ name, imageUrl, tier }] } }
             ↓
       publish to Redis → other viewers receive enriched message
```

**Caching strategy.** Chat service holds an in-process LRU keyed `(channelId, viewerId)` with 30s TTL. On miss, batch `POST /v1/loyalty/:channelId/bulk-lookup` with up to 100 viewerIds coalesced over a 50ms window. Same pattern used for user-stub fetches in DM.

**Invalidation.** Loyalty service emits `loyalty.badge_unlocked` and `loyalty.points_awarded` on `letslive.loyalty`. Chat service subscribes and drops matching cache entries. The next message from that viewer re-fetches fresh.

### Client-side rendering

Frontend renders the badges inline before the username:

```
[🎖️][⭐] coolguy: hello chat          ← badges left of name
1450 pts                                ← (optional) points chip to the right
```

Web: a small `<BadgeStrip>` component that just iterates `message.loyalty?.badges` and renders `<img src={badge.imageUrl} width=16 height=16>` each. Tier sorts ascending.

The sender's own UI shows the points chip with a "+5" micro-animation when a `loyalty.points_awarded` event targets them (subscribed via the existing DM-WebSocket presence channel or a new SSE — P2 optimization).

### What about anonymous viewers sending chat?

Anonymous viewers don't have a `viewerId`, so loyalty lookups are skipped. No badges, no points chip. Exactly the chat-command anonymous-support pattern.

---

## 8. Frontend — Web

### Route structure

- `web/app/[lng]/(main)/users/[userId]/chat.tsx` — integrate `<BadgeStrip>` and points chip into the chat row renderer.
- `web/app/[lng]/(main)/settings/channel-loyalty/page.tsx` — streamer's loyalty config: points name, icon, rules list with edit-in-place, badges grid with image upload.
- `web/app/[lng]/(main)/users/[userId]/loyalty/page.tsx` — viewer-facing channel scoreboard: my balance, progress to next badge, leaderboard (top 100), badge catalog.
- `web/app/[lng]/(main)/settings/layout.tsx` — add nav entry "Channel Loyalty" (streamer-only; hidden for viewers).

### New files

```
web/types/loyalty.ts                          LoyaltyRule, ChannelBadge, ViewerLoyalty, etc.
web/lib/api/loyalty.ts                        Fetch wrappers for all REST routes.
web/components/loyalty/badge-strip.tsx        Renders badges next to a username.
web/components/loyalty/points-chip.tsx        Number + icon with +N animation on award.
web/components/loyalty/rule-form.tsx          Streamer: dropdown of event types → unit + threshold inputs.
web/components/loyalty/badge-form.tsx         Streamer: image upload + metric picker + tier.
web/components/loyalty/leaderboard.tsx        Viewer-facing leaderboard table.
web/app/[lng]/(main)/settings/channel-loyalty/page.tsx
web/app/[lng]/(main)/users/[userId]/loyalty/page.tsx
web/lib/i18n/locales/{en,vi}/channel-loyalty.json
```

### Settings UX (the heart of flexibility)

The streamer sees two panels:

**Rules** — tiles of currently-active rules with a "+ Add rule" button. Adding opens a form:

```
┌─────────────────────────────────────────┐
│ When: [Watch stream           ▼]        │
│ Grant: [ 1 ] point per [ 60 ] seconds   │
│        watched                          │
│ Cooldown: [   0   ] seconds             │
│ Daily cap: [ 1000  ] points             │
│ ☐ Only while I'm live                   │
│ ☐ 2× points for subscribers             │
│                       [Save]  [Cancel]  │
└─────────────────────────────────────────┘
```

The event-type dropdown drives which fields appear. The form's only logic is a mapping `eventType → what fields to show` (e.g. `donation_made` shows amount-based fields, `chat_message_sent` hides magnitude). Conditions become the `conditions` JSONB.

**Badges** — grid of badge tiles with image, name, trigger text. "+ Add badge" opens a form with image uploader (to MinIO via existing `/files` pattern), name, tier number, trigger metric dropdown, threshold input.

### Live preview

As the streamer tunes rules, show a sample: "At 1 pt / 60s, a viewer watching your 4-hour stream would earn 240 points." Pure client-side math from the rule form values. No backend involved.

---

## 9. Chat service — prerequisite changes

To make `chat_message_sent` work the chat service must publish that event. This is the **only** backend surface change outside of the new loyalty service itself.

`backend/chat/src/chatServer.ts` — in the existing `ws.on('message')` handler, after validation but before publish to Redis, produce to Kafka:

```ts
// rate-limit: 1/sec per (roomId, userId) to stop spam farming points
if (shouldSample(roomId, userId)) {
    kafka.produce('letslive.chat', 'chat.message_sent', {
        channelId: roomId,    // roomId === streamerId in this codebase
        viewerId: userId,
        textLen: text.length,
        sentAt: new Date().toISOString()
    })
}
```

`shouldSample` is an in-memory sliding-window rate limiter (existing `Map` + timestamp approach). Rejected messages still go to chat; they just don't award points. This is critical — without throttling, spammers farm points.

Similarly, `backend/livestream/` needs to produce `viewer.heartbeat` every 15s from the presence-tracking loop. Producer lives where the "watching now" count is updated; piggyback on the existing tick.

Kafka producer boilerplate follows the existing `finance`/`livestream` Go eventbus usage (`eventbus.PublishEvent(...)`).

---

## 10. Frontend — Mobile (Flutter)

### Routes

- `AppRoutes.channelLoyalty(channelId)` — viewer page.
- `AppRoutes.settingsLoyalty` — streamer settings.

### New files

```
mobile/lib/models/loyalty.dart                              LoyaltyRule, ChannelBadge, ViewerLoyalty
mobile/lib/features/loyalty/data/loyalty_repository.dart    REST + bulk lookup
mobile/lib/features/loyalty/presentation/
    badge_strip.dart                                        Row of <=4 badges (ellipsis if more)
    points_chip.dart                                        Animated counter (AnimatedSwitcher)
    channel_loyalty_screen.dart                             Viewer page
    loyalty_settings_screen.dart                            Streamer config
    rule_form_sheet.dart                                    Bottom-sheet form mirroring web
    badge_form_sheet.dart                                   Image picker + trigger inputs
mobile/lib/features/livestream/presentation/livestream_screen.dart    (integration — chat row)
mobile/lib/core/network/api_endpoints.dart                  + loyalty* endpoints
mobile/lib/providers.dart                                   + loyaltyRepositoryProvider + loyaltyCacheProvider
mobile/lib/core/router/app_router.dart                      + routes
mobile/l10n/app_{en,vi}.arb                                 channelLoyalty* keys
```

Chat integration lives in `livestream_screen.dart`'s message-row builder: look up the cached `ViewerLoyalty` by `(channelId, viewerId)` and prepend a `BadgeStrip` plus trailing `PointsChip`. Cache lives in a `StateNotifier` with 30s TTL, same pattern as `chatCommandRepositoryProvider`.

---

## 11. Kafka event definitions

Add to `backend/shared/pkg/eventbus/events/loyalty.go`:

```go
package events

import "github.com/gofrs/uuid/v5"

const (
    PointsAwarded  = "loyalty.points_awarded"
    BadgeUnlocked  = "loyalty.badge_unlocked"
    PointsRedeemed = "loyalty.points_redeemed"
)

type PointsAwardedEvent struct {
    ChannelId  uuid.UUID `json:"channelId"`
    ViewerId   uuid.UUID `json:"viewerId"`
    RuleId     uuid.UUID `json:"ruleId"`
    Delta      int32     `json:"delta"`
    NewBalance int64     `json:"newBalance"`
    EventType  string    `json:"eventType"`
}

type BadgeUnlockedEvent struct {
    ChannelId uuid.UUID `json:"channelId"`
    ViewerId  uuid.UUID `json:"viewerId"`
    BadgeId   uuid.UUID `json:"badgeId"`
    BadgeName string    `json:"badgeName"`
}
```

And to `topics.go`:

```go
const TopicLoyalty = "letslive.loyalty"
// + in DefaultTopics(): {Name: TopicLoyalty, NumPartitions: 3, ReplicationFactor: 1},
```

Plus two new producers in existing services:

```
backend/shared/pkg/eventbus/events/chat.go       (new)     - ChatMessageSentEvent
backend/shared/pkg/eventbus/events/livestream.go (addition) - ViewerHeartbeatEvent
```

---

## 12. i18n

New namespace `channel-loyalty` (web) / `channelLoyalty*` keys (mobile):

```
settings.title, settings.rules_heading, settings.badges_heading,
settings.points_name_label, settings.points_icon_label,
rule.add, rule.edit, rule.delete_confirm,
rule.event.watch_seconds, rule.event.chat_message_sent,
rule.event.follow_started, rule.event.donation_made,
rule.unit.per_magnitude, rule.unit.per_event,
rule.cooldown, rule.daily_cap, rule.condition.live_only, rule.condition.subscriber_mult,
badge.add, badge.image_upload, badge.tier_label,
badge.trigger.lifetime_points, badge.trigger.watch_seconds_total,
badge.trigger.chat_messages_total, badge.trigger.donation_cents_total,
viewer.my_balance, viewer.lifetime_earned, viewer.next_badge_progress,
viewer.leaderboard_title, viewer.leaderboard_empty,
notification.badge_unlocked_title, notification.badge_unlocked_body
```

Both `en` and `vi` from day one.

The settings nav label lives in the existing `settings.navigation.channel_loyalty` key for consistency with existing nav entries.

---

## 13. Security and abuse

- **Streamer-scope auth.** All rule/badge writes: `:channelId === callerId`. Otherwise 403.
- **Rate-limit at producer.** Chat service throttles `chat.message_sent` to 1/sec per (channel, viewer) so message-spam doesn't become points-spam. Livestream heartbeats are already capped at 1/15s per session.
- **Rule evaluator caps.** Per-rule `cooldown_sec` and `daily_cap` hard-limit earn rate even if an event stream goes haywire.
- **Ledger immutability.** `points_ledger` has the same "no UPDATE, no DELETE" triggers as finance's `ledger_entries`. Corrections are new entries with negative deltas and `reason='admin_adjust'`.
- **Balance consistency.** Balance upsert + ledger insert inside a single SERIALIZABLE tx. If either fails, neither lands.
- **Badge upload surface.** Streamer badge image upload goes through the same mime-sniff + 2MB size cap as avatar uploads. MinIO path `loyalty/{channelId}/badges/{badgeId}.{ext}` — served public via Kong `/files/loyalty/...` (no auth gate: badges are visible to everyone in chat anyway).
- **Admin adjust.** `POST /adjust` is streamer-only and writes an audit row with `metadata.adminReason` so support can trace every grant/revoke.
- **No IPs / no PII.** Rules operate on userIds only.
- **Abuse surface: sockpuppeting.** A bad actor could open many accounts to farm follow-bonus (100 pts once per follow). Mitigations deferred to user service's existing signup throttling + Cloudflare turnstile at signup. Flag in the streamer UI as a known limitation ("one-time follow bonus is best-effort only").

---

## 14. Trade-offs and things deliberately not done

- **No cross-channel points.** Points live in `(channelId, viewerId)` pairs. You can't transfer points from channel A to channel B. Intentional — matches Twitch's model and avoids a cross-channel exchange-rate mess.
- **No points spending in MVP.** Redemptions are P2. Viewers can see their balance but can't do anything with it. Ledger supports negative deltas today so the schema is ready.
- **No multipliers in MVP.** Subscriber × 2 / raid × 1.5 / etc. are all expressible via `conditions` JSONB but the evaluator ignores them at P0. Flipping them on is a one-function change.
- **Watch-heartbeat granularity is 15s.** Finer granularity (client-side timer every second) would double write pressure for no user-visible benefit.
- **Badges do not expire.** "Subscriber-for-N-months" badges that revoke on lapse are a P3 concern.
- **Leaderboard is eventually consistent.** `/top` does a straight `SELECT ... ORDER BY balance DESC LIMIT 100` on `points_balance`. No cache. Cap `limit` at 100 so the query is a short index scan. Revisit if a channel's view exceeds ~10M unique viewers (it won't).
- **No streamer-to-streamer badge sharing.** Each channel maintains its own badges even if a streamer runs multiple channels. P4.
- **Badge threshold re-evaluation on rule change.** If a streamer lowers a threshold retroactively, viewers who *already* hit the new threshold aren't retro-granted until the next event. Fix with a one-off "recompute" admin button (P2).
- **No fraud detection.** Someone opening a WebSocket and never watching the video still earns watch points (heartbeat is driven by presence, not actual playback). Flag; could be tightened by also requiring "video playing" signal from client-side heartbeat.
- **Points name customizable, currency precision hardcoded.** Points are integers. No fractional points, no "2.5 stars." Simpler renderer, simpler comparisons.

---

## 15. Testing checklist

Manual test matrix once P0 + P1 ship:

- [ ] `docker compose up` — `loyalty_db` migrates, loyalty service boots, Kafka topic created.
- [ ] Visit streamer settings → default rules seed appears on first load.
- [ ] Change watch rate from 1/60s to 5/60s → save → viewer earns at new rate within next heartbeat.
- [ ] Open a stream as a viewer → wait 60s → balance shows 1 point.
- [ ] Send chat messages → points tick up according to rule; spam-send 20 msgs / 1s → still only 1 award (rate limit in chat service); spam-send 200 msgs over a day → `daily_cap` blocks further awards.
- [ ] Anonymous chat viewer → sends messages → no points earned, no badge lookup, no render blocks.
- [ ] Donation via finance → loyalty consumer receives `finance.donation_sent` → points awarded per rule; verify idempotency by replaying the same Kafka event.
- [ ] Streamer creates a badge with `lifetime_points >= 100` → earn 100 points → badge appears next to name in chat for all viewers within 30s (cache TTL).
- [ ] Streamer deletes a badge → cascade removes from `viewer_channel_badges` → disappears from chat within cache TTL.
- [ ] `GET /loyalty/:c/top?limit=10` returns balance-desc-ordered list.
- [ ] Mobile: badge strip renders, points chip animates on award event (via polling `/me`).
- [ ] Streamer auth: try to edit another channel's rule via direct API → 403.
- [ ] Switch `lng` cookie between `en` and `vi` → loyalty UI re-renders translated.

---

## 16. Rough file-count + scope

- **Backend (new `loyalty` service):** ~22 files — `cmd/main.go`, `config/`, `domains/`, `handlers/` (8 endpoints × baseHandler), `repositories/` (rules, badges, ledger, balance, cooldowns, metrics), `services/` (rule evaluator, badge checker), `api/` (router wiring), migrations (1), Dockerfile, go.mod.
- **Backend (additions):**
  - `backend/chat/src/chatServer.ts` — producer in ws message handler (+ rate-limit helper).
  - `backend/livestream/` — producer in presence tick.
  - `backend/shared/pkg/eventbus/events/loyalty.go` (new) + `chat.go` (new) + `livestream.go` (extend) + `topics.go` (extend).
  - `backend/user/` — consume `loyalty.badge_unlocked` in notification service → INSERT notification row.
- **Gateway:** `configs/kong.yml` — add `Loyalty_Routes` under a new `Loyalty` service.
- **Web:** ~11 new files + 3 edits (chat row, settings layout, home-nav entry for streamer tools).
- **Mobile:** ~11 new files + 4 edits (router, providers, l10n, api_endpoints).
- **i18n:** 2 new JSONs (web) + 2 `.arb` diffs (mobile) + `settings.navigation.channel_loyalty` in existing file.
- **docker-compose:** 1 edit — add loyalty service + `loyalty_db` postgres.

**Estimate for P0:** ~2 weeks for one engineer (service boilerplate + ledger + watch/chat rules + chat enrichment). **P1 adds ~1 week** (badge table + unlock evaluator + image upload + UI grid + chat render integration). The rule engine's flexibility comes almost for free once the evaluator is in — P3's new earn paths are ~20 LOC each.

---

## 17. Open questions (decide before implementation)

1. **Points display name customization scope.** Per-channel only, or also per-stream? Assuming per-channel.
2. **Image upload path for badges.** Reuse existing `/files` route under MinIO, or inline base64 in the badge body (simpler, worse for caching)? Assuming `/files`.
3. **Anonymous leaderboard visibility.** Should anonymous viewers see the leaderboard or just logged-in ones? Assuming public — it's marketing for the channel.
4. **Watch-points while VOD-watching.** Do VOD views earn channel points? Assuming **no** at MVP — channel points are a live-stream engagement mechanic. Easy to flip later by having VOD service emit heartbeats.
5. **Negative balance floor.** When redemptions land (P2), should balance be allowed to go negative? Assuming **no** — redemption API returns 409 if `balance - cost < 0`.
6. **Badge retroactivity.** When a streamer creates a new badge with a threshold many viewers have already met, should they be retro-unlocked? Assuming **no** at MVP — too expensive to recompute across all viewers; ship an admin "Recompute" button in P2.
7. **Chat service Kafka producer.** Chat is Node; we don't currently have a Node producer helper in this repo. Do we (a) write a thin wrapper around `kafkajs`, (b) have chat call a loyalty REST endpoint per message (bad — adds sync latency), or (c) dual-write from chat to loyalty via some shared queue? Assuming (a), wrap `kafkajs` in `backend/chat/src/lib/eventbus.ts` following the Go eventbus API shape.

---

## 18. File index

```
backend/loyalty/                                            (new service)
  cmd/main.go
  config/
  domains/rule.go, badge.go, ledger.go, metrics.go
  repositories/ruleRepo.go, badgeRepo.go, ledgerRepo.go,
                balanceRepo.go, metricsRepo.go, cooldownRepo.go
  services/evaluator.go, badgeChecker.go, loyaltyService.go
  handlers/ (general/ + basehandler/)
  api/router.go
  migrations/0001_loyalty.sql
  Dockerfile, go.mod

backend/chat/src/chatServer.ts                              (producer added)
backend/chat/src/lib/eventbus.ts                            (new — kafkajs wrapper)
backend/livestream/services/presence.go                     (producer added)

backend/shared/pkg/eventbus/events/loyalty.go               (new)
backend/shared/pkg/eventbus/events/chat.go                  (new)
backend/shared/pkg/eventbus/events/livestream.go            (+ViewerHeartbeatEvent)
backend/shared/pkg/eventbus/events/topics.go                (+TopicLoyalty)

backend/user/services/notificationService.go                (badge-unlocked consumer)

configs/kong.yml                                            (Loyalty_Routes)
docker-compose-dev.yaml                                     (loyalty service + loyalty_db)

web/types/loyalty.ts
web/lib/api/loyalty.ts
web/components/loyalty/badge-strip.tsx
web/components/loyalty/points-chip.tsx
web/components/loyalty/rule-form.tsx
web/components/loyalty/badge-form.tsx
web/components/loyalty/leaderboard.tsx
web/app/[lng]/(main)/settings/channel-loyalty/page.tsx
web/app/[lng]/(main)/users/[userId]/loyalty/page.tsx
web/app/[lng]/(main)/users/[userId]/chat.tsx               (integration)
web/app/[lng]/(main)/settings/layout.tsx                   (nav entry)
web/lib/i18n/locales/{en,vi}/channel-loyalty.json
web/lib/i18n/locales/{en,vi}/settings.json                 (navigation.channel_loyalty)

mobile/lib/models/loyalty.dart
mobile/lib/features/loyalty/data/loyalty_repository.dart
mobile/lib/features/loyalty/presentation/badge_strip.dart
mobile/lib/features/loyalty/presentation/points_chip.dart
mobile/lib/features/loyalty/presentation/channel_loyalty_screen.dart
mobile/lib/features/loyalty/presentation/loyalty_settings_screen.dart
mobile/lib/features/loyalty/presentation/rule_form_sheet.dart
mobile/lib/features/loyalty/presentation/badge_form_sheet.dart
mobile/lib/features/livestream/presentation/livestream_screen.dart   (integration)
mobile/lib/providers.dart                                   (providers added)
mobile/lib/core/network/api_endpoints.dart                  (loyalty* added)
mobile/lib/core/router/app_router.dart                      (routes added)
mobile/l10n/app_{en,vi}.arb                                 (channelLoyalty* keys)
mobile/lib/l10n/app_localizations*.dart                     (regenerated)
```
