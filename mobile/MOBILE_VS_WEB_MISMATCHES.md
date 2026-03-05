# Mobile vs Web vs Backend - Mismatch Report

This document explains every difference found between the mobile app (Flutter/Dart),
the web app (Next.js/TypeScript), and the backend (Go + Node.js) in terms of
API endpoints, models, query parameters, HTTP methods, and feature completeness.

## 9. MOBILE MISSING FEATURES (compared to web)

These are features the web app has that the mobile app hasn't implemented yet.
They're not "bugs" but feature gaps.

### VOD Comments — FIXED
| Feature | Web | Mobile |
|---------|-----|--------|
| View VOD comments | Yes | Yes — `VodCommentRepository.getComments()` + `VodCommentSection` widget |
| Create comment | Yes (`POST /vods/{id}/comments`) | Yes — `VodCommentRepository.createComment()` + comment form UI |
| Delete comment | Yes (`DELETE /vod-comments/{id}`) | Yes — `VodCommentRepository.deleteComment()` + delete UI |
| Like/Unlike comment | Yes | Yes — `VodCommentRepository.likeComment()`/`unlikeComment()` + optimistic UI |
| Get liked comment IDs | Yes | Yes — `VodCommentRepository.getLikedCommentIds()` |

### Direct Messages — MOSTLY FIXED
| Feature | Web | Mobile |
|---------|-----|--------|
| List conversations | Yes | Yes — `MessagesScreen` with pagination |
| View messages | Yes | Yes — `ConversationScreen` with cursor-based pagination |
| Send message (REST) | Yes (`POST .../messages`) | Yes — `MessageRepository.sendMessage()` + chat input UI |
| Edit message | Yes (`PATCH .../messages/{id}`) | Yes — `MessageRepository.editMessage()` + long-press edit UI |
| Delete message | Yes (`DELETE .../messages/{id}`) | Yes — `MessageRepository.deleteMessage()` + long-press delete UI |
| Update conversation | Yes (`PUT /conversations/{id}`) | Yes — `MessageRepository.updateConversation()` |
| Leave conversation | Yes (`DELETE /conversations/{id}`) | Yes — `MessageRepository.leaveConversation()` |
| Add participant | Yes | Yes — `MessageRepository.addParticipant()` |
| Remove participant | Endpoint defined | Yes — `MessageRepository.removeParticipant()` |
| DM WebSocket | Full (typing, presence, real-time) | Not implemented (REST polling only) |
| Unread counts | Yes | Yes (endpoint exists) |
| New conversation | Yes | Yes — `NewConversationDialog` with user search |

### Other
| Feature | Web | Mobile |
|---------|-----|--------|
| Google OAuth | Yes (browser redirect) | Not implemented |
| Cloudflare Turnstile | Yes (browser widget) | Not applicable (see issue #4) |
| File upload utility | Yes (`/upload-file`) | Endpoint defined but not used |

---

## 10. FIXED: Mobile Screens Now Fetch User Info Separately

Mobile's `Livestream` and `Vod` models still include nullable fields for
`username`, `displayName`, and `profilePicture` (for potential future backend JOINs),
but the UI no longer relies on them.

**Fix applied (option #2 — fetch user info separately, like web does):**
- `HomeScreen`: Fetches user info via `_fetchUsersFor()` after loading livestreams/VODs,
  caches in `_userCache`, and passes `User?` to `_LivestreamCard`/`_VodCard`.
- `LivestreamScreen`: Fetches streamer info via `_fetchStreamer()` into `_streamer`,
  uses it in `_buildStreamInfo()`.
- `VodPlayerScreen`: Fetches VOD owner info via `_fetchVodOwner()` into `_vodOwner`,
  uses it in `_buildVodInfo()` and shows profile picture + name correctly.

Remaining gaps
DM WebSocket (typing indicators, presence, real-time) — still not implemented (REST only)
Google OAuth — platform-specific, not addressed
File upload utility — endpoint exists but no UI usage