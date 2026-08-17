# Admin Dashboard — Foundation + Login Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up the new `backend/admin` Go service and `admin-web` Next.js app with a working, end-to-end admin login: an operator visits `admin-web`, logs in with an email/password seeded by hand in `admin_accounts`, lands on a protected page that proves the session round-trips through a real JWT the service itself verifies.

**Architecture:** A brand-new, fully isolated vertical slice — own Postgres database (`letslive_admin`), own Go microservice (`backend/admin`, same skeleton as `finance`/`vod`), own JWT secret (`ADMIN_JWT_SECRET`, never shared with the main app's `ACCESS_TOKEN_SECRET`), own Next.js frontend (`admin-web/`). Kong proxies `backend/admin`'s routes with **no** `jwt` plugin (precedent: `MinIO`/`Transcode` routes are also plugin-less) — the service does its own full JWT signature verification in a `RequireAdminAuth` middleware, which is what actually keeps this trust boundary separate from the main app's (Kong's single shared jwt-plugin consumer would otherwise let a main-app token through).

**Tech Stack:** Go 1.26 (net/http, pgx/v5, golang-jwt/v5, bcrypt, goose migrations via the existing `sharedutils.StartMigration`) for the backend; Next.js 16 / React 19 / Tailwind v4 / TypeScript for `admin-web`, matching `web/`'s versions exactly (no new dependency choices — same stack, just a smaller dependency set since there's no i18n/Sentry/Turnstile here).

**Spec:** [docs/superpowers/specs/2026-08-17-admin-dashboard-design.md](../specs/2026-08-17-admin-dashboard-design.md) — this plan implements the "Isolation", "Admin account creation", and "Session strategy" rows of that spec's Decisions table, plus the `admin_accounts` data model and the `POST /admin/login` endpoint. Stats and streams/VODs browsing are later plans (Plan 2, 3, 4) built on top of this foundation.

## Global Constraints

- No RBAC, no `role` column on the main `users` table — admin identity is a fully separate `admin_accounts` table in a fully separate database. (Spec: "Isolation")
- Admin accounts are seeded by hand via SQL — no signup flow, no admin-management UI. (Spec: "Admin account creation")
- Single JWT on login, ~24h expiry, no refresh-token table/rotation. (Spec: "Session strategy")
- No mock/fallback data on error — every error state shown to the user must come from a real failed request, never a fabricated placeholder. (User-global hard rule)
- No `any` in new TypeScript, no skipped type checks. (User-global hard rule)
- Never commit untested code; don't bypass hooks (`--no-verify`, `--no-gpg-sign`).

---

## Manual Prerequisite (outside this repo) — do this before Task 8

`backend/admin`'s config is pulled at boot from the private Spring Cloud Config git repo referenced by `backend/configserver`'s `CONFIG_SERVER_GIT_URI` (see `backend/configserver/src/main/resources/application.yml`) — the same mechanism every other service uses (`configServiceName = "finance_service"` etc. in each `cmd/main.go`). That repo is not part of this codebase, so this plan cannot create the file — whoever has access to it must add a config file for application name `admin_service` with (at minimum) this content, following the exact naming convention already used there for `finance_service`/`vod_service` (base file, or `-dev`/`-prod` profile variants — check the existing files in that repo for the pattern in use):

```yaml
service:
  name: admin
  hostname: admin.service.consul
  apiBindAddress: 0.0.0.0
  apiPort: 7784

database:
  migration-path: /usr/src/app/migrations
  host: main_db
  port: 5432
  name: letslive_admin
  params: ["sslmode=disable"]

tracer:
  endpoint: otel-collector:4317
  secure: false
  batchTimeout: 5000
```

(`host`/`port`/`params` for `database`, and `tracer.endpoint`, should match whatever values the existing `finance_service` config in that repo uses — copy those, only `name` differs.)

---

## File Structure

```
backend/admin/                          # new Go module, mirrors backend/finance's skeleton
  go.mod, go.sum
  cmd/main.go                           # bootstrap: config, DB, migrations, discovery, tracer, server
  config/config.go                      # Service/Database/Tracer config + PostProcess (env secrets)
  domains/admin_account.go              # AdminAccount struct + AdminAccountRepository interface
  migrations/0001_create_admin_accounts_table.sql
  repositories/repositories.go          # constructor wiring
  repositories/admin_account/admin_account.go
  repositories/admin_account/get_by_email.go
  repositories/admin_account/get_by_id.go
  response/response.go                  # type aliases -> shared/response (same pattern as finance)
  response/error.go                     # error templates, 70000-range codes for this domain
  response/success.go
  services/auth.go                      # AuthService{repo}: Login, GenerateAccessToken, GetByID
  services/auth_test.go
  types/jwt_claims.go                   # AdminClaims{AdminId, RegisteredClaims}
  utils/validator.go                    # shared go-playground validator instance
  dto/login_request.go                  # LoginRequestDTO with validate tags
  handlers/basehandler/basehandler.go   # copied verbatim from finance (WriteResponse helper)
  handlers/middleware/require_admin_auth.go
  handlers/middleware/require_admin_auth_test.go
  handlers/auth/auth.go                 # AuthHandler struct + constructor
  handlers/auth/login_public.go         # POST /v1/admin/login
  handlers/auth/get_me_private.go       # GET /v1/admin/me
  handlers/general/general.go           # GeneralHandler (health check, 404) — copied from finance
  handlers/general/route_not_found.go
  handlers/general/route_service_health.go
  api/http.go                           # APIServer: route table, ListenAndServe, Shutdown
  Dockerfile

admin-web/                              # new top-level Next.js app, sibling to web/ and backend/
  package.json, tsconfig.json, next.config.js, postcss.config.mjs, eslint.config.mjs
  Dockerfile
  app/
    layout.tsx                          # root layout: QueryProvider + globals.css
    globals.css                         # Tailwind v4 import, minimal palette
    login/page.tsx                      # email/password form -> POST /admin/login
    page.tsx                            # protected home: RequireAdminAuth -> fetch GET /admin/me
  lib/
    global.ts                           # ADMIN_API_URL from env (mirrors web/global.ts)
    api/
      admin-fetch.ts                    # fetchClient equivalent, no refresh-token logic
      api-error.ts                      # ApiError + unwrapResponse (copied from web/lib/api/api-error.ts)
      auth.ts                           # Login(), GetMe()
  types/
    fetch-response.ts                   # ApiResponse<T> (copied shape from web/types/fetch-response.ts)
    admin.ts                            # AdminMe type
  components/
    require-admin-auth.tsx              # client wrapper: redirects to /login if GetMe() fails
    query-provider.tsx                  # minimal react-query provider, no i18n toast

docker-entrypoint-initdb.d/01-create-dbs.sql   # add letslive_admin
example.env                                     # add ADMIN_DB_USER/PASSWORD, ADMIN_JWT_SECRET, NEXT_PUBLIC_ADMIN_API_URL
docker-compose-dev.yaml                         # add admin + admin-web service blocks
docker-compose.yaml                             # add admin + admin-web service blocks
configs/kong.yml                                # add Admin_Service block, /admin/login + /admin/me routes
.github/workflows/build-and-publish-images.yml  # add admin to backend matrix, new admin-web job
.github/workflows/deploy.yml                    # add ADMIN_DB_USER/PASSWORD, ADMIN_JWT_SECRET to .env
```

---

### Task 1: Backend scaffold — module, config, response package, migration, domain, repository

**Files:**
- Create: `backend/admin/go.mod`, `backend/admin/go.sum`
- Create: `backend/admin/config/config.go`
- Create: `backend/admin/response/response.go`, `backend/admin/response/error.go`, `backend/admin/response/success.go`
- Create: `backend/admin/migrations/0001_create_admin_accounts_table.sql`
- Create: `backend/admin/domains/admin_account.go`
- Create: `backend/admin/repositories/repositories.go`
- Create: `backend/admin/repositories/admin_account/admin_account.go`
- Create: `backend/admin/repositories/admin_account/get_by_email.go`
- Create: `backend/admin/repositories/admin_account/get_by_id.go`

**Interfaces:**
- Produces: `domains.AdminAccount{Id uuid.UUID, Email string, PasswordHash string, CreatedAt time.Time}`, `domains.AdminAccountRepository` interface with `GetByEmail(ctx, email string) (*AdminAccount, *response.Response[any])` and `GetByID(ctx, id uuid.UUID) (*AdminAccount, *response.Response[any])`. `repositories.NewAdminAccountRepository(conn *pgxpool.Pool) domains.AdminAccountRepository`. `response.RES_ERR_INVALID_CREDENTIALS`, `response.RES_ERR_ADMIN_NOT_FOUND`, `response.RES_SUCC_OK` templates. `config.Config`, `config.PostProcess`.

- [ ] **Step 1: Initialize the Go module**

```bash
cd backend/admin
go mod init sen1or/letslive/admin
```

- [ ] **Step 2: Write `config/config.go`**

