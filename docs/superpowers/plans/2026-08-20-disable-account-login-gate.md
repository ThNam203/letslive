# Disable-Account Login Gate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Block disabled accounts from completing login (password, Google OAuth web, Google OAuth mobile, token refresh) and, at the moment of a blocked login, let the account holder reactivate via a short-lived reactivation token instead of a cosmetic post-login dialog.

**Architecture:** All three login entry points and the refresh endpoint gain a status check (new `UserGateway.GetUserStatus`, reusing the user service's existing `GET /v1/user/{userId}`) right before a session would be issued. A disabled account gets a signed, ~10-minute reactivation token instead of cookies. A new `POST /auth/reactivate` endpoint verifies that token, flips status to `normal` via the existing internal `PUT /v1/user/{userId}`, and completes login. Zero user-service changes; all work is in the `auth` service and the web frontend.

**Tech Stack:** Go (auth service, stdlib `net/http`, `golang-jwt/jwt/v5`, stdlib `testing`), Next.js/TypeScript frontend.

**Spec:** `docs/superpowers/specs/2026-08-20-disable-account-login-gate-design.md`

## Global Constraints

- Comments: minimal. Only flow-level comments marking a distinct step in a longer function are acceptable — never comments that restate what the next line does. Match this to the existing files, which already have effectively zero comments.
- TDD is mandatory for every code task: write the failing test first, watch it fail, then implement.
- Go tests: stdlib `testing` only, table-driven where the case shape allows it, no new test dependencies (no testify/gomock — the repo doesn't use them and adding one needs to be asked first). Run with `-race`.
- Frontend: no new test dependencies (`web/` has zero test tooling today; adding one was explicitly declined for this feature). Frontend verification is manual browser testing, documented as explicit steps, not skipped silently.
- No Kong config changes needed — `configs/kong.yml`'s `Auth_Public_Routes` already matches the whole `/auth` path prefix, so the new `/auth/reactivate` endpoint is already routed.
- No user-service backend changes — reuse the existing `GET /v1/user/{userId}` and `PUT /v1/user/{userId}` endpoints exactly as they exist today. **Amended during Task 1's review (see SDD ledger):** one narrow exception — `backend/user/repositories/user/update_user.go`'s `Update` SQL wraps only `username` in `COALESCE`; `phone_number`, `bio`, and `locale` are written directly from the request, so any status-only partial update (this feature's whole design) silently NULLs them out. Task 1 also adds `COALESCE` around those three fields, matching the existing `username` pattern — fixing this at the root instead of threading full-profile payloads through the new gateway path. This also fixes a live bug in the already-shipped settings-page self-disable/reactivate dialogs, which send the same status-only partial payload today.
- `REACTIVATION_TOKEN_SECRET` needs a GitHub Actions repository secret added manually (Settings → Secrets → Actions) before the deploy workflow can inject it — this cannot be done from within the repo and is called out again at the task that needs it.

---

## File Structure

**Backend (`backend/auth`), new files:**
- `gateway/user/dto/get_user_status_response.go` — response shape for the status read
- `gateway/user/dto/update_user_status_request.go` — request shape for the status write
- `gateway/user/http/http_test.go` — httptest-backed tests for the two new gateway methods
- `internal/testutil/fakes.go` — hand-written fakes for `AuthRepository`, `RefreshTokenRepository`, `UserGateway`, shared by `services` and `handlers` package tests
- `services/auth_test.go`
- `services/jwt_test.go`
- `dto/reactivate_request.go`
- `dto/account_disabled_response.go`
- `handlers/auth_test.go`
- `handlers/google_oauth_test.go`

**Backend, modified files:**
- `gateway/user/user.go` — add status consts + 2 interface methods
- `gateway/user/http/http.go` — implement the 2 new methods
- `services/auth.go` — add `GetUserStatus`, `ReactivateUser`
- `services/jwt.go` — add `userGateway` dependency, reactivation token methods, refresh-time status gate
- `cmd/main.go` — wire the extra `NewJWTService` argument
- `response/error.go` — add `RES_ERR_ACCOUNT_DISABLED`
- `handlers/utils.go` — add `checkAccountStatus` helper
- `handlers/auth.go` — gate `LogInHandler`, add `ReactivateHandler`
- `handlers/google_oauth.go` — gate `OAuthGoogleCallBackHandler` and `OAuthGoogleMobileHandler`, add `buildDisabledRedirectURL`
- `api/http.go` — register `POST /v1/auth/reactivate`

**Repo root, modified files (secret plumbing):**
- `example.env`, `docker-compose.yaml`, `docker-compose-dev.yaml`, `.github/workflows/deploy.yml`

**Frontend (`web`), new files:**
- `components/forms/AccountDisabledDialog.tsx`

**Frontend, modified files:**
- `types/fetch-response.ts` — add the new response code
- `lib/api/auth.ts` — retype `LogIn`, add `Reactivate`
- `components/forms/LoginForm.tsx` — branch on the disabled response
- `app/[lng]/(auth)/login/page.tsx` — own the dialog's open state, read the OAuth redirect query params
- `lib/i18n/locales/en-US/auth.json`, `lib/i18n/locales/vi-VN/auth.json` — new copy

---

### Task 1: User-status gateway methods

**Files:**
- Modify: `backend/auth/gateway/user/user.go`
- Modify: `backend/auth/gateway/user/http/http.go`
- Create: `backend/auth/gateway/user/dto/get_user_status_response.go`
- Create: `backend/auth/gateway/user/dto/update_user_status_request.go`
- Test: `backend/auth/gateway/user/http/http_test.go`

**Interfaces:**
- Produces: `usergateway.UserStatusNormal`, `usergateway.UserStatusDisabled` (string consts); `UserGateway.GetUserStatus(ctx context.Context, userId string) (string, *serviceresponse.Response[any])`; `UserGateway.UpdateUserStatus(ctx context.Context, userId string, status string) *serviceresponse.Response[any]`

- [ ] **Step 1: Write the failing gateway tests**

Create `backend/auth/gateway/user/http/http_test.go`:

```go
package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sen1or/letslive/auth/gateway/user/dto"
	serviceresponse "sen1or/letslive/auth/response"
)

type fakeRegistry struct {
	addr string
}

func (f *fakeRegistry) Register(ctx context.Context, hostPort string, serviceHealthCheckURL string, serviceName string, instanceId string, tags []string) error {
	return nil
}

func (f *fakeRegistry) Deregister(ctx context.Context, serviceName string, instanceId string) error {
	return nil
}

func (f *fakeRegistry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	return []string{f.addr}, nil
}

func (f *fakeRegistry) ServiceAddress(ctx context.Context, serviceName string) (string, error) {
	return f.addr, nil
}

func TestGetUserStatus_ReturnsStatusOnSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/user/user-123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(serviceresponse.Response[dto.GetUserStatusResponseDTO]{
			Success: true,
			Data:    &dto.GetUserStatusResponseDTO{Status: "disabled"},
		})
	}))
	defer server.Close()

	g := NewUserGateway(&fakeRegistry{addr: strings.TrimPrefix(server.URL, "http://")})

	status, err := g.GetUserStatus(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if status != "disabled" {
		t.Errorf("status = %q, want %q", status, "disabled")
	}
}

func TestGetUserStatus_PropagatesNonSuccessResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(serviceresponse.Response[any]{
			Success: false,
			Code:    30000,
			Key:     "res_err_user_not_found",
		})
	}))
	defer server.Close()

	g := NewUserGateway(&fakeRegistry{addr: strings.TrimPrefix(server.URL, "http://")})

	_, err := g.GetUserStatus(context.Background(), "missing-user")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Code != 30000 {
		t.Errorf("err.Code = %d, want 30000", err.Code)
	}
}

func TestUpdateUserStatus_SendsStatusAndSucceeds(t *testing.T) {
	var receivedBody dto.UpdateUserStatusRequestDTO
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/v1/user/user-123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(serviceresponse.Response[any]{Success: true})
	}))
	defer server.Close()

	g := NewUserGateway(&fakeRegistry{addr: strings.TrimPrefix(server.URL, "http://")})

	err := g.UpdateUserStatus(context.Background(), "user-123", "normal")
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if receivedBody.Status != "normal" {
		t.Errorf("sent status = %q, want %q", receivedBody.Status, "normal")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend/auth && go test ./gateway/user/http/... -run TestGetUserStatus -v`
Expected: FAIL — `dto.GetUserStatusResponseDTO`, `g.GetUserStatus` etc. undefined.

- [ ] **Step 3: Add the DTOs**

`backend/auth/gateway/user/dto/get_user_status_response.go`:

```go
package dto

type GetUserStatusResponseDTO struct {
	Status string `json:"status"`
}
```

`backend/auth/gateway/user/dto/update_user_status_request.go`:

```go
package dto

type UpdateUserStatusRequestDTO struct {
	Status string `json:"status"`
}
```

- [ ] **Step 4: Extend the `UserGateway` interface**

Modify `backend/auth/gateway/user/user.go` to:

```go
package user

import (
	"context"
	"sen1or/letslive/auth/gateway/user/dto"
	serviceresponse "sen1or/letslive/auth/response"
)

const (
	UserStatusNormal   = "normal"
	UserStatusDisabled = "disabled"
)

type UserGateway interface {
	CreateNewUser(ctx context.Context, userRequestDTO dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, *serviceresponse.Response[any])
	GetUserStatus(ctx context.Context, userId string) (string, *serviceresponse.Response[any])
	UpdateUserStatus(ctx context.Context, userId string, status string) *serviceresponse.Response[any]
}
```

- [ ] **Step 5: Implement the two methods**

Append to `backend/auth/gateway/user/http/http.go` (same file that already implements `CreateNewUser`, same style):

```go
func (g *userGateway) GetUserStatus(ctx context.Context, userId string) (string, *serviceresponse.Response[any]) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	url := fmt.Sprintf("http://%s/v1/user/%s", addr, userId)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Errorf(ctx, "failed to create the request: %s", err)
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Errorf(ctx, "failed to create the request: %s", err)
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call request: %s", err)
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		resInfo := serviceresponse.Response[any]{}
		if err := json.NewDecoder(resp.Body).Decode(&resInfo); err != nil {
			logger.Errorf(ctx, "failed to decode error response from user service: %s", err)
			return "", serviceresponse.NewResponseFromTemplate[any](
				serviceresponse.RES_ERR_INTERNAL_SERVER,
				nil,
				nil,
				nil,
			)
		}

		return "", &resInfo
	}

	var statusRes serviceresponse.Response[dto.GetUserStatusResponseDTO]
	if err := json.NewDecoder(resp.Body).Decode(&statusRes); err != nil {
		logger.Errorf(ctx, "failed to decode resp body: %s", err)
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	return statusRes.Data.Status, nil
}

func (g *userGateway) UpdateUserStatus(ctx context.Context, userId string, status string) *serviceresponse.Response[any] {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	url := fmt.Sprintf("http://%s/v1/user/%s", addr, userId)
	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(&dto.UpdateUserStatusRequestDTO{Status: status}); err != nil {
		logger.Errorf(ctx, "failed to encode user status dto body: %s", err)
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, payloadBuf)
	if err != nil {
		logger.Errorf(ctx, "failed to create the request: %s", err)
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Errorf(ctx, "failed to create the request: %s", err)
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call request: %s", err)
		return serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		resInfo := serviceresponse.Response[any]{}
		if err := json.NewDecoder(resp.Body).Decode(&resInfo); err != nil {
			logger.Errorf(ctx, "failed to decode error response from user service: %s", err)
			return serviceresponse.NewResponseFromTemplate[any](
				serviceresponse.RES_ERR_INTERNAL_SERVER,
				nil,
				nil,
				nil,
			)
		}

		return &resInfo
	}

	return nil
}
```

(`bytes`, `fmt`, `encoding/json`, `net/http`, `logger`, `gateway` are already imported at the top of this file for `CreateNewUser` — no import changes needed.)

- [ ] **Step 6: Run tests to verify they pass**

Run: `cd backend/auth && go test ./gateway/user/http/... -race -v`
Expected: PASS (all 3 new tests plus any pre-existing ones in this package).

- [ ] **Step 7: Commit**

```bash
git add backend/auth/gateway/user/user.go backend/auth/gateway/user/http/http.go backend/auth/gateway/user/dto/get_user_status_response.go backend/auth/gateway/user/dto/update_user_status_request.go backend/auth/gateway/user/http/http_test.go
git commit -m "$(cat <<'EOF'
feat(auth): add user-status read/write to the user gateway

Refs: docs/superpowers/specs/2026-08-20-disable-account-login-gate-design.md
EOF
)"
```

---

### Task 2: `AuthService.GetUserStatus` / `ReactivateUser`

**Files:**
- Create: `backend/auth/internal/testutil/fakes.go`
- Modify: `backend/auth/services/auth.go`
- Test: `backend/auth/services/auth_test.go`

**Interfaces:**
- Consumes: `usergateway.UserGateway` (Task 1), `usergateway.UserStatusNormal`
- Produces: `testutil.FakeAuthRepository`, `testutil.FakeRefreshTokenRepository`, `testutil.FakeUserGateway` (all in `sen1or/letslive/auth/internal/testutil`) — every field is a `...Func` you set only for the methods your test exercises; `AuthService.GetUserStatus(ctx context.Context, userId uuid.UUID) (string, *serviceresponse.Response[any])`; `AuthService.ReactivateUser(ctx context.Context, userId string) *serviceresponse.Response[any]`

- [ ] **Step 1: Add the shared test fakes**

Create `backend/auth/internal/testutil/fakes.go`:

```go
package testutil

import (
	"context"
	"sen1or/letslive/auth/domains"
	usergatewaydto "sen1or/letslive/auth/gateway/user/dto"
	serviceresponse "sen1or/letslive/auth/response"

	"github.com/gofrs/uuid/v5"
)

type FakeAuthRepository struct {
	GetByEmailFunc         func(ctx context.Context, email string) (*domains.Auth, *serviceresponse.Response[any])
	GetByUserIDFunc        func(ctx context.Context, userId uuid.UUID) (*domains.Auth, *serviceresponse.Response[any])
	GetByIDFunc            func(ctx context.Context, authId uuid.UUID) (*domains.Auth, *serviceresponse.Response[any])
	CreateFunc             func(ctx context.Context, auth domains.Auth) (*domains.Auth, *serviceresponse.Response[any])
	UpdatePasswordHashFunc func(ctx context.Context, authId, newPasswordHash string) *serviceresponse.Response[any]
}

func (f *FakeAuthRepository) GetByEmail(ctx context.Context, email string) (*domains.Auth, *serviceresponse.Response[any]) {
	return f.GetByEmailFunc(ctx, email)
}

func (f *FakeAuthRepository) GetByUserID(ctx context.Context, userId uuid.UUID) (*domains.Auth, *serviceresponse.Response[any]) {
	return f.GetByUserIDFunc(ctx, userId)
}

func (f *FakeAuthRepository) GetByID(ctx context.Context, authId uuid.UUID) (*domains.Auth, *serviceresponse.Response[any]) {
	return f.GetByIDFunc(ctx, authId)
}

func (f *FakeAuthRepository) Create(ctx context.Context, auth domains.Auth) (*domains.Auth, *serviceresponse.Response[any]) {
	return f.CreateFunc(ctx, auth)
}

func (f *FakeAuthRepository) UpdatePasswordHash(ctx context.Context, authId, newPasswordHash string) *serviceresponse.Response[any] {
	return f.UpdatePasswordHashFunc(ctx, authId, newPasswordHash)
}

type FakeRefreshTokenRepository struct {
	RevokeAllTokensOfUserFunc func(ctx context.Context, userId uuid.UUID) *serviceresponse.Response[any]
	InsertFunc                func(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any]
	FindByValueFunc           func(ctx context.Context, value string) (*domains.RefreshToken, *serviceresponse.Response[any])
	UpdateFunc                func(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any]
}

func (f *FakeRefreshTokenRepository) RevokeAllTokensOfUser(ctx context.Context, userId uuid.UUID) *serviceresponse.Response[any] {
	return f.RevokeAllTokensOfUserFunc(ctx, userId)
}

func (f *FakeRefreshTokenRepository) Insert(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
	return f.InsertFunc(ctx, token)
}

func (f *FakeRefreshTokenRepository) FindByValue(ctx context.Context, value string) (*domains.RefreshToken, *serviceresponse.Response[any]) {
	return f.FindByValueFunc(ctx, value)
}

func (f *FakeRefreshTokenRepository) Update(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
	return f.UpdateFunc(ctx, token)
}

type FakeUserGateway struct {
	CreateNewUserFunc    func(ctx context.Context, requestDTO usergatewaydto.CreateUserRequestDTO) (*usergatewaydto.CreateUserResponseDTO, *serviceresponse.Response[any])
	GetUserStatusFunc    func(ctx context.Context, userId string) (string, *serviceresponse.Response[any])
	UpdateUserStatusFunc func(ctx context.Context, userId string, status string) *serviceresponse.Response[any]
}

func (f *FakeUserGateway) CreateNewUser(ctx context.Context, requestDTO usergatewaydto.CreateUserRequestDTO) (*usergatewaydto.CreateUserResponseDTO, *serviceresponse.Response[any]) {
	return f.CreateNewUserFunc(ctx, requestDTO)
}

func (f *FakeUserGateway) GetUserStatus(ctx context.Context, userId string) (string, *serviceresponse.Response[any]) {
	return f.GetUserStatusFunc(ctx, userId)
}

func (f *FakeUserGateway) UpdateUserStatus(ctx context.Context, userId string, status string) *serviceresponse.Response[any] {
	return f.UpdateUserStatusFunc(ctx, userId, status)
}
```

Each field left `nil` panics if the test path calls it — that's a deliberate loud failure for an unexpected call, not a bug to guard against.

- [ ] **Step 2: Write the failing `AuthService` tests**

Create `backend/auth/services/auth_test.go`:

```go
package services

import (
	"context"
	"testing"

	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/internal/testutil"
	serviceresponse "sen1or/letslive/auth/response"

	"github.com/gofrs/uuid/v5"
)

func TestGetUserStatus_ReturnsStatusFromGateway(t *testing.T) {
	userId := uuid.Must(uuid.NewV4())
	gateway := &testutil.FakeUserGateway{
		GetUserStatusFunc: func(ctx context.Context, id string) (string, *serviceresponse.Response[any]) {
			if id != userId.String() {
				t.Errorf("gateway called with id = %q, want %q", id, userId.String())
			}
			return "disabled", nil
		},
	}
	s := NewAuthService(&testutil.FakeAuthRepository{}, gateway)

	status, err := s.GetUserStatus(context.Background(), userId)
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if status != "disabled" {
		t.Errorf("status = %q, want %q", status, "disabled")
	}
}

func TestGetUserStatus_PropagatesGatewayError(t *testing.T) {
	wantErr := serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	gateway := &testutil.FakeUserGateway{
		GetUserStatusFunc: func(ctx context.Context, id string) (string, *serviceresponse.Response[any]) {
			return "", wantErr
		},
	}
	s := NewAuthService(&testutil.FakeAuthRepository{}, gateway)

	_, err := s.GetUserStatus(context.Background(), uuid.Must(uuid.NewV4()))
	if err != wantErr {
		t.Errorf("err = %+v, want %+v", err, wantErr)
	}
}

func TestReactivateUser_SetsStatusToNormal(t *testing.T) {
	userId := uuid.Must(uuid.NewV4())
	var gotStatus string
	gateway := &testutil.FakeUserGateway{
		UpdateUserStatusFunc: func(ctx context.Context, id string, status string) *serviceresponse.Response[any] {
			gotStatus = status
			return nil
		},
	}
	s := NewAuthService(&testutil.FakeAuthRepository{}, gateway)

	if err := s.ReactivateUser(context.Background(), userId.String()); err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if gotStatus != usergateway.UserStatusNormal {
		t.Errorf("status sent = %q, want %q", gotStatus, usergateway.UserStatusNormal)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd backend/auth && go test ./services/... -run TestGetUserStatus -v`
Expected: FAIL — `s.GetUserStatus` undefined.

- [ ] **Step 4: Implement the two methods**

Append to `backend/auth/services/auth.go` (it already imports `usergateway "sen1or/letslive/auth/gateway/user"` and `"github.com/gofrs/uuid/v5"` — no new imports):

```go
func (s AuthService) GetUserStatus(ctx context.Context, userId uuid.UUID) (string, *serviceresponse.Response[any]) {
	return s.userGateway.GetUserStatus(ctx, userId.String())
}

func (s AuthService) ReactivateUser(ctx context.Context, userId string) *serviceresponse.Response[any] {
	return s.userGateway.UpdateUserStatus(ctx, userId, usergateway.UserStatusNormal)
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd backend/auth && go test ./services/... -race -v -run 'TestGetUserStatus|TestReactivateUser'`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add backend/auth/internal/testutil/fakes.go backend/auth/services/auth.go backend/auth/services/auth_test.go
git commit -m "$(cat <<'EOF'
feat(auth): add AuthService.GetUserStatus and ReactivateUser

Refs: docs/superpowers/specs/2026-08-20-disable-account-login-gate-design.md
EOF
)"
```

---

### Task 3: Reactivation token + refresh-time status gate

**Files:**
- Modify: `backend/auth/services/jwt.go`
- Modify: `backend/auth/cmd/main.go`
- Modify: `backend/auth/response/error.go`
- Modify: `example.env`, `docker-compose.yaml`, `docker-compose-dev.yaml`, `.github/workflows/deploy.yml`
- Test: `backend/auth/services/jwt_test.go`

**Interfaces:**
- Consumes: `testutil.FakeUserGateway`, `testutil.FakeRefreshTokenRepository` (Task 2), `usergateway.UserStatusDisabled`/`UserStatusNormal` (Task 1)
- Produces: `JWTService.GenerateReactivationToken(ctx context.Context, userId string) (string, *serviceresponse.Response[any])`; `JWTService.VerifyReactivationToken(ctx context.Context, token string) (string, *serviceresponse.Response[any])`; `NewJWTService` now takes a third `userGateway usergateway.UserGateway` argument; `RefreshToken` now rejects a disabled account's refresh token

- [ ] **Step 1: Write the failing `JWTService` tests**

Create `backend/auth/services/jwt_test.go`:

```go
package services

import (
	"context"
	"testing"
	"time"

	"sen1or/letslive/auth/config"
	"sen1or/letslive/auth/domains"
	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/internal/testutil"
	serviceresponse "sen1or/letslive/auth/response"
	"sen1or/letslive/auth/types"

	"github.com/golang-jwt/jwt/v5"
)

func newTestJWTService(t *testing.T, gateway usergateway.UserGateway) *JWTService {
	t.Helper()
	t.Setenv("REACTIVATION_TOKEN_SECRET", "test-reactivation-secret")
	t.Setenv("ACCESS_TOKEN_SECRET", "test-access-secret")
	t.Setenv("REFRESH_TOKEN_SECRET", "test-refresh-secret")

	repo := &testutil.FakeRefreshTokenRepository{
		InsertFunc: func(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
			return nil
		},
	}
	cfg := config.JWT{RefreshTokenMaxAge: 3600, AccessTokenMaxAge: 900, Consumer: "test", Issuer: "test", Subject: "test"}
	return NewJWTService(repo, cfg, gateway)
}

func TestGenerateAndVerifyReactivationToken_RoundTrip(t *testing.T) {
	s := newTestJWTService(t, &testutil.FakeUserGateway{})

	token, genErr := s.GenerateReactivationToken(context.Background(), "user-123")
	if genErr != nil {
		t.Fatalf("expected no error, got %+v", genErr)
	}

	userId, verifyErr := s.VerifyReactivationToken(context.Background(), token)
	if verifyErr != nil {
		t.Fatalf("expected no error, got %+v", verifyErr)
	}
	if userId != "user-123" {
		t.Errorf("userId = %q, want %q", userId, "user-123")
	}
}

func TestVerifyReactivationToken_RejectsExpiredToken(t *testing.T) {
	s := newTestJWTService(t, &testutil.FakeUserGateway{})

	expiredClaims := types.MyClaims{
		UserId: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte("test-reactivation-secret"))
	if err != nil {
		t.Fatalf("failed to build expired token: %s", err)
	}

	if _, verifyErr := s.VerifyReactivationToken(context.Background(), expiredToken); verifyErr == nil {
		t.Fatal("expected an error for an expired token, got nil")
	}
}

func TestVerifyReactivationToken_RejectsTokenSignedWithWrongSecret(t *testing.T) {
	s := newTestJWTService(t, &testutil.FakeUserGateway{})

	claims := types.MyClaims{
		UserId: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	}
	wrongSecretToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("some-other-secret"))
	if err != nil {
		t.Fatalf("failed to build token: %s", err)
	}

	if _, verifyErr := s.VerifyReactivationToken(context.Background(), wrongSecretToken); verifyErr == nil {
		t.Fatal("expected an error for a token signed with the wrong secret, got nil")
	}
}

func TestRefreshToken_RejectsDisabledAccount(t *testing.T) {
	gateway := &testutil.FakeUserGateway{
		GetUserStatusFunc: func(ctx context.Context, id string) (string, *serviceresponse.Response[any]) {
			return usergateway.UserStatusDisabled, nil
		},
	}
	s := newTestJWTService(t, gateway)

	pair, genErr := s.GenerateTokenPair(context.Background(), "user-123")
	if genErr != nil {
		t.Fatalf("failed to generate token pair: %+v", genErr)
	}

	if _, refreshErr := s.RefreshToken(context.Background(), pair.RefreshToken); refreshErr == nil {
		t.Fatal("expected refresh to be rejected for a disabled account, got nil error")
	}
}

func TestRefreshToken_AllowsNormalAccount(t *testing.T) {
	gateway := &testutil.FakeUserGateway{
		GetUserStatusFunc: func(ctx context.Context, id string) (string, *serviceresponse.Response[any]) {
			return usergateway.UserStatusNormal, nil
		},
	}
	s := newTestJWTService(t, gateway)

	pair, genErr := s.GenerateTokenPair(context.Background(), "user-123")
	if genErr != nil {
		t.Fatalf("failed to generate token pair: %+v", genErr)
	}

	accessInfo, refreshErr := s.RefreshToken(context.Background(), pair.RefreshToken)
	if refreshErr != nil {
		t.Fatalf("expected no error, got %+v", refreshErr)
	}
	if accessInfo.AccessToken == "" {
		t.Error("expected a non-empty access token")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend/auth && go test ./services/... -run 'ReactivationToken|TestRefreshToken' -v`
Expected: FAIL — `NewJWTService` called with 3 args doesn't match the current 2-arg signature; `GenerateReactivationToken`/`VerifyReactivationToken` undefined.

- [ ] **Step 3: Add the response code these tests and the refresh gate need**

Modify `backend/auth/response/error.go` — add to the code block (after `RES_ERR_FAILED_TO_SEND_VERIFICATION_CODE = 20018`):

```go
	RES_ERR_ACCOUNT_DISABLED_CODE             = 20019
```

add to the key block (after `RES_ERR_FAILED_TO_SEND_VERIFICATION_KEY = "res_err_failed_to_send_verification"`):

```go
	RES_ERR_ACCOUNT_DISABLED_KEY             = "res_err_account_disabled"
```

add to the `var (...)` block (after `RES_ERR_FAILED_TO_SEND_VERIFICATION`):

```go
	RES_ERR_ACCOUNT_DISABLED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusForbidden,
		Code:       RES_ERR_ACCOUNT_DISABLED_CODE,
		Key:        RES_ERR_ACCOUNT_DISABLED_KEY,
		Message:    "This account has been disabled.",
	}
```

This is used by both the refresh gate below and, later, `LogInHandler`/`ReactivateHandler`/the Google OAuth handlers in Tasks 4–5.

- [ ] **Step 4: Add the `userGateway` dependency and reactivation token methods**

Modify `backend/auth/services/jwt.go`:

```go
package services

import (
	"context"
	"os"
	"sen1or/letslive/auth/config"
	"sen1or/letslive/auth/domains"
	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/shared/pkg/logger"
	serviceresponse "sen1or/letslive/auth/response"
	"sen1or/letslive/auth/types"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

const reactivationTokenMaxAge = 10 * time.Minute

type JWTService struct {
	repo        domains.RefreshTokenRepository
	config      config.JWT
	userGateway usergateway.UserGateway
}

func NewJWTService(repo domains.RefreshTokenRepository, cfg config.JWT, userGateway usergateway.UserGateway) *JWTService {
	return &JWTService{
		repo:        repo,
		config:      cfg,
		userGateway: userGateway,
	}
}
```

In `RefreshToken`, insert the status check right after the existing validity check and before generating a new access token:

```go
	} else if !parsedToken.Valid {
		logger.Errorf(ctx, "token not valid")
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		)
	}

	status, statusErr := c.userGateway.GetUserStatus(ctx, myClaims.UserId)
	if statusErr != nil {
		return nil, statusErr
	}
	if status == usergateway.UserStatusDisabled {
		return nil, serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_ACCOUNT_DISABLED,
			nil,
			nil,
			nil,
		)
	}

	accessToken, genErr := c.generateAccessToken(myClaims.UserId)
