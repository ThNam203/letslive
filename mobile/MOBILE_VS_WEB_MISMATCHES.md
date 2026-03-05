# Mobile vs Web vs Backend - Mismatch Report

This document explains every difference found between the mobile app (Flutter/Dart),
the web app (Next.js/TypeScript), and the backend (Go + Node.js) in terms of
API endpoints, models, query parameters, HTTP methods, and feature completeness.

---

## 2. CRITICAL: Mobile Uses Wrong Path for User VODs

**Backend route:**
`GET /v1/vods?userId={id}&page=0&limit=10` (query parameter style)
Registered in `backend/livestream/api/http.go` line 52.

**Web approach (correct):**
```typescript
// ui/lib/api/vod.ts
`/vods?userId=${userId}&page=${page}&limit=${limit}`
```
Uses query parameters — matches the backend.

**Mobile approach (wrong):**
```dart
// mobile/lib/core/network/api_endpoints.dart
static String userVods(String id) => '/user/$id/vods';
```
Uses a path-based URL `/user/{id}/vods` which does NOT exist on the backend.

**Impact:**
When a mobile user views another user's profile and tries to see their VODs,
the request goes to a non-existent route and returns 404. The web works fine
because it uses the correct query-parameter approach.

**Fix:** Mobile should call `/vods?userId={id}&page=...&limit=...` instead.

---

## 3. CRITICAL: Mobile Uses PATCH Instead of PUT for Profile Update

**Backend:**
`PUT /v1/user/me` — registered in `backend/user/api/http.go` line 64.
Go's `http.ServeMux` is strict about method matching.

**Web (correct):**
```typescript
// ui/lib/api/user.ts
fetchClient("/user/me", { method: "PUT", body: JSON.stringify(user) })
```

**Mobile (wrong):**
```dart
// mobile/lib/features/user/data/user_repository.dart line 92
return _client.patch(ApiEndpoints.userMe, data: data, ...);
```
Uses `PATCH` instead of `PUT`.

**Impact:**
Profile updates (display name, bio, phone number, social links) will fail
on mobile with a 405 Method Not Allowed. The web works correctly with PUT.

**Fix:** Change `_client.patch` to `_client.put` in mobile's `updateProfile()`.

---

## 4. CRITICAL: Mobile Doesn't Send CAPTCHA Token on Login

**Backend requires it:**
```go
// backend/auth/dto/login_request.go
TurnstileToken string `json:"turnstileToken" validate:"required,lte=2048"`
```
The `validate:"required"` tag means the backend will reject requests without it.

**Web (correct):**
```typescript
// ui/lib/api/auth.ts
body: JSON.stringify({ email, password, turnstileToken })
```
Web uses Cloudflare Turnstile widget to get a token and sends it.

**Mobile (missing):**
```dart
// mobile/lib/features/auth/data/auth_repository.dart
data: { 'email': email, 'password': password }
```
No `turnstileToken` field at all.

**Same issue for verify-email:**
Backend requires `turnstileToken` for the email verification request too.
Web sends it, mobile doesn't.

**Impact:**
Mobile login will always fail with a validation error (code 20000 or 20003).
Same for the email verification step during signup.

**Why this is different from web:**
Web runs in a browser where Cloudflare Turnstile can render its challenge widget.
Mobile apps can't use Turnstile the same way. The backend likely needs a separate
flow for mobile (e.g., skip captcha for mobile clients, use a different challenge,
or use an API key to identify trusted mobile clients).

**Fix options:**
1. Backend: Make `turnstileToken` optional for mobile clients (identified by User-Agent or a header)
2. Backend: Create a separate mobile auth endpoint that doesn't require captcha
3. Mobile: Integrate a WebView-based captcha solution

---

## 5. MEDIUM: Mobile Uses Wrong Pagination Param for Conversations

**Backend (chat service):**
```typescript
// backend/chat/src/handlers/conversationHandler.ts
const page = parseInt(req.query.page as string) || 0
const limit = parseInt(req.query.limit as string) || 20
```
Expects `page` and `limit`.

**Web (correct):**
```typescript
`/conversations?page=${page}&limit=${limit}`
```

