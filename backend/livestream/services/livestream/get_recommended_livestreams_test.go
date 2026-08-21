package livestream

import (
	"context"
	"testing"

	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/internal/testutil"
	"sen1or/letslive/livestream/response"
	"sen1or/letslive/shared/pkg/logger"

	"github.com/gofrs/uuid/v5"
)

func TestMain(m *testing.M) {
	logger.Init(logger.Debug)
	m.Run()
}

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
