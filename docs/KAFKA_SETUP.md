# Event Bus Setup Report

## Overview

This document describes the event bus infrastructure added to the LetsLive project. The design is **engine-agnostic** — services program against abstract `Producer`, `Consumer`, and `Admin` interfaces defined in the `eventbus` package. The actual transport (Kafka, NATS, Redis Streams, RabbitMQ, etc.) is selected at initialization time by choosing an implementation sub-package.

**Currently provided engine:** Kafka (via `eventbus/kafkabus`).

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
│   kafkabus/          │          │   (future engines)   │
│                     │          │                     │
│  NewProducer()      │          │   natsbus/          │
│  NewConsumer()      │          │   redisbus/         │
│  NewAdmin()         │          │   rabbitbus/        │
└─────────────────────┘          └─────────────────────┘
```

### Swapping Engines

The engine is chosen **once** — in `main.go` at initialization. All downstream code (services, handlers) only sees the `eventbus.Producer` and `eventbus.Consumer` interfaces.

```go
// Using Kafka:
import "sen1or/letslive/shared/pkg/eventbus/kafkabus"
producer := kafkabus.NewProducer(brokers)

// Switching to NATS (future):
import "sen1or/letslive/shared/pkg/eventbus/natsbus"
producer := natsbus.NewProducer(natsURL)

// The rest of the codebase doesn't change — same eventbus.Producer interface.
```

---

## What Was Added

### 1. Docker Infrastructure

**Files modified:**
- `docker-compose.yaml`
- `docker-compose-dev.yaml`

A Kafka broker was added using **Apache Kafka in KRaft mode** (no Zookeeper dependency).

```yaml
kafka:
  image: apache/kafka:latest
  container_name: letslive-kafka
  ports:
    - "9092:9092"
```

**Key configuration decisions:**
- **KRaft mode**: Single-node broker acting as both broker and controller — no Zookeeper needed.
- **Auto topic creation disabled** (`KAFKA_AUTO_CREATE_TOPICS_ENABLE: "false"`): Topics are created explicitly via the `Admin.EnsureTopics` interface to prevent typos from silently creating unwanted topics.
- **Log retention**: 168 hours (7 days) default retention.
- **Replication factor**: 1 (single-node setup). Increase when scaling to multiple brokers.
- **Health check**: Uses `kafka-broker-api-versions.sh` to verify the broker is ready.
- **Persistent volume**: `kafka_data` volume preserves data across container restarts.

### 2. Core Event Bus Package (engine-agnostic)

**Location:** `backend/shared/pkg/eventbus/`

| File | Purpose |
|------|---------|
| `eventbus.go` | `Event` struct, `Producer`/`Consumer`/`Admin` interfaces, `EventHandler` type, `TopicConfig` |
| `event_builder.go` | `NewEvent` helper and generic `ParseEventData[T]` for type-safe deserialization |

These files have **zero dependency** on any message broker. They define the contract that all engines must implement.

### 3. Kafka Engine Implementation

**Location:** `backend/shared/pkg/eventbus/kafkabus/`

| File | Purpose |
|------|---------|
| `producer.go` | `kafka-go` Writer-based implementation of `eventbus.Producer` |
| `consumer.go` | `kafka-go` Reader-based implementation of `eventbus.Consumer` with consumer group support |
| `admin.go` | `kafka-go` based implementation of `eventbus.Admin` with retry logic |

**Dependency:** `github.com/segmentio/kafka-go` — added to `backend/shared/go.mod`.

### 4. Shared Event Definitions (engine-agnostic)

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
kafka:
  brokers:
    - "kafka:9092"
```

In your service's `config/config.go`, add:

```go
type Kafka struct {
    Brokers []string `yaml:"brokers"`
}

type Config struct {
    Service  `yaml:"service"`
    Database `yaml:"database"`
    Tracer   `yaml:"tracer"`
    Kafka    `yaml:"kafka"`
}
```

### Step 2: Initialize in main.go

