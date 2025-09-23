package handlers

import (
	"encoding/json"
	"net/http"
	serviceresponse "sen1or/letslive/auth/response"
)

type ResponseHandler struct{}

func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

func (h ResponseHandler) WriteResponse(w http.ResponseWriter, res *serviceresponse.Response[any]) {
	res.Id = w.Header().Get("requestId")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(res.StatusCode)
	json.NewEncoder(w).Encode(res)
}

func (h ResponseHandler) RouteNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	h.WriteResponse(w, serviceresponse.NewResponseFromTemplate[any](serviceresponse.RES_ERR_ROUTE_NOT_FOUND, nil, nil, nil))
}
