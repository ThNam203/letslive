package response

import (
	sharedresponse "sen1or/letslive/shared/response"
)

// Type aliases from shared response package
type Meta = sharedresponse.Meta
type ErrorDetail = sharedresponse.ErrorDetail
type ErrorDetails = sharedresponse.ErrorDetails
type Response[T any] = sharedresponse.Response[T]
type ResponseTemplate = sharedresponse.ResponseTemplate

// NewResponseFromTemplate delegates to the shared implementation.
func NewResponseFromTemplate[T any](tpl ResponseTemplate, data *T, meta *Meta, errorDetails *ErrorDetails) *Response[T] {
	return sharedresponse.NewResponseFromTemplate[T](tpl, data, meta, errorDetails)
}

// NewResponseWithValidationErrors wraps the shared implementation, using
// the vod-specific RES_ERR_INVALID_INPUT template as the default.
func NewResponseWithValidationErrors[T any](data *T, meta *Meta, validateError error) *Response[T] {
	return sharedresponse.NewResponseWithValidationErrors[T](RES_ERR_INVALID_INPUT, data, meta, validateError)
}

// NewResponse delegates to the shared implementation.
func NewResponse[T any](success bool, statusCode int, code int, key string, message string, data *T, meta *Meta, errorDetails *ErrorDetails) *Response[T] {
	return sharedresponse.NewResponse[T](success, statusCode, code, key, message, data, meta, errorDetails)
}
