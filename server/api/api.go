package api

import (
	"log"
	"net/http"
	"sen1or/lets-live/server/domain"
	"sen1or/lets-live/server/repository"
	"time"

	"gorm.io/gorm"
)

type api struct {
	userRepo domain.UserRepository
}

func NewApi(dbConn gorm.DB) *api {
	var userRepo = repository.NewUserRepository(dbConn)

	return &api{
		userRepo: userRepo,
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

func (a *api) Routes() *http.ServeMux {
	serveMux := http.NewServeMux()

	serveMux.HandleFunc("GET /v1/users/{id}", a.GetUserByIdHandler)

	return serveMux
}

func (a *api) errorResponse(w http.ResponseWriter, status int, err error) {
	w.Header().Add("X-LetsLive-Error", err.Error())
	http.Error(w, err.Error(), status)
}
