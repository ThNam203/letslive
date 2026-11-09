package vod

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/shared/pkg/tracer"
	"sen1or/letslive/vod/domains"
	response "sen1or/letslive/vod/response"
	"time"

	"github.com/gofrs/uuid/v5"
)

type CreateVODInternalRequest struct {
	LivestreamId string `json:"livestreamId"`
	UserId       string `json:"userId"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	PlaybackURL  string `json:"playbackUrl,omitempty"`
	Duration     int64  `json:"duration"`
	// Status is "processing" while the recording is still being written
	// (it is marked ready later through the status endpoint); empty means ready.
	Status string `json:"status,omitempty"`
}

// createVODInternalStatus maps the requested status of a livestream VOD:
// empty or "ready", or "processing". Anything else is rejected.
func createVODInternalStatus(raw string) (domains.VODStatus, bool) {
	switch domains.VODStatus(raw) {
	case "", domains.VODStatusReady:
		return domains.VODStatusReady, true
	case domains.VODStatusProcessing:
		return domains.VODStatusProcessing, true
	default:
		return "", false
	}
}

func (h *VODHandler) CreateVODInternalHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	defer r.Body.Close()

	var reqBody CreateVODInternalRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	status, ok := createVODInternalStatus(reqBody.Status)
	if !ok {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	userId, err := uuid.FromString(reqBody.UserId)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	livestreamId, err := uuid.FromString(reqBody.LivestreamId)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_INPUT, nil, nil, nil))
		return
	}

	now := time.Now()
	desc := &reqBody.Description
	thumbURL := &reqBody.ThumbnailURL
	playURL := &reqBody.PlaybackURL

	if reqBody.Description == "" {
		desc = nil
	}
	if reqBody.ThumbnailURL == "" {
		thumbURL = nil
	}
	if reqBody.PlaybackURL == "" {
		playURL = nil
	}

	vodData := domains.VOD{
		LivestreamId: &livestreamId,
		UserId:       userId,
		Title:        reqBody.Title,
		Description:  desc,
		ThumbnailURL: thumbURL,
		PlaybackURL:  playURL,
		Visibility:   domains.VODPublicVisibility,
		Status:       status,
		ViewCount:    0,
		Duration:     reqBody.Duration,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	ctx, span := tracer.MyTracer.Start(ctx, "create_vod_internal_handler.vod_service.create")
	createdVOD, serviceErr := h.vodService.Create(ctx, vodData)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, response.FromError(serviceErr))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, createdVOD, nil, nil))
}
