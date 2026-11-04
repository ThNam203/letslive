package pushhandler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"sen1or/letslive/user/push"

	"github.com/gofrs/uuid/v5"
)

// recorder is a push.MessageSender that remembers every call.
type recorder struct {
	mu    sync.Mutex
	users []uuid.UUID
	msgs  []push.Message
}

func (r *recorder) SendMessage(_ context.Context, userID uuid.UUID, msg push.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users = append(r.users, userID)
	r.msgs = append(r.msgs, msg)
	return nil
}

const (
	validUserA = "11111111-1111-4111-8111-111111111111"
	validUserB = "22222222-2222-4222-8222-222222222222"
	validData  = `{"type":"dm","referenceId":"abc","actionUrl":"/messages/abc"}`
)

func post(t *testing.T, sender push.MessageSender, body string) *httptest.ResponseRecorder {
	t.Helper()
	h := NewPushHandler(sender)
	req := httptest.NewRequest(http.MethodPost, "/v1/internal/push", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.SendPushInternalHandler(rec, req)
	return rec
}

func TestInternalPushValidatesPayload(t *testing.T) {
	manyIDs := `"` + strings.Repeat(validUserA+`","`, 100) + validUserA + `"`
	oneUser := `["` + validUserA + `"]`
	cases := map[string]string{
		"bad uuid":         `{"userIds":["nope"],"title":"t","body":"b","data":` + validData + `}`,
		"no recipients":    `{"userIds":[],"title":"t","body":"b","data":` + validData + `}`,
		"too many ids":     `{"userIds":[` + manyIDs + `],"title":"t","body":"b","data":` + validData + `}`,
		"long title":       `{"userIds":` + oneUser + `,"title":"` + strings.Repeat("a", 101) + `","body":"b","data":` + validData + `}`,
		"long body":        `{"userIds":` + oneUser + `,"title":"t","body":"` + strings.Repeat("a", 501) + `","data":` + validData + `}`,
		"unknown data key": `{"userIds":` + oneUser + `,"title":"t","body":"b","data":{"type":"dm","evil":"x"}}`,
		"not json":         `{`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			rec := &recorder{}
			res := post(t, rec, body)
			if res.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (%s)", res.Code, res.Body.String())
			}
			if len(rec.users) != 0 {
				t.Fatalf("sent %d pushes for an invalid payload", len(rec.users))
			}
		})
	}
}

func TestInternalPushCallsSendMessageOncePerRecipient(t *testing.T) {
	rec := &recorder{}
	body := `{"userIds":["` + validUserA + `","` + validUserB + `"],"title":"alice","body":"hi","data":` + validData + `}`

	res := post(t, rec, body)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", res.Code, res.Body.String())
	}
	if len(rec.users) != 2 || rec.users[0].String() != validUserA || rec.users[1].String() != validUserB {
		t.Fatalf("recipients = %v", rec.users)
	}
	for _, got := range rec.msgs {
		if got.Title != "alice" || got.Body != "hi" || len(got.Data) != 3 ||
			got.Data[push.DataKeyType] != "dm" ||
			got.Data[push.DataKeyReferenceID] != "abc" ||
			got.Data[push.DataKeyActionURL] != "/messages/abc" {
			t.Fatalf("message = %+v", got)
		}
	}
}
