# API Contract Audit — 2026-09-17

Cross-check of the web client's TypeScript types against the Go DTOs / domain
structs and the Node chat service, plus the SQL schema as the source of truth.

## Security

### Sender display name was client-controlled

`dmServer.handleSendMessage` read `data.senderUsername` straight off the client
WebSocket payload (the comment claimed it came from auth, it did not), and
`dmMessageHandler` did the same with `req.body.senderUsername`. The value was
persisted on the message document and on `conversation.lastMessage`, so any
client could post under an arbitrary name. Typing events had the same hole via
`data.username`.

Fixed in two steps.

First, `dmMessageService.sendMessage` now resolves the name from the
conversation's participant record and no longer takes a `senderUsername`
argument. Typing events resolve it through the new
`conversationService.getParticipantUsername`. The client no longer sends either
field.

That alone was not enough: the participant record's `username` was itself
seeded from the request body at conversation creation
(`req.body.participantUsernames?.[id] || id`), so the creator still chose what
everyone in the conversation was called. The chat service now resolves
identities through the user service instead — see "User identity lookup" below.

### Phone numbers on unauthenticated endpoints

`GetUserPublicResponseDTO` carried `PhoneNumber`, and all four public user
queries selected `u.phone_number`. It was exposed on `GET /v1/user/{userId}`,
`GET /v1/users/search` and `GET /v1/users/recommendations`, none of which
require auth, and the web client never declared the field.

Fixed: dropped from the DTO and from every public query.

`email` is also returned by these endpoints. It is declared in the client's
`PublicUser` type and used by the UI, so it was left as is — worth a separate
decision.

## User identity lookup

The chat service previously made no outbound service calls and took every
participant's username and avatar from the client that created the
conversation. It now resolves them from the user service, which owns them:

- user service: `POST /v1/internal/users/batch` takes up to 100 ids and returns
  `{ id, username, profilePicture }` per user, backed by the existing
  `GetPublicInfosByIds` repository method
- chat service: `gateway/userService.ts` looks the user service up through
  Consul and calls that endpoint, with a 3s timeout
- `createConversation` and `addParticipant` resolve every id through the
  gateway and reject ids the user service does not know; the request body no
  longer carries `participantUsernames`, `participantProfilePictures`,
  `creatorUsername` or `creatorProfilePicture`
- the web client stopped sending those fields

If the user service is unreachable, conversation creation fails with
`res_err_internal_server` rather than silently falling back to a client-supplied
or raw-id name.

## Display name removed

`display_name` was dropped from the user service in migration
`0008_drop_display_name_and_make_username_nullable.sql`, which merged its values
into `username`. The chat service's `IParticipant.displayName` was a leftover
from before that: only ever populated from a client request body, never sent by
the web, always `null`. Removed from the model, schema, service, handlers and
the web types.

Existing conversation documents keep a stale `displayName` key in Mongo.
Mongoose ignores unknown keys, so nothing breaks; a `$unset` cleanup can be run
whenever convenient.

## Type mismatches

Fixed against the schema in `backend/*/migrations`:

| Field | Was | Now |
|---|---|---|
| `Transaction.metadata` | Go `*string` vs TS object | Go `json.RawMessage` — the column is JSONB |
| `Transaction.description` | declared in TS | removed — no such column |
| `Account.updatedAt` | declared in TS | removed — no such column |
| `Transaction.updatedAt` | declared in TS | removed — no such column |
| `Payment.updatedAt` | declared in TS | removed — no such column |
| `Transaction.actorId` | TS `string` | TS `string \| null` — `actor_id` is nullable |
| `Account.ownerId` | TS `string` | TS `string \| null` — null for platform/escrow/fee |
| `Payment.transactionId` | TS `string \| null` | TS `string` — column is NOT NULL |
| `Payment.providerReference` | TS `string \| null` | TS `string` — column is NOT NULL |
| `DepositResponse.checkoutUrl` | TS `string \| null` | TS `string` — both gateways always return one |
| `PaymentProvider` | TS, OpenAPI and i18n had `paypal` | `stripe` + `mock`, matching the registered gateways |
| `VOD.originalFileUrl` | Go `omitempty` vs TS `\| null` | Go always serializes it |
| `Notification.actionUrl/actionLabel/referenceId` | TS optional | TS `\| null` — Go has no `omitempty` |
| `CommentUser.username` | TS `string \| null` | TS `string` |
| `ConversationParticipant.username` | TS `string \| null` | TS `string` |
| `DmMessage.imageUrls` | TS optional | TS `string[]` — schema default `[]` |
| `DmMessage.replyTo` | TS optional | TS `string \| null` — schema default `null` |

## Chat event types

`types/dm-event.ts` declared `DmSendMessageEvent` / `DmTypingEvent`, but every
handler took `data: any`, so the declared contract was never enforced and the
runtime read fields the types did not have. Handlers are now typed against the
`DmClientEvent` union, with one guarded narrowing at the WebSocket entry point.

### On paypal

`paypal` was only ever a name. No PayPal gateway has existed at any point, and
`DepositService.Initiate` writes a `payments` row only after
`s.gateways[provider]` resolves, so a row with `provider = 'paypal'` cannot have
been created even while the request validator still accepted the value. The name
is now gone from the domain const, the client enum, `docs/openapi.yaml` and the
wallet locales.
