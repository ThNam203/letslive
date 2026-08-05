# Event Bus Setup Report

## Overview

This document describes the event bus infrastructure added to the LetsLive project. The design is **engine-agnostic** — services program against abstract `Producer`, `Consumer`, and `Admin` interfaces defined in the `eventbus` package. The actual transport (NATS, Kafka, Redis Streams, RabbitMQ, etc.) is selected at initialization time by choosing an implementation sub-package.

**Currently provided engine:** NATS JetStream (via `eventbus/natsbus`). An earlier iteration used Kafka (`eventbus/kafkabus`); it was replaced because nothing in this project's services was wired to it yet, and running a full Kafka broker (KRaft controller, partitions, consumer-group offset management) is disproportionate infrastructure for this project's scale. JetStream gives the same durability/replay/consumer-group properties over a single lightweight binary.

---

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                   Service Code                           │
│  (imports eventbus.Producer, eventbus.Consumer, etc.)    │
└──────────────────────┬───────────────────────────────────┘
                       │  depends on interfaces only
                       ▼
┌──────────────────────────────────────────────────────────┐
│              eventbus  (shared/pkg/eventbus/)             │
│                                                          │
│  Event struct    Producer interface    Consumer interface │
│  EventHandler    Admin interface       TopicConfig        │
│  NewEvent()      ParseEventData[T]()                     │
└──────────┬──────────────────────────────────┬────────────┘
           │                                  │
           ▼                                  ▼
┌─────────────────────┐          ┌─────────────────────┐
│   natsbus/           │          │   (future engines)   │
│                     │          │                     │
│  NewProducer()      │          │   kafkabus/          │
│  NewConsumer()      │          │   redisbus/         │
│  NewAdmin()         │          │   rabbitbus/        │
└─────────────────────┘          └─────────────────────┘
```

### Swapping Engines

The engine is chosen **once** — in `main.go` at initialization. All downstream code (services, handlers) only sees the `eventbus.Producer` and `eventbus.Consumer` interfaces.

```go
// Using NATS:
import "sen1or/letslive/shared/pkg/eventbus/natsbus"
producer, err := natsbus.NewProducer(ctx, natsURL)

// Switching to Kafka (future, if throughput/replay needs outgrow JetStream):
import "sen1or/letslive/shared/pkg/eventbus/kafkabus"
producer := kafkabus.NewProducer(brokers)

// The rest of the codebase doesn't change — same eventbus.Producer interface.
```

Note the NATS constructors return `(eventbus.X, error)` — unlike Kafka's lazy-dial clients, `nats.Connect` dials immediately, so construction can fail (or block retrying) if the server isn't reachable yet.

---

## What Was Added

### 1. Docker Infrastructure

**Files modified:**
- `docker-compose.yaml`
- `docker-compose-dev.yaml`

A single NATS server was added with JetStream enabled.

```yaml
nats:
  image: nats:2-alpine
  container_name: letslive-nats
  command: ["-js", "-sd", "/data", "-m", "8222"]
  ports:
    - "4222:4222"
```

**Key configuration decisions:**
- **JetStream (`-js`)**: enables persistence, streams, and durable consumers — the NATS equivalent of Kafka's log retention and consumer groups.
- **File-backed store (`-sd /data`)**: JetStream state survives container restarts via the `nats_data` volume.
- **Monitoring port (`-m 8222`)**: exposes `/healthz` for the compose healthcheck.
- **Single node, no clustering**: matches the prior single-broker Kafka setup (replication factor 1). Add NATS clustering when scaling beyond one instance.
- **Health check**: `wget` against `http://localhost:8222/healthz`.
- **Persistent volume**: `nats_data` volume preserves JetStream data across container restarts.

### 2. Core Event Bus Package (engine-agnostic, unchanged)

**Location:** `backend/shared/pkg/eventbus/`

| File | Purpose |
|------|---------|
| `eventbus.go` | `Event` struct, `Producer`/`Consumer`/`Admin` interfaces, `EventHandler` type, `TopicConfig` |
| `event_builder.go` | `NewEvent` helper and generic `ParseEventData[T]` for type-safe deserialization |

These files have **zero dependency** on any message broker and were not touched by the Kafka→NATS swap.

### 3. NATS Engine Implementation

**Location:** `backend/shared/pkg/eventbus/natsbus/`

