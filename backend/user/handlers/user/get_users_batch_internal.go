package user

import (
	"context"
	"encoding/json"
	"net/http"

	"sen1or/letslive/shared/pkg/tracer"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/response"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
)

func (h *UserHandler) GetUsersBatchInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var req dto.GetUsersBatchInternalRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	ids := make([]uuid.UUID, 0, len(req.Ids))
	for _, rawId := range req.Ids {
		id, err := uuid.FromString(rawId)
		if err != nil {
			h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
			return
		}
		ids = append(ids, id)
	}

	ctx, span := tracer.MyTracer.Start(ctx, "user_handler.get_users_batch_internal")
	identities, serviceErr := h.userService.GetIdentitiesByIds(ctx, ids)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, response.FromError(serviceErr))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &identities, nil, nil))
}