```

Append the two new methods at the end of the file:

```go
func (c *JWTService) GenerateReactivationToken(ctx context.Context, userId string) (string, *serviceresponse.Response[any]) {
	expiresAt := time.Now().Add(reactivationTokenMaxAge)
	myClaims := types.MyClaims{
		UserId:   userId,
		Consumer: c.config.Consumer,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    c.config.Issuer,
			Subject:   c.config.Subject,
		},
	}
	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, myClaims)

	token, err := unsignedToken.SignedString([]byte(os.Getenv("REACTIVATION_TOKEN_SECRET")))
	if err != nil {
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		)
	}

	return token, nil
}

func (c *JWTService) VerifyReactivationToken(ctx context.Context, token string) (string, *serviceresponse.Response[any]) {
	myClaims := types.MyClaims{}
	parsedToken, err := jwt.NewParser().ParseWithClaims(token, &myClaims, func(t *jwt.Token) (any, error) {
		return []byte(os.Getenv("REACTIVATION_TOKEN_SECRET")), nil
	})

	if err != nil || !parsedToken.Valid {
		return "", serviceresponse.NewResponseFromTemplate[any](
			serviceresponse.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		)
	}

	return myClaims.UserId, nil
}
```

- [ ] **Step 5: Wire the new dependency in `main.go`**

Modify `backend/auth/cmd/main.go:120`:

```go
	var jwtService = services.NewJWTService(refreshTokenRepo, cfg.JWT, userGateway)
