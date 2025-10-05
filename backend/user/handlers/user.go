package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
	"sen1or/letslive/user/services"
	"sen1or/letslive/user/types"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetUserByIdPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	authenticatedUserId, _ := getUserIdFromCookie(r)
	userId := r.PathValue("userId")
	if len(userId) == 0 {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		))
		return
	}

	userUUID, err := uuid.FromString(userId)
	if err != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_user_by_id_public_handler.user_service.get_user_public_info_by_id")
	user, serviceErr := h.userService.GetUserPublicInfoById(ctx, userUUID, authenticatedUserId)
	span.End()

	if serviceErr != nil {
		writeResponse(w, ctx, serviceErr)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, user, nil, nil))
}

func (h *UserHandler) GetAllUsersPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	page, err := strconv.Atoi(r.URL.Query().Get("page"))

	if err != nil || page < 0 {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_all_users_public_handler.user_service.get_all_users")
	users, serviceErr := h.userService.GetAllUsers(ctx, page)
	span.End()

	if serviceErr != nil {
		writeResponse(w, ctx, serviceErr)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &users, nil, nil))
}

func (h *UserHandler) SearchUsersPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	authenticatedUserId, _ := getUserIdFromCookie(r)
	username := r.URL.Query().Get("username")

	ctx, span := tracer.MyTracer.Start(ctx, "search_users_public_handler.user_service.search_users_by_username")
	users, err := h.userService.SearchUsersByUsername(ctx, username, authenticatedUserId)
	span.End()

	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &users, nil, nil))
}

func (h *UserHandler) GetUserByStreamAPIKeyInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	streamAPIKeyString := r.URL.Query().Get("streamAPIKey")
	if len(streamAPIKeyString) == 0 {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}

	streamAPIKey, err := uuid.FromString(streamAPIKeyString)
	if err != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_user_by_stream_api_key_internal_handler.user_service.get_user_by_stream_api_key")
	user, sErr := h.userService.GetUserByStreamAPIKey(ctx, streamAPIKey)
	span.End()
	if sErr != nil {
		writeResponse(w, ctx, sErr)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, user, nil, nil))
}

func (h *UserHandler) GetCurrentUserPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userUUID, cookieErr := getUserIdFromCookie(r)
	if cookieErr != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_current_user_private_handler.user_service.get_user_by_id")
	user, err := h.userService.GetUserById(ctx, *userUUID)
	span.End()

	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, user, nil, nil))
}

// INTERNAL
func (h *UserHandler) CreateUserInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var body dto.CreateUserRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "create_user_internal_handler.user_service.create_new_user")
	createdUser, err := h.userService.CreateNewUser(ctx, body)
	span.End()

	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, createdUser, nil, nil))
}

func (h *UserHandler) UpdateCurrentUserPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userUUID, cookieErr := getUserIdFromCookie(r)
	if cookieErr != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}
	defer r.Body.Close()

	var requestBody dto.UpdateUserRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		logger.Errorf(ctx, "failed to decode request body: %s", err)
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "update_current_user_private_handler.user_service.update_user")
	requestBody.Id = uuid.FromStringOrNil(userUUID.String())
	updatedUser, err := h.userService.UpdateUser(ctx, requestBody)
	span.End()

	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, updatedUser, nil, nil))
}

func (h *UserHandler) GenerateNewAPIStreamKeyPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userUUID, cookieErr := getUserIdFromCookie(r)
	if cookieErr != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}
	defer r.Body.Close()

	ctx, span := tracer.MyTracer.Start(ctx, "generate_new_api_stream_key_private_hanlder.user_service.update_user_api_key")
	newKey, err := h.userService.UpdateUserAPIKey(ctx, *userUUID)
	span.End()

	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &newKey, nil, nil))
}

func (h *UserHandler) UpdateUserProfilePicturePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	const maxUploadSize = 10 * 1024 * 1024
	userUUID, cookieErr := getUserIdFromCookie(r)
	if cookieErr != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}
	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxByteError *http.MaxBytesError
		if errors.As(err, &maxByteError) {
			writeResponse(w, ctx, response.NewResponseFromTemplate[any](
				response.RES_ERR_IMAGE_TOO_LARGE,
				nil,
				nil,
				nil,
			))
			return
		}

		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	file, fileHeader, formErr := r.FormFile("profile-picture")
	if formErr != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "update_user_profile_picture_private_handler.user_service.update_user_profile_picture")
	savedPath, err := h.userService.UpdateUserProfilePicture(ctx, file, fileHeader, *userUUID)
	span.End()

	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &savedPath, nil, nil))
}

func (h *UserHandler) UpdateUserBackgroundPicturePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	const maxUploadSize = 10 * 1024 * 1024
	userUUID, cookieErr := getUserIdFromCookie(r)
	if cookieErr != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}
	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxByteError *http.MaxBytesError
		if errors.As(err, &maxByteError) {
			writeResponse(w, ctx, response.NewResponseFromTemplate[any](
				response.RES_ERR_IMAGE_TOO_LARGE,
				nil,
				nil,
				nil,
			))
			return
		}

		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		))
		return
	}

	file, fileHeader, formErr := r.FormFile("background-picture")
	if formErr != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "update_user_background_picture_private_handler.user_service.update_user_background_picture")
	savedPath, err := h.userService.UpdateUserBackgroundPicture(ctx, file, fileHeader, *userUUID)
	span.End()

	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &savedPath, nil, nil))
}

func (h *UserHandler) UpdateUserInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userID := r.PathValue("userId")
	defer r.Body.Close()

	if len(userID) == 0 {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		))
		return
	}

	var requestBody dto.UpdateUserRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}
	requestBody.Id = uuid.FromStringOrNil(userID)

	ctx, span := tracer.MyTracer.Start(ctx, "update_user_internal_handler.user_service.update_user_internal")
	updatedUser, err := h.userService.UpdateUserInternal(ctx, requestBody)
	span.End()

	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, updatedUser, nil, nil))
}

func getUserIdFromCookie(r *http.Request) (*uuid.UUID, *response.Response[any]) {
	accessTokenCookie, err := r.Cookie("ACCESS_TOKEN")
	if err != nil || len(accessTokenCookie.Value) == 0 {
		logger.Debugf(r.Context(), "missing credentials")
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		)
	}

	myClaims := types.MyClaims{}

	// the signature should already been checked from the api gateway before going to this
	_, _, err = jwt.NewParser().ParseUnverified(accessTokenCookie.Value, &myClaims)
	if err != nil {
		logger.Debugf(r.Context(), "invalid access token: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		)
	}

	userUUID, err := uuid.FromString(myClaims.UserId)
	if err != nil {
		logger.Debugf(r.Context(), "userId not valid")
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		)
	}

	return &userUUID, nil
}

func (h *UserHandler) UploadSingleFileToMinIOHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	const maxUploadSize = 10 * 1024 * 1024
	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(0); err != nil {
		var maxByteError *http.MaxBytesError
		if errors.As(err, &maxByteError) {
			writeResponse(w, ctx, response.NewResponseFromTemplate[any](
				response.RES_ERR_IMAGE_TOO_LARGE,
				nil,
				nil,
				nil,
			))
			return
		}

		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	file, fileHeader, formErr := r.FormFile("file")
	if formErr != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "upload_single_file_to_min_io_handler.user_service.upload_file_to_min_io")
	savedPath, err := h.userService.UploadFileToMinIO(ctx, file, fileHeader)
	span.End()

	if err != nil {
		writeResponse(w, ctx, err)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &savedPath, nil, nil))
}
