package user

import (
	"context"
	"net/http"
	"sen1or/letslive/shared/pkg/tracer"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/response"
)

func (h *UserHandler) UploadSingleFileToMinIOHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	defer r.Body.Close()

	file, uploadErr := utils.ParseUploadedFile(w, r, "file", utils.GeneralUploadMaxFileBytes)
	if uploadErr != nil {
		h.WriteResponse(w, ctx, uploadErr)
		return
	}
	defer file.Close()

	ctx, span := tracer.MyTracer.Start(ctx, "upload_single_file_to_min_io_handler.user_service.upload_file_to_min_io")
	savedPath, err := h.userService.UploadFileToMinIO(ctx, file)
	span.End()

	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &savedPath, nil, nil))
}
