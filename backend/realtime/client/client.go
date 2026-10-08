package client

import (
	"context"
	"sync"
	"time"

	"sen1or/letslive/realtime/hub"
	"sen1or/letslive/realtime/protocol"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/shared/pkg/realtime"

	"github.com/coder/websocket"
)

const (
	MaxFrameBytes    = 1024
	MaxSubscriptions = 50
	SendBuffer       = 64
	WriteTimeout     = 10 * time.Second
	PingInterval     = 30 * time.Second
)

type Hub interface {
	Subscribe(ctx context.Context, s hub.Subscriber, topic realtime.Topic) error
	Unsubscribe(ctx context.Context, s hub.Subscriber, topic realtime.Topic)
}

type Presence interface {
	Connected(ctx context.Context, userID string)
	Disconnected(ctx context.Context, userID string)
}

type Client struct {
	userID   string
	send     chan []byte
	overflow func()
	// owned by the read loop goroutine; cleanup runs on the same goroutine
	topics map[realtime.Topic]struct{}
}

func (c *Client) UserID() string { return c.userID }

// Deliver never blocks the hub: a socket that cannot keep up is dropped and
// the browser reconnects and refetches.
func (c *Client) Deliver(frame []byte) {
	select {
	case c.send <- frame:
	default:
		c.overflow()
	}
}

func Serve(ctx context.Context, conn *websocket.Conn, userID string, h Hub, p Presence) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var overflowOnce sync.Once
	c := &Client{
		userID:   userID,
		send:     make(chan []byte, SendBuffer),
		overflow: func() { overflowOnce.Do(cancel) },
		topics:   make(map[realtime.Topic]struct{}),
	}
	cleanupCtx := context.WithoutCancel(ctx)

	if userID != "" {
		p.Connected(ctx, userID)
		defer p.Disconnected(cleanupCtx, userID)

		userTopic := realtime.UserTopic(userID)
		if err := h.Subscribe(ctx, c, userTopic); err != nil {
			logger.Errorf(ctx, "failed to subscribe %s: %v", userTopic.String(), err)
			if closeErr := conn.Close(websocket.StatusTryAgainLater, "realtime unavailable"); closeErr != nil {
				logger.Debugf(ctx, "close after failed subscribe: %v", closeErr)
			}
			return
		}
		defer h.Unsubscribe(cleanupCtx, c, userTopic)
	}
	defer c.unsubscribeAll(cleanupCtx, h)

	conn.SetReadLimit(MaxFrameBytes)
	go c.writeLoop(ctx, cancel, conn)
	c.readLoop(ctx, conn, h)

	if err := conn.CloseNow(); err != nil {
		logger.Debugf(ctx, "close socket: %v", err)
	}
}

func (c *Client) readLoop(ctx context.Context, conn *websocket.Conn, h Hub) {
	for {
		messageType, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if messageType != websocket.MessageText {
			c.Deliver(protocol.ErrorFrame(protocol.ErrCodeBadFrame))
			continue
		}
		c.handleFrame(ctx, data, h)
	}
}

func (c *Client) handleFrame(ctx context.Context, data []byte, h Hub) {
	frame, err := protocol.ParseClientFrame(data)
	if err != nil {
		c.Deliver(protocol.ErrorFrame(protocol.ErrCodeBadFrame))
		return
	}

	switch frame.Op {
	case protocol.OpPing:
		c.Deliver(protocol.PongFrame())
	case protocol.OpSubscribe:
		c.subscribe(ctx, frame.Topic, h)
	case protocol.OpUnsubscribe:
		c.unsubscribe(ctx, frame.Topic, h)
	}
}

func (c *Client) subscribe(ctx context.Context, rawTopic string, h Hub) {
	topic, err := realtime.ParseTopic(rawTopic)
	if err != nil {
		c.Deliver(protocol.ErrorFrame(protocol.ErrCodeInvalidTopic))
		return
	}
	if topic.Kind != realtime.TopicKindRoom {
		c.Deliver(protocol.ErrorFrame(protocol.ErrCodeForbiddenTopic))
		return
	}
	if _, already := c.topics[topic]; already {
		return
	}
	if len(c.topics) >= MaxSubscriptions {
		c.Deliver(protocol.ErrorFrame(protocol.ErrCodeTooManySubscriptions))
		return
	}
	if err := h.Subscribe(ctx, c, topic); err != nil {
		logger.Errorf(ctx, "failed to subscribe %s: %v", topic.String(), err)
		c.Deliver(protocol.ErrorFrame(protocol.ErrCodeUnavailable))
		return
	}
	c.topics[topic] = struct{}{}
}

func (c *Client) unsubscribe(ctx context.Context, rawTopic string, h Hub) {
	topic, err := realtime.ParseTopic(rawTopic)
	if err != nil {
		c.Deliver(protocol.ErrorFrame(protocol.ErrCodeInvalidTopic))
		return
	}
	if _, subscribed := c.topics[topic]; !subscribed {
		return
	}
	h.Unsubscribe(ctx, c, topic)
	delete(c.topics, topic)
}

func (c *Client) unsubscribeAll(ctx context.Context, h Hub) {
	for topic := range c.topics {
		h.Unsubscribe(ctx, c, topic)
	}
}

func (c *Client) writeLoop(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn) {
	defer cancel()

	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case frame := <-c.send:
			writeCtx, done := context.WithTimeout(ctx, WriteTimeout)
			err := conn.Write(writeCtx, websocket.MessageText, frame)
			done()
			if err != nil {
				return
			}
		case <-ticker.C:
			// the gateway drops upstream connections that are idle for too long
			pingCtx, done := context.WithTimeout(ctx, WriteTimeout)
			err := conn.Ping(pingCtx)
			done()
			if err != nil {
				return
			}
		}
	}
}