```go
package main

import (
    "sen1or/letslive/shared/pkg/eventbus"
    "sen1or/letslive/shared/pkg/eventbus/kafkabus"
    "sen1or/letslive/shared/pkg/eventbus/events"
)

func main() {
    // ... existing setup (logger, registry, config, migrations, discovery, otel) ...

    config := cfgManager.GetConfig()

    // ensure required topics exist
    admin := kafkabus.NewAdmin(config.Kafka.Brokers)
    if err := admin.EnsureTopics(ctx, events.DefaultTopics()); err != nil {
        logger.Errorf(ctx, "failed to ensure event bus topics: %v", err)
    }
    admin.Close()

    // create producer (returns eventbus.Producer — engine-agnostic)
    producer := kafkabus.NewProducer(config.Kafka.Brokers)
    defer producer.Close()

    // create consumer (returns eventbus.Consumer — engine-agnostic)
    consumer := kafkabus.NewConsumer(config.Kafka.Brokers, "livestream-service")
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

    // pass producer to your handlers/services (they accept eventbus.Producer, not kafkabus-specific types)
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

    // key is used for ordering (partition routing in Kafka, ignored by some engines)
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

Every event published through the event bus follows this format:

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

The Kafka engine additionally writes `event-type` and `event-source` as message headers for header-based filtering.

---

## Topic Map

| Topic | Producers | Consumers (suggested) |
|-------|-----------|----------------------|
| `letslive.livestream` | Livestream Service | User (notifications), VOD (stream-to-VOD) |
| `letslive.user` | User Service, Auth Service | Livestream, Notification fanout |
| `letslive.vod` | VOD Service | User (notifications), Livestream |
| `letslive.transcode` | Transcode Service | Livestream (stream status), VOD (segment tracking) |
| `letslive.finance` | Finance Service | User (notifications), Analytics |
| `letslive.notification` | Any service | User Service (notification handler) |

---

## Consumer Group Naming Convention

Use the service name as the consumer group ID:

```
{service-name}-service
```

Examples: `livestream-service`, `user-service`, `vod-service`

This ensures each service gets its own independent offset tracking — multiple services can consume the same topic independently.

---

## Files Changed / Created

```
Modified:
  docker-compose.yaml          — added kafka service + kafka_data volume
  docker-compose-dev.yaml      — added kafka service + kafka_data volume
  backend/shared/go.mod        — added github.com/segmentio/kafka-go dependency
  backend/shared/go.sum        — updated checksums

Created:
  backend/shared/pkg/eventbus/
  ├── eventbus.go              — Event struct, Producer/Consumer/Admin interfaces (engine-agnostic)
  ├── event_builder.go         — NewEvent and ParseEventData[T] helpers (engine-agnostic)
  ├── kafkabus/                — Kafka engine implementation
  │   ├── producer.go          — kafka-go Writer-based eventbus.Producer
  │   ├── consumer.go          — kafka-go Reader-based eventbus.Consumer
  │   └── admin.go             — kafka-go based eventbus.Admin with retry
  └── events/                  — Shared event definitions (engine-agnostic)
      ├── topics.go            — topic name constants + DefaultTopics()
      ├── livestream.go        — livestream event types and payloads
      ├── user.go              — user event types and payloads
      ├── vod.go               — VOD event types and payloads
      ├── transcode.go         — transcode event types and payloads
      ├── finance.go           — finance event types and payloads
      └── notification.go      — notification event types and payloads

  docs/KAFKA_SETUP.md          — this report
```

---

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Engine-agnostic interfaces** | Services depend on `eventbus.Producer`/`Consumer`, not Kafka directly. Swap engines by changing one line in `main.go`. |
| **Separate `kafkabus/` sub-package** | Kafka-specific code is isolated. Adding a new engine means adding a new sub-package (e.g., `natsbus/`), not modifying existing code. |
| **`Admin` interface** | Topic management varies by engine — Kafka needs explicit creation, NATS doesn't. The interface lets engines no-op where appropriate. |
| **KRaft mode (no Zookeeper)** | Simpler infrastructure, Kafka's recommended approach since 3.3+. |
| **`kafka-go` library** | Pure Go, no CGO dependency, well-maintained, simple API. |
| **Generic `ParseEventData[T]`** | Type-safe event deserialization using Go generics (matches project's use of generics in `ConfigManager[T]`). |
| **Explicit topic creation** | Prevents silent topic creation from typos; retry logic matches the project's pattern in discovery registration. |
| **JSON event format** | Consistent with the project's existing JSON-based REST APIs. |
| **Shared event definitions** | Single source of truth for event contracts across all services, independent of the transport engine. |

---

## Adding a New Engine

To add a new message broker (e.g., NATS):

1. Create `backend/shared/pkg/eventbus/natsbus/`
2. Implement `eventbus.Producer`, `eventbus.Consumer`, and `eventbus.Admin` interfaces
3. In `main.go`, swap the import and constructor:

```go
// Before (Kafka):
import "sen1or/letslive/shared/pkg/eventbus/kafkabus"
producer := kafkabus.NewProducer(brokers)

// After (NATS):
import "sen1or/letslive/shared/pkg/eventbus/natsbus"
producer := natsbus.NewProducer(natsURL)
```

No changes needed in services, handlers, or event definitions.

---

## Next Steps

1. **Integrate into a service** — Pick one service (suggested: Transcode or Livestream) and wire up the producer/consumer in its `main.go`
2. **Add broker config to config server** — Add the `kafka.brokers` config to each service's YAML on the config server
3. **Add `depends_on` for Kafka** — Add `kafka: condition: service_healthy` to service definitions in docker-compose that need Kafka
4. **Replace HTTP gateway calls** — Start replacing synchronous inter-service HTTP calls with event publishing where appropriate
5. **Scale topics** — Adjust partition counts based on expected throughput per topic
6. **Production readiness** — When deploying multi-broker, increase `ReplicationFactor` to 3 in `DefaultTopics()`