```go
package config

import (
	"fmt"
	neturl "net/url"
	"os"
	"strings"
)

type Service struct {
	Name           string `yaml:"name"`
	Hostname       string `yaml:"hostname"`
	APIBindAddress string `yaml:"apiBindAddress"`
	APIPort        int    `yaml:"apiPort"`
}

type Database struct {
	MigrationPath    string   `yaml:"migration-path"`
	Host             string   `yaml:"host"`
	Port             int      `yaml:"port"`
	Name             string   `yaml:"name"`
	Params           []string `yaml:"params"`
	ConnectionString string
}

type Tracer struct {
	Endpoint     string `yaml:"endpoint"`
	Secure       bool   `yaml:"secure"`
	BatchTimeout int    `yaml:"batchTimeout"` // in milli-second
}

type Config struct {
	Service  `yaml:"service"`
	Database `yaml:"database"`
	Tracer   `yaml:"tracer"`
}

// TracerConfig interface implementation
func (c Config) GetServiceName() string     { return c.Service.Name }
func (c Config) GetTracerEndpoint() string  { return c.Tracer.Endpoint }
func (c Config) GetTracerBatchTimeout() int { return c.Tracer.BatchTimeout }
func (c Config) IsSecure() bool             { return c.Tracer.Secure }

// PostProcess builds the database connection string and validates required secrets.
func PostProcess(config *Config) error {
	dbUser := os.Getenv("ADMIN_DB_USER")
	dbPassword := os.Getenv("ADMIN_DB_PASSWORD")

	dbURL := &neturl.URL{
		Scheme: "postgres",
		User:   neturl.UserPassword(dbUser, dbPassword),
		Host:   fmt.Sprintf("%s:%d", config.Database.Host, config.Database.Port),
		Path:   "/" + config.Database.Name,
	}
	if len(config.Database.Params) > 0 {
		dbURL.RawQuery = strings.Join(config.Database.Params, "&")
	}
	config.Database.ConnectionString = dbURL.String()

	if os.Getenv("ADMIN_JWT_SECRET") == "" {
		return fmt.Errorf("ADMIN_JWT_SECRET must be set")
	}

	return nil
}
```

- [ ] **Step 3: Write the response package (`response/response.go`, `response/error.go`, `response/success.go`)**

`response/response.go`:

```go
package response

import (
	sharedresponse "sen1or/letslive/shared/response"
)

type Meta = sharedresponse.Meta
type ErrorDetail = sharedresponse.ErrorDetail
type ErrorDetails = sharedresponse.ErrorDetails
type Response[T any] = sharedresponse.Response[T]
type ResponseTemplate = sharedresponse.ResponseTemplate

func NewResponseFromTemplate[T any](tpl ResponseTemplate, data *T, meta *Meta, errorDetails *ErrorDetails) *Response[T] {
	return sharedresponse.NewResponseFromTemplate[T](tpl, data, meta, errorDetails)
}

func NewResponseWithValidationErrors[T any](data *T, meta *Meta, validateError error) *Response[T] {
	return sharedresponse.NewResponseWithValidationErrors[T](RES_ERR_INVALID_INPUT, data, meta, validateError)
}

func NewResponse[T any](success bool, statusCode int, code int, key string, message string, data *T, meta *Meta, errorDetails *ErrorDetails) *Response[T] {
	return sharedresponse.NewResponse[T](success, statusCode, code, key, message, data, meta, errorDetails)
}
```

`response/error.go`:

```go
package response

import "net/http"

const (
	// Generic (shared range 20000-20017, same numbers every service uses)
	RES_ERR_INVALID_INPUT_CODE   = 20000
	RES_ERR_INVALID_PAYLOAD_CODE = 20001
	RES_ERR_UNAUTHORIZED_CODE    = 20005
	RES_ERR_FORBIDDEN_CODE       = 20008
	RES_ERR_ROUTE_NOT_FOUND_CODE = 20012
	RES_ERR_DATABASE_QUERY_CODE  = 20015
	RES_ERR_DATABASE_ISSUE_CODE  = 20016
	RES_ERR_INTERNAL_SERVER_CODE = 20017

	// Admin domain (70000-70001)
	RES_ERR_INVALID_CREDENTIALS_CODE = 70000
	RES_ERR_ADMIN_NOT_FOUND_CODE     = 70001
)

const (
	RES_ERR_INVALID_INPUT_KEY   = "res_err_invalid_input"
	RES_ERR_INVALID_PAYLOAD_KEY = "res_err_invalid_payload"
	RES_ERR_UNAUTHORIZED_KEY    = "res_err_unauthorized"
	RES_ERR_FORBIDDEN_KEY       = "res_err_forbidden"
	RES_ERR_ROUTE_NOT_FOUND_KEY = "res_err_route_not_found"
	RES_ERR_DATABASE_QUERY_KEY  = "res_err_database_query"
	RES_ERR_DATABASE_ISSUE_KEY  = "res_err_database_issue"
	RES_ERR_INTERNAL_SERVER_KEY = "res_err_internal_server"

	RES_ERR_INVALID_CREDENTIALS_KEY = "res_err_invalid_credentials"
	RES_ERR_ADMIN_NOT_FOUND_KEY     = "res_err_admin_not_found"
)

var (
	RES_ERR_INVALID_INPUT = ResponseTemplate{
		Success: false, StatusCode: http.StatusBadRequest, Code: RES_ERR_INVALID_INPUT_CODE,
		Key: RES_ERR_INVALID_INPUT_KEY, Message: "Input invalid.",
	}
	RES_ERR_INVALID_PAYLOAD = ResponseTemplate{
		Success: false, StatusCode: http.StatusBadRequest, Code: RES_ERR_INVALID_PAYLOAD_CODE,
		Key: RES_ERR_INVALID_PAYLOAD_KEY, Message: "Payload invalid.",
	}
	RES_ERR_UNAUTHORIZED = ResponseTemplate{
		Success: false, StatusCode: http.StatusUnauthorized, Code: RES_ERR_UNAUTHORIZED_CODE,
		Key: RES_ERR_UNAUTHORIZED_KEY, Message: "Unauthorized.",
	}
	RES_ERR_FORBIDDEN = ResponseTemplate{
		Success: false, StatusCode: http.StatusForbidden, Code: RES_ERR_FORBIDDEN_CODE,
		Key: RES_ERR_FORBIDDEN_KEY, Message: "Forbidden.",
	}
	RES_ERR_ROUTE_NOT_FOUND = ResponseTemplate{
		Success: false, StatusCode: http.StatusNotFound, Code: RES_ERR_ROUTE_NOT_FOUND_CODE,
		Key: RES_ERR_ROUTE_NOT_FOUND_KEY, Message: "Requested endpoint not found.",
	}
	RES_ERR_DATABASE_QUERY = ResponseTemplate{
		Success: false, StatusCode: http.StatusInternalServerError, Code: RES_ERR_DATABASE_QUERY_CODE,
		Key: RES_ERR_DATABASE_QUERY_KEY, Message: "Error querying database, please try again.",
	}
	RES_ERR_DATABASE_ISSUE = ResponseTemplate{
		Success: false, StatusCode: http.StatusInternalServerError, Code: RES_ERR_DATABASE_ISSUE_CODE,
		Key: RES_ERR_DATABASE_ISSUE_KEY, Message: "Database issue, please try again.",
	}
	RES_ERR_INTERNAL_SERVER = ResponseTemplate{
		Success: false, StatusCode: http.StatusInternalServerError, Code: RES_ERR_INTERNAL_SERVER_CODE,
		Key: RES_ERR_INTERNAL_SERVER_KEY, Message: "Something went wrong.",
	}

	// Deliberately the same message/code for "unknown email" and "wrong password" — no user enumeration.
	RES_ERR_INVALID_CREDENTIALS = ResponseTemplate{
		Success: false, StatusCode: http.StatusUnauthorized, Code: RES_ERR_INVALID_CREDENTIALS_CODE,
		Key: RES_ERR_INVALID_CREDENTIALS_KEY, Message: "Invalid email or password.",
	}
	RES_ERR_ADMIN_NOT_FOUND = ResponseTemplate{
		Success: false, StatusCode: http.StatusNotFound, Code: RES_ERR_ADMIN_NOT_FOUND_CODE,
		Key: RES_ERR_ADMIN_NOT_FOUND_KEY, Message: "Admin account not found.",
	}
)
```

`response/success.go`:

```go
package response

import "net/http"

const (
	RES_SUCC_OK_CODE = 100000
	RES_SUCC_OK_KEY  = "res_succ_ok"
)

var RES_SUCC_OK = ResponseTemplate{
	Success: true, StatusCode: http.StatusOK, Code: RES_SUCC_OK_CODE, Key: RES_SUCC_OK_KEY,
}
```

- [ ] **Step 4: Write the migration**

`migrations/0001_create_admin_accounts_table.sql`:

```sql
-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "admin_accounts" (
  "id" uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
  "email" text NOT NULL,
  "password_hash" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT current_timestamp,
  CONSTRAINT "uni_admin_accounts_email" UNIQUE ("email")
);

CREATE INDEX IF NOT EXISTS "idx_admin_accounts_email" ON "admin_accounts"("email");

-- +goose Down
DROP INDEX IF EXISTS "idx_admin_accounts_email";
DROP TABLE IF EXISTS "admin_accounts";
DROP EXTENSION IF EXISTS "uuid-ossp";
```

