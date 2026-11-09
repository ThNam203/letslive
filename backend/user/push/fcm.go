package push

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

const (
	// FCM accepts at most 500 tokens per multicast.
	maxTokensPerBatch = 500
	soundDefault      = "default"
)

// multicaster is the slice of *messaging.Client used here.
type multicaster interface {
	SendEachForMulticast(ctx context.Context, m *messaging.MulticastMessage) (*messaging.BatchResponse, error)
}

// FCMGateway implements Gateway on top of Firebase Cloud Messaging.
type FCMGateway struct {
	client multicaster
}

// NewFCMGateway validates the service-account JSON and builds the client. It
// does not touch the network: a malformed credential fails here, an
// unreachable FCM only fails individual sends. projectID may be empty, the
// credential's own project_id is then used.
func NewFCMGateway(ctx context.Context, serviceAccountJSON, projectID string) (*FCMGateway, error) {
	creds, err := parseServiceAccount(serviceAccountJSON)
	if err != nil {
		return nil, err
	}
	if projectID == "" {
		projectID = creds.ProjectID
	}
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID},
		option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(serviceAccountJSON)))
	if err != nil {
		return nil, fmt.Errorf("%s is not a usable credential: %w", EnvServiceAccountJSON, err)
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("create fcm client: %w", err)
	}
	return &FCMGateway{client: client}, nil
}

type serviceAccount struct {
	Type        string `json:"type"`
	ProjectID   string `json:"project_id"`
	PrivateKey  string `json:"private_key"`
	ClientEmail string `json:"client_email"`
}

func parseServiceAccount(raw string) (serviceAccount, error) {
	var sa serviceAccount
	if err := json.Unmarshal([]byte(raw), &sa); err != nil {
		return sa, fmt.Errorf("%s is not valid JSON: %w", EnvServiceAccountJSON, err)
	}
	var missing []string
	if sa.Type != "service_account" {
		missing = append(missing, `type="service_account"`)
	}
	if sa.ProjectID == "" {
		missing = append(missing, "project_id")
	}
	if sa.PrivateKey == "" {
		missing = append(missing, "private_key")
	}
	if sa.ClientEmail == "" {
		missing = append(missing, "client_email")
	}
	if len(missing) > 0 {
		return sa, fmt.Errorf("%s is not a service account key, missing %s", EnvServiceAccountJSON, strings.Join(missing, ", "))
	}
	return sa, nil
}

// SendToTokens sends msg to every token (in batches of at most 500) and
// reports a status per token. A non-nil error means the whole call failed;
// results may then be partial.
func (g *FCMGateway) SendToTokens(ctx context.Context, tokens []string, msg Message) ([]TokenResult, error) {
	results := make([]TokenResult, 0, len(tokens))
	for start := 0; start < len(tokens); start += maxTokensPerBatch {
		end := min(start+maxTokensPerBatch, len(tokens))
		chunk := tokens[start:end]

		resp, err := g.client.SendEachForMulticast(ctx, buildMulticast(chunk, msg))
		if err != nil {
			return results, err
		}
		if len(resp.Responses) != len(chunk) {
			return results, fmt.Errorf("fcm returned %d responses for %d tokens", len(resp.Responses), len(chunk))
		}
		for i, r := range resp.Responses {
			results = append(results, TokenResult{Token: chunk[i], Status: classify(r), Err: r.Error})
		}
	}
	return results, nil
}

func classify(r *messaging.SendResponse) TokenStatus {
	switch {
	case r.Success:
		return StatusSent
	case messaging.IsUnregistered(r.Error):
		return StatusUnregistered
	case messaging.IsInvalidArgument(r.Error):
		return StatusInvalidArgument
	default:
		return StatusFailed
	}
}

func buildMulticast(tokens []string, msg Message) *messaging.MulticastMessage {
	return &messaging.MulticastMessage{
		Tokens:       tokens,
		Notification: &messaging.Notification{Title: msg.Title, Body: msg.Body},
		Data:         msg.Data,
		Android: &messaging.AndroidConfig{
			Priority:     "high",
			Notification: &messaging.AndroidNotification{Sound: soundDefault},
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{"apns-priority": "10"},
			Payload: &messaging.APNSPayload{Aps: &messaging.Aps{
				Alert: &messaging.ApsAlert{Title: msg.Title, Body: msg.Body},
				Sound: soundDefault,
			}},
		},
	}
}
