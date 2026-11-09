package rtmp

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/gofrs/uuid/v5"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/transcode/domains"
	livestreamdto "sen1or/letslive/transcode/gateway/livestream/dto"
	"sen1or/letslive/transcode/live"
	"sen1or/letslive/transcode/response"
)

// gatewayTimeout bounds each call to the livestream and vod services.
const gatewayTimeout = 10 * time.Second

// livestreamEnder is the part of the livestream gateway used when a stream ends.
type livestreamEnder interface {
	EndLivestream(ctx context.Context, streamId string, endDTO livestreamdto.EndLivestreamRequestDTO) (*uuid.UUID, *response.Response[any])
	UpdateVODStatus(ctx context.Context, vodId string, status domains.VODStatus, playbackUrl string, thumbnailUrl string, duration *int64) error
}

// liveSession is the part of live.Session used when a stream ends.
type liveSession interface {
	Duration() time.Duration
	Close(ctx context.Context) error
}

// finishStream ends a livestream whose RTMP connection is gone. The
// livestream is ended at once, with its VOD created as "processing"; the
// session is then closed, which writes the VOD, and the VOD is marked
// "ready" with its playback URL, or "failed" when nothing was written.
// ctx only carries log fields: every call gets its own timeout and still runs
// during a service shutdown.
func finishStream(ctx context.Context, ender livestreamEnder, sess liveSession, streamId, vodPlaybackPrefix string) {
	base := context.WithoutCancel(ctx)
	duration := int64(math.Ceil(sess.Duration().Seconds()))

	endCtx, cancelEnd := context.WithTimeout(base, gatewayTimeout)
	vodId, endErr := ender.EndLivestream(endCtx, streamId, livestreamdto.EndLivestreamRequestDTO{
		EndedAt:   time.Now(),
		Duration:  duration,
		VODStatus: string(domains.VODStatusProcessing),
	})
	cancelEnd()
	if endErr != nil {
		logger.Errorf(ctx, "livestream %s: failed to end livestream: %s", streamId, endErr.Message)
	}

	closeCtx, cancelClose := context.WithTimeout(base, sessionCloseTimeout)
	closeErr := sess.Close(closeCtx)
	cancelClose()
	switch {
	case errors.Is(closeErr, live.ErrNothingRecorded):
		logger.Infof(ctx, "livestream %s: %v", streamId, closeErr)
	case closeErr != nil:
		logger.Errorf(ctx, "livestream %s: failed to close session: %v", streamId, closeErr)
	}

	if endErr != nil {
		return
	}
	if vodId == nil {
		logger.Errorf(ctx, "livestream %s: livestream service returned no VOD id, VOD status not updated", streamId)
		return
	}

	status, playbackURL, vodDuration := domains.VODStatusReady, fmt.Sprintf("%s/%s/index.m3u8", vodPlaybackPrefix, streamId), &duration
	if closeErr != nil {
		status, playbackURL, vodDuration = domains.VODStatusFailed, "", nil
	}

	updateCtx, cancelUpdate := context.WithTimeout(base, gatewayTimeout)
	defer cancelUpdate()
	if err := ender.UpdateVODStatus(updateCtx, vodId.String(), status, playbackURL, "", vodDuration); err != nil {
		logger.Errorf(ctx, "livestream %s: failed to mark VOD %s %s: %v", streamId, vodId, status, err)
	}
}
