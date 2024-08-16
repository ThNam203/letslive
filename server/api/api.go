package api

import (
	"log"
	"net"
	"net/http"
	"sen1or/lets-live/server/config"
	"sen1or/lets-live/server/domain"
	"sen1or/lets-live/server/repository"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type api struct {
	logger *zap.Logger
	db     *gorm.DB // For raw sql queries

	userRepo         domain.UserRepository
	refreshTokenRepo domain.RefreshTokenRepository
}

func NewApi(dbConn gorm.DB) *api {
	var userRepo = repository.NewUserRepository(dbConn)
	var refreshTokenRepo = repository.NewRefreshTokenRepository(dbConn)
	var logger, _ = zap.NewProduction()

	return &api{
		logger: logger,
		db:     &dbConn,

		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (a *api) ListenAndServe() {
	server := &http.Server{
		Addr:         ":8000",
		Handler:      a.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("server ending: ", server.ListenAndServe())
}

func (a *api) Routes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/v1/users/{id}", a.GetUserByIdHandler).Methods("GET")

	router.HandleFunc("/v1/auth/signup", a.SignUpHandler).Methods("POST")
	router.HandleFunc("/v1/auth/login", a.LogInHandler).Methods("POST")
	router.HandleFunc("/v1/auth/google", a.OAuthGoogleLogin).Methods("GET")
	router.HandleFunc("/v1/auth/google/callback", a.OAuthGoogleCallBack).Methods("GET")

	router.Use(a.corsMiddleware)
	router.Use(a.loggingMiddleware)

	return router
}

// Set the error to the custom "X-LetsLive-Error" header
// The function doesn't end the request, if so call errorResponse
func (a *api) setError(w http.ResponseWriter, err error) {
	w.Header().Add("X-LetsLive-Error", err.Error())
}

// Set error to the custom header and write the error to the request
// After calling, the request will end and no other write should be done
func (a *api) errorResponse(w http.ResponseWriter, status int, err error) {
	w.Header().Add("X-LetsLive-Error", err.Error())
	http.Error(w, err.Error(), status)
}

func (a *api) setTokens(w http.ResponseWriter, refreshToken string, accessToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:  "refreshToken",
		Value: refreshToken,

		Expires:  time.Now().Add(config.RefreshTokenExpiresDuration),
		MaxAge:   config.RefreshTokenMaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:  "accessToken",
		Value: accessToken,

		Expires:  time.Now().Add(config.AccessTokenExpiresDuration),
		MaxAge:   config.AccessTokenMaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	})
}

func (a *api) addGlobalMiddlewares(r *http.ServeMux, middlewares ...func(next http.Handler) http.Handler) http.Handler {
	var wrappedHandler http.Handler
	wrappedHandler = r

	for _, mw := range middlewares {
		wrappedHandler = mw(wrappedHandler)
	}

	return wrappedHandler
}

func (a *api) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5000")
		next.ServeHTTP(w, r)
	})
}

type LoggingResponseWriter struct {
	w          http.ResponseWriter
	statusCode int
	bytes      int
}

func (lrw *LoggingResponseWriter) Header() http.Header {
	return lrw.w.Header()
}

func (lrw *LoggingResponseWriter) Write(data []byte) (int, error) {
	wb, err := lrw.w.Write(data)
	lrw.bytes += wb
	return wb, err
}

func (lrw *LoggingResponseWriter) WriteHeader(statusCode int) {
	lrw.w.WriteHeader(statusCode)
	lrw.statusCode = statusCode
}

func (a *api) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()
		lrw := &LoggingResponseWriter{w: w}
		next.ServeHTTP(lrw, r)

		duration := time.Since(timeStart).Milliseconds()
		remoteAddr := r.Header.Get("X-Forwarded-For")
		if remoteAddr == "" {
			if ip, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
				remoteAddr = "unknown address"
			} else {
				remoteAddr = ip
			}
		}

		fields := []zap.Field{
			zap.Int64("duration", duration),
			zap.String("method", r.Method),
			zap.String("remote#addr", remoteAddr),
			zap.Int("response#bytes", lrw.bytes),
			zap.Int("response#status", lrw.statusCode),
			zap.String("uri", r.RequestURI),
		}

		if lrw.statusCode == 200 {
			a.logger.Info("Server: ", fields...)
		} else {
			err := lrw.w.Header().Get("X-LetsLive-Error")
			if len(err) == 0 {
				a.logger.Info("Server: ", fields...)
			} else {
				a.logger.Error(err, fields...)

			}
		}

		// TODO: prometheus
	})
}