- [ ] **Step 5: Write the domain**

`domains/admin_account.go`:

```go
package domains

import (
	"context"
	"time"

	"sen1or/letslive/admin/response"

	"github.com/gofrs/uuid/v5"
)

type AdminAccount struct {
	Id           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

type AdminAccountRepository interface {
	GetByEmail(ctx context.Context, email string) (*AdminAccount, *response.Response[any])
	GetByID(ctx context.Context, id uuid.UUID) (*AdminAccount, *response.Response[any])
}
```

- [ ] **Step 6: Write the repository**

`repositories/admin_account/admin_account.go`:

```go
package adminaccount

import (
	"sen1or/letslive/admin/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresAdminAccountRepo struct {
	dbConn *pgxpool.Pool
}

func NewAdminAccountRepository(conn *pgxpool.Pool) domains.AdminAccountRepository {
	return &postgresAdminAccountRepo{dbConn: conn}
}
```

`repositories/admin_account/get_by_email.go`:

```go
package adminaccount

import (
	"context"
	"errors"

	"sen1or/letslive/admin/domains"
	"sen1or/letslive/admin/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5"
)

func (r *postgresAdminAccountRepo) GetByEmail(ctx context.Context, email string) (*domains.AdminAccount, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT *
		FROM admin_accounts
		WHERE email = $1
	`, email)
	if err != nil {
		logger.Errorf(ctx, "failed to get admin account by email: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY, nil, nil,
			&response.ErrorDetails{response.ErrorDetail{"email": email}},
		)
	}
	defer rows.Close()

	account, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.AdminAccount])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_ADMIN_NOT_FOUND, nil, nil,
				&response.ErrorDetails{response.ErrorDetail{"email": email}},
			)
		}
		logger.Errorf(ctx, "failed to collect row: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE, nil, nil,
			&response.ErrorDetails{response.ErrorDetail{"email": email}},
		)
	}

	return &account, nil
}
```

`repositories/admin_account/get_by_id.go`:

```go
package adminaccount

import (
	"context"
	"errors"

	"sen1or/letslive/admin/domains"
	"sen1or/letslive/admin/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresAdminAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*domains.AdminAccount, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT *
		FROM admin_accounts
		WHERE id = $1
	`, id)
	if err != nil {
		logger.Errorf(ctx, "failed to get admin account by id: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY, nil, nil,
			&response.ErrorDetails{response.ErrorDetail{"id": id}},
		)
	}
	defer rows.Close()

	account, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.AdminAccount])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_ADMIN_NOT_FOUND, nil, nil,
				&response.ErrorDetails{response.ErrorDetail{"id": id}},
			)
		}
		logger.Errorf(ctx, "failed to collect row: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE, nil, nil,
			&response.ErrorDetails{response.ErrorDetail{"id": id}},
		)
	}

	return &account, nil
}
```

`repositories/repositories.go`:

```go
package repositories

import (
	adminaccount "sen1or/letslive/admin/repositories/admin_account"
	"sen1or/letslive/admin/domains"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewAdminAccountRepository(conn *pgxpool.Pool) domains.AdminAccountRepository {
	return adminaccount.NewAdminAccountRepository(conn)
}
```

- [ ] **Step 7: Add dependencies and verify the module builds**

```bash
cd backend/admin
go get github.com/go-playground/validator/v10@v10.30.1
go get github.com/gofrs/uuid/v5@v5.4.0
go get github.com/golang-jwt/jwt/v5@v5.3.1
go get github.com/jackc/pgx/v5@v5.8.0
go get golang.org/x/crypto@v0.48.0
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@v0.70.0
go get go.uber.org/zap@v1.28.0
```

Then edit `go.mod` to add the local `shared` module (same as every other service):

```
require sen1or/letslive/shared v0.0.0

replace sen1or/letslive/shared v0.0.0 => ../shared
```

```bash
go mod tidy
go build ./...
```

Expected: builds cleanly (no `cmd/`, `handlers/`, `services/`, `api/` yet, so `go build ./...` only compiles `config`, `response`, `domains`, `repositories` — that's expected and fine at this point).

- [ ] **Step 8: Commit**

```bash
git add backend/admin/go.mod backend/admin/go.sum backend/admin/config backend/admin/response backend/admin/migrations backend/admin/domains backend/admin/repositories
git commit -m "feat(admin): scaffold admin service — config, response package, admin_accounts migration, repository

Refs: docs/superpowers/specs/2026-08-17-admin-dashboard-design.md"
```

---

### Task 2: AuthService — Login, GenerateAccessToken, GetByID

**Files:**
- Create: `backend/admin/types/jwt_claims.go`
- Create: `backend/admin/services/auth.go`
- Test: `backend/admin/services/auth_test.go`

**Interfaces:**
- Consumes: `domains.AdminAccount`, `domains.AdminAccountRepository` (Task 1), `response.RES_ERR_INVALID_CREDENTIALS`, `response.RES_ERR_INTERNAL_SERVER` (Task 1).
- Produces: `types.AdminClaims{AdminId string, jwt.RegisteredClaims}`. `services.AuthService{repo}`, `services.NewAuthService(repo domains.AdminAccountRepository) *AuthService`, methods `Login(ctx, email, password string) (*domains.AdminAccount, *response.Response[any])`, `GenerateAccessToken(adminId string) (token string, expiresAt time.Time, errResp *response.Response[any])`, `GetByID(ctx, adminIdStr string) (*domains.AdminAccount, *response.Response[any])`. Consumed by Task 4's handlers and Task 3's middleware (`types.AdminClaims`).

- [ ] **Step 1: Write `types/jwt_claims.go`**

```go
package types

import "github.com/golang-jwt/jwt/v5"

type AdminClaims struct {
	AdminId string `json:"adminId"`
	jwt.RegisteredClaims
}
```

- [ ] **Step 2: Write the failing test for `AuthService.Login`**

`services/auth_test.go`:

```go
package services

import (
	"context"
	"testing"
	"time"

	"sen1or/letslive/admin/domains"
	"sen1or/letslive/admin/response"
	"sen1or/letslive/admin/types"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type fakeAdminAccountRepo struct {
	byEmail map[string]*domains.AdminAccount
	byID    map[uuid.UUID]*domains.AdminAccount
}

func newFakeRepo(accounts ...*domains.AdminAccount) *fakeAdminAccountRepo {
	r := &fakeAdminAccountRepo{
		byEmail: map[string]*domains.AdminAccount{},
		byID:    map[uuid.UUID]*domains.AdminAccount{},
	}
	for _, a := range accounts {
		r.byEmail[a.Email] = a
		r.byID[a.Id] = a
	}
	return r
}

func (r *fakeAdminAccountRepo) GetByEmail(ctx context.Context, email string) (*domains.AdminAccount, *response.Response[any]) {
	if a, ok := r.byEmail[email]; ok {
		return a, nil
	}
	return nil, response.NewResponseFromTemplate[any](response.RES_ERR_ADMIN_NOT_FOUND, nil, nil, nil)
}

func (r *fakeAdminAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*domains.AdminAccount, *response.Response[any]) {
	if a, ok := r.byID[id]; ok {
		return a, nil
	}
	return nil, response.NewResponseFromTemplate[any](response.RES_ERR_ADMIN_NOT_FOUND, nil, nil, nil)
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %s", err)
	}
	return string(hash)
}

func TestAuthService_Login(t *testing.T) {
	knownAdmin := &domains.AdminAccount{
		Id:           uuid.Must(uuid.NewV4()),
		Email:        "admin@letslive.app",
		PasswordHash: mustHash(t, "correct-password"),
		CreatedAt:    time.Now(),
	}

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{name: "correct credentials", email: "admin@letslive.app", password: "correct-password", wantErr: false},
		{name: "wrong password", email: "admin@letslive.app", password: "wrong-password", wantErr: true},
		{name: "unknown email", email: "nobody@letslive.app", password: "correct-password", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAuthService(newFakeRepo(knownAdmin))
			account, errResp := svc.Login(context.Background(), tt.email, tt.password)

			if tt.wantErr {
				if errResp == nil {
					t.Fatalf("expected error, got account %+v", account)
				}
				if errResp.Code != response.RES_ERR_INVALID_CREDENTIALS_CODE {
					t.Fatalf("expected RES_ERR_INVALID_CREDENTIALS_CODE, got %d", errResp.Code)
				}
				return
			}

			if errResp != nil {
				t.Fatalf("unexpected error: %+v", errResp)
			}
			if account.Id != knownAdmin.Id {
				t.Fatalf("got account id %s, want %s", account.Id, knownAdmin.Id)
			}
		})
	}
}

func TestAuthService_GenerateAccessToken(t *testing.T) {
	t.Setenv("ADMIN_JWT_SECRET", "test-secret")

	svc := NewAuthService(newFakeRepo())
	adminId := uuid.Must(uuid.NewV4()).String()

	token, expiresAt, errResp := svc.GenerateAccessToken(adminId)
	if errResp != nil {
		t.Fatalf("unexpected error: %+v", errResp)
	}
	if time.Until(expiresAt) <= 0 || time.Until(expiresAt) > 25*time.Hour {
		t.Fatalf("expiresAt %v not within expected ~24h window", expiresAt)
	}

	claims := &types.AdminClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte("test-secret"), nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("generated token does not parse/verify: %s", err)
	}
	if claims.AdminId != adminId {
		t.Fatalf("got adminId %q, want %q", claims.AdminId, adminId)
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

```bash
cd backend/admin
go test ./services/... -v
```

Expected: FAIL — `NewAuthService`/`Login`/`GenerateAccessToken` undefined.

- [ ] **Step 4: Write `services/auth.go`**

```go
package services

import (
	"context"
	"os"
	"time"

	"sen1or/letslive/admin/domains"
	"sen1or/letslive/admin/response"
	"sen1or/letslive/admin/types"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const AccessTokenMaxAge = 24 * time.Hour

type AuthService struct {
	repo domains.AdminAccountRepository
}

func NewAuthService(repo domains.AdminAccountRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Login returns the same RES_ERR_INVALID_CREDENTIALS for an unknown email and a wrong
// password on purpose — distinguishing the two would let a caller enumerate admin emails.
func (s *AuthService) Login(ctx context.Context, email, password string) (*domains.AdminAccount, *response.Response[any]) {
	account, errResp := s.repo.GetByEmail(ctx, email)
	if errResp != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_CREDENTIALS, nil, nil, nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_CREDENTIALS, nil, nil, nil)
	}

	return account, nil
}

func (s *AuthService) GenerateAccessToken(adminId string) (string, time.Time, *response.Response[any]) {
	expiresAt := time.Now().Add(AccessTokenMaxAge)
	claims := types.AdminClaims{
		AdminId: adminId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := unsignedToken.SignedString([]byte(os.Getenv("ADMIN_JWT_SECRET")))
	if err != nil {
		return "", time.Time{}, response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}

	return signedToken, expiresAt, nil
}

func (s *AuthService) GetByID(ctx context.Context, adminIdStr string) (*domains.AdminAccount, *response.Response[any]) {
	adminId, err := uuid.FromString(adminIdStr)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_UNAUTHORIZED, nil, nil, nil)
	}
	return s.repo.GetByID(ctx, adminId)
}
```

- [ ] **Step 5: Run the test to verify it passes**

```bash
go test ./services/... -v
```

Expected: PASS — all subtests of `TestAuthService_Login` and `TestAuthService_GenerateAccessToken`.

- [ ] **Step 6: Commit**

```bash
git add backend/admin/types backend/admin/services
git commit -m "feat(admin): add AuthService (login, JWT issuance) with unit tests

Refs: docs/superpowers/specs/2026-08-17-admin-dashboard-design.md"
```

---

### Task 3: RequireAdminAuth middleware

**Files:**
- Create: `backend/admin/handlers/middleware/require_admin_auth.go`
- Test: `backend/admin/handlers/middleware/require_admin_auth_test.go`

**Interfaces:**
- Consumes: `types.AdminClaims` (Task 2), `response.RES_ERR_UNAUTHORIZED` (Task 1).
- Produces: `middleware.RequireAdminAuth(next http.HandlerFunc) http.HandlerFunc`, `middleware.AdminIdContextKey` (context key used to read the verified admin id in handlers). Consumed by Task 4's `GetMePrivateHandler` and Task 4's route wiring.

- [ ] **Step 1: Write the failing test**

`handlers/middleware/require_admin_auth_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sen1or/letslive/admin/types"

	"github.com/golang-jwt/jwt/v5"
)

