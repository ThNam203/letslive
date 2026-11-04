package pushhandler

import (
	"encoding/json"
	"net/http"

	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/push"
	"sen1or/letslive/user/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
)

// SendPushInternalHandler handles POST /v1/internal/push. It is internal only:
// envoy has no route for it, so only services on the cluster network reach it.
// Delivery is best-effort, so a per-recipient failure is logged and the
// request still succeeds once the payload is accepted.
func (h *PushHandler) SendPushInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.SendPushInternalRequestDTO
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // rejects data keys other than the three allowed
	if err := decoder.Decode(&req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}
	if err := validator.New().Struct(req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	msg := push.Message{Title: req.Title, Body: req.Body, Data: dataMap(req.Data)}
	for _, id := range req.UserIds {
		userID, err := uuid.FromString(id) // already validated as a uuid
		if err != nil {
			continue
		}
		if err := h.sender.SendMessage(ctx, userID, msg); err != nil {
			logger.Errorf(ctx, "failed to enqueue push for user %s: %v", userID, err)
		}
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_SUCC_OK, nil, nil, nil))
}

// dataMap converts the DTO to FCM data, omitting empty values.
func dataMap(d dto.PushDataDTO) map[string]string {
	data := map[string]string{push.DataKeyType: d.Type}
	if d.ReferenceID != "" {
		data[push.DataKeyReferenceID] = d.ReferenceID
	}
	if d.ActionURL != "" {
		data[push.DataKeyActionURL] = d.ActionURL
	}
	return data
}