**Mobile (wrong):**
```dart
// mobile/lib/features/messages/data/message_repository.dart
queryParameters: { 'page': page, 'page_size': pageSize }
```
Sends `page_size` instead of `limit`.

**Impact:**
The `page_size` parameter is ignored by the backend, which defaults `limit` to 20.
Pagination will technically "work" but always returns 20 items regardless of what
the mobile app requests. If the app requests a different page size, it won't be respected.

**Fix:** Change `page_size` to `limit` in mobile's `getConversations()`.

---

## 6. CRITICAL: Mobile Uses Offset Pagination for DM Messages, Backend Uses Cursor

**Backend (chat service):**
```typescript
// backend/chat/src/handlers/dmMessageHandler.ts
const before = req.query.before as string | undefined  // message ID cursor
const limit = parseInt(req.query.limit as string) || 50
```
Uses **cursor-based** pagination: pass `before={lastMessageId}` to get older messages.

**Web (correct):**
```typescript
// ui/lib/api/dm.ts
let url = `/conversations/${id}/messages?limit=${limit}`;
if (before) url += `&before=${before}`;
```
Correctly uses cursor-based pagination.

**Mobile (wrong):**
```dart
// mobile/lib/features/messages/data/message_repository.dart
queryParameters: { 'page': page, 'page_size': pageSize }
```
Uses offset-based pagination with `page` and `page_size`.

**Impact:**
- `page` and `page_size` are completely ignored by the backend
- Mobile will always get the latest 50 messages (the default)
- Scrolling up to load older messages will keep returning the same 50 messages
- DM message history is effectively broken on mobile

**Why they're different:**
Cursor pagination is better for real-time chat (messages can be added/deleted
between requests). The web correctly implements this. Mobile needs to track the
oldest message ID and pass it as `before` to load older messages.

**Fix:** Change mobile's `getConversationMessages()` to accept a `before` parameter
(the `_id` of the oldest loaded message) and pass it as a query parameter instead
of `page`/`page_size`.

---

## 7. HIGH: Mobile Uses PATCH Instead of POST for Mark-Read

**Backend:**
`POST /v1/conversations/:id/read` — registered in `backend/chat/src/index.ts` line 91.

**Web (correct):**
```typescript
// ui/lib/api/dm.ts
fetchClient(`/conversations/${id}/read`, { method: "POST", ... })
```

**Mobile (wrong):**
```dart
// mobile/lib/features/messages/data/message_repository.dart
return _client.patch(ApiEndpoints.conversationRead(id));
```

**Impact:**
Marking conversations as read will fail on mobile. Unread counts will never
decrease, and read receipts won't be sent.

**Fix:** Change `_client.patch` to `_client.post` in mobile's `markConversationRead()`.

---

## 8. HIGH: Mobile Uses Wrong Query Param for User Search

**Backend:**
```go
// backend/user/handlers/user/search_users_public.go
username := r.URL.Query().Get("username")
```
Expects `?username=...`

**Web (correct):**
```typescript
// ui/lib/api/user.ts
`/users/search?username=${encodeURIComponent(query)}`
```

**Mobile (wrong):**
```dart
// mobile/lib/features/user/data/user_repository.dart
queryParameters: { 'query': query, 'page': page, 'page_size': pageSize }
```
Sends `query` instead of `username`.

**Impact:**
The backend reads the `username` query param, which mobile never sends.
The search will always receive an empty string and return no results (or all users).
Search is completely broken on mobile.

**Additional note:** The backend doesn't support `page`/`page_size` for search either —
it returns all matching results in one response. The web doesn't send pagination
params for search. Mobile sends them but they're ignored.

**Fix:** Change `'query'` to `'username'` in mobile's `searchUsers()`.

---

## 9. MOBILE MISSING FEATURES (compared to web)

These are features the web app has that the mobile app hasn't implemented yet.
They're not "bugs" but feature gaps.

### VOD Comments (partially missing)
| Feature | Web | Mobile |
|---------|-----|--------|
| View VOD comments | Yes | Endpoint defined but repo not implemented |
| Create comment | Yes (`POST /vods/{id}/comments`) | Not implemented |
| Delete comment | Yes (`DELETE /vod-comments/{id}`) | Not implemented |
| Like/Unlike comment | Yes | Endpoint defined, repo not implemented |
| Get liked comment IDs | Yes | Endpoint defined, repo not implemented |