func signToken(t *testing.T, secret string, adminId string, expiresAt time.Time) string {
	t.Helper()
	claims := types.AdminClaims{
		AdminId: adminId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %s", err)
	}
	return token
}

func TestRequireAdminAuth(t *testing.T) {
	t.Setenv("ADMIN_JWT_SECRET", "correct-secret")

	tests := []struct {
		name       string
		cookie     *http.Cookie
		wantStatus int
		wantNext   bool
	}{
		{
			name:       "valid token",
			cookie:     &http.Cookie{Name: "ADMIN_ACCESS_TOKEN", Value: signToken(t, "correct-secret", "admin-1", time.Now().Add(time.Hour))},
			wantStatus: http.StatusOK,
			wantNext:   true,
		},
		{
			name:       "expired token",
			cookie:     &http.Cookie{Name: "ADMIN_ACCESS_TOKEN", Value: signToken(t, "correct-secret", "admin-1", time.Now().Add(-time.Hour))},
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
		{
			name:       "wrong secret (e.g. a main-app token)",
			cookie:     &http.Cookie{Name: "ADMIN_ACCESS_TOKEN", Value: signToken(t, "some-other-secret", "admin-1", time.Now().Add(time.Hour))},
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
		{
			name:       "malformed token",
			cookie:     &http.Cookie{Name: "ADMIN_ACCESS_TOKEN", Value: "not-a-jwt"},
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
		{
			name:       "missing cookie",
			cookie:     nil,
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			}

			req := httptest.NewRequest(http.MethodGet, "/v1/admin/me", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rec := httptest.NewRecorder()

			RequireAdminAuth(next)(rec, req)

			if nextCalled != tt.wantNext {
				t.Fatalf("next called = %v, want %v", nextCalled, tt.wantNext)
			}
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
cd backend/admin
go test ./handlers/middleware/... -v
```

Expected: FAIL — `RequireAdminAuth` undefined.

- [ ] **Step 3: Write `handlers/middleware/require_admin_auth.go`**

```go
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"sen1or/letslive/admin/response"
	"sen1or/letslive/admin/types"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const AdminIdContextKey contextKey = "adminId"

// RequireAdminAuth does full JWT signature verification against ADMIN_JWT_SECRET —
// unlike the main app's services, which trust Kong's jwt plugin and only decode
// claims unverified. Kong has no jwt plugin on /admin/* routes (see plan/spec), so
// this middleware IS the auth boundary.
func RequireAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("ADMIN_ACCESS_TOKEN")
		if err != nil || len(cookie.Value) == 0 {
			writeUnauthorized(w, r)
			return
		}

		claims := types.AdminClaims{}
		token, err := jwt.ParseWithClaims(cookie.Value, &claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("ADMIN_JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			writeUnauthorized(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), AdminIdContextKey, claims.AdminId)
		next(w, r.WithContext(ctx))
	}
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request) {
	res := response.NewResponseFromTemplate[any](response.RES_ERR_UNAUTHORIZED, nil, nil, nil)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(res.StatusCode)
	_ = json.NewEncoder(w).Encode(res)
}
```

- [ ] **Step 4: Run the test to verify it passes**

```bash
go test ./handlers/middleware/... -v
```

Expected: PASS — all 5 subtests.

- [ ] **Step 5: Commit**

```bash
git add backend/admin/handlers/middleware
git commit -m "feat(admin): add RequireAdminAuth middleware with full JWT verification

Refs: docs/superpowers/specs/2026-08-17-admin-dashboard-design.md"
```

---

### Task 4: Handlers, DTO, validator, general handler, API wiring, main.go, Dockerfile

**Files:**
- Create: `backend/admin/utils/validator.go`
- Create: `backend/admin/dto/login_request.go`
- Create: `backend/admin/handlers/basehandler/basehandler.go`
- Create: `backend/admin/handlers/general/general.go`, `route_not_found.go`, `route_service_health.go`
- Create: `backend/admin/handlers/auth/auth.go`, `login_public.go`, `get_me_private.go`
- Create: `backend/admin/api/http.go`
- Create: `backend/admin/cmd/main.go`
- Create: `backend/admin/Dockerfile`

**Interfaces:**
- Consumes: `services.AuthService` (Task 2), `middleware.RequireAdminAuth`, `middleware.AdminIdContextKey` (Task 3), `response.*` (Task 1), `repositories.NewAdminAccountRepository` (Task 1).
- Produces: running `backend/admin` service exposing `POST /v1/admin/login` and `GET /v1/admin/me` — consumed by Task 5 (infra wiring) for the end-to-end curl test, and by `admin-web` (Tasks 6-8) as its API.

- [ ] **Step 1: Write the validator and DTO**

`utils/validator.go`:

```go
package utils

import "github.com/go-playground/validator/v10"

var Validator = validator.New(validator.WithRequiredStructEnabled())
```

`dto/login_request.go`:

```go
package dto

type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email,lte=320"`
	Password string `json:"password" validate:"required"`
}
```

- [ ] **Step 2: Write the base handler and general handler (health check, 404)**

`handlers/basehandler/basehandler.go` — copied verbatim from `backend/finance/handlers/basehandler/basehandler.go`, package path adjusted:

```go
package basehandler

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
)

type BaseHandler struct{}

func (b *BaseHandler) WriteResponse(w http.ResponseWriter, ctx context.Context, res any) {
	resValue := reflect.ValueOf(res).Elem()

	if ctxRequestId, ok := ctx.Value("requestId").(string); ok && len(ctxRequestId) > 0 {
		requestIdField := resValue.FieldByName("RequestId")
		if requestIdField.IsValid() && requestIdField.CanSet() {
			requestIdField.SetString(ctxRequestId)
		}
	}

	var statusCode int
	statusCodeField := resValue.FieldByName("StatusCode")
	if statusCodeField.IsValid() && statusCodeField.CanInt() {
		statusCode = int(statusCodeField.Int())
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if statusCode > 0 {
		w.WriteHeader(statusCode)
	}
	json.NewEncoder(w).Encode(res)
}
```

`handlers/general/general.go`:

```go
package general

import (
	"sen1or/letslive/admin/handlers/basehandler"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GeneralHandler struct {
	basehandler.BaseHandler
	DB *pgxpool.Pool
}

func NewGeneralHandler(db *pgxpool.Pool) *GeneralHandler {
	return &GeneralHandler{DB: db}
}
```

`handlers/general/route_not_found.go`:

```go
package general

import (
	"net/http"
	"sen1or/letslive/admin/response"
)

func (h *GeneralHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.WriteResponse(w, r.Context(), response.NewResponseFromTemplate[any](response.RES_ERR_ROUTE_NOT_FOUND, nil, nil, nil))
}
```

`handlers/general/route_service_health.go`:

```go
package general

import (
	"encoding/json"
	"net/http"
)

func (h *GeneralHandler) RouteServiceHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "ok"
	dbStatus := "ok"

	if err := h.DB.Ping(r.Context()); err != nil {
		status = "degraded"
		dbStatus = "unavailable"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status": status,
		"checks": map[string]string{"database": dbStatus},
	})
}
```

- [ ] **Step 3: Write the auth handlers**

`handlers/auth/auth.go`:

```go
package auth

import (
	"sen1or/letslive/admin/handlers/basehandler"
	"sen1or/letslive/admin/services"
)

type AuthHandler struct {
	basehandler.BaseHandler
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}
```

`handlers/auth/login_public.go`:

```go
package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"sen1or/letslive/admin/dto"
	"sen1or/letslive/admin/response"
	"sen1or/letslive/admin/utils"
)

func (h *AuthHandler) LoginPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var req dto.LoginRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}
	if err := utils.Validator.Struct(req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	account, loginErr := h.authService.Login(ctx, req.Email, req.Password)
	if loginErr != nil {
		h.WriteResponse(w, ctx, loginErr)
		return
	}

	accessToken, expiresAt, tokenErr := h.authService.GenerateAccessToken(account.Id.String())
	if tokenErr != nil {
		h.WriteResponse(w, ctx, tokenErr)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "ADMIN_ACCESS_TOKEN",
		Value:    accessToken,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}
```

`handlers/auth/get_me_private.go`:

```go
package auth

import (
	"context"
	"net/http"

	"sen1or/letslive/admin/handlers/middleware"
	"sen1or/letslive/admin/response"
)

type MeResponseDTO struct {
	Email string `json:"email"`
}

func (h *AuthHandler) GetMePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	adminId, ok := r.Context().Value(middleware.AdminIdContextKey).(string)
	if !ok {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_UNAUTHORIZED, nil, nil, nil))
		return
	}

	account, errResp := h.authService.GetByID(ctx, adminId)
	if errResp != nil {
		h.WriteResponse(w, ctx, errResp)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &MeResponseDTO{Email: account.Email}, nil, nil))
}
```

- [ ] **Step 4: Write `api/http.go`**

```go
package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"sen1or/letslive/admin/config"
	"sen1or/letslive/admin/handlers/auth"
	"sen1or/letslive/admin/handlers/general"
	"sen1or/letslive/admin/handlers/middleware"
	"sen1or/letslive/shared/middlewares"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type APIServer struct {
	httpServer *http.Server
	logger     *zap.SugaredLogger
	config     *config.Config

	generalHandler *general.GeneralHandler
	authHandler    *auth.AuthHandler
}

