package user

import (
	"context"
	"net/http"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func (h *UserHandler) GetUserByStreamAPIKeyInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	streamAPIKeyString := r.URL.Query().Get("streamAPIKey")
	if len(streamAPIKeyString) == 0 {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}

	streamAPIKey, err := uuid.FromString(streamAPIKeyString)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_user_by_stream_api_key_internal_handler.user_service.get_user_by_stream_api_key")
	user, sErr := h.userService.GetUserByStreamAPIKey(ctx, streamAPIKey)
	span.End()
	if sErr != nil {
		h.WriteResponse(w, ctx, sErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, user, nil, nil))
}