```

(`userGateway` is already constructed two lines above this call.)

- [ ] **Step 6: Add `REACTIVATION_TOKEN_SECRET` alongside the existing token secrets**

`example.env` (next to line 32, `REFRESH_TOKEN_SECRET=refresh_token_secret`):

```
REACTIVATION_TOKEN_SECRET=reactivation_token_secret
```

`docker-compose.yaml` and `docker-compose-dev.yaml` (next to the `REFRESH_TOKEN_SECRET=${REFRESH_TOKEN_SECRET}` line under the `auth` service's `environment:` block in each file):

```yaml
      - REACTIVATION_TOKEN_SECRET=${REACTIVATION_TOKEN_SECRET}
```

`.github/workflows/deploy.yml` (next to the `REFRESH_TOKEN_SECRET=${{ secrets.REFRESH_TOKEN_SECRET }}` line inside the generated `.env`):

```
          REACTIVATION_TOKEN_SECRET=${{ secrets.REACTIVATION_TOKEN_SECRET }}
```

**Manual step, cannot be done from the repo:** add a `REACTIVATION_TOKEN_SECRET` repository secret in GitHub (Settings → Secrets and variables → Actions) before the next deploy — otherwise the workflow renders an empty value for it.

- [ ] **Step 7: Run tests to verify they pass**

Run: `cd backend/auth && go test ./... -race -v`
Expected: PASS across the whole module — this task leaves `backend/auth` fully compiling and green.

- [ ] **Step 8: Commit**

```bash
git add backend/auth/services/jwt.go backend/auth/services/jwt_test.go backend/auth/cmd/main.go backend/auth/response/error.go example.env docker-compose.yaml docker-compose-dev.yaml .github/workflows/deploy.yml
git commit -m "$(cat <<'EOF'
feat(auth): add reactivation tokens and gate refresh on account status

