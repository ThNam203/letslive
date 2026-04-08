# Security Audit — letslive

_Date: 2026-04-08_

## 🔴 CRITICAL

1. **JWT signature NOT verified in services.** All Go services use `jwt.ParseUnverified()`; chat uses `jwt.decode()`. Files: [user/handlers/utils/cookie.go:28](user/handlers/utils/cookie.go#L28), [livestream/handlers/utils/cookie.go:28](livestream/handlers/utils/cookie.go#L28), [vod/handlers/utils/cookie.go:28](vod/handlers/utils/cookie.go#L28), [backend/chat/src/middlewares/auth.ts:14](backend/chat/src/middlewares/auth.ts#L14). Anyone reaching a service directly (bypassing Kong) can forge any user's identity.

## 🟠 HIGH

2. **Unauthenticated file upload** at `POST /v1/upload-file` ([user/api/http.go:55](user/api/http.go#L55)) — no auth, no MIME/extension validation, MinIO bucket public-read. Malware hosting and possible RCE.

3. **CORS wildcard with credentials** in [kong.yml:431](kong.yml#L431) (`origins: ["*"]`, `credentials: true`). Any website can issue authenticated requests and read responses.

4. **No TLS on backend services** — `useTLS` path hard-errors ([auth/api/http.go:58](auth/api/http.go#L58)). JWT cookies and payloads in plaintext between Kong and services (and to clients if Kong isn't fronted by TLS).

5. **Broken rate limiting** — OTP set to 123/min with `#TODO: change to one` ([kong.yml:76-80](kong.yml#L76-L80)). No brute-force protection on login/signup.

## 🟡 MEDIUM

6. **Finance service routed but unimplemented** — [finance/api/http.go](finance/api/http.go) only exposes `/health` while Kong already routes `/finance`. When endpoints land they must enforce per-user authz, atomic balance updates, and reject negative amounts to avoid double-spend.

7. **Chat conversation updates lack role checks** — any participant can rename/modify ([backend/chat/src/services/conversationService.ts:143](backend/chat/src/services/conversationService.ts#L143)).

8. **No input validation/sanitization** on conversation `name`, `displayName`, `profilePicture` URLs in chat handlers — stored XSS risk on frontends that render them as HTML.

9. **Swagger UI publicly exposed** on port 11111 ([docker-compose.yaml:191-203](docker-compose.yaml#L191-L203)) — leaks full API surface.

10. **CSRF**: combined with finding #3, cookie-based JWT auth has no CSRF token. Fixing CORS mitigates most of it; otherwise add `SameSite=strict` or CSRF tokens.

## 🟢 LOW

11. **Missing security headers** at the gateway: HSTS, X-Frame-Options, X-Content-Type-Options, CSP.

12. **Default admin credentials** for local Grafana/MinIO in [docker-compose.yaml:62-63](docker-compose.yaml#L62-L63) — fine for dev, must not ship to prod.

---

**Suggested fix order:** #1 → #2 → #3 → #5 → #4.
