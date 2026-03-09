package vod

import (
	"context"
	"net/http"
	"sen1or/letslive/livestream/handlers/utils"
	"sen1or/letslive/livestream/pkg/tracer"
	response "sen1or/letslive/livestream/response"
)

const maxUploadSize = 2 << 30 // 2GB

func (h *VODHandler) UploadVODPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userId, err := utils.GetUserIdFromCookie(r)
	if err != nil {
		h.WriteResponse(w, ctx, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if parseErr := r.ParseMultipartForm(32 << 20); parseErr != nil { // 32MB memory buffer
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}

	file, header, fileErr := r.FormFile("file")
	if fileErr != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil))
		return
	}
	defer file.Close()

	title := r.FormValue("title")
	if title == "" {
		title = header.Filename
	}
	description := r.FormValue("description")
	visibility := r.FormValue("visibility")
	if visibility == "" {
		visibility = "public"
	}

	ctx, span := tracer.MyTracer.Start(ctx, "upload_vod_private_handler.vod_service.upload_vod")
	createdVOD, serviceErr := h.vodService.UploadVOD(ctx, *userId, title, description, visibility, header.Filename, header.Size, file)
	span.End()

	if serviceErr != nil {
		h.WriteResponse(w, ctx, serviceErr)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, createdVOD, nil, nil))
}
