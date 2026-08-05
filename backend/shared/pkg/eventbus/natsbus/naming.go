package natsbus

import "strings"

// streamNameFor derives a JetStream-safe stream name from a topic name.
// Stream and consumer names cannot contain '.' in NATS, but topics in this
// project are dot-separated (e.g. "letslive.livestream"). The topic string
// itself is still used verbatim as the NATS subject.
func streamNameFor(topic string) string {
	return strings.ReplaceAll(topic, ".", "_")
}

// durableNameFor derives a JetStream-safe durable consumer name from a
// consumer group id. Same character restrictions as stream names apply.
func durableNameFor(groupID string) string {
	r := strings.NewReplacer(".", "_", " ", "_")
	return r.Replace(groupID)
}
