# Chat Commands

Slash chat-command system for the live chat. Lets viewers run built-in chat-command shortcuts (e.g. `/me`, `/roll`) and lets users define their own custom chat commands at two scopes:

- **User-scoped** — created by a user, available to that user in any channel they chat in.
- **Channel-scoped** — created by a streamer, available to any viewer chatting in that streamer's channel.

Built-in chat commands are hardcoded on the client and always available.

---

## Architecture

```
                                                                          ┌──────────────┐
       Sender's browser                                                   │ Other viewer │
┌───────────────────────────┐                                             └──────┬───────┘
│  ChatPanel (chat.tsx)     │                                                    │
│  ┌─────────────────────┐  │                                                    │
│  │ user types "/roll"  │  │                                                    │
│  └──────────┬──────────┘  │                                                    │
│             ▼             │                                                    │
│  ┌───────────────────────┐│   GET /v1/chat-commands?roomId=<streamerId>        │
│  │ chat-command-parser.ts│◄┼──────────── (REST, on mount) ──────────────►      │
│  │  - built-ins          ││                                                    │
│  │  - custom registry    ││                                                    │
│  └──────────┬────────────┘│                                                    │
│             │             │                                                    │
│             │ expands to plain text ("🎲 rolled 42 (1-100)")                   │
│             ▼             │                                                    │
│  ┌─────────────────────┐  │   wss /v1/ws  →  ChatServer  →  Redis pub/sub      │
│  │ wsRef.current.send  │──┼──────────────────────────────────────────────────► │
│  └─────────────────────┘  │                                                    │
└───────────────────────────┘                                                    │
                                                                                 ▼
                                                                  receiver renders text
```

**Key design choice:** all chat-command parsing happens client-side. Chat commands expand to plain message text *before* hitting the WebSocket. The chat backend never learns that chat commands exist on the realtime path — the only new backend surface is a CRUD API for the custom chat-command registry.

This keeps the realtime path untouched, eliminates a class of message-shape changes, and avoids introducing a new "system message" type. Trade-off: no server-authoritative behavior (e.g. `/roll` is trivially riggable by a malicious client). Acceptable for this scale.

---

## Files added / changed

### Backend (`backend/chat/`)

| File | Purpose |
|---|---|
| `src/models/ChatCommand.ts` | Mongoose schema. Indexed `(scope, ownerId, name)` unique. Mongo collection `chat_commands`. |
| `src/services/chatCommandService.ts` | `listForRoom`, `listMine`, `create`, `delete` with ownership + validation. |
| `src/handlers/chatCommandHandler.ts` | Express handlers. |
| `src/index.ts` | Route wiring. |

### Gateway

| File | Change |
|---|---|
| `configs/kong.yml` | Added `Chat_Commands` route under the `Chat` service exposing `/chat-commands`. |

### Frontend (`web/`)

| File | Purpose |
|---|---|
| `types/chat-command.ts` | `ChatCommand`, `ChatCommandScope`, `MyChatCommands` types. |
| `lib/api/chat-command.ts` | Fetch wrappers: `GetRoomChatCommands`, `GetMyChatCommands`, `CreateChatCommand`, `DeleteChatCommand`. |
| `utils/chat-parser/command.ts` | Built-ins, registry merging, parser, autocomplete filter, help renderer. All localizable strings deferred to consumer via `t()`. |
| `utils/chat-parser/emote.tsx` | Emote shortcode parser (was `utils/emote-parser.tsx` — moved here so all chat-text parsing lives in one folder). |
| `utils/chat-parser/index.ts` | Barrel re-export for both. |
| `components/chat-command-suggestions.tsx` | Autocomplete dropdown. |
| `app/[lng]/(main)/users/[userId]/chat.tsx` | Integrates parser + autocomplete + local system messages + italic `_text_` rendering. |
| `app/[lng]/(main)/settings/chat-commands/page.tsx` | Manage personal + channel chat commands; lists built-ins for reference. |
| `app/[lng]/(main)/settings/layout.tsx` | New nav entry. |
| `lib/i18n/locales/en/chat-commands.json` + `vi/chat-commands.json` | New i18n namespace. |
| `lib/i18n/locales/{en,vi}/settings.json` | New `navigation.chat_commands` key. |

### Mobile (`mobile/`)

