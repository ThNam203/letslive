package response

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

type Meta struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
	Total    int `json:"total,omitempty"`
}

type ErrorDetail map[string]any

type ErrorDetails []ErrorDetail

type Response[T any] struct {
	RequestId    string        `json:"requestId"`
	Success      bool          `json:"success"`
	StatusCode   int           `json:"-"`
	Code         int           `json:"code"`
	Key          string        `json:"key"`
	Message      string        `json:"message"`
	Data         *T            `json:"data,omitempty"`
	Meta         *Meta         `json:"meta,omitempty"`
	ErrorDetails *ErrorDetails `json:"errorDetails,omitempty"`
}

type ResponseTemplate struct {
	Success    bool
	StatusCode int
	Code       int
	Key        string
	Message    string
}

func NewResponseFromTemplate[T any](tpl ResponseTemplate, data *T, meta *Meta, errorDetails *ErrorDetails) *Response[T] {
	return &Response[T]{
		Success:      tpl.Success,
		StatusCode:   tpl.StatusCode,
		Code:         tpl.Code,
		Key:          tpl.Key,
		Message:      tpl.Message,
		Data:         data,
		Meta:         meta,
		ErrorDetails: errorDetails,
	}
}

func NewResponseWithValidationErrors[T any](data *T, meta *Meta, validateError error) *Response[T] {
	errorDetails := ErrorDetails{}
	var validateErrors validator.ValidationErrors
	if errors.As(validateError, &validateErrors) {
		for _, e := range validateErrors {
			errorDetails = append(errorDetails, ErrorDetail{
				"Field": e.Field(),
				"Tag":   e.Tag(),
				"Param": e.Param(),
			})
		}
	}
	return &Response[T]{
		Success:      RES_ERR_INVALID_INPUT.Success,
		StatusCode:   RES_ERR_INVALID_INPUT.StatusCode,
		Code:         RES_ERR_INVALID_INPUT.Code,
		Key:          RES_ERR_INVALID_INPUT.Key,
		Message:      RES_ERR_INVALID_INPUT.Message,
		Data:         data,
		Meta:         meta,
		ErrorDetails: &errorDetails,
	}
}

func NewResponse[T any](success bool, statusCode int, code int, key string, message string, data *T, meta *Meta, errorDetails *ErrorDetails) *Response[T] {
	return &Response[T]{
		Success:      success,
		StatusCode:   statusCode,
		Code:         code,
		Key:          key,
		Message:      message,
		Data:         data,
		Meta:         meta,
		ErrorDetails: errorDetails,
	}
}
