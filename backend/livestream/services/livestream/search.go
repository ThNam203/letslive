package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/shared/pkg/logger"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
)

const (
	searchStreamerLimit = 10
	// the user search tolerates several typos, so very short queries would
	// match almost every streamer and flood the results with their streams
	searchStreamerMinQueryLength = 3
	searchStreamerTimeout        = 2 * time.Second
)

func (s *LivestreamService) SearchLiveLivestreams(ctx context.Context, query string, page int, limit int) ([]domains.Livestream, error) {
	return s.livestreamRepo.SearchLive(ctx, query, s.searchStreamerIds(ctx, query), page, limit)
}

// A slow or failing user service must not fail the search: it falls back to
// title/description matching only.
func (s *LivestreamService) searchStreamerIds(ctx context.Context, query string) []uuid.UUID {
	if utf8.RuneCountInString(query) < searchStreamerMinQueryLength {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, searchStreamerTimeout)
	defer cancel()

	streamerIds, err := s.userGateway.SearchUserIds(ctx, query, searchStreamerLimit)
	if err != nil {
		logger.Warnf(ctx, "failed to search streamers, searching livestreams by title/description only: %v", err)
		return nil
	}
	return streamerIds
}