func NewAPIServer(authHandler *auth.AuthHandler, cfg *config.Config, db *pgxpool.Pool) *APIServer {
	return &APIServer{
		logger:         logger.Logger,
		config:         cfg,
		generalHandler: general.NewGeneralHandler(db),
		authHandler:    authHandler,
	}
}

func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()

	wrap := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		sm.Handle(pattern, http.HandlerFunc(handlerFunc))
	}

	// Public routes
	wrap("POST /v1/admin/login", a.authHandler.LoginPublicHandler)

	// Private routes (verified in-service — see handlers/middleware/require_admin_auth.go)
	wrap("GET /v1/admin/me", middleware.RequireAdminAuth(a.authHandler.GetMePrivateHandler))

	// Health check
	wrap("GET /v1/health", a.generalHandler.RouteServiceHealth)
	wrap("GET /", a.generalHandler.RouteNotFoundHandler)

	finalHandler := otelhttp.NewHandler(sm, "/", otelhttp.WithFilter(func(r *http.Request) bool {
		return r.URL.Path != "/v1/health"
	}))
	finalHandler = middlewares.LoggingMiddleware(finalHandler)
	finalHandler = middlewares.RequestIDMiddleware(finalHandler)

	return finalHandler
}

func (a *APIServer) ListenAndServe(ctx context.Context, useTLS bool) error {
	addr := fmt.Sprintf("%s:%d", a.config.Service.APIBindAddress, a.config.Service.APIPort)

	a.httpServer = &http.Server{
		Addr:         addr,
		Handler:      a.getHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	var err error
	if useTLS {
		err = fmt.Errorf("TLS not implemented")
	} else {
		err = a.httpServer.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		logger.Errorf(ctx, "server listener error: %v", err)
		return err
	}
	return nil
}

func (a *APIServer) Shutdown(ctx context.Context) error {
	if a.httpServer == nil {
		logger.Warnf(ctx, "server instance not found, cannot shutdown.")
		return nil
	}
	if err := a.httpServer.Shutdown(ctx); err != nil {
		logger.Errorf(ctx, "server shutdown failed: %v", err)
		return err
	}
	return nil
}
```

- [ ] **Step 5: Write `cmd/main.go`**

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sen1or/letslive/admin/api"
	authhandler "sen1or/letslive/admin/handlers/auth"
	cfg "sen1or/letslive/admin/config"
	"sen1or/letslive/admin/repositories"
	"sen1or/letslive/admin/services"

	sharedconfig "sen1or/letslive/shared/config"
	"sen1or/letslive/shared/pkg/discovery"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/tracer"
	sharedutils "sen1or/letslive/shared/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	configServiceName = "admin_service"
	configProfile     = os.Getenv("CONFIG_SERVER_PROFILE")

	shutdownTimeout = 15 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Init(logger.LogLevel(logger.Debug))

	registry, err := discovery.NewConsulRegistry(os.Getenv("REGISTRY_SERVICE_ADDRESS"))
	if err != nil {
		logger.Panicf(ctx, "failed to start discovery mechanism: %s", err)
	}

	cfgManager, err := sharedconfig.NewConfigManager[cfg.Config](ctx, registry, configServiceName, configProfile, cfg.PostProcess)
	if err != nil {
		logger.Panicf(ctx, "failed to set up config manager: %s", err)
	}
	defer cfgManager.Stop()

	config := cfgManager.GetConfig()
	sharedutils.StartMigration(config.Database.ConnectionString, config.Database.MigrationPath)

	serviceName := config.Service.Name
	instanceId := discovery.GenerateInstanceID(serviceName)
	go sharedutils.RegisterToDiscoveryService(ctx, registry, serviceName, instanceId, config.Service.Hostname, config.Service.APIPort)

	otelShutdownFunc, err := tracer.SetupOTelSDK(ctx, *config)
	if err != nil {
		logger.Panicf(ctx, "failed to setup otel sdk: %v", err)
	}

	dbConn := sharedutils.ConnectDB(ctx, config.Database.ConnectionString)
	defer dbConn.Close()

	server := SetupServer(dbConn, config)
	go func() {
		logger.Infof(ctx, "starting server on %s:%d...", config.Service.Hostname, config.Service.APIPort)
		server.ListenAndServe(ctx, false)
		stop()
	}()

	logger.Infof(ctx, "server started.")
	<-ctx.Done()

	logger.Infof(ctx, "shutdown signal received, starting graceful shutdown...")

	shutdownCtx, cancelShutdown := context.WithTimeout(ctx, shutdownTimeout)
	defer cancelShutdown()

	var shutdownWg sync.WaitGroup

	shutdownWg.Add(1)
	go (func() {
		if err := server.Shutdown(shutdownCtx); err != nil {
			if err == context.DeadlineExceeded {
				logger.Errorf(shutdownCtx, "server shutdown timed out.")
			}
		}
		shutdownWg.Done()
	})()

	shutdownWg.Add(1)
	go (func() {
		sharedutils.DeregisterDiscoveryService(shutdownCtx, registry, serviceName, instanceId)
		shutdownWg.Done()
	})()

	shutdownWg.Add(1)
	go (func() {
		otelShutdownFunc(shutdownCtx)
		shutdownWg.Done()
	})()

	shutdownWg.Wait()
	logger.Infof(shutdownCtx, "service shut down complete.")
}

func SetupServer(dbConn *pgxpool.Pool, cfg *cfg.Config) *api.APIServer {
	adminAccountRepo := repositories.NewAdminAccountRepository(dbConn)
	authService := services.NewAuthService(adminAccountRepo)
	authHandler := authhandler.NewAuthHandler(authService)

	return api.NewAPIServer(authHandler, cfg, dbConn)
}
```

- [ ] **Step 6: Write `Dockerfile`**

```dockerfile
# STAGE 1
FROM golang:1.26.5-alpine AS builder

WORKDIR /usr/src/app

COPY shared/go.mod shared/go.sum ./shared/
COPY admin/go.mod admin/go.sum ./admin/
RUN go work init ./admin
RUN go work edit -replace=sen1or/letslive/shared=./shared

WORKDIR /usr/src/app/admin
RUN go mod download

WORKDIR /usr/src/app
COPY shared/ ./shared/
COPY admin/ ./admin/

WORKDIR /usr/src/app/admin
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o /app ./cmd/

# STAGE 2
FROM alpine:latest AS final

WORKDIR /usr/src/app
COPY --from=builder /usr/src/app/admin/migrations /usr/src/app/migrations

COPY --from=builder /app /usr/local/bin/app

CMD ["app"]
```

- [ ] **Step 7: Verify the whole module builds and all tests still pass**

```bash
cd backend/admin
go build ./...
go vet ./...
go test ./... -v
```

Expected: builds cleanly, `go vet` clean, all tests from Tasks 2-3 still PASS.

- [ ] **Step 8: Commit**

```bash
git add backend/admin
git commit -m "feat(admin): wire login/me handlers, API routes, main.go bootstrap, Dockerfile

Refs: docs/superpowers/specs/2026-08-17-admin-dashboard-design.md"
```

---

### Task 5: Backend infra wiring — DB, docker-compose, Kong

**Files:**
- Modify: `docker-entrypoint-initdb.d/01-create-dbs.sql`
- Modify: `example.env`
- Modify: `docker-compose-dev.yaml`
- Modify: `docker-compose.yaml`
- Modify: `configs/kong.yml`

**Interfaces:**
- Consumes: `backend/admin`'s Dockerfile and API routes (Task 4). Requires the Manual Prerequisite (config-server YAML) to already be in place.
- Produces: a running `admin` service reachable through Kong at `POST /admin/login` / `GET /admin/me` — consumed by Task 8's `admin-web` login flow.

- [ ] **Step 1: Add the database**

In `docker-entrypoint-initdb.d/01-create-dbs.sql`, add after the existing `CREATE DATABASE` statements:

```sql
-- 6. Create the 'letslive_admin' database
CREATE DATABASE letslive_admin;
```

- [ ] **Step 2: Add env vars to `example.env`**

Add near the other `*_DB_USER`/`*_DB_PASSWORD` pairs:

```
ADMIN_DB_USER=postgres
ADMIN_DB_PASSWORD=postgres
ADMIN_JWT_SECRET=change_me_admin_jwt_secret
```

- [ ] **Step 3: Add the `admin` service block to `docker-compose-dev.yaml`**

Insert after the `finance` block, following the exact shape of every other service block there:

```yaml
  admin:
    build:
      context: ./backend/
      dockerfile: admin/Dockerfile
    container_name: letslive-admin
    restart: always
    ports:
      - "7784:7784"
    expose:
      - "7784"
    environment:
      - CONFIG_SERVER_PROFILE=${CONFIG_SERVER_PROFILE}
      - CONFIG_SERVER_INTERVAL=${CONFIG_SERVER_INTERVAL}
      - REGISTRY_SERVICE_ADDRESS=${REGISTRY_SERVICE_ADDRESS}
      - ADMIN_DB_USER=${ADMIN_DB_USER}
      - ADMIN_DB_PASSWORD=${ADMIN_DB_PASSWORD}
      - ADMIN_JWT_SECRET=${ADMIN_JWT_SECRET}
    networks:
      general_network:
    depends_on:
      consul:
        condition: service_healthy
```

- [ ] **Step 4: Add the same `admin` service block to `docker-compose.yaml`** (prod)

Following the prod `finance` block's shape (image reference instead of build):

```yaml
  admin:
    image: sen1or/letslive-admin:latest
    container_name: letslive-admin
    restart: always
    ports:
      - "7784:7784"
    expose:
      - "7784"
    environment:
      - CONFIG_SERVER_PROFILE=${CONFIG_SERVER_PROFILE}
      - CONFIG_SERVER_INTERVAL=${CONFIG_SERVER_INTERVAL}
      - REGISTRY_SERVICE_ADDRESS=${REGISTRY_SERVICE_ADDRESS}
      - ADMIN_DB_USER=${ADMIN_DB_USER}
      - ADMIN_DB_PASSWORD=${ADMIN_DB_PASSWORD}
      - ADMIN_JWT_SECRET=${ADMIN_JWT_SECRET}
    networks:
      general_network:
    depends_on:
      consul:
        condition: service_healthy
```

- [ ] **Step 5: Add the Kong service + routes to `configs/kong.yml`**

Insert a new service block (alongside `MinIO`/`Transcode_Service` — no `jwt` plugin, matching those two):

```yaml
  - name: Admin_Service
    host: admin.service.consul
    port: 7784
    path: /v1
    connect_timeout: 60000
    read_timeout: 60000
    write_timeout: 60000
    routes:
      - name: Admin_Login_Route
        protocols:
          - http
          - https
        paths:
          - /admin/login
        strip_path: false
      - name: Admin_Private_Routes
        protocols:
          - http
          - https
        paths:
          - /admin/me
        strip_path: false
```

- [ ] **Step 6: End-to-end smoke test**

```bash
docker compose -f docker-compose-dev.yaml up -d --build admin
docker compose -f docker-compose-dev.yaml logs -f admin   # wait for "server started."
```

In another terminal, generate a real bcrypt hash for the test password `test-password-123` (uses the `golang.org/x/crypto/bcrypt` dependency already added in Task 1 Step 7, so the cost factor matches what the service itself uses):

```bash
cd backend/admin
HASH=$(cat <<'EOF' | go run -
package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash, _ := bcrypt.GenerateFromPassword([]byte("test-password-123"), bcrypt.DefaultCost)
	fmt.Println(string(hash))
}
EOF
)
echo "$HASH"
```

Then insert the test admin account with that hash:

```bash
docker compose -f docker-compose-dev.yaml exec main_db psql -U postgres -d letslive_admin -c \
  "INSERT INTO admin_accounts (email, password_hash) VALUES ('admin@letslive.app', '$HASH');"
```

Then:

```bash
curl -i -c /tmp/admin-cookies.txt -X POST http://localhost:8000/admin/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@letslive.app","password":"test-password-123"}'
# Expected: HTTP/1.1 200, Set-Cookie: ADMIN_ACCESS_TOKEN=...

curl -i -b /tmp/admin-cookies.txt http://localhost:8000/admin/me
# Expected: HTTP/1.1 200, {"success":true,...,"data":{"email":"admin@letslive.app"}}

curl -i http://localhost:8000/admin/me
# Expected: HTTP/1.1 401 (no cookie)
```

- [ ] **Step 7: Commit**

```bash
git add docker-entrypoint-initdb.d/01-create-dbs.sql example.env docker-compose-dev.yaml docker-compose.yaml configs/kong.yml
git commit -m "feat(admin): wire admin service into local/prod compose and Kong

Refs: docs/superpowers/specs/2026-08-17-admin-dashboard-design.md"
```

---

### Task 6: admin-web scaffold

**Files:**
- Create: `admin-web/package.json`, `admin-web/tsconfig.json`, `admin-web/next.config.js`, `admin-web/postcss.config.mjs`, `admin-web/eslint.config.mjs`
- Create: `admin-web/lib/global.ts`
- Create: `admin-web/app/globals.css`
- Create: `admin-web/app/layout.tsx`
- Create: `admin-web/components/query-provider.tsx`

**Interfaces:**
- Produces: a buildable, lintable Next.js app skeleton with a `QueryProvider` — consumed by Task 7 (API client hooks run inside this provider) and Task 8 (pages render inside this layout).

- [ ] **Step 1: `package.json`**

```json
{
    "name": "letslive-admin-web",
    "version": "0.1.0",
    "private": true,
    "scripts": {
        "dev": "next dev",
        "build": "next build",
        "start": "next start",
        "lint": "eslint ."
    },
    "dependencies": {
        "@next/env": "16.3.0",
        "@tanstack/react-query": "^5.101.4",
        "next": "16.3.0",
        "react": "19.2.8",
        "react-dom": "19.2.8"
    },
    "devDependencies": {
        "@tailwindcss/postcss": "^4.3.3",
        "@types/node": "^26",
        "@types/react": "19.2.18",
        "@types/react-dom": "19.2.4",
        "eslint": "^9.39.1",
        "eslint-config-next": "16.3.0",
        "postcss": "^8",
        "tailwindcss": "^4.2.1",
        "typescript": "^5"
    }
}
```

- [ ] **Step 2: `tsconfig.json`**

```json
{
    "compilerOptions": {
        "lib": ["dom", "dom.iterable", "esnext"],
        "allowJs": true,
        "skipLibCheck": true,
        "strict": true,
        "noEmit": true,
        "esModuleInterop": true,
        "module": "esnext",
        "moduleResolution": "bundler",
        "resolveJsonModule": true,
        "isolatedModules": true,
        "jsx": "react-jsx",
        "incremental": true,
        "plugins": [{ "name": "next" }],
        "paths": { "@/*": ["./*"] },
        "target": "ES2017"
    },
    "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
    "exclude": ["node_modules"]
}
```

- [ ] **Step 3: `next.config.js`, `postcss.config.mjs`, `eslint.config.mjs`**

`next.config.js`:

```js
/** @type {import('next').NextConfig} */
const nextConfig = {
    reactStrictMode: false,
    output: "standalone",
};

module.exports = nextConfig;
```

`postcss.config.mjs`:

```js
/** @type {import('postcss-load-config').Config} */
const config = {
    plugins: {
        "@tailwindcss/postcss": {},
    },
};

export default config;
```

`eslint.config.mjs`:

```js
import next from "eslint-config-next/core-web-vitals";

const config = [...next];

export default config;
```

- [ ] **Step 4: `lib/global.ts`**

```ts
function getAdminApiUrl(): string {
    const url = process.env.NEXT_PUBLIC_ADMIN_API_URL?.trim();

    if (!url) {
        if (typeof window !== "undefined") {
            console.error("Missing NEXT_PUBLIC_ADMIN_API_URL environment variable");
        }
        throw new Error("Missing required environment variable: NEXT_PUBLIC_ADMIN_API_URL");
    }

    return url;
}

const GLOBAL = Object.freeze({
    ADMIN_API_URL: getAdminApiUrl(),
});

export default GLOBAL;
```

- [ ] **Step 5: `app/globals.css`**

```css
@import "tailwindcss";

@theme {
    --color-background: #0b0d12;
    --color-foreground: #e6e8eb;
    --color-muted: #9aa1ac;
    --color-border: #262b33;
    --color-primary: #4f7cff;
    --color-primary-foreground: #ffffff;
    --color-destructive: #ef4444;
}

body {
    background: var(--color-background);
    color: var(--color-foreground);
}
```

- [ ] **Step 6: `components/query-provider.tsx`**

```tsx
"use client";

import { useState } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

export default function QueryProvider({
    children,
}: {
    children: React.ReactNode;
}) {
    const [queryClient] = useState(
        () =>
            new QueryClient({
                defaultOptions: {
                    queries: {
                        staleTime: 30_000,
                        retry: 1,
                    },
                },
            }),
    );

    return (
        <QueryClientProvider client={queryClient}>
            {children}
        </QueryClientProvider>
    );
}
```

- [ ] **Step 7: `app/layout.tsx`**

```tsx
import "@/app/globals.css";
import QueryProvider from "@/components/query-provider";

export const metadata = {
    title: "letslive admin",
};

export default function RootLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <html lang="en">
            <body>
                <QueryProvider>{children}</QueryProvider>
            </body>
        </html>
    );
}
```

- [ ] **Step 8: Install and verify the build**

```bash
cd admin-web
npm install
npm run build
```

Expected: build fails only because `app/page.tsx` doesn't exist yet — that's Task 8. If any other error appears, fix it before moving on.

- [ ] **Step 9: Commit**

```bash
git add admin-web/package.json admin-web/package-lock.json admin-web/tsconfig.json admin-web/next.config.js admin-web/postcss.config.mjs admin-web/eslint.config.mjs admin-web/lib admin-web/app admin-web/components
git commit -m "feat(admin-web): scaffold Next.js app — config, layout, query provider

Refs: docs/superpowers/specs/2026-08-17-admin-dashboard-design.md"
```

---

### Task 7: admin-web API client and auth hook

**Files:**
- Create: `admin-web/types/fetch-response.ts`
- Create: `admin-web/types/admin.ts`
- Create: `admin-web/lib/api/api-error.ts`
- Create: `admin-web/lib/api/admin-fetch.ts`
- Create: `admin-web/lib/api/auth.ts`

**Interfaces:**
- Consumes: `GLOBAL.ADMIN_API_URL` (Task 6).
- Produces: `Login(email, password): Promise<ApiResponse<null>>`, `GetMe(): Promise<ApiResponse<AdminMe>>` — consumed by Task 8's login page and protected home page. (No `Logout` — there's no server-side session to invalidate yet since the token is a single non-rotating JWT; clearing the cookie client-side isn't possible since it's httpOnly. Out of scope here, same as the spec's "no refresh-token rotation" simplification — a real logout endpoint is a small future addition, not needed for this plan's login demo.)

