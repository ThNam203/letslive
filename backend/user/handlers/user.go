package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sen1or/lets-live/user/dto"
	livestreamgateway "sen1or/lets-live/user/gateway/livestream"
	"sen1or/lets-live/user/repositories"
	"sen1or/lets-live/user/services"
	"sen1or/lets-live/user/types"
	"sen1or/lets-live/user/utils"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

type UserHandler struct {
	ErrorHandler
	minioService      *services.MinIOService
	ctrl              services.UserController
	livestreamGateway livestreamgateway.LivestreamGateway
}

func NewUserHandler(ctrl services.UserController, livestreamGateway livestreamgateway.LivestreamGateway, minioService *services.MinIOService) *UserHandler {
	return &UserHandler{
		ctrl:              ctrl,
		livestreamGateway: livestreamGateway,
		minioService:      minioService,
	}
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("id")
	if len(userId) == 0 {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("missing user id"))
		return
	}

	userUUID, err := uuid.FromString(userId)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("userId not valid"))
		return
	}

	user, err := h.ctrl.GetById(userUUID)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.New("user not found"))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	userVODs, errRes := h.livestreamGateway.GetUserLivestreams(context.Background(), userId)
	if errRes != nil {
		h.WriteErrorResponse(w, errRes.StatusCode, errors.New(errRes.Message))
		return
	}

	user.VODs = userVODs

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.ctrl.GetAll()
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	for _, user := range users {
		userVODs, errRes := h.livestreamGateway.GetUserLivestreams(context.Background(), user.Id.String())
		if errRes != nil {
			continue // what should be done?
		}

		user.VODs = userVODs
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetUserByQueries(w http.ResponseWriter, r *http.Request) {
	isOnline := r.URL.Query().Get("isOnline")

	if len(isOnline) == 0 {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("missing query parameters"))
		return
	}

	users, err := h.ctrl.GetStreamingUsers()
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// internal
func (h *UserHandler) GetUserByStreamAPIKey(w http.ResponseWriter, r *http.Request) {
	streamAPIKeyString := r.URL.Query().Get("streamAPIKey")

	if len(streamAPIKeyString) == 0 {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("missing stream api key"))
		return
	}

	streamAPIKey, err := uuid.FromString(streamAPIKeyString)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("stream api key not valid"))
		return
	}

	user, err := h.ctrl.GetByStreamAPIKey(streamAPIKey)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("user not found for stream key - %s", streamAPIKey))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetCurrentUserInfo(w http.ResponseWriter, r *http.Request) {
	userUUID, err := getUserIdFromCookie(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}
	user, err := h.ctrl.GetUserFullProfile(*userUUID)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.New("user not found"))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body dto.CreateUserRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error decoding request body: %s", err.Error()))
		return
	}

	if err := utils.Validator.Struct(&body); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error validating payload: %s", err))
		return
	}

	createdUser, err := h.ctrl.Create(body)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createdUser)
}

// INTERNAL
func (h *UserHandler) SetUserVerified(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	if len(userId) == 0 {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("missing user id"))
		return
	}

	userUUID, err := uuid.FromString(userId)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("user id not valid"))
		return
	}

	h.ctrl.UpdateUserVerified(userUUID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	userUUID, err := getUserIdFromCookie(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}
	defer r.Body.Close()

	var requestBody dto.UpdateUserRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error decoding request body: %s", err.Error()))
		return
	}

	requestBody.Id = uuid.FromStringOrNil(userUUID.String())

	if err := utils.Validator.Struct(&requestBody); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error validating payload: %s", err))
		return
	}

	updatedUser, err := h.ctrl.Update(requestBody)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.New("user not found"))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}

func (h *UserHandler) GenerateNewAPIStreamKey(w http.ResponseWriter, r *http.Request) {
	userUUID, err := getUserIdFromCookie(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}
	defer r.Body.Close()

	newKey, err := h.ctrl.UpdateStreamAPIKey(*userUUID)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.New("user not found"))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(newKey))
}

func (h *UserHandler) UpdateUserProfilePicture(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 10 * 1024 * 1024
	userUUID, err := getUserIdFromCookie(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}
	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxByteError *http.MaxBytesError
		if errors.As(err, &maxByteError) {
			h.WriteErrorResponse(w, http.StatusRequestEntityTooLarge, err)
			return
		}

		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error decoding request body: %s", err.Error()))
		return
	}

	file, fileHeader, err := r.FormFile("profile-picture")
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	savedPath, err := h.minioService.AddFile(file, fileHeader, "profile-pictures")
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("failed to save the picture: %s", savedPath))
		return
	}

	err = h.ctrl.UpdateProfilePicture(*userUUID, savedPath)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.New("user not found"))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(savedPath))
}

func (h *UserHandler) UpdateUserBackgroundPicture(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 10 * 1024 * 1024
	userUUID, err := getUserIdFromCookie(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}
	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxByteError *http.MaxBytesError
		if errors.As(err, &maxByteError) {
			h.WriteErrorResponse(w, http.StatusRequestEntityTooLarge, err)
			return
		}

		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error parsing request body: %s", err.Error()))
		return
	}

	file, fileHeader, err := r.FormFile("background-picture")
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	savedPath, err := h.minioService.AddFile(file, fileHeader, "background-pictures")
	if err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("failed to save the picture: %s", savedPath))
		return

	}

	err = h.ctrl.UpdateBackgroundPicture(*userUUID, savedPath)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.New("user not found"))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(savedPath))
}

func getUserIdFromCookie(r *http.Request) (*uuid.UUID, error) {
	accessTokenCookie, err := r.Cookie("ACCESS_TOKEN")
	if err != nil || len(accessTokenCookie.Value) == 0 {
		return nil, errors.New("missing credentials")
	}

	myClaims := types.MyClaims{}

	// the signature should already been checked from the api gateway before going to this
	_, _, err = jwt.NewParser().ParseUnverified(accessTokenCookie.Value, &myClaims)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %s", err)
	}

	userUUID, err := uuid.FromString(myClaims.UserId)
	if err != nil {
		return nil, errors.New("userId not valid")
	}

	return &userUUID, nil
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	defer r.Body.Close()

	if len(userID) == 0 {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("missing user id"))
		return
	}

	var requestBody dto.UpdateUserRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error decoding request body: %s", err.Error()))
		return
	}
	requestBody.Id = uuid.FromStringOrNil(userID)

	if err := utils.Validator.Struct(&requestBody); err != nil {
		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error validating payload: %s", err))
		return
	}

	updatedUser, err := h.ctrl.Update(requestBody)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.New("user not found"))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}
