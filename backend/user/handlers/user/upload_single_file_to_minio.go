package user

import (
	"context"
	"errors"
	"net/http"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
)

func (h *UserHandler) UploadSingleFileToMinIOHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	const maxUploadSize = 10 * 1024 * 1024
	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(0); err != nil {
		var maxByteError *http.MaxBytesError
		if errors.As(err, &maxByteError) {
			h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
				response.RES_ERR_IMAGE_TOO_LARGE,
				nil,
				nil,
				nil,
			))
			return
		}

		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	file, fileHeader, formErr := r.FormFile("file")
	if formErr != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
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
		h.WriteResponse(w, ctx, err)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &savedPath, nil, nil))
}
