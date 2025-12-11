package general

import (
	"net/http"
)

func (h *GeneralHandler) RouteServiceHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
