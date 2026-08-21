# Disabled-Account Enforcement Extension Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend disabled-account enforcement (built in the login-gate feature) past login into read paths (profile lookup, following list, livestream/VOD discovery) and write paths (follow, stream start, VOD comments), across the `user`, `livestream`, `vod`, and `transcode` Go services.

**Architecture:** `user` service gets one new internal batch endpoint, `POST /v1/internal/users/statuses`, accepting a list of userIds and returning a status map — this avoids N+1 status-check calls on listing pages. `livestream` and `vod` (separate databases, no existing status visibility) each get a new gateway method calling that endpoint, used to post-filter listing results and gate VOD-comment creation. `user` service also gains same-database `WHERE status != 'disabled'` filters on two more of its own queries (same pattern as the already-shipped search/recommendations filters), and a same-database status check in its own follow flow. `transcode` (which already calls `user` for stream-key verification) gets one new field on an existing DTO plus one new check in the RTMP connect handler. Chat is explicitly out of scope — deferred to a separate branch per product decision (the live chat WebSocket has no real authentication today, so gating it by status would be bypassable and misleading; fixing that is a bigger, separate problem).

**Tech Stack:** Go (user, livestream, vod, transcode services — stdlib `net/http`, stdlib `testing`), PostgreSQL (pgx).

**Design record:** This plan's design was settled through direct conversation (no separate spec doc — the decisions below were confirmed interactively, not written up first):
- Batch status endpoint over per-item live checks or DB denormalization (avoids N+1 on listing pages without the bigger lift of event-driven sync).
- Chat is out of scope this round (separate branch; main chat WebSocket has no real auth, so status-gating it now would be bypassable).
- On a gateway failure during **read-path filtering** (discovery/listing pages), fail open (log, return the unfiltered list) — these are visibility/UX concerns, not authentication boundaries; breaking an entire discovery page because one dependency blipped is worse than a disabled account transiently still appearing in it.
- On a gateway failure during **write-path gating** (follow, comment, stream start), fail closed (reject the action) — consistent with the login-gate's existing fail-closed philosophy for anything that changes state.

## Global Constraints

- Comments: minimal. Only flow-level comments marking a distinct step in a longer function are acceptable — never comments that restate what the next line does. Match the zero-comment style already established in every file this plan touches.
- TDD is mandatory for every code task: write the failing test first, watch it fail, then implement. Go tests: stdlib `testing` only, table-driven where the case shape allows it, no new test dependencies. Run with `-race`.
- `user`, `livestream`, and `vod` services currently have **zero test files and no test-fakes package** (confirmed: only `backend/auth` has one, built in the prior feature). This plan builds a minimal `internal/testutil` fakes package in each service that needs one, following the exact same hand-written-function-field-fakes pattern as `backend/auth/internal/testutil/fakes.go` — no testify/gomock, no new dependency.
- New internal endpoint goes at `/v1/internal/users/statuses` (Go mux path `POST /v1/internal/users/statuses`), matching the established safe convention already used by `/v1/internal/inventory/add` and `/v1/internal/gifts/create` — confirmed via `configs/kong.yml` that no Kong route matches any `/internal/*` path, so this is unreachable from the public gateway by construction, unlike the ambiguous `/user/{id}`-shaped internal routes that caused a real gap fixed earlier in this branch (see the Kong `Users` route fix, commit `06c57f8`). Do not register this under `/v1/users/...` or `/v1/user/...` — that space already has ambiguous Kong route overlap.
- `domains.User` (user service) already has a `Status UserStatus` field with `db:"status"` — adding `u.status` to a SQL `SELECT` list is enough to populate it via `pgx.RowToStructByNameLax`; no struct changes needed for `get_user_by_api_key.go`.
- Reuse the exact gateway file-layout convention already established twice in this codebase (`backend/livestream/gateway/user`, `backend/vod/gateway/user` — both interface file + `http/` impl file + local `userServiceResponse` wrapper struct) for the new methods added to those same packages. Do not create new gateway packages — extend the existing ones.

## Out of Scope (deliberate, not silently dropped)

- **`backend/livestream/handlers/livestream/create_livestream_internal.go` gets no independent status check.** `transcode`'s `onConnect` (Task 3) is the only caller of this internal endpoint and already gates before calling it — adding a second check here would mean `livestream` service also needs user-status visibility for a case with no other caller today. If a second caller of this internal endpoint is ever added, it needs its own gating decision at that time, not an assumption that transcode's check still covers it.
- **Livestream's `GetByUser` (a channel's current live status) and VOD's `GetPublicVODsByUser` (a channel's VOD list) are not filtered.** Both are single-user lookups on an already-known channel page, not discovery/listing surfaces — the "hide disabled accounts from discovery" goal this plan targets doesn't clearly extend to "a disabled user's own channel page shows nothing," and login-gate action-gating (Task 3) already stops a disabled account from starting new streams, so the main way `GetByUser` could return a disabled author's stream is a narrow edge case (already live when disabled). Revisit if this turns out to matter in practice.
- **No frontend changes.** Everything in this plan is backend-only; nothing here changes an API response shape a frontend currently renders (the two filtered read paths already omit results for other reasons like `RES_ERR_USER_NOT_FOUND`, so a disabled author simply looks like "not found" rather than surfacing a new state the frontend needs to handle).

---

## File Structure

**`user` service, new files:**
- `dto/get_users_statuses_request.go`, `dto/get_users_statuses_response.go`
- `handlers/user/get_users_statuses_internal.go`
- `internal/testutil/fakes.go`
- `repositories/user/get_statuses_by_ids_test.go` is not possible (no DB test infra) — service/handler-level tests instead: `services/user_status_test.go`, `handlers/follow/follow_test.go`

