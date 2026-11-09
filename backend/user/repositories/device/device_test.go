package devicerepo

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMain(m *testing.M) {
	logger.Init(logger.LogLevel(logger.Debug))
	os.Exit(m.Run())
}

// These tests run the real SQL against a Postgres that has migration 0012
// applied. Set DEVICE_TEST_DATABASE_URL to enable them, for example
// postgres://postgres:postgres@localhost:9999/letslive_user
func newTestRepo(t *testing.T) (domains.DeviceTokenRepository, *pgxpool.Pool) {
	t.Helper()
	url := os.Getenv("DEVICE_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("DEVICE_TEST_DATABASE_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return NewDeviceTokenRepository(pool), pool
}

func createTestUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	suffix := uuid.Must(uuid.NewV4()).String()
	var id uuid.UUID
	err := pool.QueryRow(ctx,
		`INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id`,
		"dev"+suffix[:12], suffix+"@device-test.invalid",
	).Scan(&id)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id) })
	return id
}

func uniqueToken() string {
	return fmt.Sprintf("tok-%s", uuid.Must(uuid.NewV4()))
}

func strPtr(s string) *string { return &s }

func TestRegisteringSameTokenForAnotherUserReassignsIt(t *testing.T) {
	repo, pool := newTestRepo(t)
	ctx := context.Background()
	userA, userB := createTestUser(t, pool), createTestUser(t, pool)
	token := uniqueToken()

	first, err := repo.Upsert(ctx, domains.DeviceToken{UserId: userA, Token: token, Platform: "ios", Locale: strPtr("en")})
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	second, err := repo.Upsert(ctx, domains.DeviceToken{UserId: userB, Token: token, Platform: "android", Locale: strPtr("vi")})
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	if second.Id != first.Id || second.UserId != userB || second.Platform != "android" || *second.Locale != "vi" {
		t.Fatalf("token not reassigned/updated: %+v", second)
	}
	if !second.LastSeenAt.After(first.LastSeenAt) {
		t.Errorf("last_seen_at not refreshed: %v -> %v", first.LastSeenAt, second.LastSeenAt)
	}
	if got, _ := repo.ListTokensByUserID(ctx, userA); len(got) != 0 {
		t.Errorf("previous owner still has the token: %+v", got)
	}
}

func TestInvalidPlatformIsRejectedByDatabase(t *testing.T) {
	repo, pool := newTestRepo(t)
	userA := createTestUser(t, pool)

	_, err := repo.Upsert(context.Background(), domains.DeviceToken{UserId: userA, Token: uniqueToken(), Platform: "windows"})
	if err == nil {
		t.Fatal("expected check constraint error for invalid platform")
	}
}

func TestDeleteRemovesOnlyCallersToken(t *testing.T) {
	repo, pool := newTestRepo(t)
	ctx := context.Background()
	owner, other := createTestUser(t, pool), createTestUser(t, pool)
	token := uniqueToken()
	if _, err := repo.Upsert(ctx, domains.DeviceToken{UserId: owner, Token: token, Platform: "ios"}); err != nil {
		t.Fatal(err)
	}

	if err := repo.DeleteByUserAndToken(ctx, other, token); err != nil {
		t.Fatalf("foreign delete must be a no-op success: %v", err)
	}
	if got, _ := repo.ListTokensByUserID(ctx, owner); len(got) != 1 {
		t.Fatalf("foreign delete removed the token: %+v", got)
	}

	if err := repo.DeleteByUserAndToken(ctx, owner, token); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.ListTokensByUserID(ctx, owner); len(got) != 0 {
		t.Fatalf("owner delete did not remove the token: %+v", got)
	}
	if err := repo.DeleteByUserAndToken(ctx, owner, token); err != nil {
		t.Fatalf("deleting a missing token must succeed: %v", err)
	}
}

func TestDeleteByTokenIsUnscoped(t *testing.T) {
	repo, pool := newTestRepo(t)
	ctx := context.Background()
	owner := createTestUser(t, pool)
	token := uniqueToken()
	if _, err := repo.Upsert(ctx, domains.DeviceToken{UserId: owner, Token: token, Platform: "android"}); err != nil {
		t.Fatal(err)
	}

	if err := repo.DeleteByToken(ctx, token); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.ListTokensByUserID(ctx, owner); len(got) != 0 {
		t.Fatalf("token still present: %+v", got)
	}
}

func TestListReturnsTokensForUser(t *testing.T) {
	repo, pool := newTestRepo(t)
	ctx := context.Background()
	userA, userB := createTestUser(t, pool), createTestUser(t, pool)
	tokenA1, tokenA2, tokenB := uniqueToken(), uniqueToken(), uniqueToken()
	for _, d := range []domains.DeviceToken{
		{UserId: userA, Token: tokenA1, Platform: "ios"},
		{UserId: userA, Token: tokenA2, Platform: "android", Locale: strPtr("vi")},
		{UserId: userB, Token: tokenB, Platform: "ios"},
	} {
		if _, err := repo.Upsert(ctx, d); err != nil {
			t.Fatal(err)
		}
	}

	got, err := repo.ListTokensByUserID(ctx, userA)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 tokens for user A, got %d: %+v", len(got), got)
	}
	seen := map[string]bool{}
	for _, d := range got {
		if d.UserId != userA {
			t.Errorf("token of another user listed: %+v", d)
		}
		seen[d.Token] = true
	}
	if !seen[tokenA1] || !seen[tokenA2] {
		t.Errorf("missing tokens: %v", seen)
	}
}
