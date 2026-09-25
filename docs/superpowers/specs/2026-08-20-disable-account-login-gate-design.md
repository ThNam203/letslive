# Disable-Account Login Gate Design

**Date:** 2026-08-20
**Branch:** feat/disable-account-enforcement
**Author:** ThNam203

---

## Overview

Two prior commits (`de7188c`, `8a328bb`) built self-service account disable/reactivate but left the security-critical half missing: a disabled account can still log in and get a fully valid session on every path (password, Google OAuth web, Google OAuth mobile, token refresh). The post-login `ReactivateAccountDialog` is currently cosmetic — nothing backend-enforced sits behind it.

This design closes that gap: a disabled account must not be able to log in normally. On any login attempt, the account holder is asked whether they want to reactivate. There is no hard time limit on disablement — they can choose to reactivate at any point, no expiry (verified: no cron/TTL/cleanup job anywhere in the codebase touches `disabled` status; it is a plain flag with no expiry mechanism).

---

## Decisions Made

| Question | Decision |
|---|---|
| Mechanism for prompting reactivation at login | Short-lived, stateless reactivation token (JWT), not resend-credentials |
| Why not resend-credentials | Google web OAuth's authorization `code` is single-use — can't resubmit it. Token approach is the only one that works uniformly across password/Google-web/Google-mobile |
| User-service changes required | None — reuse existing `GET /v1/user/{userId}` (status read) and `PUT /v1/user/{userId}` (status write, already internal/unauthenticated at the network trust boundary) via direct service-discovery calls, same pattern as the existing `CreateNewUser` gateway call |
| Reactivation token TTL | ~10 minutes — bounds only this login attempt, not how long the account can stay disabled. If it expires, the user just logs in again and gets a fresh one |
| Refresh-token gating | Included — closes the gap where a second device stays logged in indefinitely after disable, by reusing the same status-check gateway call |
| Existing post-login `ReactivateAccountDialog` / self-disable flow | Untouched — still valid defense for the narrow window where a second device's already-issued access token hasn't expired yet |
| CAPTCHA/rate-limit on new `/auth/reactivate` endpoint | Not added — the token is unguessable (signed JWT) and short-lived; it's only reachable after a prior login attempt already passed Turnstile (password) or Google's own auth (OAuth) |

---

## Architecture

No new services, no new database tables/columns. All changes are in the `auth` service (Go) and the web frontend. The `user` service's HTTP surface is unchanged — the two internal endpoints it already exposes are reused as-is.

### Status check point (new: `UserGateway.GetUserStatus`)

All three login paths already resolve a `userId` after validating credentials/identity, before calling `setAuthJWTsInCookie`:

- `AuthHandler.LogInHandler` → `AuthService.GetUserFromCredentials` → `auth.UserId`
- `AuthHandler.OAuthGoogleCallBackHandler` → `GoogleAuthService.CallbackHandler` → `createdAuth.UserId`
- `AuthHandler.OAuthGoogleMobileHandler` → `GoogleAuthService.VerifyIDTokenAndGetUser` → `createdAuth.UserId`

A new `UserGateway.GetUserStatus(ctx, userId) (string, *Response)` method calls the user service's existing `GET /v1/user/{userId}` directly via `registry.ServiceAddress(ctx, "user")` (same as `CreateNewUser` today — bypasses Kong, no new user-service code) and extracts `.Status` from the response DTO.

### Reactivation token (new: two `JWTService` methods)

Same file/pattern as the existing `GenerateTokenPair`/`RefreshToken`:

```go
func (c *JWTService) GenerateReactivationToken(ctx context.Context, userId string) (string, *serviceresponse.Response[any])
func (c *JWTService) VerifyReactivationToken(ctx context.Context, token string) (userId string, err *serviceresponse.Response[any])
```

