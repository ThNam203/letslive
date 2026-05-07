# System Design Q&A: LETS LIVE (Livestreaming Platform)

---

## 1. Walk me through the overall architecture of this system.

**Answer:** LETS LIVE is a microservices-based livestreaming platform — think Twitch. The backend has 7 Go services (Auth, User, Livestream, VOD, Transcode, Finance) plus a Node.js/TypeScript Chat service. All traffic enters through a **Kong API Gateway**, which handles JWT validation, rate limiting, CORS, and routes requests to internal services discovered via **HashiCorp Consul**. Config is externalized to a **Spring Cloud Config Server**. For observability, the stack includes OpenTelemetry tracing (Grafana Tempo), log aggregation (Loki + Promtail), and Grafana dashboards.

---

## 2. Why use an API Gateway, and why Kong specifically?

**Answer:** An API gateway centralizes cross-cutting concerns — auth validation, rate limiting, CORS, tracing — so each service doesn't implement them independently. Kong was chosen because it has a rich plugin ecosystem (JWT, rate-limiting, OpenTelemetry, correlation-id are all declarative config), it's cluster-ready, and routes are version-controlled as YAML. The alternative would be a custom middleware layer, which adds maintenance burden. The trade-off is Kong is heavier than a simple reverse proxy like Nginx.

---

## 3. How does livestream video work end-to-end?

**Answer:**
1. A creator pushes an RTMP stream (from OBS) to the **Transcode service**
2. The Transcode service uses **FFmpeg** to produce HLS segments at 360p/720p/1080p
3. HLS files are written to **MinIO** (S3-compatible object storage)
4. A file watcher detects new segments and makes them available via MinIO's HTTP endpoint
5. Viewers fetch the HLS playlist (`.m3u8`) and segments directly from MinIO
6. When the stream ends, the VOD is archived — a Kafka event triggers the VOD service to record the archive

This separates ingestion from delivery, and using HLS + MinIO allows horizontal scaling without stateful media servers.

---

## 4. Why two message brokers — Kafka AND Redis?

**Answer:** They solve different problems:
- **Redis Pub/Sub** is used for live chat — it has sub-millisecond latency and is ideal for fire-and-forget fan-out to WebSocket connections. Messages don't need to be durable.
- **Kafka** is used for cross-service business events (stream started, VOD created, payment made). Kafka gives durability, replay capability, and decoupled consumers.

Using Redis for chat avoids Kafka's overhead (partitions, consumer groups, offsets) for something that needs to be fast but not durable. The trade-off is Redis pub/sub has no message persistence — if a subscriber drops, it misses messages.

---

## 5. Why MongoDB for chat and PostgreSQL for everything else?

**Answer:** Chat has semi-structured, high-volume, append-heavy data (messages, conversations with variable metadata) that maps naturally to documents. MongoDB's flexible schema handles slash commands and conversation metadata without rigid migrations. PostgreSQL is used for transactional data with strong relational integrity — users, auth tokens, followers, finance — where ACID guarantees and foreign keys matter. The trade-off is operational complexity from running two DB systems.

---

## 6. How does authentication work, and how is it enforced at scale?

**Answer:** Users authenticate via email+password (with OTP signup verification) or Google OAuth. On success, they receive two JWTs: a short-lived **access token** and a long-lived **refresh token** stored in httpOnly cookies. The access token is validated at the **Kong gateway level** using the JWT plugin — services themselves don't verify tokens, reducing redundant crypto work. Refresh tokens are stored in PostgreSQL and can be revoked (e.g., on logout, all tokens for a user are invalidated). The trade-off: since Kong validates JWTs statically (signature + expiry), there's no real-time revocation for access tokens — a revoked token remains valid until it expires.

---

## 7. How does service discovery work?

**Answer:** Each service self-registers with **Consul** on startup, providing its name, address, port, and a health check endpoint (`/v1/health`). Kong uses Consul's DNS interface to resolve service addresses dynamically. If a service crashes, Consul marks it unhealthy and removes it from DNS, so Kong stops routing to it. This avoids hardcoding service IPs and supports multiple instances of the same service. The exponential backoff retry logic in each service ensures they recover from temporary Consul unavailability on startup.