Refs: docs/superpowers/specs/2026-08-20-disable-account-login-gate-design.md
EOF
)"
```

---

### Task 4: Gate password login, add `/auth/reactivate`

**Files:**
- Create: `backend/auth/dto/reactivate_request.go`
- Create: `backend/auth/dto/account_disabled_response.go`
- Modify: `backend/auth/handlers/utils.go`
- Modify: `backend/auth/handlers/auth.go`
- Modify: `backend/auth/api/http.go`
- Test: `backend/auth/handlers/auth_test.go`

**Interfaces:**
- Consumes: `AuthService.GetUserStatus`/`ReactivateUser` (Task 2), `JWTService.GenerateReactivationToken`/`VerifyReactivationToken` (Task 3), `usergateway.UserStatusDisabled` (Task 1)
- Produces: `serviceresponse.RES_ERR_ACCOUNT_DISABLED`; `(*AuthHandler).checkAccountStatus(ctx context.Context, userId uuid.UUID) (isDisabled bool, reactivationToken string, errRes *serviceresponse.Response[any])`; `(*AuthHandler).ReactivateHandler`; route `POST /v1/auth/reactivate`

- [ ] **Step 1: Write the failing handler tests**

Create `backend/auth/handlers/auth_test.go`:

```go
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sen1or/letslive/auth/config"
	"sen1or/letslive/auth/domains"
	"sen1or/letslive/auth/dto"
	usergateway "sen1or/letslive/auth/gateway/user"
	"sen1or/letslive/auth/internal/testutil"
	serviceresponse "sen1or/letslive/auth/response"
	"sen1or/letslive/auth/services"

	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
)