- [ ] **Step 1: `types/fetch-response.ts`**

```ts
export type ErrorDetail = Record<string, unknown>;
export type ErrorDetails = ErrorDetail[];

export type ApiResponse<T> = {
    requestId: string;
    success: boolean;
    statusCode: number;
    code: number;
    key: string;
    message: string;
    data?: T;
    errorDetails?: ErrorDetails;
};
```

- [ ] **Step 2: `types/admin.ts`**

```ts
export type AdminMe = {
    email: string;
};
```

- [ ] **Step 3: `lib/api/api-error.ts`**

```ts
import { ApiResponse } from "@/types/fetch-response";

export class ApiError extends Error {
    key: string;
    requestId: string;

    constructor(res: Pick<ApiResponse<unknown>, "key" | "requestId" | "message">) {
        super(res.message);
        this.name = "ApiError";
        this.key = res.key;
        this.requestId = res.requestId;
    }
}

export function unwrapResponse<T>(res: ApiResponse<T>): T {
    if (!res.success) {
        throw new ApiError(res);
    }
    return res.data as T;
}
```

- [ ] **Step 4: `lib/api/admin-fetch.ts`**

```ts
import { ApiResponse } from "@/types/fetch-response";
import GLOBAL from "@/lib/global";

type WithStatusCode<T> = T & { statusCode: number };

export async function adminFetch<T>(
    path: string,
    options: RequestInit = {},
): Promise<WithStatusCode<ApiResponse<T>>> {
    const url = path.startsWith("http") ? path : GLOBAL.ADMIN_API_URL + path;

    const response = await fetch(url, {
        credentials: "include",
        ...options,
        headers: {
            "Content-Type": "application/json",
            ...(options.headers ?? {}),
        },
    });

    const data = (await response.json()) as ApiResponse<T>;
    return { ...data, statusCode: response.status };
}
```

