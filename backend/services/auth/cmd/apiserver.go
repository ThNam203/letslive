package main

import (
	"log"
	"net/http"
	"os"
	"sen1or/lets-live/auth/config"
	"sen1or/lets-live/auth/controllers"

	// TODO: add swagger _ "sen1or/lets-live/auth/docs"
	"sen1or/lets-live/auth/handlers"
	"sen1or/lets-live/auth/middlewares"
	_ "sen1or/lets-live/auth/migrations"
	"sen1or/lets-live/auth/repositories"
	"time"

	"github.com/jackc/pgx/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type APIServer struct {
	logger    *zap.Logger
	dbConn    *pgx.Conn // For raw sql queries
	serverURL string

	authHandler  *handlers.AuthHandler
	errorHandler *handlers.ErrorHandler

	loggingMiddleware middlewares.Middleware
	corsMiddleware    middlewares.Middleware
}

// TODO: make tls usable
func NewAPIServer(dbConn *pgx.Conn, authServerURL string) *APIServer {
	var userRepo = repositories.NewUserRepository(dbConn)
	var refreshTokenRepo = repositories.NewRefreshTokenRepository(dbConn)
	var verifyTokenRepo = repositories.NewVerifyTokenRepo(dbConn)

	var userCtrl = controllers.NewUserController(userRepo)
	var refreshTokenCtrl = controllers.NewRefreshTokenController(refreshTokenRepo)
	var verifyTokenCtrl = controllers.NewVerifyTokenController(verifyTokenRepo)

	var authHandler = handlers.NewAuthHandler(refreshTokenCtrl, userCtrl, verifyTokenCtrl, authServerURL)

	var logger, _ = zap.NewProduction()

	return &APIServer{
		logger:    logger,
		dbConn:    dbConn,
		serverURL: authServerURL,

		authHandler:  authHandler,
		errorHandler: &handlers.ErrorHandler{},

		loggingMiddleware: middlewares.NewLoggingMiddleware(logger),
		corsMiddleware:    middlewares.NewCORSMiddleware(),
	}
}

func (a *APIServer) ListenAndServe(useTLS bool) {
	server := &http.Server{
		Addr:         a.serverURL,
		Handler:      a.getHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if useTLS {
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

	log.Printf("server running on addr: %s", server.Addr)
}

// @title           Let's Live API
// @version         0.1
// @description     The server API

// @contact.name   Nam Huynh
// @contact.email  hthnam203@gmail.com

// @host      localhost:8000
// @BasePath  /v1
func (a *APIServer) getHandler() http.Handler {
	sm := http.NewServeMux()

	sm.HandleFunc("POST /v1/auth/signup", a.authHandler.SignUpHandler)
	sm.HandleFunc("POST /v1/auth/login", a.authHandler.LogInHandler)
	sm.HandleFunc("GET /v1/auth/google", a.authHandler.OAuthGoogleLogin)
	sm.HandleFunc("GET /v1/auth/google/callback", a.authHandler.OAuthGoogleCallBack)
	sm.HandleFunc("GET /v1/auth/verify", a.authHandler.VerifyEmailHandler)

	sm.HandleFunc("GET /v1/swagger", httpSwagger.Handler(
		httpSwagger.URL(a.serverURL+"/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	sm.HandleFunc("GET /", a.errorHandler.RouteNotFoundHandler)

	finalHandler := a.corsMiddleware.GetMiddleware(sm)
	finalHandler = a.loggingMiddleware.GetMiddleware(finalHandler)

	return finalHandler
}
