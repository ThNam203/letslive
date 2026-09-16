package vod

import (
	"context"
	"sen1or/letslive/vod/domains"

	"github.com/gofrs/uuid/v5"
)

const (
	minWatchSeconds    int64   = 15
	minWatchPercentage float64 = 0.10
)

func (s *VODService) RegisterView(ctx context.Context, vodId uuid.UUID, watchedSeconds int64) error {
	// Fetch VOD to get the stored duration
	vod, errResp := s.vodRepo.GetById(ctx, vodId)
	if errResp != nil {
		return errResp
	}

	// Validate watch time threshold: at least 15 seconds OR 10% of video duration (whichever is smaller)
	threshold := minWatchSeconds
	tenPercent := int64(float64(vod.Duration) * minWatchPercentage)
	if tenPercent < threshold {
		threshold = tenPercent
	}
	// For very short videos (< 1s), allow any watch time
	if vod.Duration > 0 && threshold < 1 {
		threshold = 1
	}

	if watchedSeconds < threshold {
		return domains.ErrVODViewThreshold
	}

	return s.vodRepo.IncrementViewCount(ctx, vodId)
}