- [ ] **Step 5: `lib/api/auth.ts`**

```ts
import { adminFetch } from "@/lib/api/admin-fetch";
import { ApiResponse } from "@/types/fetch-response";
import { AdminMe } from "@/types/admin";

export async function Login(
    email: string,
    password: string,
): Promise<ApiResponse<null>> {
    return adminFetch<null>("/admin/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
    });
}

export async function GetMe(): Promise<ApiResponse<AdminMe>> {
    return adminFetch<AdminMe>("/admin/me");
}
```

- [ ] **Step 6: Verify the build**

```bash
cd admin-web
npx tsc --noEmit
```

Expected: no type errors.

- [ ] **Step 7: Commit**

```bash
git add admin-web/types admin-web/lib/api
git commit -m "feat(admin-web): add admin API client (login, me)

Refs: docs/superpowers/specs/2026-08-17-admin-dashboard-design.md"
```

---

### Task 8: admin-web login page + protected home page

**Files:**
- Create: `admin-web/components/require-admin-auth.tsx`
- Create: `admin-web/app/login/page.tsx`
- Create: `admin-web/app/page.tsx`

**Interfaces:**
- Consumes: `Login`, `GetMe` (Task 7).
- Produces: the full browser-testable login flow this plan set out to build.

- [ ] **Step 1: `components/require-admin-auth.tsx`**

```tsx
"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { GetMe } from "@/lib/api/auth";
import { unwrapResponse } from "@/lib/api/api-error";

export default function RequireAdminAuth({
    children,
}: {
    children: React.ReactNode;
}) {
    const router = useRouter();
    const [redirected, setRedirected] = useState(false);

    const { data: me, isLoading, isError } = useQuery({
        queryKey: ["admin", "me"],
        queryFn: async () => unwrapResponse(await GetMe()),
        retry: false,
    });

    useEffect(() => {
        if (!isLoading && isError && !redirected) {
            setRedirected(true);
            router.replace("/login");
        }
    }, [isLoading, isError, redirected, router]);

    if (isLoading) {
        return <div className="p-8 text-muted">Loading...</div>;
    }

    if (isError || !me) {
        return null;
    }

    return <>{children}</>;
}
```

- [ ] **Step 2: `app/login/page.tsx`**

