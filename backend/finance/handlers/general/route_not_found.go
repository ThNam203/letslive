package general

import (
	"net/http"
	"sen1or/letslive/finance/response"
	sharedresponse "sen1or/letslive/shared/response"
)

func (h *GeneralHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.WriteResponse(w, r.Context(), sharedresponse.NewResponseFromTemplate[any](response.RES_ERR_ROUTE_NOT_FOUND, nil, nil, nil))
}