- Claims: just `UserId` + standard registered claims (`ExpiresAt` ~10 min out).
- Signed with a new secret `REACTIVATION_TOKEN_SECRET` (own secret, not shared with access/refresh tokens, so it can't be confused with or substituted for a session token).
- Stateless — no DB row, unlike refresh tokens (no revocation needed; it's single-purpose and short-lived).

### Gated login paths

For each of the three login handlers, after resolving `userId` and before `setAuthJWTsInCookie`:

```go
status, err := h.userGateway.GetUserStatus(ctx, userId)
if err != nil { /* existing error handling */ }
if status == "disabled" {
    token, err := h.jwtService.GenerateReactivationToken(ctx, userId)
    // password + mobile: return RES_ERR_ACCOUNT_DISABLED with token in Data
    // web OAuth: redirect to {CLIENT_URL}/login?accountDisabled=true&reactivationToken=...
    return // no cookies set
}
// existing success path unchanged
```

### Refresh-token gating

`JWTService.RefreshToken` already parses `myClaims.UserId` from the refresh token before minting a new access token. Insert the same `GetUserStatus` check there; if disabled, return `RES_ERR_ACCOUNT_DISABLED` instead of a new access token — same code as the login paths, so any current or future frontend handling can treat "disabled" as one consistent signal rather than a generic 401. The frontend's refresh-failure path already clears local session state on any refresh error, so a disabled account's other device gets logged out on its next silent refresh and, if it wants back in, goes through the same login gate.

### New endpoint: `POST /auth/reactivate`

Request: `{ "reactivationToken": string }`

1. `jwtService.VerifyReactivationToken` → `userId` (rejects expired/invalid/wrong-secret tokens).
2. `userGateway.UpdateUserStatus(ctx, userId, "normal")` → wraps the existing internal `PUT /v1/user/{userId}` with `{status: "normal"}` (same call shape as `CreateNewUser`/the existing internal update handler — no new user-service code).
3. On success: `setAuthJWTsInCookie` (identical to a normal login completion) + `RES_SUCC_LOGIN`.
4. On invalid/expired token: `RES_ERR_UNAUTHORIZED` — frontend tells the user to just log in again.

### Frontend

- **Login page** (`web/app/[lng]/(auth)/login/page.tsx`): new `useEffect` reads `accountDisabled` + `reactivationToken` query params (same pattern already used there for `errorMessage`/`redirectUrl`), opens a new dialog.
- **New component** `AccountDisabledDialog` (sibling to the existing `ReactivateAccountDialog`, but pre-session — it doesn't read from `useUser`, it holds the token from the query param / API response in local state):
  - Confirm → `POST /auth/reactivate` → on success, same completion as `LogInForm`'s success path (`GetMeProfile()` + `setUser` + redirect).
  - Decline → close dialog, reset form. No API call — nothing was ever created server-side, so there's nothing to undo.
- **`LogInForm`** and the mobile Google sign-in call site: on receiving `RES_ERR_ACCOUNT_DISABLED` with a `reactivationToken` in `data`, open the same dialog with that token instead of toasting a generic error.
- **`web/lib/api/auth.ts`**: new `Reactivate({reactivationToken})` calling `POST /auth/reactivate`.
- Existing `ReactivateAccountDialog` (post-login) and `disable-account-dialog.tsx` (settings self-disable): no changes.

### New response code

`backend/auth/response/error.go`: `RES_ERR_ACCOUNT_DISABLED` (code `20019`, key `res_err_account_disabled`, HTTP 403), following the exact shape of the existing `RES_ERR_*` templates in that file.

### i18n

New keys for the account-disabled-at-login dialog, added to `web/lib/i18n/locales/{en-US,vi-VN}/auth.json` — `LogInForm` already loads the `auth` namespace, so this dialog's copy lives there rather than in `settings.json`. Mirrors the existing `reactivate.dialog.*` keys added in `8a328bb`, since this is the same choice (reactivate vs. decline) at a different point in the flow.

### Config/deployment

`REACTIVATION_TOKEN_SECRET` needs to be added alongside `REFRESH_TOKEN_SECRET`/`ACCESS_TOKEN_SECRET` in: `example.env`, `docker-compose.yaml`, `docker-compose-dev.yaml`, `.github/workflows/deploy.yml`.

---

## Data Flow (password login, disabled account)

```
LogInForm.handleSubmit
  → POST /auth/login {email, password, turnstileToken}
  → AuthHandler.LogInHandler
      → AuthService.GetUserFromCredentials  (bcrypt check, as today)
      → UserGateway.GetUserStatus(userId)     [NEW]
      → status == "disabled"
      → JWTService.GenerateReactivationToken(userId)   [NEW]
      ← RES_ERR_ACCOUNT_DISABLED { data: { reactivationToken } }
  ← LogInForm shows AccountDisabledDialog instead of a toast
      Confirm → POST /auth/reactivate {reactivationToken}
        → JWTService.VerifyReactivationToken   [NEW]
        → UserGateway.UpdateUserStatus(userId, "normal")   [NEW]
        → setAuthJWTsInCookie (cookies set — login completes)
        ← RES_SUCC_LOGIN
      → GetMeProfile() + setUser + router.push("/")
      Decline → dialog closes, no request sent
```

Google web OAuth differs only in transport: the disabled branch redirects to `/login?accountDisabled=true&reactivationToken=...` instead of returning JSON, and the login page's existing query-param effect picks it up.

---

## Error Handling

- Invalid/expired reactivation token at `/auth/reactivate` → `RES_ERR_UNAUTHORIZED`, no partial state change (status is only updated after the token verifies).
- `GetUserStatus` gateway call itself fails (network/user-service down) → existing `RES_ERR_INTERNAL_SERVER` pattern, login attempt fails closed (no cookies set) rather than failing open — a temporarily-unreachable user service must not let a disabled account slip through.
- `UpdateUserStatus` fails after a valid token → `RES_ERR_INTERNAL_SERVER`, no cookies set, user can retry (token may still be valid within its TTL).

---

## Testing Plan

Repo currently has zero test coverage in `backend/auth`, `backend/user` (status paths), or anywhere in `web/`. Per the mandatory TDD workflow, tests are written first for every new/changed unit:

**Go (table-driven, `-race`, target 80%+ on touched packages):**
- `AuthService`/handler tests: login succeeds for `normal` status (existing behavior preserved), blocked with `RES_ERR_ACCOUNT_DISABLED` + valid token for `disabled` status, fails closed when the status-gateway call errors.
- `JWTService` tests: `GenerateReactivationToken`/`VerifyReactivationToken` round-trip, rejects expired token, rejects token signed with the wrong secret, rejects a well-formed access/refresh token presented at the reactivate endpoint (wrong purpose/secret).
- `RefreshTokenHandler` test: refresh rejected for a disabled account's refresh token.
- New `/auth/reactivate` handler test: valid token → status flipped + cookies set; expired/invalid token → 401, no status change.
- Google OAuth web/mobile handler tests: same disabled-branch behavior, mocking `GoogleAuthService`.

**Frontend:**
- `AccountDisabledDialog` unit test: renders on the right trigger, confirm calls `Reactivate` and completion path, decline calls nothing.
- `LogInForm` test: `RES_ERR_ACCOUNT_DISABLED` response opens the dialog instead of toasting.
- Manual verification in-browser for both password and Google web OAuth paths (Google's OAuth redirect flow is impractical to fully automate; documented as a manual test step).

---

## Out of Scope (explicitly deferred, not silently dropped)

- Instant revocation of an already-issued, unexpired access token on a second device at the moment of disable (pre-existing property of the JWT scheme, not something this feature introduces or is positioned to fix — would require access-token revocation infrastructure, a separate and larger piece of work).
- The pre-existing `// TODO: revoke refresh token` in `LogOutHandler` and the unused `JWTService.RevokeAllTokensOfUser` — real gap, but predates this feature and is separate from it.
- Read-path filtering for other still-unfiltered user listings (profile lookup, following list, livestream/VOD discovery) and action gating (stream start, chat, comments, follow) — separate punch-list items, not part of the login gate.
- Any admin/moderator disable capability or a distinct "banned" status — explicitly deferred by product decision; only `normal`/`disabled` exist.
