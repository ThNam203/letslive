package utils

import (
	"errors"
	"mime/multipart"
	"net/http"
	"sen1or/letslive/user/response"
)

const (
	// matches FILE_SIZE_LIMIT_MB_UNIT on the web client, which checks file.size
	MaxUploadFileBytes     = 10 * 1024 * 1024
	multipartOverheadBytes = 64 * 1024
)

func ParseUploadedFile(w http.ResponseWriter, r *http.Request, field string) (multipart.File, *multipart.FileHeader, *response.Response[any]) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadFileBytes+multipartOverheadBytes)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return nil, nil, response.NewResponseFromTemplate[any](response.RES_ERR_IMAGE_TOO_LARGE, nil, nil, nil)
		}
		return nil, nil, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil)
	}

	file, fileHeader, err := r.FormFile(field)
	if err != nil {
		return nil, nil, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil)
	}

	if fileHeader.Size > MaxUploadFileBytes {
		file.Close()
		return nil, nil, response.NewResponseFromTemplate[any](response.RES_ERR_IMAGE_TOO_LARGE, nil, nil, nil)
	}

	return file, fileHeader, nil
}
