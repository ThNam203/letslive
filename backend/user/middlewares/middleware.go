package middlewares

import "net/http"

type Middleware interface {
	GetMiddleware(http.Handler) http.Handler
}