---

## 8. How would you scale this system under high load?

**Answer:**
- **Stateless services** (all Go/Node services) scale horizontally behind Kong
- **Transcode** is CPU-intensive — scale with a worker pool; each worker handles one RTMP stream
- **Chat** bottleneck is Redis pub/sub — at scale, replace with Redis Cluster or switch to a distributed messaging system
- **PostgreSQL** needs read replicas for query-heavy services (User, Livestream); write throughput goes through the primary
- **Kafka** moves from single-broker to a multi-broker cluster with replication
- **MinIO** clusters (erasure coding mode) or migrate to S3 for HLS delivery + CloudFront CDN in front of it
- **Kong** can be clustered with a shared database (PostgreSQL/Cassandra)

The current docker-compose setup is single-instance everything — designed for dev, not production HA.

---

## 9. How does real-time chat work technically?

**Answer:**
1. Client connects via WebSocket to the Chat service (`/ws?roomId=xxx`)
2. Chat service subscribes to a Redis channel for that room (`chat:room:{roomId}`)
3. When a message is sent, it's saved to MongoDB and published to the Redis channel
4. All connected WebSocket clients subscribed to that channel receive the broadcast
5. For DMs, JWT-authenticated WebSocket (`/dm-ws`) with per-conversation Redis channels

This is a classic **fan-out on write** pattern. The Chat service acts as a stateful WebSocket server, but the pub/sub state lives in Redis, so multiple Chat service instances can handle the same room.

---

## 10. Why a Spring Cloud Config Server in a Go-centric stack?

**Answer:** Spring Cloud Config provides mature, production-grade externalized configuration with Git-backed versioning, environment profiles (dev/staging/prod), and dynamic refresh without rebuilds. Go doesn't have an equivalent first-class solution. The trade-off is running a JVM service purely for config, which adds memory overhead. The alternative would be Consul KV (already in the stack) or Vault, but Spring Cloud Config was likely chosen for familiarity and feature completeness.

---

## 11. What are the main failure points and how are they handled?

**Answer:**
- **Kong down**: All traffic blocked — it's a SPOF in this setup. Production fix: Kong clustering + load balancer
- **Consul down**: Service discovery fails, no new routing. Services use exponential backoff to reconnect. Cached DNS can sustain briefly.
- **Kafka down**: Async events are lost — VOD archiving, notifications, finance events fail silently. Production fix: local retry queues + DLTs
- **Redis down**: Live chat goes dark. Chat service WebSockets still connect but no fan-out. Fix: Redis Sentinel/Cluster
- **MinIO down**: HLS delivery fails, streams go black. Fix: MinIO clustering or S3 fallback
- **PostgreSQL down**: Most services degrade to read-only or fail entirely — no mitigation currently

---

## 12. How does the Transcode/HLS pipeline handle private vs public streams?

**Answer:** HLS segments are written to separate MinIO paths: `/transcode/hls/public/` for public streams (accessible without auth) and `/transcode/hls/private/` for private ones. A file watcher detects when private VODs are ready and moves them. Access control happens at the Kong level — private VOD routes require a valid JWT. The trade-off: since HLS segments are static files served from MinIO, enforcing per-segment auth is hard; in production, this would be done with signed URLs or a CDN token system.

---

## 13. How do notifications work end-to-end?

**Answer:** Notifications are event-driven via Kafka. For example, when a followed user starts a stream, the Livestream service publishes to the `letslive.livestream` topic. The User service consumes this event and inserts notification records into PostgreSQL for each follower. The frontend polls or receives notifications through the User service REST API. The trade-off of this approach: notification fanout to many followers is done synchronously in the consumer — for a creator with millions of followers, this would need to be batched or offloaded to a dedicated notification worker.

---

## 14. How are WebSocket connections kept alive and cleaned up?

**Answer:** The Chat service implements a keep-alive mechanism using periodic ping/pong frames over WebSocket. A background goroutine/timer sends pings at a fixed interval; if a pong isn't received within the deadline, the connection is marked dead and closed. This prevents ghost connections from accumulating (e.g., when a mobile client loses signal without sending a TCP FIN). Without this, the server would hold open file descriptors for dead sockets indefinitely, exhausting OS limits. On disconnect (clean or dead), the service unsubscribes the socket from its Redis pub/sub channel.