| File | Purpose |
|------|---------|
| `conn.go` | Shared connect-with-retry helper (mirrors the old kafkabus admin's backoff loop — 2s initial delay, doubling, capped at 30s, 10 attempts) |
| `naming.go` | Derives JetStream-safe stream/durable names from dot-separated topic names (`letslive.livestream` → stream `letslive_livestream`; the topic itself is still used verbatim as the NATS subject) |
| `producer.go` | JetStream `Publish`-based implementation of `eventbus.Producer`, deduped on `event.ID` via `Nats-Msg-Id` |
| `consumer.go` | JetStream durable-consumer implementation of `eventbus.Consumer` — one durable consumer per topic, named after the consumer group id |
| `admin.go` | JetStream stream management (`CreateOrUpdateStream`) implementation of `eventbus.Admin` |

**Dependency:** `github.com/nats-io/nats.go` (includes the `jetstream` subpackage) — added to `backend/shared/go.mod`. `github.com/segmentio/kafka-go` was removed.

### 4. Shared Event Definitions (engine-agnostic, unchanged)

**Location:** `backend/shared/pkg/eventbus/events/`

| File | Event Types |
|------|-------------|
| `topics.go` | Topic name constants and `DefaultTopics()` |
| `livestream.go` | `livestream.started`, `livestream.ended`, `livestream.updated` |
| `user.go` | `user.created`, `user.updated`, `user.followed`, `user.unfollowed` |
| `vod.go` | `vod.created`, `vod.ready`, `vod.transcode_failed` |
| `transcode.go` | `transcode.stream_connected`, `transcode.stream_disconnected`, `transcode.segment_uploaded` |
| `finance.go` | `finance.payment_completed`, `finance.payment_failed`, `finance.donation_sent` |
| `notification.go` | `notification.requested` |

---

## How to Use in a Service

### Step 1: Add Broker Config to Service Config

In your service's config YAML (on the config server), add:

```yaml
nats:
  url: "nats://nats:4222"
```

In your service's `config/config.go`, add:

```go
type Nats struct {
    URL string `yaml:"url"`
}

type Config struct {
    Service  `yaml:"service"`
    Database `yaml:"database"`
    Tracer   `yaml:"tracer"`
    Nats     `yaml:"nats"`
}
```

### Step 2: Initialize in main.go

```go
package main

import (
    "sen1or/letslive/shared/pkg/eventbus"
    "sen1or/letslive/shared/pkg/eventbus/natsbus"
    "sen1or/letslive/shared/pkg/eventbus/events"
)

func main() {
    // ... existing setup (logger, registry, config, migrations, discovery, otel) ...

    config := cfgManager.GetConfig()

    // ensure required streams exist
    admin, err := natsbus.NewAdmin(ctx, config.Nats.URL)
    if err != nil {
        logger.Panicf(ctx, "failed to connect nats admin: %v", err)
    }
    if err := admin.EnsureTopics(ctx, events.DefaultTopics()); err != nil {
        logger.Errorf(ctx, "failed to ensure event bus topics: %v", err)
    }
    admin.Close()

    // create producer (returns eventbus.Producer — engine-agnostic)
    producer, err := natsbus.NewProducer(ctx, config.Nats.URL)
    if err != nil {
        logger.Panicf(ctx, "failed to connect nats producer: %v", err)
    }
    defer producer.Close()

    // create consumer (returns eventbus.Consumer — engine-agnostic)
    consumer, err := natsbus.NewConsumer(ctx, config.Nats.URL, "livestream-service")
    if err != nil {
        logger.Panicf(ctx, "failed to connect nats consumer: %v", err)
    }
    defer consumer.Close()

    // start consuming in a goroutine
    go func() {
        consumer.Subscribe(ctx, []string{events.TopicTranscode}, func(ctx context.Context, event eventbus.Event) error {
            switch event.Type {
            case events.TranscodeStreamConnected:
                data, err := eventbus.ParseEventData[events.TranscodeStreamConnectedEvent](event)
                if err != nil {
                    return err
                }
                // handle the event...
                _ = data
            }
            return nil
        })
    }()

    // pass producer to your handlers/services (they accept eventbus.Producer, not natsbus-specific types)
    server := SetupServer(dbConn, registry, config, producer)

    // ... existing shutdown logic ...
}
```

### Step 3: Publish Events from Services

Services depend only on `eventbus.Producer` — they have no knowledge of the underlying engine.

```go
package services

import (
    "sen1or/letslive/shared/pkg/eventbus"
    "sen1or/letslive/shared/pkg/eventbus/events"
)

type LivestreamService struct {
    repo     LivestreamRepository
    producer eventbus.Producer  // engine-agnostic interface
}

func (s *LivestreamService) StartLivestream(ctx context.Context, userId uuid.UUID, title string) error {
    // ... create livestream in DB ...

    // publish event
    event, err := eventbus.NewEvent(
        events.LivestreamStarted,
        "livestream-service",
        events.LivestreamStartedEvent{
            LivestreamId: livestream.Id,
            UserId:       userId,
            Title:        title,
            StartedAt:    livestream.StartedAt,
        },
    )
    if err != nil {
        return err
    }

    // key has no NATS routing equivalent (JetStream routes by subject, not
    // by key) — natsbus carries it as an "Event-Key" header for parity/debugging
    return s.producer.Publish(ctx, events.TopicLivestream, userId.String(), event)
}
```

### Step 4: Graceful Shutdown

Add cleanup to the existing shutdown `WaitGroup` in `main.go`:

```go
shutdownWg.Add(1)
go func() {
    producer.Close()
    shutdownWg.Done()
}()

shutdownWg.Add(1)
go func() {
    consumer.Close()
    shutdownWg.Done()
}()
```

---

## Event Structure

Every event published through the event bus follows this format (unchanged from the Kafka iteration):

```json
{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "livestream.started",
    "source": "livestream-service",
    "timestamp": "2026-04-13T10:30:00Z",
    "data": {
        "livestreamId": "...",
        "userId": "...",
        "title": "My Stream",
        "startedAt": "2026-04-13T10:30:00Z"
    }
}
```

The NATS engine additionally writes `Event-Type`, `Event-Source`, and `Event-Key` as message headers, and sets `Nats-Msg-Id` to `event.ID` so JetStream can dedupe redelivered publishes within its default dedup window.

---

## Topic Map

Topic names are unchanged from the Kafka iteration — each becomes both a JetStream stream (name with dots replaced by underscores) and the subject used to publish/subscribe.

| Topic | JetStream Stream | Producers | Consumers (suggested) |
|-------|-------------------|-----------|------------------------|
| `letslive.livestream` | `letslive_livestream` | Livestream Service | User (notifications), VOD (stream-to-VOD) |
| `letslive.user` | `letslive_user` | User Service, Auth Service | Livestream, Notification fanout |
| `letslive.vod` | `letslive_vod` | VOD Service | User (notifications), Livestream |
| `letslive.transcode` | `letslive_transcode` | Transcode Service | Livestream (stream status), VOD (segment tracking) |
| `letslive.finance` | `letslive_finance` | Finance Service | User (notifications), Analytics |
| `letslive.notification` | `letslive_notification` | Any service | User Service (notification handler) |

---

## Consumer Group Naming Convention

Use the service name as the consumer group id, same as before:

```
{service-name}-service
```

Examples: `livestream-service`, `user-service`, `vod-service`

Each becomes the **durable name** of a JetStream consumer on the relevant stream, giving each service its own independent progress tracking — multiple services can consume the same topic independently, and multiple instances of the same service share one durable consumer's progress (the JetStream analog of a Kafka consumer group).

---

## Files Changed / Created

```
Modified:
  docker-compose.yaml          — kafka service/volume replaced with nats service/volume
  docker-compose-dev.yaml      — kafka service/volume replaced with nats service/volume
  backend/shared/go.mod        — removed github.com/segmentio/kafka-go, added github.com/nats-io/nats.go
  backend/shared/go.sum        — updated checksums
  backend/transcode/cmd/main.go — updated a stray "kafka" comment to "nats" (not wired to any engine)

Removed:
  backend/shared/pkg/eventbus/kafkabus/   — kafka-go based implementation

Created:
  backend/shared/pkg/eventbus/natsbus/
  ├── conn.go                  — shared connect-with-retry helper
  ├── naming.go                — topic → stream/durable name derivation
  ├── producer.go              — JetStream-based eventbus.Producer
  ├── consumer.go              — JetStream durable-consumer eventbus.Consumer
  └── admin.go                 — JetStream stream management eventbus.Admin

  docs/NATS_SETUP.md            — this report (replaces docs/KAFKA_SETUP.md)
```

---

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Engine-agnostic interfaces (unchanged)** | Services depend on `eventbus.Producer`/`Consumer`, not NATS directly. Swap engines by changing one line in `main.go`. |
| **Separate `natsbus/` sub-package** | NATS-specific code is isolated, same pattern the Kafka implementation used. |
| **JetStream over core NATS pub/sub** | Core NATS has no persistence or replay — matching Kafka's durability/replay value prop requires JetStream, not plain pub/sub. |
| **Dot-to-underscore stream naming** | JetStream stream/consumer names reject `.`; topics in this project use dots as a namespace separator, so the topic string itself stays the NATS *subject* (dots allowed there) while a sanitized copy becomes the *stream* name. |
| **Durable name = consumer group id** | Keeps the exact same "one shared progress cursor per service" semantics as the Kafka consumer-group convention this project already documents. |
| **Retry-with-backoff on connect** | `nats.Connect` dials immediately (unlike kafka-go's lazy writer/reader) — reused the exact backoff shape from the old `kafkaAdmin.EnsureTopics` so behavior during `docker-compose` startup races is unchanged. |
| **`Nats-Msg-Id` dedup on publish** | Free idempotency for redelivered publishes, directly relevant to the same duplicate-event concerns raised for the (still-unbuilt) loyalty/points feature design. |
| **JSON event format (unchanged)** | Consistent with the project's existing JSON-based REST APIs; no reason to change on engine swap. |

---

## Next Steps

1. **Integrate into a service** — Pick one service (suggested: Transcode or Livestream) and wire up the producer/consumer in its `main.go`
2. **Add broker config to config server** — Add the `nats.url` config to each service's YAML on the config server
3. **Add `depends_on` for NATS** — Add `nats: condition: service_healthy` to service definitions in docker-compose that need it
4. **Replace HTTP gateway calls** — Start replacing synchronous inter-service HTTP calls with event publishing where appropriate (e.g. finance→user gift/inventory calls — see `docs/ISSUES.md` F1/F2)
5. **Production readiness** — When deploying beyond a single node, cluster NATS (3+ nodes) and set stream `Replicas` > 1 in `DefaultTopics()`
