package events

import "sen1or/letslive/shared/pkg/eventbus"

// Topic names for the event bus.
const (
	TopicLivestream   = "letslive.livestream"
	TopicUser         = "letslive.user"
	TopicVOD          = "letslive.vod"
	TopicTranscode    = "letslive.transcode"
	TopicFinance      = "letslive.finance"
	TopicNotification = "letslive.notification"
)

// DefaultTopics returns the full list of topics that should be created on startup.
func DefaultTopics() []eventbus.TopicConfig {
	return []eventbus.TopicConfig{
		{Name: TopicLivestream, NumPartitions: 3, ReplicationFactor: 1},
		{Name: TopicUser, NumPartitions: 3, ReplicationFactor: 1},
		{Name: TopicVOD, NumPartitions: 3, ReplicationFactor: 1},
		{Name: TopicTranscode, NumPartitions: 3, ReplicationFactor: 1},
		{Name: TopicFinance, NumPartitions: 3, ReplicationFactor: 1},
		{Name: TopicNotification, NumPartitions: 3, ReplicationFactor: 1},
	}
}
