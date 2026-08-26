package user

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/response"
	"sen1or/letslive/user/utils"
)

func (h *UserHandler) GetUsersStatusesInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var reqBody dto.GetUsersStatusesRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	if err := utils.Validator.Struct(&reqBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseWithValidationErrors[any](nil, nil, err))
		return
	}

	statuses, sErr := h.userService.GetUsersStatuses(ctx, reqBody.UserIds)
	if sErr != nil {
		h.WriteResponse(w, ctx, sErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(
		response.RES_SUCC_OK,
		&dto.GetUsersStatusesResponseDTO{Statuses: statuses},
		nil,
		nil,
	))
}
