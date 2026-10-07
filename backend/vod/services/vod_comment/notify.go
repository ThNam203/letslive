package vodcomment

import (
	"context"
	"fmt"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/vod/domains"
	usergateway "sen1or/letslive/vod/gateway/user"
	"time"
)

const (
	notifyTimeout = 10 * time.Second
	// must stay within the user service's notification message limit
	notificationMessageMaxRunes = 500
)

// notifyNewComment tells the VOD owner about a top-level comment, or the parent
// comment's author about a reply. It is best-effort: failures are only logged.
func (s *VODCommentService) notifyNewComment(ctx context.Context, vod domains.VOD, parent *domains.VODComment, comment domains.VODComment) {
	ctx, cancel := context.WithTimeout(ctx, notifyTimeout)
	defer cancel()

	recipientId := vod.UserId
	if parent != nil {
		if parent.IsDeleted {
			return
		}
		recipientId = parent.UserId
	}

	if recipientId == comment.UserId {
		return
	}

	commenterName := "Someone"
	if info, err := s.userGateway.GetUserPublicInfo(ctx, comment.UserId); err == nil && info != nil && info.Username != "" {
		commenterName = info.Username
	}

	notifType := usergateway.NotificationTypeVODComment
	title := "New comment on your video"
	message := fmt.Sprintf("%s commented on \"%s\": %s", commenterName, vod.Title, comment.Content)
	if parent != nil {
		notifType = usergateway.NotificationTypeVODCommentReply
		title = "New reply to your comment"
		message = fmt.Sprintf("%s replied to your comment on \"%s\": %s", commenterName, vod.Title, comment.Content)
	}

	actionURL := fmt.Sprintf("/users/%s/vods/%s", vod.UserId, vod.Id)
	referenceId := comment.Id.String()
	err := s.userGateway.CreateNotification(ctx, usergateway.CreateNotificationRequest{
		UserId:      recipientId.String(),
		Type:        notifType,
		Title:       title,
		Message:     truncateRunes(message, notificationMessageMaxRunes),
		ActionUrl:   &actionURL,
		ReferenceId: &referenceId,
	})
	if err != nil {
		logger.Errorf(ctx, "failed to create notification for comment %s: %v", comment.Id, err)
	}
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