---

## 15. How does the Finance service ensure transaction integrity?

**Answer:** The Finance service uses PostgreSQL transactions for all deposit and withdrawal operations. A database trigger (`allow_transaction_status_update_only`) enforces that once a transaction is created, only its status can be updated — fields like amount and type are immutable. This prevents accidental double-updates and ensures an audit trail. Kafka events are published after a successful commit, so downstream services (e.g., notification) only react to committed transactions, not in-flight ones. The trade-off: there's no outbox pattern, so if the service crashes after the DB commit but before publishing to Kafka, the event is lost.

---

## 16. How does the system handle adaptive bitrate streaming?

**Answer:** The Transcode service uses FFmpeg to produce three quality renditions from a single RTMP input: 360p, 720p, and 1080p. FFmpeg generates a master HLS playlist (`.m3u8`) that lists all renditions with their bandwidth metadata. The video player on the client (HLS.js or native) reads the master playlist, measures available bandwidth, and automatically switches between quality levels. This happens client-side without any server involvement after the files are generated. The trade-off: all three renditions are always encoded even if no viewer needs 1080p, which wastes CPU on the Transcode service.

---

## 17. How does service-to-service communication work?

**Answer:** Services communicate synchronously over HTTP, using Consul DNS for address resolution (e.g., `http://user-service/v1/users/{id}`). There's no gRPC or service mesh sidecar — just plain HTTP with JSON. Consul health checks ensure only healthy instances receive traffic. For async flows (VOD archiving, notifications), services communicate via Kafka topics instead. The trade-off of HTTP over gRPC: simpler debugging and no schema compilation step, but larger payload sizes and no streaming support. For a platform at Twitch scale, gRPC with protobuf would be more efficient.

---

## 18. How is observability implemented across services?

**Answer:** The stack has three pillars:
- **Tracing**: Every service instruments with OpenTelemetry SDK, exporting spans via OTLP to Grafana Tempo. Kong injects a `correlation-id` header and propagates W3C trace context, so a single user request can be traced across Kong → Auth → User → Livestream.
- **Logging**: Services use structured loggers (Zap in Go, Pino in Node). Promtail tails container stdout/stderr and ships to Grafana Loki, labeled by service name.
- **Dashboards**: Grafana connects to both Tempo and Loki, enabling trace-to-log correlation — you can click a slow span and jump directly to the logs from that service at that timestamp.

In dev, 100% of traces are sampled. In production this would be reduced (e.g., 1–5%) to control storage costs.

---

## 19. How does config management work and how do services pick up changes?

**Answer:** A Spring Cloud Config Server reads configuration from a Git repository and exposes it over HTTP. Each service fetches its config on startup at `http://config-server/{service-name}/{profile}`. Services poll the config server on a configurable interval (default 30 min) to pick up changes without restarting. The profile (dev/staging/prod) is passed as an environment variable, so the same service binary gets different config in each environment. The trade-off: a 30-minute polling interval means config changes aren't instant. For faster propagation, Spring Cloud Bus (backed by Kafka) can push refresh events — but that's not wired up here.

---

## 20. How do you prevent unauthorized RTMP streams?

**Answer:** Each user has a unique **stream key** stored in the User service. When an RTMP connection arrives at the Transcode service, it extracts the stream key from the RTMP URL path and validates it against the User service before accepting the stream. If the key is invalid or the user has no active livestream record, the connection is rejected. This means a streamer must have created a livestream session (via the API) before going live. The stream key should be treated like a password — it's not shown in the UI after initial generation and can be rotated.

---

## 21. What's the data flow when a user follows another user?

**Answer:**
1. Client sends `POST /v1/users/{id}/follow` → Kong validates JWT → User service
2. User service inserts a row into the `followers` table (follower_id, followee_id) in PostgreSQL
3. User service publishes a `user.followed` event to the `letslive.user` Kafka topic
4. Downstream consumers (e.g., notification service) react to the event — e.g., to notify the followee or update recommendation signals
5. The follower count on the followee's profile is either computed on read (COUNT query) or maintained as a denormalized counter column

