// Package push delivers mobile push notifications through FCM.
//
// A push flows through three layers:
//
//	Async   bounded worker pool; enqueueing never blocks the caller and each
//	        delivery runs detached from the request with a timeout.
//	Service skips users with a live realtime connection (PresenceChecker),
//	        loads their device tokens and removes tokens FCM reports as dead.
//	Gateway the FCM boundary (FCMGateway); fakes replace it in tests.
//
// NewFromEnv wires the layers. When FIREBASE_SERVICE_ACCOUNT_JSON is unset it
// returns a Disabled sender instead: sends are logged no-ops and nothing
// connects to NATS or Firebase. A malformed credential fails at startup.
//
// Reusable entry point: build a Message (FromNotification maps a stored
// notification, other features build their own) and call
// MessageSender.SendMessage.
package push
