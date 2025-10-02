package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/pkg/tracer"
	response "sen1or/letslive/livestream/response"
	"sen1or/letslive/livestream/services/vod"

	"github.com/gofrs/uuid/v5"
)

type VODHandler struct {
	vodService *vod.VODService
}

func NewVODHandler(vodService *vod.VODService) *VODHandler {
	return &VODHandler{
		vodService: vodService,
	}
}

func (h VODHandler) GetVODByIdPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	streamId := r.PathValue("vodId")
	if len(streamId) == 0 {
		WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	vodUUID, err := uuid.FromString(streamId)
	if err != nil {
		WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_vod_by_id_public_handler.vod_service.get_vod_by_id")
	vod, serviceErr := h.vodService.GetVODById(ctx, vodUUID)
	span.End()

	if serviceErr != nil {
		WriteResponse(w, ctx, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vod)
}

func (h VODHandler) GetVODsOfUserPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userId := r.URL.Query().Get("userId")
	if len(userId) == 0 {
		WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	userUUID, err := uuid.FromString(userId)
	if err != nil {
		WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	page, limit := getPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_vods_of_user_public_handler.vod_service.get_public_vods_by_user")
	vods, serviceErr := h.vodService.GetPublicVODsByUser(ctx, userUUID, page, limit)
	span.End()

	if serviceErr != nil {
		WriteResponse(w, ctx, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vods)
}

func (h VODHandler) GetVODsOfAuthorPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userUUID, err := getUserIdFromCookie(r)
	if err != nil {
		WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	page, limit := getPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_vods_of_author_private_handler.vod_service.get_all_vods_by_user")
	livestreams, serviceErr := h.vodService.GetAllVODsByUser(ctx, *userUUID, page, limit)
	span.End()

	if serviceErr != nil {
		WriteResponse(w, ctx, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(livestreams)
}

// TODO: recommendation system
func (h VODHandler) GetRecommendedVODsPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	page, limit := getPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_recommended_vods_public_handler.vod_service.get_recommended_vods")
	vods, serviceErr := h.vodService.GetRecommendedVODs(ctx, page, limit)
	span.End()

	if serviceErr != nil {
		WriteResponse(w, ctx, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vods)
}

func (h VODHandler) UpdateVODMetadataPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userId, err := getUserIdFromCookie(r)
	if err != nil {
		WriteResponse(w, ctx, err)
		return
	}

	rawStreamId := r.PathValue("vodId")
	streamId, er := uuid.FromString(rawStreamId)
	if er != nil {
		WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}
	defer r.Body.Close()

	var requestBody dto.UpdateVODRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "update_vod_metadata_private_handler.vod_service.update_vod_metadata")
	updatedvod, serviceErr := h.vodService.UpdateVODMetadata(ctx, requestBody, streamId, *userId)
	span.End()

	if serviceErr != nil {
		WriteResponse(w, ctx, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedvod)
}

func (h VODHandler) DeleteVODPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	rawVODId := r.PathValue("vodId")
	vodId, err := uuid.FromString(rawVODId)

	if err != nil {
		WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	userUUID, cErr := getUserIdFromCookie(r)
	if cErr != nil {
		WriteResponse(w, ctx, cErr)
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "delete_vod_private_handler.vod_service.delete")
	serviceErr := h.vodService.Delete(ctx, vodId, *userUUID)
	span.End()

	if serviceErr != nil {
		WriteResponse(w, ctx, serviceErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
