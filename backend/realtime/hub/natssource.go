package hub

import "github.com/nats-io/nats.go"

type NATSSource struct {
	conn *nats.Conn
}

func NewNATSSource(conn *nats.Conn) *NATSSource {
	return &NATSSource{conn: conn}
}

func (s *NATSSource) Subscribe(subject string, handler func(data []byte)) (Subscription, error) {
	subscription, err := s.conn.Subscribe(subject, func(msg *nats.Msg) { handler(msg.Data) })
	if err != nil {
		return nil, err
	}
	return subscription, nil
}