### Direct Messages (mostly missing)
| Feature | Web | Mobile |
|---------|-----|--------|
| List conversations | Yes | Yes (but wrong pagination param) |
| View messages | Yes | Yes (but wrong pagination strategy) |
| Send message (REST) | Yes (`POST .../messages`) | Not implemented |
| Edit message | Yes (`PATCH .../messages/{id}`) | Not implemented |
| Delete message | Yes (`DELETE .../messages/{id}`) | Not implemented |
| Update conversation | Yes (`PUT /conversations/{id}`) | Not implemented |
| Leave conversation | Yes (`DELETE /conversations/{id}`) | Not implemented |
| Add participant | Yes | Not implemented |
| Remove participant | Endpoint defined | Not implemented |
| DM WebSocket | Full (typing, presence, real-time) | Not implemented |
| Unread counts | Yes | Yes (endpoint exists) |

### Other
| Feature | Web | Mobile |
|---------|-----|--------|
| Google OAuth | Yes (browser redirect) | Not implemented |
| Cloudflare Turnstile | Yes (browser widget) | Not applicable (see issue #4) |
| File upload utility | Yes (`/upload-file`) | Endpoint defined but not used |

---

## 10. LOW: Mobile Models Have Extra Fields Not in Backend Response

Mobile's `Livestream` and `Vod` models include:
- `username`
- `displayName`
- `profilePicture`

These fields are intended to show the streamer's info alongside the stream/VOD.
However, the backend's `Livestream` and `VOD` domain structs do NOT include
these fields — they only contain the `userId`.

**Backend Livestream struct** (`backend/livestream/domains/livestream.go`):
Only has `userId`, no username/displayName/profilePicture.

**Backend VOD struct** (`backend/livestream/domains/vod.go`):
Only has `userId`, no username/displayName/profilePicture.

**Web's approach:**
The web fetches user info separately when needed (e.g., on the profile page,
the user data is already loaded, and streams are shown in that context).

**Mobile's approach:**
Expects these fields to come in the API response, but they'll always be `null`.

**Impact:**
Not a crash — the fields are nullable. But mobile can't display the streamer's
name or avatar next to livestreams/VODs unless it fetches user info separately.

**Fix options:**
1. Backend: Add a JOIN query that includes user info in livestream/VOD responses
2. Mobile: Fetch user info separately and combine client-side (like web does)

---

## Summary Table

| # | Issue | Mobile | Web | Severity |
|---|-------|--------|-----|----------|
| 1 | `/auth/verify-otp` missing | Broken | Broken | Critical |
| 2 | User VODs wrong path | `/user/{id}/vods` (wrong) | `/vods?userId={id}` (correct) | Critical |
| 3 | Profile update method | PATCH (wrong) | PUT (correct) | Critical |
| 4 | No captcha token | Missing | Sends it | Critical |
| 5 | Conversation pagination | `page_size` (wrong) | `limit` (correct) | Medium |
| 6 | DM messages pagination | Offset-based (wrong) | Cursor-based (correct) | Critical |
| 7 | Mark-read method | PATCH (wrong) | POST (correct) | High |
| 8 | Search query param | `query` (wrong) | `username` (correct) | High |
| 9 | Missing features | Many gaps | Complete | Low |
| 10 | Extra model fields | Always null | N/A | Low |

---

## Recommended Fix Priority

1. **Fix captcha issue** (#4) — Without this, no mobile user can log in at all
2. **Fix profile update method** (#3) — PUT instead of PATCH
3. **Fix search param** (#8) — `username` instead of `query`
4. **Fix user VODs endpoint** (#2) — Use query param style
5. **Fix mark-read method** (#7) — POST instead of PATCH
6. **Fix DM pagination** (#6) — Switch to cursor-based
7. **Fix conversation pagination** (#5) — `limit` instead of `page_size`
8. **Investigate verify-otp** (#1) — May need backend endpoint added
9. **Handle extra model fields** (#10) — Fetch user info separately
10. **Implement missing features** (#9) — Incremental feature parity