```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Login } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/api-error";

export default function LoginPage() {
    const router = useRouter();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [submitting, setSubmitting] = useState(false);

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault();
        setError(null);
        setSubmitting(true);

        try {
            const res = await Login(email, password);
            if (!res.success) {
                throw new ApiError(res);
            }
            router.replace("/");
        } catch (err) {
            setError(err instanceof ApiError ? err.message : "Login failed.");
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <div className="flex min-h-screen items-center justify-center">
            <form
                onSubmit={handleSubmit}
                className="border-border w-full max-w-sm space-y-4 rounded-lg border p-6"
            >
                <h1 className="text-xl font-semibold">Admin login</h1>

                <div className="space-y-1">
                    <label className="text-muted text-sm" htmlFor="email">
                        Email
                    </label>
                    <input
                        id="email"
                        type="email"
                        required
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        className="border-border w-full rounded border bg-transparent px-3 py-2"
                    />
                </div>

                <div className="space-y-1">
                    <label className="text-muted text-sm" htmlFor="password">
                        Password
                    </label>
                    <input
                        id="password"
                        type="password"
                        required
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="border-border w-full rounded border bg-transparent px-3 py-2"
                    />
                </div>

                {error && <p className="text-destructive text-sm">{error}</p>}

                <button
                    type="submit"
                    disabled={submitting}
                    className="bg-primary text-primary-foreground w-full rounded px-3 py-2 disabled:opacity-50"
                >
                    {submitting ? "Logging in..." : "Log in"}
                </button>
            </form>
        </div>
    );
}
```

- [ ] **Step 3: `app/page.tsx`**

```tsx
"use client";

import { useQuery } from "@tanstack/react-query";
import RequireAdminAuth from "@/components/require-admin-auth";
import { GetMe } from "@/lib/api/auth";
import { unwrapResponse } from "@/lib/api/api-error";

function HomeContent() {
    const { data: me } = useQuery({
        queryKey: ["admin", "me"],
        queryFn: async () => unwrapResponse(await GetMe()),
    });

    return (
        <div className="p-8">
            <h1 className="text-xl font-semibold">letslive admin</h1>
            <p className="text-muted mt-2">Logged in as {me?.email}.</p>
        </div>
    );
}

export default function HomePage() {
    return (
        <RequireAdminAuth>
            <HomeContent />
        </RequireAdminAuth>
    );
}
```

- [ ] **Step 4: Verify the build**

```bash
cd admin-web
npx tsc --noEmit
NEXT_PUBLIC_ADMIN_API_URL=http://localhost:8000 npm run build
```

Expected: both succeed with zero errors. `NEXT_PUBLIC_ADMIN_API_URL` must be set for the `build` step specifically — `lib/global.ts` (Task 6) throws if it's missing, and Next.js executes that module during static generation of the `"use client"` pages added in this task (Steps 1-3), not only in the browser. `tsc --noEmit` is a pure type check and never executes the module, so it needs no env var.

- [ ] **Step 5: Manual browser end-to-end test**

With Task 5's local stack running (`docker compose -f docker-compose-dev.yaml up -d`) and the test admin account from Task 5 Step 6 already seeded:

```bash
cd admin-web
NEXT_PUBLIC_ADMIN_API_URL=http://localhost:8000 npm run dev
```

- Visit `http://localhost:3000/` while logged out → redirected to `/login`.
- Log in with the seeded test account's email/password → redirected to `/`, page shows "Logged in as admin@letslive.app."
- Refresh `/` → still logged in (cookie persists).
- Clear the `ADMIN_ACCESS_TOKEN` cookie via devtools, refresh `/` → redirected to `/login` again.
- Try logging in with a wrong password → inline error shown, stays on `/login`.

- [ ] **Step 6: Commit**

```bash
git add admin-web/components/require-admin-auth.tsx admin-web/app/login admin-web/app/page.tsx
git commit -m "feat(admin-web): add login page and protected home page

Refs: docs/superpowers/specs/2026-08-17-admin-dashboard-design.md"
```

---

### Task 9: admin-web infra — Dockerfile, compose entries, CI

**Files:**
- Create: `admin-web/Dockerfile`
- Modify: `docker-compose-dev.yaml`
- Modify: `docker-compose.yaml`
- Modify: `example.env`
- Modify: `.github/workflows/build-and-publish-images.yml`
- Modify: `.github/workflows/deploy.yml`

**Interfaces:**
- Consumes: `admin-web`'s app (Task 8), `backend/admin`'s Dockerfile/compose entry (Tasks 4-5).
- Produces: buildable/deployable images for both new services — closes out this plan.

- [ ] **Step 1: `admin-web/Dockerfile`**

```dockerfile
FROM node:22-alpine

EXPOSE 5001

ENV PORT=5001

WORKDIR /home/nextjs/app

ARG NEXT_PUBLIC_ADMIN_API_URL

ENV NEXT_PUBLIC_ADMIN_API_URL=${NEXT_PUBLIC_ADMIN_API_URL}

COPY package.json .
COPY package-lock.json .

RUN npm install
RUN npx next telemetry disable

COPY . .

RUN npm run build

RUN addgroup -g 1001 -S nodejs
RUN adduser -S nextjs -u 1001

USER nextjs

CMD [ "npm", "start" ]
```

- [ ] **Step 2: Add `admin-web` to `docker-compose-dev.yaml`** (uncomment-style block, mirroring the commented-out `web` block at the top of the file)

```yaml
  admin-web:
    build:
      context: ./admin-web
      dockerfile: Dockerfile
      args:
        NEXT_PUBLIC_ADMIN_API_URL: ${NEXT_PUBLIC_ADMIN_API_URL}
    container_name: letslive_admin_web
    network_mode: "host"
    ports:
      - "5001:5001"
```

- [ ] **Step 3: Add `admin-web` to `docker-compose.yaml`** (prod)

```yaml
  admin-web:
    image: sen1or/letslive-admin-web:latest
    container_name: letslive_admin_web
    network_mode: "host"
    ports:
      - "5001:5001"
```

Note: wiring a real domain/subdomain to port 5001 is server-level reverse-proxy config (see `configs/nginx.conf`'s comment — the equivalent for `web`'s port 5000 already lives outside this repo). Document this as a manual deploy step; it's not something this plan can commit.

- [ ] **Step 4: Add `NEXT_PUBLIC_ADMIN_API_URL` to `example.env`**

```
NEXT_PUBLIC_ADMIN_API_URL=http://localhost:8000
```

- [ ] **Step 5: Add `admin` to the backend matrix in `build-and-publish-images.yml`**

```yaml
            {
              path: "./backend/admin",
              name: "letslive-admin",
              context: "./backend",
            },
```

- [ ] **Step 6: Add a new `build-and-publish-admin-web` job in the same file**

```yaml
  build-and-publish-admin-web:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v7

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v4

      - name: Login to DockerHub
        uses: docker/login-action@v4
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - name: Build and push
        uses: docker/build-push-action@v7
        with:
          context: ./admin-web
          file: ./admin-web/Dockerfile
          push: true
          tags: ${{ secrets.DOCKERHUB_USERNAME }}/letslive-admin-web:latest
          build-args: |
            NEXT_PUBLIC_ADMIN_API_URL=${{ vars.NEXT_PUBLIC_ADMIN_API_URL }}
```

- [ ] **Step 7: Add the new secrets to `deploy.yml`'s `.env` generation**

Add to the `Service database credentials` block:

```
ADMIN_DB_USER=${{ secrets.ADMIN_DB_USER }}
ADMIN_DB_PASSWORD=${{ secrets.ADMIN_DB_PASSWORD }}
```

Add a new block near `# Authentication tokens`:

```
# Admin service
ADMIN_JWT_SECRET=${{ secrets.ADMIN_JWT_SECRET }}
```

Add to `# Frontend environment variables`:

```
NEXT_PUBLIC_ADMIN_API_URL=${{ secrets.NEXT_PUBLIC_ADMIN_API_URL }}
```

Manual step (GitHub repo settings, outside this repo's files): add `ADMIN_DB_USER`, `ADMIN_DB_PASSWORD`, `ADMIN_JWT_SECRET`, and (as a repo variable, matching how `NEXT_PUBLIC_BACKEND_PROTOCOL` etc. are set up) `NEXT_PUBLIC_ADMIN_API_URL` in the repo's Actions secrets/variables before the next deploy run.

- [ ] **Step 8: Commit**

```bash
git add admin-web/Dockerfile docker-compose-dev.yaml docker-compose.yaml example.env .github/workflows/build-and-publish-images.yml .github/workflows/deploy.yml
git commit -m "feat(admin-web): add Dockerfile, compose entries, and CI build/deploy wiring

Refs: docs/superpowers/specs/2026-08-17-admin-dashboard-design.md"
```

---

## Definition of Done

- `docker compose -f docker-compose-dev.yaml up` brings up a healthy `admin` service; `POST /admin/login` and `GET /admin/me` work as shown in Task 5 Step 6.
- `admin-web` runs locally (`npm run dev`), and the full flow from Task 8 Step 5 works in a real browser: login, protected page, refresh persists session, cookie clear forces re-login, wrong password shows an inline error.
- All Go unit tests (Tasks 2, 3) pass; `go vet ./...` is clean.
- `npx tsc --noEmit` and `npm run build` are clean in `admin-web`.
- CI (`build-and-publish-images.yml`) builds and pushes both new images once secrets/vars are added (Task 9 Step 7's manual step).

**Next:** Plan 2 (stats overview) builds on this — it adds `GET /admin/stats` behind the same `RequireAdminAuth` middleware, plus the internal `/v1/internal/stats` endpoints on `user`/`livestream`/`vod`/`finance`.
