package utils

import (
	"errors"
	"mime/multipart"
	"net/http"
	"sen1or/letslive/user/response"
)

const (
	// matches FILE_SIZE_LIMIT_MB_UNIT on the web client, which checks file.size
	maxUploadFileBytes     = 10 * 1024 * 1024
	multipartOverheadBytes = 64 * 1024
)

// ParseUploadForm parses a multipart body carrying one file of up to
// maxUploadFileBytes, with room left for the boundaries and small text fields.
func ParseUploadForm(w http.ResponseWriter, r *http.Request) *response.Response[any] {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadFileBytes+multipartOverheadBytes)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return response.NewResponseFromTemplate[any](response.RES_ERR_IMAGE_TOO_LARGE, nil, nil, nil)
		}
		return response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil)
	}

	return nil
}

// UploadedFile returns a file part of a form parsed by ParseUploadForm.
func UploadedFile(r *http.Request, field string) (multipart.File, *multipart.FileHeader, *response.Response[any]) {
	file, fileHeader, err := r.FormFile(field)
	if err != nil {
		return nil, nil, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_PAYLOAD, nil, nil, nil)
	}

	if fileHeader.Size > maxUploadFileBytes {
		file.Close()
		return nil, nil, response.NewResponseFromTemplate[any](response.RES_ERR_IMAGE_TOO_LARGE, nil, nil, nil)
	}

	return file, fileHeader, nil
}

func ParseUploadedFile(w http.ResponseWriter, r *http.Request, field string) (multipart.File, *multipart.FileHeader, *response.Response[any]) {
	if errRes := ParseUploadForm(w, r); errRes != nil {
		return nil, nil, errRes
	}
	return UploadedFile(r, field)
}
