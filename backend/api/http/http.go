package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sen1or/lets-live/api/config"
	"sen1or/lets-live/api/domains"
	"sen1or/lets-live/api/repositories"
	_ "sen1or/lets-live/docs"
	"strings"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type APIServer struct {
	logger    *zap.Logger
	db        *gorm.DB // For raw sql queries
	serverURL string

	userRepo         domains.UserRepository
	refreshTokenRepo domains.RefreshTokenRepository
	verifyTokenRepo  domains.VerifyTokenRepository
}

// TODO: make tls usable
func NewAPIServer(dbConn gorm.DB) *APIServer {
	var userRepo = repositories.NewUserRepository(dbConn)
	var refreshTokenRepo = repositories.NewRefreshTokenRepository(dbConn)
	var verifyTokenRepo = repositories.NewVerifyTokenRepo(dbConn)
	var logger, _ = zap.NewProduction()

	return &APIServer{
		logger:    logger,
		db:        &dbConn,
		serverURL: "http://localhost:8000",

		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		verifyTokenRepo:  verifyTokenRepo,
	}
}

func (a *APIServer) ListenAndServeTLS() {
	server := &http.Server{
		Addr:         ":8000",
		Handler:      a.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if strings.Contains(a.serverURL, "https") {
		if _, err := os.Stat(config.SERVER_CRT_FILE); err != nil {
			log.Panic("error loading server cert file", err.Error())
		}

		if _, err := os.Stat(config.SERVER_KEY_FILE); err != nil {
			log.Panic("error loading server key file", err.Error())
		}

		go (func() {
			log.Panic("server ends: ", server.ListenAndServeTLS(config.SERVER_CRT_FILE, config.SERVER_KEY_FILE))
		})()
	} else {
		go (func() {
			log.Panic("server ends: ", server.ListenAndServe())
		})()
	}

	log.Println("server api started")
}

// @title           Let's Live API
// @version         0.1
// @description     The server API

// @contact.name   Nam Huynh
// @contact.email  hthnam203@gmail.com

// @host      localhost:8000
// @BasePath  /v1
func (a *APIServer) Routes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/v1/users/{id}", a.GetUserByIdHandler).Methods(http.MethodGet)

	router.HandleFunc("/v1/streams/{apiKey}/online", a.SetUserStreamOnline).Methods(http.MethodPatch)
	router.HandleFunc("/v1/streams/{apiKey}/offline", a.SetUserStreamOffline).Methods(http.MethodPatch)
	router.HandleFunc("/v1/streams", a.GetOnlineStreams).Methods(http.MethodGet)

	router.HandleFunc("/v1/auth/signup", a.SignUpHandler).Methods(http.MethodPost)
	router.HandleFunc("/v1/auth/login", a.LogInHandler).Methods(http.MethodPost)
	router.HandleFunc("/v1/auth/google", a.OAuthGoogleLogin).Methods(http.MethodGet)
	router.HandleFunc("/v1/auth/google/callback", a.OAuthGoogleCallBack).Methods(http.MethodGet)
	router.HandleFunc("/v1/auth/verify", a.verifyEmailHandler).Methods(http.MethodGet)

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(a.serverURL+"/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)

	router.PathPrefix("/").HandlerFunc(a.RouteNotFound)

	router.Use(a.loggingMiddleware)
	router.Use(a.corsMiddleware)

	return router
}

func (a *APIServer) RouteNotFound(w http.ResponseWriter, r *http.Request) {
	a.errorResponse(w, http.StatusNotFound, fmt.Errorf("route not found"))
}

// Set the error to the custom "X-LetsLive-Error" header
// The function doesn't end the request, if so call errorResponse
func (a *APIServer) setError(w http.ResponseWriter, err error) {
	w.Header().Add("X-LetsLive-Error", err.Error())
}

type HTTPErrorResponse struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"internal server error"`
}

// Set error to the custom header and write the error to the request
// After calling, the request will end and no other write should be done
func (a *APIServer) errorResponse(w http.ResponseWriter, status int, err error) {
	w.Header().Add("X-LetsLive-Error", err.Error())
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(HTTPErrorResponse{
		Message: err.Error(),
	})
}

func (a *APIServer) setTokens(w http.ResponseWriter, refreshToken string, accessToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:  "refreshToken",
		Value: refreshToken,

		Expires:  time.Now().Add(config.REFRESH_TOKEN_EXPIRES_DURATION),
		MaxAge:   config.REFRESH_TOKEN_MAX_AGE,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:  "accessToken",
		Value: accessToken,

		Expires:  time.Now().Add(config.ACCESS_TOKEN_EXPIRES_DURATION),
		MaxAge:   config.ACCESS_TOKEN_MAX_AGE,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	})
}

func (a *APIServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5000")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

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

func (a *APIServer) loggingMiddleware(next http.Handler) http.Handler {
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

		if lrw.statusCode/100 == 2 {
			a.logger.Info("success api call", fields...)
		} else {
			err := lrw.w.Header().Get("X-LetsLive-Error")
			if len(err) == 0 {
				a.logger.Info("failed api call", fields...)
			} else {
				a.logger.Error("failed api call: "+err, fields...)

			}
		}

		// TODO: prometheus
	})
}
