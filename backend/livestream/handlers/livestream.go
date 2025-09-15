package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/pkg/tracer"
	serviceresponse "sen1or/letslive/livestream/responses"
	"sen1or/letslive/livestream/services/livestream"

	"github.com/gofrs/uuid/v5"
)

type LivestreamHandler struct {
	ResponseHandler
	livestreamService *livestream.LivestreamService
}

func NewLivestreamHandler(livestreamService *livestream.LivestreamService) *LivestreamHandler {
	return &LivestreamHandler{
		livestreamService: livestreamService,
	}
}

func (h LivestreamHandler) GetLivestreamOfUserPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userId := r.URL.Query().Get("userId")
	if len(userId) == 0 {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidPath)
		return
	}

	userUUID, err := uuid.FromString(userId)
	if err != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidInput)
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_livestream_of_user_public_handler.livestream_service.get_livestream_of_user")
	vod, serviceErr := h.livestreamService.GetLivestreamOfUser(ctx, userUUID)
	span.End()

	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vod)
}

func (h *LivestreamHandler) CreateLivestreamInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	var body dto.CreateLivestreamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidPayload)
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "create_livestream_internal_handler.livestream_service.create")
	createdLivestream, err := h.livestreamService.Create(ctx, body)
	span.End()

	if err != nil {
		h.WriteErrorResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createdLivestream)
}

func (h *LivestreamHandler) GetRecommendedLivestreamsPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	page, limit := getPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_recommended_livestreams_public_handler.livestream_service.get_recommended_livestreams")
	livestreams, serviceErr := h.livestreamService.GetRecommendedLivestreams(ctx, page, limit)
	span.End()

	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(livestreams)
}

func (h *LivestreamHandler) EndLivestreamAndCreateVODInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	rawStreamId := r.PathValue("livestreamId")
	streamId, err := uuid.FromString(rawStreamId)
	if err != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidInput)
		return
	}
	defer r.Body.Close()

	var requestBody dto.EndLivestreamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidPayload)
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "end_livestream_and_create_vod_internal_handler.livestream_service.end_livestream_and_create_vod")
	serviceErr := h.livestreamService.EndLivestreamAndCreateVOD(ctx, streamId, requestBody)
	span.End()

	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *LivestreamHandler) UpdateLivestreamPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userUUID, e := getUserIdFromCookie(r)
	if e != nil {
		h.WriteErrorResponse(w, e)
		return
	}

	rawStreamId := r.PathValue("livestreamId")
	streamId, err := uuid.FromString(rawStreamId)
	if err != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidInput)
		return
	}
	defer r.Body.Close()

	var requestBody dto.UpdateLivestreamRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidPayload)
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "update_livestream_private_handler.livestream_service.update")
	updatedLivestream, serviceErr := h.livestreamService.Update(ctx, requestBody, streamId, *userUUID)
	span.End()

	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedLivestream)
}

//func (h *LivestreamHandler) DeleteLivestreamPrivateHandler(w http.ResponseWriter, r *http.Request) {
//	ctx, cancelCtx := context.WithCancel(r.Context())
//	defer cancelCtx()
//
//	rawStreamId := r.PathValue("livestreamId")
//	streamId, err := uuid.FromString(rawStreamId)
//
//	if err != nil {
//		h.WriteErrorResponse(w, serviceresponse.ErrInvalidPath)
//		return
//	}
//
//	userUUID, cErr := getUserIdFromCookie(r)
//	if cErr != nil {
//		h.WriteErrorResponse(w, cErr)
//		return
//	}
//
//	serviceErr := h.livestreamService.Delete(ctx, streamId, *userUUID)
//	if serviceErr != nil {
//		h.WriteErrorResponse(w, serviceErr)
//		return
//	}
//
//	w.WriteHeader(http.StatusNoContent)
//}
