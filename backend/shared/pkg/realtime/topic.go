package realtime

import (
	"errors"
	"regexp"
	"strings"
)

const (
	TopicKindUser = "user"
	TopicKindRoom = "room"

	PresenceSubject = "rt.presence"

	subjectPrefix = "rt"
)

var (
	ErrInvalidTopic = errors.New("invalid topic")

	// ids become NATS subject tokens; anything outside this set could inject
	// wildcards ("*", ">") or extra tokens (".")
	topicIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,36}$`)
)

type Topic struct {
	Kind string
	ID   string
}

func UserTopic(userID string) Topic { return Topic{Kind: TopicKindUser, ID: userID} }

func RoomTopic(roomID string) Topic { return Topic{Kind: TopicKindRoom, ID: roomID} }

func (t Topic) String() string { return t.Kind + ":" + t.ID }

func (t Topic) Subject() string { return subjectPrefix + "." + t.Kind + "." + t.ID }

func (t Topic) Valid() bool {
	return (t.Kind == TopicKindUser || t.Kind == TopicKindRoom) && topicIDPattern.MatchString(t.ID)
}

func ParseTopic(raw string) (Topic, error) {
	kind, id, found := strings.Cut(raw, ":")
	topic := Topic{Kind: kind, ID: id}
	if !found || !topic.Valid() {
		return Topic{}, ErrInvalidTopic
	}
	return topic, nil
}