The follow relationship is a simple join table, which is efficient for checking "does user A follow user B" and for fetching follower/following lists with pagination.

---

## 22. How does the VOD system work after a stream ends?

**Answer:**
1. Stream ends → Transcode service publishes to `letslive.transcode` Kafka topic with the HLS file location in MinIO
2. VOD service consumes the event and creates a VOD record in PostgreSQL pointing to the MinIO path
3. The private HLS directory is cleaned up or moved to the VOD path in MinIO
4. Users can then browse VODs, which are served as static HLS files from MinIO
5. VOD comments are stored in PostgreSQL as a separate `vod_comments` table linked to the VOD ID

The VOD is essentially a snapshot of the HLS output — no re-encoding is needed. The trade-off: HLS segment files from the live stream are kept as-is for VOD playback, which is efficient but means the VOD quality tiers match whatever was transcoded live.

---

## 23. How would you add a CDN in front of this system?

**Answer:** The main candidate for CDN is HLS delivery from MinIO. Currently, viewers fetch segments directly from MinIO — at scale this creates a hot origin. The fix:
1. Put CloudFront (or any CDN) in front of the MinIO public bucket
2. Update the HLS playlist URLs to point to the CDN hostname instead of MinIO directly
3. For live streams, set a short TTL (2–5s) on segment files so CDN doesn't serve stale chunks
4. For VODs, use a long TTL (days/weeks) since files are immutable

For the API layer, Kong already handles rate limiting so a CDN isn't needed there. The config change would be in the Transcode service where it constructs the playback URL returned to clients.

---

## 24. What are the main security considerations in this system?

**Answer:**
- **JWT in httpOnly cookies**: Prevents XSS from stealing tokens via `document.cookie`
- **Kong JWT plugin**: Centralized token validation — services trust that Kong has already verified the caller
- **OTP email verification**: Prevents throwaway account creation at signup
- **CAPTCHA (Cloudflare Turnstile)**: Protects login/signup from bots on web; skipped for mobile clients via User-Agent detection
- **Rate limiting**: Kong applies global (100 req/s) and per-route limits to prevent abuse
- **Stream key rotation**: Users can regenerate their RTMP stream key if compromised
- **CORS**: Kong CORS plugin restricts which origins can make credentialed requests

Gaps: no mutual TLS between services (traffic inside Docker network is unencrypted), no secrets management (secrets in env vars rather than Vault), and the JWT access token can't be revoked before expiry.

---

## 25. Walk me through the email verification (OTP) flow at signup.

