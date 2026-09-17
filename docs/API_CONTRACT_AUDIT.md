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

Fixed: `dmMessageService.sendMessage` now resolves the name from the
conversation's participant record and no longer takes a `senderUsername`
argument. Typing events resolve it through the new
`conversationService.getParticipantUsername`. The client no longer sends either
field.

### Phone numbers on unauthenticated endpoints

`GetUserPublicResponseDTO` carried `PhoneNumber`, and all four public user
queries selected `u.phone_number`. It was exposed on `GET /v1/user/{userId}`,
`GET /v1/users/search` and `GET /v1/users/recommendations`, none of which
require auth, and the web client never declared the field.

Fixed: dropped from the DTO and from every public query.

`email` is also returned by these endpoints. It is declared in the client's
`PublicUser` type and used by the UI, so it was left as is — worth a separate
decision.

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
| `PaymentProvider` | TS had `paypal` | `stripe` + `mock`, matching the registered gateways |
| `VOD.originalFileUrl` | Go `omitempty` vs TS `\| null` | Go always serializes it |
| `Notification.actionUrl/actionLabel/referenceId` | TS optional | TS `\| null` — Go has no `omitempty` |
| `CommentUser.username` | TS `string \| null` | TS `string` |
| `ConversationParticipant.username` | TS `string \| null` | TS `string` |
| `ConversationParticipant.displayName` | undeclared | declared `string \| null` |
| `DmMessage.imageUrls` | TS optional | TS `string[]` — schema default `[]` |
| `DmMessage.replyTo` | TS optional | TS `string \| null` — schema default `null` |

## Chat event types

`types/dm-event.ts` declared `DmSendMessageEvent` / `DmTypingEvent`, but every
handler took `data: any`, so the declared contract was never enforced and the
runtime read fields the types did not have. Handlers are now typed against the
`DmClientEvent` union, with one guarded narrowing at the WebSocket entry point.