const testPassword = "Password123!"

func newTestAuthHandler(t *testing.T, gateway *testutil.FakeUserGateway, authRepo *testutil.FakeAuthRepository, refreshRepo *testutil.FakeRefreshTokenRepository) *AuthHandler {
	t.Helper()
	t.Setenv("ACCESS_TOKEN_SECRET", "test-access-secret")
	t.Setenv("REFRESH_TOKEN_SECRET", "test-refresh-secret")
	t.Setenv("REACTIVATION_TOKEN_SECRET", "test-reactivation-secret")

	jwtCfg := config.JWT{RefreshTokenMaxAge: 3600, AccessTokenMaxAge: 900, Consumer: "test", Issuer: "test", Subject: "test"}

	authService := services.NewAuthService(authRepo, gateway)
	googleAuthService := services.NewGoogleAuthService(authRepo, gateway)
	jwtService := services.NewJWTService(refreshRepo, jwtCfg, gateway)

	return NewAuthHandler(*jwtService, *authService, services.VerificationService{}, *googleAuthService, "")
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %s", err)
	}
	return string(hash)
}

func TestLogInHandler_NormalAccountLogsInSuccessfully(t *testing.T) {
	userId := uuid.Must(uuid.NewV4())
	passwordHash := hashPassword(t, testPassword)

	authRepo := &testutil.FakeAuthRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*domains.Auth, *serviceresponse.Response[any]) {
			return &domains.Auth{UserId: &userId, Email: email, PasswordHash: passwordHash}, nil
		},
	}
	gateway := &testutil.FakeUserGateway{
		GetUserStatusFunc: func(ctx context.Context, id string) (string, *serviceresponse.Response[any]) {
			return usergateway.UserStatusNormal, nil
		},
	}
	refreshRepo := &testutil.FakeRefreshTokenRepository{
		InsertFunc: func(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
			return nil
		},
	}

	h := newTestAuthHandler(t, gateway, authRepo, refreshRepo)

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": testPassword})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("User-Agent", "letslive-mobile-test")
	rec := httptest.NewRecorder()

	h.LogInHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if _, ok := rec.Result().Header["Set-Cookie"]; !ok {
		t.Error("expected Set-Cookie header on successful login")
	}
}

func TestLogInHandler_DisabledAccountReturnsReactivationTokenWithoutCookies(t *testing.T) {
	userId := uuid.Must(uuid.NewV4())
	passwordHash := hashPassword(t, testPassword)

	authRepo := &testutil.FakeAuthRepository{
		GetByEmailFunc: func(ctx context.Context, email string) (*domains.Auth, *serviceresponse.Response[any]) {
			return &domains.Auth{UserId: &userId, Email: email, PasswordHash: passwordHash}, nil
		},
	}
	gateway := &testutil.FakeUserGateway{
		GetUserStatusFunc: func(ctx context.Context, id string) (string, *serviceresponse.Response[any]) {
			return usergateway.UserStatusDisabled, nil
		},
	}
	h := newTestAuthHandler(t, gateway, authRepo, &testutil.FakeRefreshTokenRepository{})

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": testPassword})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("User-Agent", "letslive-mobile-test")
	rec := httptest.NewRecorder()

	h.LogInHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
	if _, ok := rec.Result().Header["Set-Cookie"]; ok {
		t.Error("expected no Set-Cookie header when account is disabled")
	}

	var res serviceresponse.Response[dto.AccountDisabledResponseDTO]
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response body: %s", err)
	}
	if res.Data == nil || res.Data.ReactivationToken == "" {
		t.Fatalf("expected a non-empty reactivationToken, got %+v", res.Data)
	}
}