**`user` service, modified files:**
- `domains/user.go` (add `GetStatusesByIds` to `UserRepository` interface)
- `repositories/user/get_statuses_by_ids.go` (new repo method file)
- `repositories/user/get_public_info_by_id.go`, `repositories/user/get_public_infos_by_ids.go` (add status filter)
- `repositories/user/get_user_by_api_key.go` (add `u.status` to SELECT)
- `services/user.go` (add `GetUsersStatuses` method)
- `services/follow.go` (gate `Follow`)
- `response/error.go` (add `RES_ERR_ACCOUNT_DISABLED`, mirroring auth's)
- `api/http.go` (register the new route)
- `cmd/main.go` — no change needed (no new dependency injected into anything in this service)

**`livestream` service:**
- Modify: `gateway/user/user.go` (add `GetUsersStatuses` to interface), `gateway/user/http/http.go` (implement it)
- Modify: `services/livestream/livestream.go` or wherever `LivestreamService` struct/constructor lives (add `userGateway` field/param — see Task 5 for the exact current constructor)
- Modify: `services/livestream/get_recommended_livestreams.go` (post-filter)
- Modify: `cmd/main.go` (construct + inject the user gateway)
- New: `internal/testutil/fakes.go`, `services/livestream/get_recommended_livestreams_test.go`

**`vod` service:**
- Modify: `gateway/user/user.go`, `gateway/user/http/http.go` (add `GetUsersStatuses`, same as livestream)
- Modify: `services/vod/vod.go` or wherever `VODService` struct/constructor lives (add `userGateway`)
- Modify: `services/vod/get_recommended.go` (post-filter)
- Modify: `services/vod_comment/create.go` (gate `CreateComment`)
- Modify: `cmd/main.go` (inject user gateway into `VODService` too — it's already injected into `VODCommentService`)
- New: `internal/testutil/fakes.go`, `services/vod/get_recommended_test.go`, `services/vod_comment/create_test.go`

**`transcode` service:**
- Modify: `gateway/user/dto.go` (add `Status` field)
- Modify: `rtmp/rtmp.go` (`onConnect` gate)
- No new tests — see Task 4's testability note.

---

### Task 1: User-service batch status endpoint

**Files:**
- Create: `backend/user/dto/get_users_statuses_request.go`, `backend/user/dto/get_users_statuses_response.go`
- Create: `backend/user/repositories/user/get_statuses_by_ids.go`
- Create: `backend/user/handlers/user/get_users_statuses_internal.go`
- Create: `backend/user/internal/testutil/fakes.go`
- Modify: `backend/user/domains/user.go`, `backend/user/services/user.go`, `backend/user/api/http.go`
- Test: `backend/user/services/user_status_test.go`, `backend/user/handlers/user/get_users_statuses_internal_test.go`

**Interfaces:**
- Produces: `UserRepository.GetStatusesByIds(ctx, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any])`; `UserService.GetUsersStatuses(ctx, userIds []uuid.UUID) (map[string]string, *response.Response[any])` (string-keyed map for a guaranteed-safe JSON shape); route `POST /v1/internal/users/statuses`; `testutil.FakeUserRepository` implementing the full `domains.UserRepository` interface.

- [ ] **Step 1: Add the shared test fakes**

Create `backend/user/internal/testutil/fakes.go`. This must implement the FULL `domains.UserRepository` interface (11 methods, from `backend/user/domains/user.go`) plus `domains.FollowRepository` (3 methods, from `backend/user/domains/follower.go`), since later tasks in this plan reuse this same file. Follow the exact function-field-fake pattern from `backend/auth/internal/testutil/fakes.go` (each method a `...Func` field, panics if called unset — that's intentional):

```go
package testutil

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

type FakeUserRepository struct {
	GetByIdFunc              func(ctx context.Context, userId uuid.UUID) (*domains.User, *response.Response[any])
	GetAllFunc               func(ctx context.Context, page int) ([]domains.User, *response.Response[any])
	GetByUsernameFunc        func(ctx context.Context, username string) (*domains.User, *response.Response[any])
	GetByEmailFunc           func(ctx context.Context, email string) (*domains.User, *response.Response[any])
	GetByAPIKeyFunc          func(ctx context.Context, apiKey uuid.UUID) (*domains.User, *response.Response[any])
	GetPublicInfoByIdFunc    func(ctx context.Context, userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, *response.Response[any])
	GetPublicInfosByIdsFunc  func(ctx context.Context, ids []uuid.UUID, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *response.Response[any])
	GetRecommendedPublicFunc func(ctx context.Context, excludeUserId *uuid.UUID, page, limit int) ([]dto.GetUserPublicResponseDTO, *response.Response[any])
	SearchUsersByUsernameFunc func(ctx context.Context, username string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *response.Response[any])
	CreateFunc               func(ctx context.Context, username string, email string, authProvider domains.AuthProvider) (*domains.User, *response.Response[any])
	UpdateFunc               func(ctx context.Context, user dto.UpdateUserRequestDTO) (*domains.User, *response.Response[any])
	UpdateStreamAPIKeyFunc   func(ctx context.Context, userId uuid.UUID, newKey string) *response.Response[any]
	UpdateProfilePictureFunc func(ctx context.Context, userId uuid.UUID, newProfilePictureURL string) *response.Response[any]
	UpdateBackgroundPictureFunc func(ctx context.Context, userId uuid.UUID, newBackgroundPictureURL string) *response.Response[any]
	GetStatusesByIdsFunc     func(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any])
}

func (f *FakeUserRepository) GetById(ctx context.Context, userId uuid.UUID) (*domains.User, *response.Response[any]) {
	return f.GetByIdFunc(ctx, userId)
}
func (f *FakeUserRepository) GetAll(ctx context.Context, page int) ([]domains.User, *response.Response[any]) {
	return f.GetAllFunc(ctx, page)
}
func (f *FakeUserRepository) GetByUsername(ctx context.Context, username string) (*domains.User, *response.Response[any]) {
	return f.GetByUsernameFunc(ctx, username)
}
func (f *FakeUserRepository) GetByEmail(ctx context.Context, email string) (*domains.User, *response.Response[any]) {
	return f.GetByEmailFunc(ctx, email)
}
func (f *FakeUserRepository) GetByAPIKey(ctx context.Context, apiKey uuid.UUID) (*domains.User, *response.Response[any]) {
	return f.GetByAPIKeyFunc(ctx, apiKey)
}
func (f *FakeUserRepository) GetPublicInfoById(ctx context.Context, userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, *response.Response[any]) {
	return f.GetPublicInfoByIdFunc(ctx, userId, authenticatedUserId)
}
func (f *FakeUserRepository) GetPublicInfosByIds(ctx context.Context, ids []uuid.UUID, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *response.Response[any]) {
	return f.GetPublicInfosByIdsFunc(ctx, ids, authenticatedUserId)
}
func (f *FakeUserRepository) GetRecommendedPublic(ctx context.Context, excludeUserId *uuid.UUID, page, limit int) ([]dto.GetUserPublicResponseDTO, *response.Response[any]) {
	return f.GetRecommendedPublicFunc(ctx, excludeUserId, page, limit)
}
func (f *FakeUserRepository) SearchUsersByUsername(ctx context.Context, username string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *response.Response[any]) {
	return f.SearchUsersByUsernameFunc(ctx, username, authenticatedUserId)
}
func (f *FakeUserRepository) Create(ctx context.Context, username string, email string, authProvider domains.AuthProvider) (*domains.User, *response.Response[any]) {
	return f.CreateFunc(ctx, username, email, authProvider)
}
func (f *FakeUserRepository) Update(ctx context.Context, user dto.UpdateUserRequestDTO) (*domains.User, *response.Response[any]) {
	return f.UpdateFunc(ctx, user)
}
func (f *FakeUserRepository) UpdateStreamAPIKey(ctx context.Context, userId uuid.UUID, newKey string) *response.Response[any] {
	return f.UpdateStreamAPIKeyFunc(ctx, userId, newKey)
}
func (f *FakeUserRepository) UpdateProfilePicture(ctx context.Context, userId uuid.UUID, newProfilePictureURL string) *response.Response[any] {
	return f.UpdateProfilePictureFunc(ctx, userId, newProfilePictureURL)
}
func (f *FakeUserRepository) UpdateBackgroundPicture(ctx context.Context, userId uuid.UUID, newBackgroundPictureURL string) *response.Response[any] {
	return f.UpdateBackgroundPictureFunc(ctx, userId, newBackgroundPictureURL)
}
func (f *FakeUserRepository) GetStatusesByIds(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any]) {
	return f.GetStatusesByIdsFunc(ctx, userIds)
}

type FakeFollowRepository struct {
	FollowUserFunc          func(ctx context.Context, followUser, followedUser uuid.UUID) *response.Response[any]
	UnfollowUserFunc        func(ctx context.Context, followUser, followedUser uuid.UUID) *response.Response[any]
	GetFollowedUserIdsFunc  func(ctx context.Context, followerId uuid.UUID) ([]uuid.UUID, *response.Response[any])
}

func (f *FakeFollowRepository) FollowUser(ctx context.Context, followUser, followedUser uuid.UUID) *response.Response[any] {
	return f.FollowUserFunc(ctx, followUser, followedUser)
}
func (f *FakeFollowRepository) UnfollowUser(ctx context.Context, followUser, followedUser uuid.UUID) *response.Response[any] {
	return f.UnfollowUserFunc(ctx, followUser, followedUser)
}
func (f *FakeFollowRepository) GetFollowedUserIds(ctx context.Context, followerId uuid.UUID) ([]uuid.UUID, *response.Response[any]) {
	return f.GetFollowedUserIdsFunc(ctx, followerId)
}
```

Note: `GetStatusesByIds` is added directly to the `UserRepository` interface in Step 4 below — this fake is written against the interface's final shape (including that new method) so it doesn't need revisiting later in this task.

- [ ] **Step 2: Write the failing service test**

Create `backend/user/services/user_status_test.go`:

```go
package services

import (
	"context"
	"testing"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/internal/testutil"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func TestGetUsersStatuses_ReturnsStringKeyedMap(t *testing.T) {
	userA := uuid.Must(uuid.NewV4())
	userB := uuid.Must(uuid.NewV4())

	userRepo := &testutil.FakeUserRepository{
		GetStatusesByIdsFunc: func(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any]) {
			return map[uuid.UUID]domains.UserStatus{
				userA: domains.UserStatusDisabled,
				userB: domains.UserStatusNormal,
			}, nil
		},
	}
	s := NewUserService(userRepo, nil, nil, nil, MinIOService{})

	statuses, err := s.GetUsersStatuses(context.Background(), []uuid.UUID{userA, userB})
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if statuses[userA.String()] != "disabled" {
		t.Errorf("statuses[%s] = %q, want %q", userA, statuses[userA.String()], "disabled")
	}
	if statuses[userB.String()] != "normal" {
		t.Errorf("statuses[%s] = %q, want %q", userB, statuses[userB.String()], "normal")
	}
}

func TestGetUsersStatuses_PropagatesRepoError(t *testing.T) {
	wantErr := response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_QUERY, nil, nil, nil)
	userRepo := &testutil.FakeUserRepository{
		GetStatusesByIdsFunc: func(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any]) {
			return nil, wantErr
		},
	}
	s := NewUserService(userRepo, nil, nil, nil, MinIOService{})

	_, err := s.GetUsersStatuses(context.Background(), []uuid.UUID{uuid.Must(uuid.NewV4())})
	if err != wantErr {
		t.Errorf("err = %+v, want %+v", err, wantErr)
	}
}
```

(`NewUserService`'s other three repo params — `livestreamInformationRepo`, `notificationRepo`, `followRepo` — and `minioService` are irrelevant to this test; `nil` interfaces and a zero-value `MinIOService{}` are fine since `GetUsersStatuses` never touches them.)

- [ ] **Step 2b: Write the failing handler test**

Create `backend/user/handlers/user/get_users_statuses_internal_test.go`:

```go
package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/internal/testutil"
	"sen1or/letslive/user/response"
	"sen1or/letslive/user/services"

	"github.com/gofrs/uuid/v5"
)

func TestGetUsersStatusesInternalHandler_ReturnsStatusMap(t *testing.T) {
	userA := uuid.Must(uuid.NewV4())
	userRepo := &testutil.FakeUserRepository{
		GetStatusesByIdsFunc: func(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any]) {
			return map[uuid.UUID]domains.UserStatus{userA: domains.UserStatusDisabled}, nil
		},
	}
	userService := services.NewUserService(userRepo, nil, nil, nil, services.MinIOService{})
	h := NewUserHandler(*userService)

	body, _ := json.Marshal(dto.GetUsersStatusesRequestDTO{UserIds: []uuid.UUID{userA}})
	req := httptest.NewRequest(http.MethodPost, "/v1/internal/users/statuses", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.GetUsersStatusesInternalHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var res response.Response[dto.GetUsersStatusesResponseDTO]
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %s", err)
	}
	if res.Data == nil || res.Data.Statuses[userA.String()] != "disabled" {
		t.Fatalf("expected statuses[%s] = disabled, got %+v", userA, res.Data)
	}
}

func TestGetUsersStatusesInternalHandler_InvalidPayloadReturnsError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/internal/users/statuses", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()

	userService := services.NewUserService(&testutil.FakeUserRepository{}, nil, nil, nil, services.MinIOService{})
	h := NewUserHandler(*userService)

	h.GetUsersStatusesInternalHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
```

(`NewUserHandler(userService services.UserService) *UserHandler` — confirmed at `backend/user/handlers/user/user.go:13` — so `NewUserHandler(*userService)` above is correct, dereferencing the `*UserService` that `services.NewUserService(...)` returns.)

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd backend/user && go build ./... && go test ./services/... ./handlers/user/... -run 'TestGetUsersStatuses' -v`
Expected: FAIL to build — `domains.UserStatus` has no matching repo method yet, `UserService.GetUsersStatuses`, `dto.GetUsersStatusesRequestDTO/ResponseDTO`, and `GetUsersStatusesInternalHandler` are all undefined.

- [ ] **Step 4: Add the DTOs**

`backend/user/dto/get_users_statuses_request.go`:

```go
package dto

import "github.com/gofrs/uuid/v5"

type GetUsersStatusesRequestDTO struct {
	UserIds []uuid.UUID `json:"userIds" validate:"required,min=1"`
}
```

`backend/user/dto/get_users_statuses_response.go`:

```go
package dto

type GetUsersStatusesResponseDTO struct {
	Statuses map[string]string `json:"statuses"`
}
```

- [ ] **Step 5: Add the repository method**

Modify `backend/user/domains/user.go` — add to the `UserRepository` interface (after `GetByAPIKey`):

```go
	GetStatusesByIds(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]UserStatus, *response.Response[any])
```

Create `backend/user/repositories/user/get_statuses_by_ids.go`:

```go
package user

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresUserRepo) GetStatusesByIds(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT id, status FROM users WHERE id = ANY($1::uuid[])
	`, userIds)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	defer rows.Close()

	statuses := make(map[uuid.UUID]domains.UserStatus, len(userIds))
	for rows.Next() {
		var id uuid.UUID
		var status domains.UserStatus
		if err := rows.Scan(&id, &status); err != nil {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_DATABASE_ISSUE,
				nil,
				nil,
				nil,
			)
		}
		statuses[id] = status
	}

	return statuses, nil
}
```

- [ ] **Step 6: Add the service method**

Append to `backend/user/services/user.go`:

```go
func (s *UserService) GetUsersStatuses(ctx context.Context, userIds []uuid.UUID) (map[string]string, *response.Response[any]) {
	statuses, err := s.userRepo.GetStatusesByIds(ctx, userIds)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(statuses))
	for id, status := range statuses {
		result[id.String()] = string(status)
	}

	return result, nil
}
```

- [ ] **Step 7: Add the handler and route**

Create `backend/user/handlers/user/get_users_statuses_internal.go` (match the exact `UserHandler` receiver/import style already used by sibling files in this package, e.g. `get_user_by_stream_api_key_internal.go`):

```go
package user

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/response"
	"sen1or/letslive/user/utils"
)

func (h *UserHandler) GetUsersStatusesInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var reqBody dto.GetUsersStatusesRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	if err := utils.Validator.Struct(&reqBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	statuses, sErr := h.userService.GetUsersStatuses(ctx, reqBody.UserIds)
	if sErr != nil {
		h.WriteResponse(w, ctx, sErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(
		response.RES_SUCC_OK,
		&dto.GetUsersStatusesResponseDTO{Statuses: statuses},
		nil,
		nil,
	))
}
```

(`utils.Validator` is `sen1or/letslive/user/utils` — confirmed by reading its existing use in `services/user.go`'s `UpdateUser`, same import path as used above.)

Modify `backend/user/api/http.go` — add right after the other internal routes (near `PUT /v1/user/{userId}` / `GET /v1/verify-stream-key`):

```go
	wrap("POST /v1/internal/users/statuses", a.userHandler.GetUsersStatusesInternalHandler) // internal
```

- [ ] **Step 8: Run tests to verify they pass**

Run: `cd backend/user && go build ./... && go test ./... -race -v`
Expected: PASS across the whole module.

- [ ] **Step 9: Commit**

```bash
git add backend/user/dto/get_users_statuses_request.go backend/user/dto/get_users_statuses_response.go backend/user/repositories/user/get_statuses_by_ids.go backend/user/handlers/user/get_users_statuses_internal.go backend/user/internal/testutil/fakes.go backend/user/domains/user.go backend/user/services/user.go backend/user/api/http.go backend/user/services/user_status_test.go backend/user/handlers/user/get_users_statuses_internal_test.go
git commit -m "$(cat <<'EOF'
feat(user): add internal batch user-status endpoint

Refs: docs/superpowers/plans/2026-08-21-disabled-account-enforcement-extension.md
EOF
)"
```

---

### Task 2: User-service read-path filtering + follow gating + stream-key status

**Files:**
- Modify: `backend/user/repositories/user/get_public_info_by_id.go`, `backend/user/repositories/user/get_public_infos_by_ids.go`, `backend/user/repositories/user/get_user_by_api_key.go`
- Modify: `backend/user/response/error.go`, `backend/user/services/follow.go`
- Test: `backend/user/services/follow_test.go`

**Interfaces:**
- Consumes: `testutil.FakeUserRepository`, `testutil.FakeFollowRepository` (Task 1)
- Produces: `response.RES_ERR_ACCOUNT_DISABLED` (user-service's own copy, mirroring auth's); `FollowService.Follow` now rejects when either party is disabled

- [ ] **Step 1: Write the failing follow-gating tests**

Create `backend/user/services/follow_test.go`:

```go
package services

import (
	"context"
	"testing"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/internal/testutil"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func TestFollow_RejectsWhenFollowerIsDisabled(t *testing.T) {
	follower := uuid.Must(uuid.NewV4())
	followed := uuid.Must(uuid.NewV4())

	userRepo := &testutil.FakeUserRepository{
		GetByIdFunc: func(ctx context.Context, id uuid.UUID) (*domains.User, *response.Response[any]) {
			if id == follower {
				return &domains.User{Id: id, Status: domains.UserStatusDisabled}, nil
			}
			return &domains.User{Id: id, Status: domains.UserStatusNormal}, nil
		},
	}
	called := false
	followRepo := &testutil.FakeFollowRepository{
		FollowUserFunc: func(ctx context.Context, a, b uuid.UUID) *response.Response[any] {
			called = true
			return nil
		},
	}
	s := NewFollowService(followRepo, userRepo)

	err := s.Follow(context.Background(), follower.String(), followed.String())
	if err == nil {
		t.Fatal("expected an error when the follower is disabled, got nil")
	}
	if called {
		t.Error("FollowUser should not be called when the follower is disabled")
	}
}

func TestFollow_RejectsWhenTargetIsDisabled(t *testing.T) {
	follower := uuid.Must(uuid.NewV4())
	followed := uuid.Must(uuid.NewV4())

	userRepo := &testutil.FakeUserRepository{
		GetByIdFunc: func(ctx context.Context, id uuid.UUID) (*domains.User, *response.Response[any]) {
			if id == followed {
				return &domains.User{Id: id, Status: domains.UserStatusDisabled}, nil
			}
			return &domains.User{Id: id, Status: domains.UserStatusNormal}, nil
		},
	}
	called := false
	followRepo := &testutil.FakeFollowRepository{
		FollowUserFunc: func(ctx context.Context, a, b uuid.UUID) *response.Response[any] {
			called = true
			return nil
		},
	}
	s := NewFollowService(followRepo, userRepo)

	err := s.Follow(context.Background(), follower.String(), followed.String())
	if err == nil {
		t.Fatal("expected an error when the target is disabled, got nil")
	}
	if called {
		t.Error("FollowUser should not be called when the target is disabled")
	}
}

func TestFollow_AllowsWhenBothNormal(t *testing.T) {
	follower := uuid.Must(uuid.NewV4())
	followed := uuid.Must(uuid.NewV4())

	userRepo := &testutil.FakeUserRepository{
		GetByIdFunc: func(ctx context.Context, id uuid.UUID) (*domains.User, *response.Response[any]) {
			return &domains.User{Id: id, Status: domains.UserStatusNormal}, nil
		},
	}
	called := false
	followRepo := &testutil.FakeFollowRepository{
		FollowUserFunc: func(ctx context.Context, a, b uuid.UUID) *response.Response[any] {
			called = true
			return nil
		},
	}
	s := NewFollowService(followRepo, userRepo)

	if err := s.Follow(context.Background(), follower.String(), followed.String()); err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if !called {
		t.Error("FollowUser should be called when both accounts are normal")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd backend/user && go test ./services/... -run TestFollow -v`
Expected: FAIL to build — `NewFollowService(followRepo, userRepo)` doesn't match the current one-argument constructor.

- [ ] **Step 3: Add the response code**

Modify `backend/user/response/error.go`. Confirmed current highest 300xx code in this file is `RES_ERR_INSUFFICIENT_INVENTORY_CODE = 30004`, so the new code is `30005`. Add to the code const block:

```go
	RES_ERR_ACCOUNT_DISABLED_CODE = 30005
```

Add to the key const block:

```go
	RES_ERR_ACCOUNT_DISABLED_KEY = "res_err_account_disabled"
```

Add to the `var (...)` template block:

```go
	RES_ERR_ACCOUNT_DISABLED = ResponseTemplate{
		Success:    false,
		StatusCode: http.StatusForbidden,
		Code:       RES_ERR_ACCOUNT_DISABLED_CODE,
		Key:        RES_ERR_ACCOUNT_DISABLED_KEY,
		Message:    "This account has been disabled.",
	}
```

- [ ] **Step 4: Gate `Follow`**

Modify `backend/user/services/follow.go` — add `userRepo domains.UserRepository` to `FollowService` and its constructor, and check both parties' status before delegating to the repo:

```go
package services

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

type FollowService struct {
	followRepo domains.FollowRepository
	userRepo   domains.UserRepository
}

func NewFollowService(
	followRepo domains.FollowRepository,
	userRepo domains.UserRepository,
) *FollowService {
	return &FollowService{
		followRepo: followRepo,
		userRepo:   userRepo,
	}
}

func (s FollowService) Follow(ctx context.Context, followId, followedId string) *response.Response[any] {
	followUUID, err1 := uuid.FromString(followId)
	followedUUID, err2 := uuid.FromString(followedId)
	if err1 != nil || err2 != nil || followId == followedId {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}

	follower, followerErr := s.userRepo.GetById(ctx, followUUID)
	if followerErr != nil {
		return followerErr
	}
	if follower.Status == domains.UserStatusDisabled {
		return response.NewResponseFromTemplate[any](response.RES_ERR_ACCOUNT_DISABLED, nil, nil, nil)
	}

	followed, followedErr := s.userRepo.GetById(ctx, followedUUID)
	if followedErr != nil {
		return followedErr
	}
	if followed.Status == domains.UserStatusDisabled {
		return response.NewResponseFromTemplate[any](response.RES_ERR_ACCOUNT_DISABLED, nil, nil, nil)
	}

	err := s.followRepo.FollowUser(ctx, followUUID, followedUUID)
	if err != nil {
		return err
	}

	return nil
}

func (s FollowService) Unfollow(ctx context.Context, followId, followedId string) *response.Response[any] {
	followUUID, err1 := uuid.FromString(followId)
	followedUUID, err2 := uuid.FromString(followedId)
	if err1 != nil || err2 != nil || followId == followedId {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}
	err := s.followRepo.UnfollowUser(ctx, followUUID, followedUUID)
	if err != nil {
		return err
	}

	return nil
}
```

Update the one call site — `backend/user/cmd/main.go`'s `SetupServer` currently has `var followService = services.NewFollowService(followRepo)`; change to `services.NewFollowService(followRepo, userRepo)` (`userRepo` is already constructed earlier in the same function).

- [ ] **Step 5: Filter the two read-path queries**

Modify `backend/user/repositories/user/get_public_info_by_id.go` — change the query's `WHERE` clause from `WHERE u.id = $1` to `WHERE u.id = $1 AND u.status != 'disabled'`.

Modify `backend/user/repositories/user/get_public_infos_by_ids.go` — change its `WHERE u.id = ANY($1::uuid[])` to `WHERE u.id = ANY($1::uuid[]) AND u.status != 'disabled'`.

(`get_user_by_username.go`'s `GetByUsername` is confirmed dead code — zero callers anywhere in the service — do not modify it as part of this plan.)

- [ ] **Step 6: Add status to stream-key verification**

Modify `backend/user/repositories/user/get_user_by_api_key.go` — add `u.status` to the `SELECT` column list:

```go
		SELECT u.id, u.username, u.email, u.status, u.created_at, u.stream_api_key, u.phone_number, u.bio, u.profile_picture, u.background_picture, l.user_id, l.title, l.description, l.thumbnail_url
		FROM users u
		JOIN livestream_information l ON u.id = l.user_id
		WHERE u.stream_api_key = $1
```

No other change needed in this file or its handler/service — `domains.User.Status` already has the right `db:"status"` tag and the handler already serializes the whole `*domains.User`, so the `status` field will appear in the `GET /v1/verify-stream-key` JSON response automatically.

- [ ] **Step 7: Run tests to verify they pass**

Run: `cd backend/user && go build ./... && go test ./... -race -v`
Expected: PASS across the whole module.

- [ ] **Step 8: Commit**

```bash
git add backend/user/services/follow.go backend/user/services/follow_test.go backend/user/response/error.go backend/user/repositories/user/get_public_info_by_id.go backend/user/repositories/user/get_public_infos_by_ids.go backend/user/repositories/user/get_user_by_api_key.go backend/user/cmd/main.go
git commit -m "$(cat <<'EOF'
feat(user): filter disabled users from profile/following reads, gate follow, expose status on stream-key verify

Refs: docs/superpowers/plans/2026-08-21-disabled-account-enforcement-extension.md
EOF
)"
```

---

### Task 3: Transcode — gate stream start

**Files:**
- Modify: `backend/transcode/gateway/user/dto.go`, `backend/transcode/rtmp/rtmp.go`

**Interfaces:**
- Consumes: `status` field now present in `GET /v1/verify-stream-key`'s JSON response (Task 2, Step 6)
- Produces: `GetUserResponseDTO.Status string`; `onConnect` now returns an error for a disabled account before creating a livestream

**Testability note (documented, not silently skipped):** `onConnect` is deeply coupled to the `joy5` RTMP library's connection types and this plan does not restructure it to be independently unit-testable — that would be a larger refactor than this feature warrants (same category of judgment call as the Google-OAuth-handler testability note in the login-gate plan: the new logic here is a two-line status comparison, not complex branching logic, and the surrounding function was already untested before this change). Verify by direct code reading and the full-module build/vet passing, not a new test.

- [ ] **Step 1: Add the `Status` field**

Modify `backend/transcode/gateway/user/dto.go` — add to `GetUserResponseDTO` (non-pointer, no `omitempty`, matching how `domains.User.Status` is represented — this field is a plain `string` here since this DTO doesn't import the `user` service's domain package):

```go
type GetUserResponseDTO struct {
	Id                uuid.UUID `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"createdAt"`
	StreamAPIKey      uuid.UUID `json:"streamAPIKey"`
	PhoneNumber       *string   `json:"phoneNumber,omitempty"`
	Bio               *string   `json:"bio,omitempty"`
	ProfilePicture    *string   `json:"profilePicture,omitempty"`
	BackgroundPicture *string   `json:"backgroundPicture,omitempty"`

	LivestreamInformationResponseDTO `json:"livestreamInformation"`
}
```

- [ ] **Step 2: Gate `onConnect`**

Modify `backend/transcode/rtmp/rtmp.go` — in `onConnect`, right after the existing `userInfo, errRes := s.userGateway.GetUserInformation(reqCtx, streamingKey)` error check and before building `streamDTO`:

```go
	userInfo, errRes := s.userGateway.GetUserInformation(reqCtx, streamingKey)
	if errRes != nil {
		return "", "", fmt.Errorf("failed to get user information: %s", errRes.Message)
	}

	if userInfo.Data.Status == "disabled" {
		return "", "", fmt.Errorf("account is disabled")
	}

	thumb := userInfo.Data.LivestreamInformationResponseDTO.ThumbnailURL
```

(The rest of `onConnect` is unchanged — this fails closed: `HandleConnection`'s existing caller already closes the connection on any non-nil error from `onConnect`, so no new error-handling path is needed.)

- [ ] **Step 3: Verify**

Run: `cd backend/transcode && go build ./... && go vet ./...`
Expected: clean build, no vet warnings.

- [ ] **Step 4: Commit**

```bash
git add backend/transcode/gateway/user/dto.go backend/transcode/rtmp/rtmp.go
git commit -m "$(cat <<'EOF'
fix(transcode): reject stream start for disabled accounts

Refs: docs/superpowers/plans/2026-08-21-disabled-account-enforcement-extension.md
EOF
)"
```

---

### Task 4: Livestream — batch gateway + discovery filtering

**Files:**
- Modify: `backend/livestream/gateway/user/user.go`, `backend/livestream/gateway/user/http/http.go`
- Modify: `backend/livestream/services/livestream/livestream.go` (confirmed struct/constructor file)
- Modify: `backend/livestream/services/livestream/get_recommended_livestreams.go`
- Modify: `backend/livestream/cmd/main.go`
- Create: `backend/livestream/internal/testutil/fakes.go`
- Test: `backend/livestream/services/livestream/get_recommended_livestreams_test.go`

**Interfaces:**
- Produces: `UserGateway.GetUsersStatuses(ctx, userIds []uuid.UUID) (map[string]string, error)`; `LivestreamService` gains a `userGateway` dependency; `GetRecommendedLivestreams` now excludes disabled authors, fail-open on gateway error (logs and returns the unfiltered list)

- [ ] **Step 1: Add the fakes**

Create `backend/livestream/internal/testutil/fakes.go`:

```go
package testutil

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

type FakeLivestreamRepository struct {
	GetByIdFunc                    func(ctx context.Context, id uuid.UUID) (*domains.Livestream, *response.Response[any])
	GetByUserFunc                  func(ctx context.Context, userId uuid.UUID) (*domains.Livestream, *response.Response[any])
	GetRecommendedLivestreamsFunc  func(ctx context.Context, page, limit int) ([]domains.Livestream, *response.Response[any])
	CreateFunc                     func(ctx context.Context, ls domains.Livestream) (*domains.Livestream, *response.Response[any])
	UpdateFunc                     func(ctx context.Context, ls domains.Livestream) (*domains.Livestream, *response.Response[any])
	DeleteFunc                     func(ctx context.Context, id uuid.UUID) *response.Response[any]
}

func (f *FakeLivestreamRepository) GetById(ctx context.Context, id uuid.UUID) (*domains.Livestream, *response.Response[any]) {
	return f.GetByIdFunc(ctx, id)
}
func (f *FakeLivestreamRepository) GetByUser(ctx context.Context, userId uuid.UUID) (*domains.Livestream, *response.Response[any]) {
	return f.GetByUserFunc(ctx, userId)
}
func (f *FakeLivestreamRepository) GetRecommendedLivestreams(ctx context.Context, page, limit int) ([]domains.Livestream, *response.Response[any]) {
	return f.GetRecommendedLivestreamsFunc(ctx, page, limit)
}
func (f *FakeLivestreamRepository) Create(ctx context.Context, ls domains.Livestream) (*domains.Livestream, *response.Response[any]) {
	return f.CreateFunc(ctx, ls)
}
func (f *FakeLivestreamRepository) Update(ctx context.Context, ls domains.Livestream) (*domains.Livestream, *response.Response[any]) {
	return f.UpdateFunc(ctx, ls)
}
func (f *FakeLivestreamRepository) Delete(ctx context.Context, id uuid.UUID) *response.Response[any] {
	return f.DeleteFunc(ctx, id)
}

type FakeUserGateway struct {
	GetUsersStatusesFunc func(ctx context.Context, userIds []uuid.UUID) (map[string]string, error)
}

func (f *FakeUserGateway) GetUsersStatuses(ctx context.Context, userIds []uuid.UUID) (map[string]string, error) {
	return f.GetUsersStatusesFunc(ctx, userIds)
}
```

This matches `backend/livestream/domains/livestream.go`'s `LivestreamRepository` interface exactly (6 methods: `GetById`, `GetByUser`, `GetRecommendedLivestreams`, `Create`, `Update`, `Delete`) — confirmed by reading that file directly, not assumed.

- [ ] **Step 2: Write the failing test**

Create `backend/livestream/services/livestream/get_recommended_livestreams_test.go`:

```go
package livestream

import (
	"context"
	"testing"

	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/internal/testutil"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func TestGetRecommendedLivestreams_ExcludesDisabledAuthors(t *testing.T) {
	normalUser := uuid.Must(uuid.NewV4())
	disabledUser := uuid.Must(uuid.NewV4())

	livestreamRepo := &testutil.FakeLivestreamRepository{
		GetRecommendedLivestreamsFunc: func(ctx context.Context, page, limit int) ([]domains.Livestream, *response.Response[any]) {
			return []domains.Livestream{
				{UserId: normalUser},
				{UserId: disabledUser},
			}, nil
		},
	}
	userGateway := &testutil.FakeUserGateway{
		GetUsersStatusesFunc: func(ctx context.Context, userIds []uuid.UUID) (map[string]string, error) {
			return map[string]string{
				normalUser.String():   "normal",
				disabledUser.String(): "disabled",
			}, nil
		},
	}
	s := NewLivestreamService(livestreamRepo, nil, userGateway)

	result, err := s.GetRecommendedLivestreams(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("expected no error, got %+v", err)
	}
	if len(result) != 1 || result[0].UserId != normalUser {
		t.Errorf("result = %+v, want exactly one livestream from %s", result, normalUser)
	}
}

func TestGetRecommendedLivestreams_FailsOpenOnGatewayError(t *testing.T) {
	someUser := uuid.Must(uuid.NewV4())
	livestreamRepo := &testutil.FakeLivestreamRepository{
		GetRecommendedLivestreamsFunc: func(ctx context.Context, page, limit int) ([]domains.Livestream, *response.Response[any]) {
			return []domains.Livestream{{UserId: someUser}}, nil
		},
	}
	userGateway := &testutil.FakeUserGateway{
		GetUsersStatusesFunc: func(ctx context.Context, userIds []uuid.UUID) (map[string]string, error) {
			return nil, errUnavailable
		},
	}
	s := NewLivestreamService(livestreamRepo, nil, userGateway)

	result, err := s.GetRecommendedLivestreams(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("expected no error (fail-open), got %+v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected the unfiltered list on gateway failure, got %+v", result)
	}
}

var errUnavailable = &testError{}

type testError struct{}

func (e *testError) Error() string { return "user service unavailable" }
```

`backend/livestream/services/livestream/livestream.go` currently defines:
```go
type LivestreamService struct {
	livestreamRepo domains.LivestreamRepository
	vodGateway     vodgateway.VODGateway
}

func NewLivestreamService(livestreamRepo domains.LivestreamRepository, vodGateway vodgateway.VODGateway) *LivestreamService {
```
Step 5 below adds `userGateway` as a third field/param, giving `NewLivestreamService(livestreamRepo, vodGateway, userGateway)` — so in the test above, `NewLivestreamService(livestreamRepo, nil, userGateway)` passes `nil` for `vodGateway` (untouched by `GetRecommendedLivestreams`) with the real, final parameter order.

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd backend/livestream && go test ./services/livestream/... -run TestGetRecommendedLivestreams -v`
Expected: FAIL to build — `NewLivestreamService` doesn't yet take a `userGateway` param, `GetUsersStatuses` doesn't exist on the gateway interface.

- [ ] **Step 4: Add the gateway method**

Modify `backend/livestream/gateway/user/user.go` — add to `UserGateway`:

```go
	GetUsersStatuses(ctx context.Context, userIds []uuid.UUID) (map[string]string, error)
```

Modify `backend/livestream/gateway/user/http/http.go` — append (matching the file's existing style exactly, including the local `userServiceResponse`-shaped wrapper for this new response):

```go
type getUsersStatusesRequest struct {
	UserIds []uuid.UUID `json:"userIds"`
}

type getUsersStatusesResponse struct {
	Success bool               `json:"success"`
	Data    *getUsersStatusesData `json:"data,omitempty"`
}

type getUsersStatusesData struct {
	Statuses map[string]string `json:"statuses"`
}

func (g *userHTTPGateway) GetUsersStatuses(ctx context.Context, userIds []uuid.UUID) (map[string]string, error) {
	addr, err := g.registry.ServiceAddress(ctx, "user")
	if err != nil {
		logger.Errorf(ctx, "failed to get user service address: %v", err)
		return nil, fmt.Errorf("user service unavailable")
	}

	payloadBuf := new(bytes.Buffer)
	if err := json.NewEncoder(payloadBuf).Encode(&getUsersStatusesRequest{UserIds: userIds}); err != nil {
		logger.Errorf(ctx, "failed to encode statuses request: %v", err)
		return nil, fmt.Errorf("failed to encode request")
	}

	url := fmt.Sprintf("http://%s/v1/internal/users/statuses", addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payloadBuf)
	if err != nil {
		logger.Errorf(ctx, "failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")

	if err := gateway.SetRequestIDHeader(ctx, req); err != nil {
		logger.Warnf(ctx, "failed to set request id header: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Errorf(ctx, "failed to call user service: %v", err)
		return nil, fmt.Errorf("failed to call user service")
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	var result getUsersStatusesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf(ctx, "failed to decode user service response: %v", err)
		return nil, fmt.Errorf("failed to decode user service response")
	}
	if result.Data == nil {
		return nil, fmt.Errorf("user service returned no data")
	}

	return result.Data.Statuses, nil
}
```

Add `"bytes"` to this file's import block (not currently imported — every other import in the block is already present per the file dump in this plan's design record).

- [ ] **Step 5: Wire `userGateway` into `LivestreamService` and filter**

Modify `backend/livestream/services/livestream/livestream.go`:

```go
package livestream

import (
	"sen1or/letslive/livestream/domains"
	usergateway "sen1or/letslive/livestream/gateway/user"
	vodgateway "sen1or/letslive/livestream/gateway/vod"
)

type LivestreamService struct {
	livestreamRepo domains.LivestreamRepository
	vodGateway     vodgateway.VODGateway
	userGateway    usergateway.UserGateway
}

func NewLivestreamService(livestreamRepo domains.LivestreamRepository, vodGateway vodgateway.VODGateway, userGateway usergateway.UserGateway) *LivestreamService {
	return &LivestreamService{
		livestreamRepo: livestreamRepo,
		vodGateway:     vodGateway,
		userGateway:    userGateway,
	}
}
```

Modify `backend/livestream/services/livestream/get_recommended_livestreams.go`:

```go
package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

func (s *LivestreamService) GetRecommendedLivestreams(ctx context.Context, page int, limit int) ([]domains.Livestream, *response.Response[any]) {
	if page < 0 {
		page = 0
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 50 {
		limit = 50
	}

	livestreams, err := s.livestreamRepo.GetRecommendedLivestreams(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	if len(livestreams) == 0 {
		return livestreams, nil
	}

	userIdSet := make(map[uuid.UUID]struct{}, len(livestreams))
	for _, l := range livestreams {
		userIdSet[l.UserId] = struct{}{}
	}
	userIds := make([]uuid.UUID, 0, len(userIdSet))
	for id := range userIdSet {
		userIds = append(userIds, id)
	}

	statuses, statusErr := s.userGateway.GetUsersStatuses(ctx, userIds)
	if statusErr != nil {
		logger.Errorf(ctx, "failed to fetch author statuses, returning unfiltered results: %v", statusErr)
		return livestreams, nil
	}

	filtered := make([]domains.Livestream, 0, len(livestreams))
	for _, l := range livestreams {
		if statuses[l.UserId.String()] != "disabled" {
			filtered = append(filtered, l)
		}
	}

	return filtered, nil
}
```

- [ ] **Step 6: Wire in `main.go`**

Modify `backend/livestream/cmd/main.go` — construct a user gateway and pass it into the service constructor:

```go
	usergatewayhttp "sen1or/letslive/livestream/gateway/user/http"
```

(add to imports, alongside the existing `vodgatewayhttp` import)

```go
	var userGateway = usergatewayhttp.NewUserGateway(registry)
	var livestreamService = livestreamService.NewLivestreamService(livestreamRepo, vodGateway, userGateway)
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `cd backend/livestream && go build ./... && go test ./... -race -v`
Expected: PASS across the whole module.

- [ ] **Step 8: Commit**

```bash
git add backend/livestream/gateway/user/user.go backend/livestream/gateway/user/http/http.go backend/livestream/services/livestream backend/livestream/cmd/main.go backend/livestream/internal/testutil/fakes.go
git commit -m "$(cat <<'EOF'
feat(livestream): filter disabled authors out of stream discovery

Refs: docs/superpowers/plans/2026-08-21-disabled-account-enforcement-extension.md
EOF
)"
```

---

### Task 5: VOD — batch gateway + popular-VOD filtering + comment gating

**Files:**
- Modify: `backend/vod/gateway/user/user.go`, `backend/vod/gateway/user/http/http.go`
- Modify: `backend/vod/services/vod/vod.go` (confirmed struct/constructor file)
- Modify: `backend/vod/services/vod/get_recommended.go`
- Modify: `backend/vod/services/vod_comment/create.go`
- Modify: `backend/vod/cmd/main.go`
- Create: `backend/vod/internal/testutil/fakes.go`
- Test: `backend/vod/services/vod/get_recommended_test.go`, `backend/vod/services/vod_comment/create_test.go`

**Interfaces:**
- Consumes: same `GetUsersStatuses(ctx, userIds []uuid.UUID) (map[string]string, error)` shape as Task 4, added to `vod`'s own (currently byte-identical) `gateway/user` package
- Produces: `VODService` gains a `userGateway` dependency; `GetRecommendedVODs` excludes disabled authors (fail-open); `CreateComment` rejects when the commenting user is disabled (fail-closed)

- [ ] **Step 1: Add the fakes**

Create `backend/vod/internal/testutil/fakes.go`. This matches `backend/vod/domains/vod.go`'s `VODRepository` (9 methods) and `backend/vod/domains/vod_comment.go`'s `VODCommentRepository` (11 methods, including `WithTx` which the fake satisfies by returning itself) and `VODCommentLikeRepository` (6 methods) — all confirmed by reading those two files directly:

```go
package testutil

import (
	"context"
	"sen1or/letslive/vod/domains"
	"sen1or/letslive/vod/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

type FakeUserGateway struct {
	GetUsersStatusesFunc func(ctx context.Context, userIds []uuid.UUID) (map[string]string, error)
}

func (f *FakeUserGateway) GetUsersStatuses(ctx context.Context, userIds []uuid.UUID) (map[string]string, error) {
	return f.GetUsersStatusesFunc(ctx, userIds)
}

type FakeVODRepository struct {
	GetByIdFunc            func(ctx context.Context, id uuid.UUID) (*domains.VOD, *response.Response[any])
	GetByUserFunc          func(ctx context.Context, userId uuid.UUID, page int, limit int) ([]domains.VOD, *response.Response[any])
	GetPublicVODsByUserFunc func(ctx context.Context, userId uuid.UUID, page int, limit int) ([]domains.VOD, *response.Response[any])
	GetPopularFunc         func(ctx context.Context, page int, limit int) ([]domains.VOD, *response.Response[any])
	IncrementViewCountFunc func(ctx context.Context, id uuid.UUID) *response.Response[any]
	CreateFunc             func(ctx context.Context, vod domains.VOD) (*domains.VOD, *response.Response[any])
	UpdateFunc             func(ctx context.Context, vod domains.VOD) (*domains.VOD, *response.Response[any])
	UpdateStatusFunc       func(ctx context.Context, vodId uuid.UUID, status domains.VODStatus, playbackUrl *string, thumbnailUrl *string) *response.Response[any]
	DeleteFunc             func(ctx context.Context, id uuid.UUID) *response.Response[any]
}

func (f *FakeVODRepository) GetById(ctx context.Context, id uuid.UUID) (*domains.VOD, *response.Response[any]) {
	return f.GetByIdFunc(ctx, id)
}
func (f *FakeVODRepository) GetByUser(ctx context.Context, userId uuid.UUID, page int, limit int) ([]domains.VOD, *response.Response[any]) {
	return f.GetByUserFunc(ctx, userId, page, limit)
}
func (f *FakeVODRepository) GetPublicVODsByUser(ctx context.Context, userId uuid.UUID, page int, limit int) ([]domains.VOD, *response.Response[any]) {
	return f.GetPublicVODsByUserFunc(ctx, userId, page, limit)
}
func (f *FakeVODRepository) GetPopular(ctx context.Context, page int, limit int) ([]domains.VOD, *response.Response[any]) {
	return f.GetPopularFunc(ctx, page, limit)
}
func (f *FakeVODRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) *response.Response[any] {
	return f.IncrementViewCountFunc(ctx, id)
}
func (f *FakeVODRepository) Create(ctx context.Context, vod domains.VOD) (*domains.VOD, *response.Response[any]) {
	return f.CreateFunc(ctx, vod)
}
func (f *FakeVODRepository) Update(ctx context.Context, vod domains.VOD) (*domains.VOD, *response.Response[any]) {
	return f.UpdateFunc(ctx, vod)
}
func (f *FakeVODRepository) UpdateStatus(ctx context.Context, vodId uuid.UUID, status domains.VODStatus, playbackUrl *string, thumbnailUrl *string) *response.Response[any] {
	return f.UpdateStatusFunc(ctx, vodId, status, playbackUrl, thumbnailUrl)
}
func (f *FakeVODRepository) Delete(ctx context.Context, id uuid.UUID) *response.Response[any] {
	return f.DeleteFunc(ctx, id)
}

type FakeVODCommentRepository struct {
	GetByVODIdFunc          func(ctx context.Context, vodId uuid.UUID, page int, limit int) ([]domains.VODComment, *response.Response[any])
	CountByVODIdFunc        func(ctx context.Context, vodId uuid.UUID) (int, *response.Response[any])
	CountRepliesFunc        func(ctx context.Context, parentId uuid.UUID) (int, *response.Response[any])
	GetRepliesFunc          func(ctx context.Context, parentId uuid.UUID, page int, limit int) ([]domains.VODComment, *response.Response[any])
	GetByIdFunc             func(ctx context.Context, id uuid.UUID) (*domains.VODComment, *response.Response[any])
	CreateFunc              func(ctx context.Context, comment domains.VODComment) (*domains.VODComment, *response.Response[any])
	IncrementReplyCountFunc func(ctx context.Context, commentId uuid.UUID) *response.Response[any]
	DecrementReplyCountFunc func(ctx context.Context, commentId uuid.UUID) *response.Response[any]
	SoftDeleteFunc          func(ctx context.Context, id uuid.UUID) *response.Response[any]
	HardDeleteFunc          func(ctx context.Context, id uuid.UUID) *response.Response[any]
}

func (f *FakeVODCommentRepository) WithTx(tx pgx.Tx) domains.VODCommentRepository { return f }
func (f *FakeVODCommentRepository) GetByVODId(ctx context.Context, vodId uuid.UUID, page int, limit int) ([]domains.VODComment, *response.Response[any]) {
	return f.GetByVODIdFunc(ctx, vodId, page, limit)
}
func (f *FakeVODCommentRepository) CountByVODId(ctx context.Context, vodId uuid.UUID) (int, *response.Response[any]) {
	return f.CountByVODIdFunc(ctx, vodId)
}
func (f *FakeVODCommentRepository) CountReplies(ctx context.Context, parentId uuid.UUID) (int, *response.Response[any]) {
	return f.CountRepliesFunc(ctx, parentId)
}
func (f *FakeVODCommentRepository) GetReplies(ctx context.Context, parentId uuid.UUID, page int, limit int) ([]domains.VODComment, *response.Response[any]) {
	return f.GetRepliesFunc(ctx, parentId, page, limit)
}
func (f *FakeVODCommentRepository) GetById(ctx context.Context, id uuid.UUID) (*domains.VODComment, *response.Response[any]) {
	return f.GetByIdFunc(ctx, id)
}
func (f *FakeVODCommentRepository) Create(ctx context.Context, comment domains.VODComment) (*domains.VODComment, *response.Response[any]) {
	return f.CreateFunc(ctx, comment)
}
func (f *FakeVODCommentRepository) IncrementReplyCount(ctx context.Context, commentId uuid.UUID) *response.Response[any] {
	return f.IncrementReplyCountFunc(ctx, commentId)
}
func (f *FakeVODCommentRepository) DecrementReplyCount(ctx context.Context, commentId uuid.UUID) *response.Response[any] {
	return f.DecrementReplyCountFunc(ctx, commentId)
}
func (f *FakeVODCommentRepository) SoftDelete(ctx context.Context, id uuid.UUID) *response.Response[any] {
	return f.SoftDeleteFunc(ctx, id)
}
func (f *FakeVODCommentRepository) HardDelete(ctx context.Context, id uuid.UUID) *response.Response[any] {
	return f.HardDeleteFunc(ctx, id)
}

type FakeVODCommentLikeRepository struct {
	GetUserLikedCommentIdsFunc func(ctx context.Context, commentIds []uuid.UUID, userId uuid.UUID) ([]uuid.UUID, *response.Response[any])
	InsertLikeFunc             func(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) *response.Response[any]
	DeleteLikeFunc             func(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) *response.Response[any]
	IncrementLikeCountFunc     func(ctx context.Context, commentId uuid.UUID) *response.Response[any]
	DecrementLikeCountFunc     func(ctx context.Context, commentId uuid.UUID) *response.Response[any]
}

func (f *FakeVODCommentLikeRepository) WithTx(tx pgx.Tx) domains.VODCommentLikeRepository { return f }
func (f *FakeVODCommentLikeRepository) GetUserLikedCommentIds(ctx context.Context, commentIds []uuid.UUID, userId uuid.UUID) ([]uuid.UUID, *response.Response[any]) {
	return f.GetUserLikedCommentIdsFunc(ctx, commentIds, userId)
}
func (f *FakeVODCommentLikeRepository) InsertLike(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) *response.Response[any] {
	return f.InsertLikeFunc(ctx, commentId, userId)
}
func (f *FakeVODCommentLikeRepository) DeleteLike(ctx context.Context, commentId uuid.UUID, userId uuid.UUID) *response.Response[any] {
	return f.DeleteLikeFunc(ctx, commentId, userId)
}
func (f *FakeVODCommentLikeRepository) IncrementLikeCount(ctx context.Context, commentId uuid.UUID) *response.Response[any] {
	return f.IncrementLikeCountFunc(ctx, commentId)
}
func (f *FakeVODCommentLikeRepository) DecrementLikeCount(ctx context.Context, commentId uuid.UUID) *response.Response[any] {
	return f.DecrementLikeCountFunc(ctx, commentId)
}
```

- [ ] **Step 2: Write the failing tests**

Create `backend/vod/services/vod/get_recommended_test.go` — mirror Task 4's `get_recommended_livestreams_test.go` structure exactly (same two cases: excludes disabled authors, fails open on gateway error), using `testutil.FakeVODRepository{GetPopularFunc: ...}`, `testutil.FakeUserGateway`, and `domains.VOD{UserId: ...}`. `NewVODService`'s real signature (confirmed at `backend/vod/services/vod/vod.go`) is `(vodRepo domains.VODRepository, transcodeJobRepo domains.TranscodeJobRepository, minioStorage *miniostorage.MinIOStorage)`; Step 5 below adds `userGateway` as a fourth parameter, so the test constructs with `NewVODService(vodRepo, nil, nil, userGateway)` (both `transcodeJobRepo` and `minioStorage` are untouched by `GetRecommendedVODs`, `nil` is safe for both).

Create `backend/vod/services/vod_comment/create_test.go`:

```go
package vodcomment

import (
	"context"
	"testing"

	"sen1or/letslive/vod/domains"
	"sen1or/letslive/vod/dto"
	"sen1or/letslive/vod/internal/testutil"
	"sen1or/letslive/vod/response"

	"github.com/gofrs/uuid/v5"
)

func TestCreateComment_RejectsWhenCommenterIsDisabled(t *testing.T) {
	commenterId := uuid.Must(uuid.NewV4())
	vodId := uuid.Must(uuid.NewV4())

	vodRepo := &testutil.FakeVODRepository{
		GetByIdFunc: func(ctx context.Context, id uuid.UUID) (*domains.VOD, *response.Response[any]) {
			return &domains.VOD{Id: id}, nil
		},
	}
	commentRepoCalled := false
	commentRepo := &testutil.FakeVODCommentRepository{
		CreateFunc: func(ctx context.Context, comment domains.VODComment) (*domains.VODComment, *response.Response[any]) {
			commentRepoCalled = true
			return &comment, nil
		},
	}
	userGateway := &testutil.FakeUserGateway{
		GetUsersStatusesFunc: func(ctx context.Context, userIds []uuid.UUID) (map[string]string, error) {
			return map[string]string{commenterId.String(): "disabled"}, nil
		},
	}
	s := NewVODCommentService(commentRepo, &testutil.FakeVODCommentLikeRepository{}, vodRepo, userGateway, nil)

	_, err := s.CreateComment(context.Background(), dto.CreateVODCommentRequestDTO{Content: "hello"}, vodId, commenterId)
	if err == nil {
		t.Fatal("expected an error when the commenter is disabled, got nil")
	}
	if commentRepoCalled {
		t.Error("comment should not be created when the commenter is disabled")
	}
}

func TestCreateComment_FailsClosedOnGatewayError(t *testing.T) {
	commenterId := uuid.Must(uuid.NewV4())
	vodId := uuid.Must(uuid.NewV4())

	vodRepo := &testutil.FakeVODRepository{
		GetByIdFunc: func(ctx context.Context, id uuid.UUID) (*domains.VOD, *response.Response[any]) {
			return &domains.VOD{Id: id}, nil
		},
	}
	userGateway := &testutil.FakeUserGateway{
		GetUsersStatusesFunc: func(ctx context.Context, userIds []uuid.UUID) (map[string]string, error) {
			return nil, errUnavailable
		},
	}
	s := NewVODCommentService(&testutil.FakeVODCommentRepository{}, &testutil.FakeVODCommentLikeRepository{}, vodRepo, userGateway, nil)

	_, err := s.CreateComment(context.Background(), dto.CreateVODCommentRequestDTO{Content: "hello"}, vodId, commenterId)
	if err == nil {
		t.Fatal("expected an error when the status check itself fails (fail-closed), got nil")
	}
}

var errUnavailable = &testError{}

type testError struct{}

func (e *testError) Error() string { return "user service unavailable" }
```

`NewVODCommentService`'s real signature (`backend/vod/services/vod_comment/vod_comment.go`) is `(commentRepo domains.VODCommentRepository, commentLikeRepo domains.VODCommentLikeRepository, vodRepo domains.VODRepository, userGateway usergateway.UserGateway, dbPool *pgxpool.Pool)` — `dbPool` stays `nil` here since `CreateComment`'s non-reply path (used by both tests above, no `ParentId` set) never touches it; only `createReplyWithTransaction` does.

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd backend/vod && go build ./... && go test ./services/vod/... ./services/vod_comment/... -run 'TestGetRecommendedVODs|TestCreateComment' -v`
Expected: FAIL to build — `GetUsersStatuses` doesn't exist on the gateway yet, `NewVODService`/`NewVODCommentService` don't take the params these tests assume, `CreateComment` doesn't check status yet.

- [ ] **Step 4: Add the gateway method**

Modify `backend/vod/gateway/user/user.go` and `backend/vod/gateway/user/http/http.go` — identical change to Task 4 Step 4, applied to `vod`'s copy of this package (same byte-for-byte starting content, confirmed in the design record above).

- [ ] **Step 5: Wire `userGateway` into `VODService` and filter**

Modify `backend/vod/services/vod/vod.go`:

```go
package vod

import (
	"sen1or/letslive/vod/domains"
	usergateway "sen1or/letslive/vod/gateway/user"
	miniostorage "sen1or/letslive/vod/storage/minio"
)

type VODService struct {
	vodRepo          domains.VODRepository
	transcodeJobRepo domains.TranscodeJobRepository
	minioStorage     *miniostorage.MinIOStorage
	userGateway      usergateway.UserGateway
}

func NewVODService(vodRepo domains.VODRepository, transcodeJobRepo domains.TranscodeJobRepository, minioStorage *miniostorage.MinIOStorage, userGateway usergateway.UserGateway) *VODService {
	return &VODService{
		vodRepo:          vodRepo,
		transcodeJobRepo: transcodeJobRepo,
		minioStorage:     minioStorage,
		userGateway:      userGateway,
	}
}
```

Modify `backend/vod/services/vod/get_recommended.go` — same post-fetch-filter pattern as Task 4 Step 5, adapted to `domains.VOD` and this file's existing `GetPopular` call:

```go
package vod

import (
	"context"
	"sen1or/letslive/vod/domains"
	response "sen1or/letslive/vod/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) GetRecommendedVODs(ctx context.Context, page int, limit int) ([]domains.VOD, *response.Response[any]) {
	if page < 0 {
		page = 0
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 50 {
		limit = 50
	}

	vods, err := s.vodRepo.GetPopular(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	if len(vods) == 0 {
		return vods, nil
	}

	userIdSet := make(map[uuid.UUID]struct{}, len(vods))
	for _, v := range vods {
		userIdSet[v.UserId] = struct{}{}
	}
	userIds := make([]uuid.UUID, 0, len(userIdSet))
	for id := range userIdSet {
		userIds = append(userIds, id)
	}

	statuses, statusErr := s.userGateway.GetUsersStatuses(ctx, userIds)
	if statusErr != nil {
		logger.Errorf(ctx, "failed to fetch author statuses, returning unfiltered results: %v", statusErr)
		return vods, nil
	}

	filtered := make([]domains.VOD, 0, len(vods))
	for _, v := range vods {
		if statuses[v.UserId.String()] != "disabled" {
			filtered = append(filtered, v)
		}
	}

	return filtered, nil
}
```

- [ ] **Step 6: Gate `CreateComment`**

Modify `backend/vod/services/vod_comment/create.go` — add the status check right after the existing VOD-exists check:

```go
	// verify VOD exists
	_, vodErr := s.vodRepo.GetById(ctx, vodId)
	if vodErr != nil {
		return nil, vodErr
	}

	statuses, statusErr := s.userGateway.GetUsersStatuses(ctx, []uuid.UUID{userId})
	if statusErr != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_FORBIDDEN, nil, nil, nil)
	}
	if statuses[userId.String()] == "disabled" {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_FORBIDDEN, nil, nil, nil)
	}
```

(`response.RES_ERR_FORBIDDEN` already exists in `backend/vod/response/error.go:80` — confirmed by reading the file directly. No new response code needed for this task.)

- [ ] **Step 7: Wire in `main.go`**

Modify `backend/vod/cmd/main.go` — `userGateway` is already constructed (`var userGateway = usergatewayhttp.NewUserGateway(registry)`) and already passed to `vodCommentService`; add it to the `vodService` constructor call too:

```go
	var vodService = vodService.NewVODService(vodRepo, transcodeJobRepo, minio, userGateway)
```

- [ ] **Step 8: Run tests to verify they pass**

Run: `cd backend/vod && go build ./... && go test ./... -race -v`
Expected: PASS across the whole module.

- [ ] **Step 9: Commit**

```bash
git add backend/vod/gateway/user/user.go backend/vod/gateway/user/http/http.go backend/vod/services/vod backend/vod/services/vod_comment/create.go backend/vod/cmd/main.go backend/vod/internal/testutil/fakes.go
git commit -m "$(cat <<'EOF'
feat(vod): filter disabled authors from popular VODs, gate comment creation

Refs: docs/superpowers/plans/2026-08-21-disabled-account-enforcement-extension.md
EOF
)"
```
