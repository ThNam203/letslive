package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sen1or/lets-live/user/controllers"
	"sen1or/lets-live/user/dto"
	transcodegateway "sen1or/lets-live/user/gateway/transcode"
	"sen1or/lets-live/user/repositories"
	"sen1or/lets-live/user/types"
	"sen1or/lets-live/user/utils"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

type UserHandler struct {
	ErrorHandler
	ctrl             controllers.UserController
	transcodeGateway transcodegateway.TranscodeGateway
}

func NewUserHandler(ctrl controllers.UserController, transcodeGateway transcodegateway.TranscodeGateway) *UserHandler {
	return &UserHandler{
		ctrl:             ctrl,
		transcodeGateway: transcodeGateway,
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

	userVODs, errRes := h.transcodeGateway.GetUserVODs(context.Background(), userId)
	if errRes != nil {
		h.WriteErrorResponse(w, errRes.StatusCode, errors.New(errRes.Message))
		return
	}

	user.VODs = userVODs.Data

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
		userVODs, errRes := h.transcodeGateway.GetUserVODs(context.Background(), user.Id.String())
		if errRes != nil {
			continue // what should be done?
		}

		user.VODs = userVODs.Data
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// TODO: refactor to use verifyStreamAPIKey
// get user by using path query '/user?streamAPIKey=123123123'
func (h *UserHandler) GetUserByQueries(w http.ResponseWriter, r *http.Request) {
	streamAPIKeyString := r.URL.Query().Get("streamAPIKey")
	isOnline := r.URL.Query().Get("isOnline")

	if len(streamAPIKeyString) == 0 && len(isOnline) == 0 {
		h.WriteErrorResponse(w, http.StatusBadRequest, errors.New("missing query parameters, at least has one"))
		return
	}

	if len(streamAPIKeyString) > 0 {
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

	} else {
		users, err := h.ctrl.GetStreamingUsers()
		if err != nil {
			h.WriteErrorResponse(w, http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	}
}

func (h *UserHandler) GetCurrentUserInfo(w http.ResponseWriter, r *http.Request) {
	accessTokenCookie, err := r.Cookie("ACCESS_TOKEN")
	if err != nil || len(accessTokenCookie.Value) == 0 {
		h.WriteErrorResponse(w, http.StatusForbidden, errors.New("missing credentials"))
		return
	}

	myClaims := types.MyClaims{}

	// the signature should already been checked from the api gateway before going to this
	_, _, err = jwt.NewParser().ParseUnverified(accessTokenCookie.Value, &myClaims)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusForbidden, fmt.Errorf("invalid access token: %s", err))
		return
	}

	userUUID, err := uuid.FromString(myClaims.UserId)
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

	updatedUser, err := h.ctrl.Create(body)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
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