func TestReactivateHandler_ValidTokenReactivatesAndLogsIn(t *testing.T) {
	userId := uuid.Must(uuid.NewV4())

	var updatedStatus string
	gateway := &testutil.FakeUserGateway{
		UpdateUserStatusFunc: func(ctx context.Context, id string, status string) *serviceresponse.Response[any] {
			updatedStatus = status
			return nil
		},
	}
	refreshRepo := &testutil.FakeRefreshTokenRepository{
		InsertFunc: func(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
			return nil
		},
	}
	h := newTestAuthHandler(t, gateway, &testutil.FakeAuthRepository{}, refreshRepo)

	token, tokenErr := h.jwtService.GenerateReactivationToken(context.Background(), userId.String())
	if tokenErr != nil {
		t.Fatalf("failed to generate reactivation token: %+v", tokenErr)
	}

	body, _ := json.Marshal(map[string]string{"reactivationToken": token})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/reactivate", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.ReactivateHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if updatedStatus != usergateway.UserStatusNormal {
		t.Errorf("updated status = %q, want %q", updatedStatus, usergateway.UserStatusNormal)
	}
	if _, ok := rec.Result().Header["Set-Cookie"]; !ok {
		t.Error("expected Set-Cookie header after successful reactivation")
	}
}

func TestReactivateHandler_InvalidTokenIsRejectedWithoutUpdatingStatus(t *testing.T) {
	called := false
	gateway := &testutil.FakeUserGateway{
		UpdateUserStatusFunc: func(ctx context.Context, id string, status string) *serviceresponse.Response[any] {
			called = true
			return nil
		},
	}
	h := newTestAuthHandler(t, gateway, &testutil.FakeAuthRepository{}, &testutil.FakeRefreshTokenRepository{})

	body, _ := json.Marshal(map[string]string{"reactivationToken": "not-a-real-token"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/reactivate", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.ReactivateHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	if called {
		t.Error("UpdateUserStatus should not be called for an invalid token")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend/auth && go test ./handlers/... -v -run 'LogInHandler|ReactivateHandler'`
Expected: FAIL — `dto.AccountDisabledResponseDTO`, `h.ReactivateHandler`, `RES_ERR_ACCOUNT_DISABLED` undefined.

- [ ] **Step 3: Add the DTOs**

(`RES_ERR_ACCOUNT_DISABLED`, used below, was already added to `backend/auth/response/error.go` in Task 3.)

`backend/auth/dto/reactivate_request.go`:

```go
package dto

type ReactivateRequestDTO struct {
	ReactivationToken string `json:"reactivationToken" validate:"required"`
}
```

`backend/auth/dto/account_disabled_response.go`:

```go
package dto

type AccountDisabledResponseDTO struct {
	ReactivationToken string `json:"reactivationToken"`
}
```

- [ ] **Step 4: Add the `checkAccountStatus` helper**

Modify `backend/auth/handlers/utils.go` — add `usergateway "sen1or/letslive/auth/gateway/user"` to the import block, then append:

```go
func (h *AuthHandler) checkAccountStatus(ctx context.Context, userId uuid.UUID) (isDisabled bool, reactivationToken string, errRes *serviceresponse.Response[any]) {
	status, statusErr := h.authService.GetUserStatus(ctx, userId)
	if statusErr != nil {
		return false, "", statusErr
	}

	if status != usergateway.UserStatusDisabled {
		return false, "", nil
	}

	token, tokenErr := h.jwtService.GenerateReactivationToken(ctx, userId.String())
	if tokenErr != nil {
		return false, "", tokenErr
	}

	return true, token, nil
}
```

- [ ] **Step 5: Gate `LogInHandler` and add `ReactivateHandler`**

Modify `backend/auth/handlers/auth.go` — replace the body of `LogInHandler` between the credentials check and the success response with:

```go
	auth, err := h.authService.GetUserFromCredentials(ctx, userCredentials)
	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	isDisabled, reactivationToken, statusErr := h.checkAccountStatus(ctx, *auth.UserId)
	if statusErr != nil {
		writeResponse(w, ctx, statusErr)
		return
	}

	if isDisabled {
		disabledData := any(dto.AccountDisabledResponseDTO{ReactivationToken: reactivationToken})
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate(serviceresponse.RES_ERR_ACCOUNT_DISABLED, &disabledData, nil, nil))
		return
	}

	if err := h.setAuthJWTsInCookie(ctx, auth.UserId.String(), w); err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
		serviceresponse.RES_SUCC_LOGIN,
		nil,
		nil,
		nil,
	))
```

Add the new handler (near `LogInHandler`):

```go
func (h *AuthHandler) ReactivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var reqBody dto.ReactivateRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	if err := utils.Validator.Struct(&reqBody); err != nil {
		writeResponse(w, ctx, serviceresponse.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	userId, tokenErr := h.jwtService.VerifyReactivationToken(ctx, reqBody.ReactivationToken)
	if tokenErr != nil {
		writeResponse(w, ctx, tokenErr)
		return
	}

	if statusErr := h.authService.ReactivateUser(ctx, userId); statusErr != nil {
		writeResponse(w, ctx, statusErr)
		return
	}

	if err := h.setAuthJWTsInCookie(ctx, userId, w); err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
		serviceresponse.RES_SUCC_LOGIN,
		nil,
		nil,
		nil,
	))
}
```

(`dto` and `utils` are already imported in this file.)

- [ ] **Step 6: Register the route**

Modify `backend/auth/api/http.go` — add right after the login route:

```go
	wrap("POST /v1/auth/reactivate", a.authHandler.ReactivateHandler)
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `cd backend/auth && go test ./... -race -v`
Expected: PASS across the whole module.

- [ ] **Step 8: Commit**

```bash
git add backend/auth/dto/reactivate_request.go backend/auth/dto/account_disabled_response.go backend/auth/handlers/utils.go backend/auth/handlers/auth.go backend/auth/api/http.go backend/auth/handlers/auth_test.go
git commit -m "$(cat <<'EOF'
feat(auth): gate password login on account status, add /auth/reactivate

Refs: docs/superpowers/specs/2026-08-20-disable-account-login-gate-design.md
EOF
)"
```

---

### Task 5: Gate Google OAuth (web + mobile)

**Files:**
- Modify: `backend/auth/handlers/google_oauth.go`
- Test: `backend/auth/handlers/google_oauth_test.go`

**Interfaces:**
- Consumes: `(*AuthHandler).checkAccountStatus` (Task 4)
- Produces: `buildDisabledRedirectURL(clientAddr, reactivationToken string) string`

**Known limitation (documented, not silently skipped):** `GoogleAuthService.CallbackHandler` and `VerifyIDTokenAndGetUser` make real network calls to Google with no injected HTTP client, so they can't be exercised in a unit test without a larger, separate refactor of pre-existing code that's out of scope here. The status-check-and-branch logic itself is the same `checkAccountStatus` helper already fully tested in Task 4; what's new and testable in isolation here is the redirect-URL construction. The full callback path is verified manually in Step 5.

- [ ] **Step 1: Write the failing test for the redirect URL builder**

Create `backend/auth/handlers/google_oauth_test.go`:

```go
package handlers

import "testing"

func TestBuildDisabledRedirectURL_EscapesReactivationToken(t *testing.T) {
	got := buildDisabledRedirectURL("https://letslive.app", "abc.def+ghi")
	want := "https://letslive.app/login?accountDisabled=true&reactivationToken=abc.def%2Bghi"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend/auth && go test ./handlers/... -run TestBuildDisabledRedirectURL -v`
Expected: FAIL — `buildDisabledRedirectURL` undefined.

- [ ] **Step 3: Gate both Google handlers**