**Answer:** When a user signs up with email+password, the Auth service generates a 6-digit numeric OTP using `crypto/rand` (not `math/rand`, so it's cryptographically secure). The OTP is stored in PostgreSQL (`sign_up_otp` table) with a 5-minute TTL and an `email` foreign key. The OTP is delivered via SMTP using Go's `net/smtp`. To verify, the client posts the code; the service looks up by `(code, email)`, checks `used_at IS NULL` and `expires_at > now()`, and stamps `used_at` on success. The trade-off: storing OTPs in the same Postgres as auth credentials simplifies operations but couples OTP read load to the auth DB; a dedicated Redis store with native TTL would scale better and avoid manual expiry checks.

---

## 26. Why does the transcode service watch the filesystem with a polling library instead of `fsnotify` / inotify?

**Answer:** The HLS pipeline uses `github.com/radovskyb/watcher`, which polls the filesystem on an interval rather than subscribing to kernel events. The reason is portability and reliability across container filesystems — `fsnotify` events can be lost or misreported on overlay filesystems, network mounts, and some Docker volume drivers, which is exactly the environment FFmpeg writes into here. Polling trades latency (segments are detected on poll tick, not instantly) for correctness. The watcher uses a strategy interface (`OnMaster`/`OnVariant`/`OnSegment`/`OnThumbnail`/`OnCreate`) so different lifecycle stages dispatch to different handlers. Trade-off: at high segment-rates, polling cost grows linearly with number of files in the watched tree.

---

## 27. How does the system handle uploaded VOD files (not live-streamed)?

**Answer:** Users can upload pre-recorded videos that get transcoded into HLS. The flow:
1. Client uploads raw video file to a private MinIO bucket
2. The Transcode service runs a background worker (`TranscodeWorker`) that polls every 5 seconds for new objects in the raw bucket
3. When found, FFmpeg transcodes the file into the same multi-quality HLS layout used for live streams
4. The result is uploaded to the public HLS path in MinIO and the Livestream service is notified via HTTP gateway to update the VOD record's `playback_url`/`thumbnail_url` and status

The trade-off of polling vs an event-driven trigger (e.g., MinIO bucket-notification webhook): polling is simpler and tolerates restarts well, but introduces up to 5 seconds of latency before transcoding starts.

---

## 28. How are database migrations managed across services?

**Answer:** Each service owns its own migration directory (e.g., `backend/auth/migrations/`, `backend/user/migrations/`) with sequentially numbered SQL files (`0001_*.sql`, `0002_*.sql`, ...). Migrations run on container startup before the service binary starts. There is no shared schema and no cross-service foreign key — services reference each other only by IDs (typically UUIDs). The trade-off: this gives each service true schema ownership and lets teams ship migrations independently, but cross-service joins are impossible — anything that needs another service's data goes through HTTP gateways or denormalized event-sourced copies.

---

## 29. What's the frontend stack and why those choices?

**Answer:** The web app is **Next.js 16 (App Router) + React 19 + TypeScript + Tailwind**, with **Radix UI primitives** + **shadcn/ui**-style components for accessibility, **Zustand** for client state, **react-i18next** for localization, **react-player** for HLS playback, **Zod** for runtime validation, **Sentry** for error tracking, **react-turnstile** for the Cloudflare CAPTCHA, and **MSW** for mocked dev mode (`dev:mock` script). The mock mode lets the frontend run without the backend stack — useful for UI iteration. App Router was chosen for its server component model, which keeps auth-sensitive logic (cookie reads) on the server.

---

## 30. What's the mobile architecture, and how does it share auth with the web?

**Answer:** The mobile app is **Flutter 3.41+** using **Riverpod** for state, **GoRouter** for navigation, **Dio + dio_cookie_manager + cookie_jar** for HTTP, **flutter_secure_storage** for keychain-backed persistence, **video_player + chewie** for HLS playback, **web_socket_channel** for chat, and **forui** for UI components. Critically, mobile uses the same **httpOnly cookie JWT** flow as web — `dio_cookie_manager` persists the cookies in `cookie_jar`, which means the backend doesn't need a separate `Authorization: Bearer` code path. The trade-off: mobile must handle CSRF separately from web (web relies on `SameSite`), and CAPTCHA is skipped for mobile via User-Agent sniffing in the auth handler.

---

## 31. How does Kong's JWT plugin map to per-user identity?

**Answer:** Kong's `jwt` plugin in this setup uses a **single shared consumer** (`authenticated users`) with a single secret keyed `access_token_secret`. All users' JWTs are signed with that same key, and Kong only validates signature + expiry — it does *not* perform per-user lookup. The actual user identity is carried in JWT claims (`user_id`), which Kong forwards in headers to upstream services. The trade-off: this is fast and stateless, but it means revoking a single user's tokens requires either rotating the global secret (invalidating *all* tokens) or accepting that revoked access tokens stay valid until expiry. The Auth service does maintain per-user refresh tokens in PostgreSQL that *can* be revoked individually.

---

## 32. How is request correlation done across services for tracing/debugging?

**Answer:** A `RequestIDMiddleware` in `shared/middlewares/correlation_id.go` reads the `X-Request-ID` header (or generates a UUID v4 if absent), echoes it back in the response, attaches it to the active OpenTelemetry span as `http.request_id`, and stores it in the request context. Kong injects this header at the gateway, so the same ID flows from gateway → service A → gateway → service B. In Grafana, you can search Loki logs by the request ID, then jump to the matching trace in Tempo. The trade-off vs. relying purely on OTel trace IDs: a separate `X-Request-ID` is human-friendlier in log queries and survives even when tracing is disabled.

---

## 33. How does Google OAuth differ from email-password signup?

**Answer:** The Auth service uses `golang.org/x/oauth2/google` with a configured client ID/secret and redirect URL. Flow:
1. Frontend hits `/v1/auth/google`, which redirects to Google's consent page with a CSRF `state` param
2. Google redirects back to the configured callback URL with a `code`
3. Auth service exchanges the code for an access token, then fetches `userinfo` (email, verified_email, name, picture)
4. If the email is already linked to an existing auth row, the user is logged in; otherwise a new auth row is created and the User service is asked to create a corresponding profile
5. Critically, the username is left **NULL** for OAuth users — they're redirected to `/account-setup` to choose one, because deriving a username from the email local-part would leak PII (e.g., `john.smith@gmail.com → john.smith-gg1234`)

Same JWT cookies are issued at the end. The trade-off of separate flows: more code paths to maintain, but it cleanly separates verified-email (Google) from to-be-verified (OTP).

---

## 34. How does the chat service support custom slash commands?

**Answer:** The Chat service exposes a `ChatCommand` model (MongoDB) with two scopes: `CHAT_COMMAND_SCOPE_USER` (personal commands) and `CHAT_COMMAND_SCOPE_CHANNEL` (room-wide commands). Each command has a name (regex `^[a-z0-9_-]{1,32}$`), a static response (max 500 chars), an optional description, and an `ownerId`. There's a 50-commands-per-owner-per-scope cap. When a chat message starts with `/`, the service looks up a matching command (channel scope first, then user scope) and replaces or augments the broadcast with the command's response. The trade-off: only static-response commands are supported here — moderation commands like `/ban` or `/timeout` would require a separate role/permission layer that doesn't yet exist.

---

## 35. What's the testing strategy, and what are its gaps?

**Answer:** There is essentially no automated test coverage in the Go services — no `*_test.go` files exist across the seven Go modules. The Chat service has a single `chatserver.test.ts`. Validation happens manually via `docker-compose-dev.yaml` running the full stack locally. The trade-off is honest: this is a personal/portfolio project optimized for breadth of system integration over test rigor. In production, you'd want at minimum: unit tests for each service's domain logic, integration tests that spin up Postgres + Kafka via `testcontainers`, and contract tests at gateway boundaries (e.g., between Auth and User).

---

## 36. Why does the transcode service write HLS to a local directory and *then* upload to MinIO, rather than letting FFmpeg write directly to S3-compatible storage?

**Answer:** FFmpeg's HLS muxer is filesystem-native — it writes the master playlist, variant playlists, and segment files as a tightly coupled tree, often updating the playlist atomically as new segments arrive. Streaming this directly to S3-compatible storage is fragile because S3's PUT-based API doesn't support partial writes or directory-level atomic rename. So FFmpeg writes to a local watched directory, and the file watcher uploads complete segments to MinIO once they're closed. The trade-off: this introduces local disk pressure on the transcode host (segments accumulate until uploaded), but gives clean atomicity and lets MinIO/S3 stay simple.

---

## 37. How are stream keys generated and stored?

**Answer:** The stream key is a UUID stored in the `users.stream_api_key` column in the User service's PostgreSQL. The user can regenerate it via a `POST /v1/user/me/stream-key` endpoint, which simply runs `UPDATE users SET stream_api_key = $1 WHERE id = $2` with a freshly generated UUID. The Transcode service validates incoming RTMP connections by querying `GET /v1/internal/users/by-stream-key/{key}` against the User service. The trade-off of UUID over a hashed random secret: UUIDs are fine for opaque tokens, but if the database leaks, all stream keys leak in plaintext — a hashed-secret model (store hash, compare on lookup) would be more secure but slightly slower per RTMP handshake.

---

## 38. What is the account-setup flow and why does it exist?

**Answer:** When a user signs in via Google OAuth for the first time, the User service creates their profile with `username = NULL` (the column is nullable, with a UNIQUE index that permits multiple NULLs). The web and mobile clients both detect this state on the post-login user fetch and redirect to `/account-setup`, which forces the user to choose a username before accessing the rest of the app. A global GoRouter redirect on mobile and a layout-level redirect on web enforce this. The reason: previously, the system auto-derived a username from the email local-part (e.g., `john.smith@gmail.com → john.smith-gg1234`), which leaked PII to all viewers. The trade-off: an extra step in the OAuth onboarding funnel, but no PII leakage and users own their public identity.
