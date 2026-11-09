// Package pushhandler serves the internal push endpoint used by peer services.
package pushhandler

import (
	"sen1or/letslive/user/handlers/basehandler"
	"sen1or/letslive/user/push"
)

// PushHandler exposes push.MessageSender over HTTP.
type PushHandler struct {
	basehandler.BaseHandler
	sender push.MessageSender
}

// NewPushHandler builds a PushHandler that delivers through sender.
func NewPushHandler(sender push.MessageSender) *PushHandler {
	return &PushHandler{sender: sender}
}