Modify `backend/auth/handlers/google_oauth.go` — add `neturl "net/url"` and `"sen1or/letslive/auth/dto"` to the imports, then update `OAuthGoogleCallBackHandler`:

```go
	createdAuth, handleErr := h.googleAuthService.CallbackHandler(ctx, r.FormValue("code"))
	if handleErr != nil {
		http.Redirect(w, r, GetRedirectURLOnFail(handleErr.Message), http.StatusTemporaryRedirect)
		return
	}

	isDisabled, reactivationToken, statusErr := h.checkAccountStatus(ctx, *createdAuth.UserId)
	if statusErr != nil {
		http.Redirect(w, r, GetRedirectURLOnFail(statusErr.Message), http.StatusTemporaryRedirect)
		return
	}

	if isDisabled {
		http.Redirect(w, r, buildDisabledRedirectURL(os.Getenv("CLIENT_URL"), reactivationToken), http.StatusTemporaryRedirect)
		return
	}

	if err := h.setAuthJWTsInCookie(ctx, createdAuth.UserId.String(), w); err != nil {
		http.Redirect(w, r, GetRedirectURLOnFail(err.Message), http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, GetRedirectURLOnSuccess("/account-setup"), http.StatusMovedPermanently)
}

func buildDisabledRedirectURL(clientAddr, reactivationToken string) string {
	return fmt.Sprintf("%s/login?accountDisabled=true&reactivationToken=%s", clientAddr, neturl.QueryEscape(reactivationToken))
}
```

Update `OAuthGoogleMobileHandler`:

```go
	createdAuth, authErr := h.googleAuthService.VerifyIDTokenAndGetUser(ctx, body.IDToken)
	if authErr != nil {
		writeResponse(w, ctx, authErr)
		return
	}

	isDisabled, reactivationToken, statusErr := h.checkAccountStatus(ctx, *createdAuth.UserId)
	if statusErr != nil {
		writeResponse(w, ctx, statusErr)
		return
	}

	if isDisabled {
		disabledData := any(dto.AccountDisabledResponseDTO{ReactivationToken: reactivationToken})
		writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate(serviceresponse.RES_ERR_ACCOUNT_DISABLED, &disabledData, nil, nil))
		return
	}

	if err := h.setAuthJWTsInCookie(ctx, createdAuth.UserId.String(), w); err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, serviceresponse.NewResponseFromTemplate[any](
		serviceresponse.RES_SUCC_LOGIN, nil, nil, nil,
	))
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend/auth && go test ./... -race -v`
Expected: PASS across the whole module.

- [ ] **Step 5: Manual verification (documented, not automated)**

1. Disable a test account via the settings page.
2. Attempt Google web login for that account → expect redirect to `/login?accountDisabled=true&reactivationToken=...` and the (still-to-be-built, Task 7) dialog to appear, not a completed login.
3. Confirm reactivation in that dialog → expect to land logged in as that account.
4. Repeat steps 1–3 ending in decline instead of confirm → expect to stay on the login page, logged out.

- [ ] **Step 6: Commit**

```bash
git add backend/auth/handlers/google_oauth.go backend/auth/handlers/google_oauth_test.go
git commit -m "$(cat <<'EOF'
feat(auth): gate Google OAuth web and mobile login on account status

Refs: docs/superpowers/specs/2026-08-20-disable-account-login-gate-design.md
EOF
)"
```

---

### Task 6: Frontend — response code + API client

**Files:**
- Modify: `web/types/fetch-response.ts`
- Modify: `web/lib/api/auth.ts`

**Interfaces:**
- Produces: `ApiCode.RES_ERR_ACCOUNT_DISABLED`; `LogIn(): Promise<ApiResponse<{ reactivationToken?: string }>>`; `Reactivate(body: { reactivationToken: string }): Promise<ApiResponse<void>>`

- [ ] **Step 1: Add the response code**

Modify `web/types/fetch-response.ts` — in `ApiCode`, after `RES_ERR_FAILED_TO_SEND_VERIFICATION = 20018,`:

```ts
    RES_ERR_ACCOUNT_DISABLED = 20019,
```

in `ApiKey`, after `RES_ERR_FAILED_TO_SEND_VERIFICATION = "res_err_failed_to_send_verification",`:

```ts
    RES_ERR_ACCOUNT_DISABLED = "res_err_account_disabled",
```

- [ ] **Step 2: Retype `LogIn`, add `Reactivate`**

Modify `web/lib/api/auth.ts` — change `LogIn`'s signature:

```ts
export async function LogIn(body: {
    email: string;
    password: string;
    turnstileToken: string;
}): Promise<ApiResponse<{ reactivationToken?: string }>> {
    return fetchClient<ApiResponse<{ reactivationToken?: string }>>(
        "/auth/login",
        {
            method: "POST",
            body: JSON.stringify({
                email: body.email,
                password: body.password,
                turnstileToken: body.turnstileToken,
            }),
        },
    );
}
```

Append:

```ts
export async function Reactivate(body: {
    reactivationToken: string;
}): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>("/auth/reactivate", {
        method: "POST",
        body: JSON.stringify({
            reactivationToken: body.reactivationToken,
        }),
    });
}
```

- [ ] **Step 3: Type-check**