| File | Purpose |
|---|---|
| `lib/models/chat_command.dart` | `ChatCommand`, `ChatCommandScope`, `MyChatCommands`. |
| `lib/features/livestream/data/chat_command_repository.dart` | REST repo: `listForRoom`, `listMine`, `create`, `delete`. |
| `lib/core/network/api_endpoints.dart` | Added `chatCommands*` endpoint constants. |
| `lib/providers.dart` | Added `chatCommandRepositoryProvider` (Riverpod). |
| `lib/core/chat_parser/command.dart` | Built-ins, parser, suggestion index, help renderer. Strings deferred to consumer via `AppLocalizations`. |
| `lib/core/chat_parser/emote.dart` | Emote parser (was `lib/core/emotes/emote_parser.dart` — moved to mirror web's `chat_parser/` folder). |
| `lib/shared/widgets/chat_command_suggestions.dart` | Autocomplete dropdown widget. |
| `lib/features/settings/presentation/chat_commands_settings_screen.dart` | Manage personal + channel chat commands; lists built-ins. |
| `lib/features/settings/presentation/settings_screen.dart` | New tile. |
| `lib/core/router/app_router.dart` | New `settingsChatCommands` route. |
| `lib/features/livestream/presentation/livestream_screen.dart` | Integrates parser + autocomplete + local system lines + italic `_text_` rendering. |
| `l10n/app_en.arb` + `l10n/app_vi.arb` | All new `chatCommands*` and `settingsNavChatCommands` keys. Regenerated dart bindings via `flutter gen-l10n`. |

---

## Data model

```
collection: chat_commands  (MongoDB)
  scope:       "user" | "channel"           (indexed)
  ownerId:     string                        (indexed; userId for both scopes)
  name:        string  /^[a-z0-9_-]{1,32}$/
  response:    string  ≤ 500 chars
  description: string  ≤ 120 chars (optional)
  createdAt:   Date

unique index: (scope, ownerId, name)
```

`ownerId` semantics:
- `scope: "user"`   → the creator's userId. The chat command activates only when *that* user is the sender.
- `scope: "channel"` → the streamer's userId. The chat command activates for any sender in the streamer's chat room. (`roomId === streamerId` in this codebase — verified at `web/app/[lng]/(main)/users/[userId]/page.tsx:155`.)

Cap: 50 chat commands per `(scope, ownerId)` pair.

---

## REST API

All routes mounted on the chat service (`/v1`), exposed through Kong.

| Method | Path | Auth | Purpose |
|---|---|---|---|
| `GET` | `/v1/chat-commands?roomId=<id>` | optional | Returns merged list: channel chat commands for `roomId` + (if cookie present) the caller's user chat commands. |
| `GET` | `/v1/chat-commands/mine` | required | Returns `{ user: [...], channel: [...] }` for the caller's own chat commands. |
| `POST` | `/v1/chat-commands` | required | Body `{ scope, name, response, description? }`. Creator owns the chat command. |
| `PATCH` | `/v1/chat-commands/:id` | required | Body `{ name?, response?, description? }`. Caller must be the owner; scope is immutable. |
| `DELETE` | `/v1/chat-commands/:id` | required | Caller must be the owner; otherwise 403. |

Auth uses the existing `authMiddleware` (cookie-based JWT). The Kong route does **not** enable the `jwt` plugin so the public `GET` works for anonymous viewers.

---

## Built-in chat commands

| Chat command | Behavior |
|---|---|
| `/me <text>` | Wraps text in `_..._`; renderer shows it italic, name without colon. |
| `/shrug [text]` | Appends `¯\_(ツ)_/¯`. |
| `/tableflip [text]` | Appends `(╯°□°)╯︵ ┻━┻`. |
| `/unflip [text]` | Appends `┬─┬ ノ( ゜-゜ノ)`. |
| `/roll [max]` | Sends `🎲 rolled N (1-max)`. Default max 100, capped at 1,000,000. Localized at send time using the sender's locale. |
| `/help` | Renders a local-only system message listing all available chat commands. Never sent to the wire. |

Custom chat commands are pure text expansions: `/discord` → "Join us at discord.gg/example".

---

## Client-side flow

1. **On chat mount** — `GET /v1/chat-commands?roomId=<roomId>` populates a local registry of channel + user chat commands. Anonymous viewers still get channel chat commands.
2. **On keystroke** — if input starts with `/` and contains no space, `filterChatCommandSuggestions` matches by prefix (max 8) and renders the autocomplete dropdown. Arrow keys cycle, Tab applies, Esc dismisses.
3. **On submit** —
   - Plain text → sent as a normal chat message (existing path).
   - Starts with `/` → `parseChatCommand`:
     - `noop` → swallow (e.g. empty `/me`).
     - `error` → render a local-only system message visible only to sender.
     - `help` → render local help text.
     - `send` → forward expanded text via WebSocket like any other message.
4. **On render** — text matching `^_(.+)_$` is shown italic with the colon dropped after the username (the `/me` style).

`buildChatCommandIndex` and `buildChatCommandHelpText` accept a `t` function so all surfaced strings (descriptions, headers, errors, the `/roll` message, the suggestion source label) are localized.

---

## Internationalization

A dedicated `chat-commands` namespace lives at `web/lib/i18n/locales/{en,vi}/chat-commands.json`. Includes:

- `page.*` — settings-page strings (titles, labels, placeholders, toasts).
- `builtin.*` — descriptions for each built-in chat command.
- `errors.*` — `unknown`, `roll_usage`.
- `help.header`, `help.custom_header`.
- `roll_message` — interpolates `{value, max}`.
- `source.{builtin,user,channel}` — autocomplete source label.

The settings nav label lives in the existing `settings.navigation.chat_commands` key for consistency with the other nav entries.

The parser exports a `descriptionKey` per built-in chat command instead of a translated string, so consumers (chat panel, settings page, suggestions component) call `t(c.descriptionKey)` themselves. This keeps the parser pure and avoids forcing it to import an i18n instance.

---

## Validation

**Client (`page.tsx`):** name regex `^[a-z0-9_-]{1,32}$`, response non-empty, description ≤ 120 chars. Toasts on failure.

**Server (`chatCommandService.ts`):** identical name regex, length caps, scope enum check, per-scope creation cap (50), uniqueness via the Mongo index (catches `code: 11000` and returns invalid-input).

Both layers enforce the same rules so a malicious client can't bypass.

---

## Trade-offs and things deliberately not done

- **Client-side `/roll`** — not server-authoritative; a determined user could rig it. Trivial to address later by moving roll execution to the chat server, but out of scope for "simple feature."
- **No moderation hooks** — no block-list of reserved names, no rate limiting, no abuse reporting on custom chat commands. The 500-char `response` cap and 50-per-scope cap are the only protections.
- **Edit endpoint** — `PATCH /v1/chat-commands/:id` updates name/response/description in place; scope and ownerId are immutable. Settings UI exposes it via a pencil button on each tile.
- **Custom chat-command shadowing** — built-ins win over user/channel chat commands of the same name. There is no warning when creating a custom chat command that shadows a user-scoped one with the same name in the channel-scope set (or vice versa); the parser just checks built-ins first, then customs in list order.
- **No "kind" field on chat messages** — the existing `ChatMessage` shape (`type: join|leave|message`, plus `text`) is unchanged. `/me` italic styling rides on a marker convention (`_text_`) in the message body. Pragmatic, but means anyone typing `_foo_` gets italicized too.
- **Help/error messages are local-only** — sender sees them; no one else does.

---

## Testing checklist

- [ ] `docker compose up` and visit a stream's chat as a logged-in viewer.
- [ ] Type `/` — autocomplete dropdown appears with built-in chat commands.
- [ ] `/me waves` — appears as italic action to all viewers.
- [ ] `/roll 6` — renders localized rolled message; reload in `vi` and re-roll to confirm sender locale.
- [ ] `/help` — local-only system message listing chat commands; not visible to other viewers.
- [ ] `/unknownthing` — local-only error.
- [ ] Visit `/settings/chat-commands` — both scopes show, built-ins listed.
- [ ] Add a channel chat command `/discord` → response → reload chat → `/discord` works for a different viewer.
- [ ] Try to delete another user's chat command via direct API call → 403.
- [ ] Switch `lng` cookie between `en` and `vi` → all chat-command UI re-renders translated.

---

## File index

```
backend/chat/src/models/ChatCommand.ts
backend/chat/src/services/chatCommandService.ts
backend/chat/src/handlers/chatCommandHandler.ts
backend/chat/src/index.ts                                     (routes wired)
configs/kong.yml                                              (Chat_Commands route)

web/utils/chat-parser/index.ts                                (barrel)
web/utils/chat-parser/command.ts
web/utils/chat-parser/emote.tsx                               (moved from utils/emote-parser.tsx)
web/types/chat-command.ts
web/lib/api/chat-command.ts
web/components/chat-command-suggestions.tsx
web/app/[lng]/(main)/users/[userId]/chat.tsx                  (integration)
web/app/[lng]/(main)/settings/chat-commands/page.tsx          (new page)
web/app/[lng]/(main)/settings/layout.tsx                      (nav entry)
web/lib/i18n/locales/en/chat-commands.json
web/lib/i18n/locales/vi/chat-commands.json
web/lib/i18n/locales/en/settings.json                         (navigation.chat_commands)
web/lib/i18n/locales/vi/settings.json                         (navigation.chat_commands)

mobile/lib/core/chat_parser/command.dart
mobile/lib/core/chat_parser/emote.dart                        (moved from core/emotes/emote_parser.dart)
mobile/lib/models/chat_command.dart
mobile/lib/features/livestream/data/chat_command_repository.dart
mobile/lib/core/network/api_endpoints.dart                    (chatCommands* added)
mobile/lib/providers.dart                                     (provider added)
mobile/lib/shared/widgets/chat_command_suggestions.dart
mobile/lib/features/settings/presentation/chat_commands_settings_screen.dart
mobile/lib/features/settings/presentation/settings_screen.dart  (new tile)
mobile/lib/core/router/app_router.dart                        (route added)
mobile/lib/features/livestream/presentation/livestream_screen.dart  (integration)
mobile/l10n/app_en.arb
mobile/l10n/app_vi.arb
mobile/lib/l10n/app_localizations*.dart                       (regenerated)
```