Run: `cd web && npx tsc --noEmit`
Expected: no new errors (the `LogIn` call site in `LoginForm.tsx` is updated in Task 7 in the same PR — if you run this before Task 7, `LoginForm.tsx`'s existing `.then((res) => ...)` still compiles fine since it only narrows on `res.success`/`res.key`, not on the removed `void` data shape).

- [ ] **Step 4: Commit**

```bash
git add web/types/fetch-response.ts web/lib/api/auth.ts
git commit -m "$(cat <<'EOF'
feat(web): add RES_ERR_ACCOUNT_DISABLED code and Reactivate API call

Refs: docs/superpowers/specs/2026-08-20-disable-account-login-gate-design.md
EOF
)"
```

---

### Task 7: Frontend — disabled-login dialog

**Files:**
- Create: `web/components/forms/AccountDisabledDialog.tsx`
- Modify: `web/components/forms/LoginForm.tsx`
- Modify: `web/app/[lng]/(auth)/login/page.tsx`
- Modify: `web/lib/i18n/locales/en-US/auth.json`, `web/lib/i18n/locales/vi-VN/auth.json`

**Interfaces:**
- Consumes: `ApiCode.RES_ERR_ACCOUNT_DISABLED`, `Reactivate` (Task 6)
- Produces: `<AccountDisabledDialog reactivationToken={string | null} onClose={() => void} />`; `<LogInForm onAccountDisabled={(token: string) => void} />`

No automated tests for this task (per Global Constraints — no new frontend test dependency for this feature). Verified manually in Step 5.

- [ ] **Step 1: Add the i18n copy**

Modify `web/lib/i18n/locales/en-US/auth.json` — add after `"account_setup_submit": "Continue"` (remember the trailing comma on that line):

```json
    "account_disabled_dialog_title": "Your account is disabled",
    "account_disabled_dialog_description": "This account is currently disabled. Log back in to reactivate it, or cancel to stay logged out.",
    "account_disabled_dialog_confirm": "Reactivate my account",
    "account_disabled_dialog_decline": "Cancel"
```

Modify `web/lib/i18n/locales/vi-VN/auth.json` — add after `"account_setup_submit": "Tiếp tục"`:

```json
    "account_disabled_dialog_title": "Tài khoản của bạn đang bị vô hiệu hóa",
    "account_disabled_dialog_description": "Tài khoản này hiện đang bị vô hiệu hóa. Đăng nhập lại để kích hoạt lại, hoặc hủy để tiếp tục đăng xuất.",
    "account_disabled_dialog_confirm": "Kích hoạt lại tài khoản",
    "account_disabled_dialog_decline": "Hủy"
```

- [ ] **Step 2: Create the dialog component**

Create `web/components/forms/AccountDisabledDialog.tsx`:

```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import IconLoader from "@/components/icons/loader";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { Reactivate } from "@/lib/api/auth";
import { GetMeProfile } from "@/lib/api/user";

interface AccountDisabledDialogProps {
    reactivationToken: string | null;
    onClose: () => void;
}

export default function AccountDisabledDialog({
    reactivationToken,
    onClose,
}: AccountDisabledDialogProps) {
    const [isReactivating, setIsReactivating] = useState(false);
    const { t } = useT(["auth", "api-response", "fetch-error"]);
    const { setUser } = useUser();
    const router = useRouter();

    const handleReactivate = async () => {
        if (!reactivationToken) return;

        setIsReactivating(true);
        await Reactivate({ reactivationToken })
            .then((res) => {
                if (!res.success) {
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                    onClose();
                    return;
                }

                GetMeProfile().then((meRes) => {
                    if (meRes.success && meRes.data) {
                        setUser(meRes.data);
                        router.push("/");
                    }
                });
                onClose();
            })
            .catch((_) => {
                toast(t("fetch-error:client_fetch_error"), {
                    toastId: "client-fetch-error-id",
                    type: "error",
                });
            })
            .finally(() => setIsReactivating(false));
    };

    return (
        <Dialog open={reactivationToken !== null}>
            <DialogContent
                className="bg-background text-foreground"
                showCloseButton={false}
                onInteractOutside={(e) => e.preventDefault()}
                onEscapeKeyDown={(e) => e.preventDefault()}
            >
                <DialogHeader>
                    <DialogTitle>
                        {t("auth:account_disabled_dialog_title")}
                    </DialogTitle>
                    <DialogDescription>
                        {t("auth:account_disabled_dialog_description")}
                    </DialogDescription>
                </DialogHeader>

                <DialogFooter>
                    <Button
                        variant="outline"
                        disabled={isReactivating}
                        onClick={onClose}
                    >
                        {t("auth:account_disabled_dialog_decline")}
                    </Button>
                    <Button disabled={isReactivating} onClick={handleReactivate}>
                        {t("auth:account_disabled_dialog_confirm")}
                        {isReactivating && <IconLoader className="ml-1" />}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
```

- [ ] **Step 3: Wire `LoginForm` to bubble up the disabled response**

Modify `web/components/forms/LoginForm.tsx` — add the import `import { ApiCode } from "@/types/fetch-response";`, change the component signature to accept the callback:

```tsx
export default function LogInForm({
    onAccountDisabled,
}: {
    onAccountDisabled: (reactivationToken: string) => void;
}) {
```

and change the `LogIn(...).then(...)` branch:

```tsx
            .then((res) => {
                if (!res.success) {
                    if (
                        res.code === ApiCode.RES_ERR_ACCOUNT_DISABLED &&
                        res.data?.reactivationToken
                    ) {
                        onAccountDisabled(res.data.reactivationToken);
                        return;
                    }

                    turnstile.reset();
                    setTurnstileToken("");
                    toast.error(t(`api-response:${res.key}`), {
                        toastId: res.requestId,
                    });
                } else {
                    GetMeProfile().then((res) => {
                        if (res.success && res.data) {
                            setUser(res.data);
                            router.push("/");
                        }
                    });
                }
            })
```

- [ ] **Step 4: Own the dialog state on the login page**

Modify `web/app/[lng]/(auth)/login/page.tsx`:

```tsx
"use client";

import Link from "next/link";
import IconGoogle from "@/components/icons/google";
import LogInForm from "@/components/forms/LoginForm";
import AccountDisabledDialog from "@/components/forms/AccountDisabledDialog";
import GLOBAL from "@/global";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";

export default function LogInPage() {
    const { t } = useT(["auth", "common"]);
    const searchParams = useSearchParams();
    const router = useRouter();
    const user = useUser((userState) => userState.user);
    const [reactivationToken, setReactivationToken] = useState<string | null>(
        null,
    );

    useEffect(() => {
        const err = searchParams.get("errorMessage");
        if (err) {
            toast(err, {
                type: "error",
            });
        }
    }, [searchParams, router]);

    useEffect(() => {
        const disabled = searchParams.get("accountDisabled");
        const token = searchParams.get("reactivationToken");
        if (disabled === "true" && token) {
            setReactivationToken(token);
        }
    }, [searchParams]);

    useEffect(() => {
        if (!user) return;
        const redirectUrl = searchParams.get("redirectUrl");
        if (
            redirectUrl &&
            redirectUrl.startsWith("/") &&
            !redirectUrl.startsWith("//")
        ) {
            router.push(redirectUrl);
            return;
        }
        router.push("/");
    }, [user, searchParams, router]);

    return (
        <>
            <h1 className="mb-1 text-2xl font-bold">{t("login_title")}</h1>
            <p className="text-md">{t("login_subtitle")}</p>
            <div className="mt-4 mb-2 flex gap-2">
                <div className="w-full">
                    <Link
                        href={GLOBAL.API_URL + "/auth/google"}
                        className="border-border flex h-12 flex-1 flex-row items-center justify-center gap-4 rounded-lg border bg-white py-2 text-black hover:bg-[#ebebeb]"
                    >
                        <IconGoogle /> Google
                    </Link>
                </div>
            </div>
            <div className="mt-2 mb-4 flex w-full items-center justify-center">
                <hr className="bg-border h-[2px] flex-1" />
                <p className="text-foreground mx-4 text-center">
                    {t("common:or")}
                </p>
                <hr className="bg-border h-[2px] flex-1" />
            </div>
            <LogInForm onAccountDisabled={setReactivationToken} />
            <AccountDisabledDialog
                reactivationToken={reactivationToken}
                onClose={() => setReactivationToken(null)}
            />
            <p className="mt-4 text-end text-sm opacity-80">
                {t("no_account")}
                <Link
                    href="/signup"
                    className="ml-2 font-bold text-blue-400 hover:text-blue-600"
                >
                    {t("signup")}
                </Link>
            </p>
        </>
    );
}
```

- [ ] **Step 5: Manual verification**

Run: `cd web && npm run dev`, then in a browser:

1. Type-check first: `cd web && npx tsc --noEmit` — expect no errors.
2. Log in normally with a `normal`-status test account → still lands on `/` as before.
3. Disable that account from the settings page (self-disable), then attempt to log back in with the same password → expect the `AccountDisabledDialog` to appear instead of a completed login, and no `ACCESS_TOKEN`/`REFRESH_TOKEN` cookies set (check DevTools → Application → Cookies).
4. Click "Reactivate my account" → expect to land on `/`, logged in, and the account's status back to `normal` (check via the settings page).
5. Repeat steps 3–4 but click "Cancel" instead → expect to stay on the login page, logged out, status still `disabled`.
6. Repeat the Google OAuth web flow from Task 5's Step 5 now that the dialog exists to receive the redirect.

- [ ] **Step 6: Commit**

```bash
git add web/components/forms/AccountDisabledDialog.tsx web/components/forms/LoginForm.tsx "web/app/[lng]/(auth)/login/page.tsx" web/lib/i18n/locales/en-US/auth.json web/lib/i18n/locales/vi-VN/auth.json
git commit -m "$(cat <<'EOF'
feat(web): show reactivation dialog when login is blocked for a disabled account

Refs: docs/superpowers/specs/2026-08-20-disable-account-login-gate-design.md
EOF
)"
```
