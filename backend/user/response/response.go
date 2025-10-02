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
	RequestId    string        `json:"requestId"` // can be used as trace_id, well it is trace_id (request_id)
	Success      bool          `json:"success"`
	StatusCode   int           `json:"-"`       // no include
	Code         int           `json:"code"`    // business level
	Key          string        `json:"key"`     // for i18n key
	Message      string        `json:"message"` // default to english, should be handled by frontend so can be ignored
	Data         *T            `json:"data,omitempty"`
	Meta         *Meta         `json:"meta,omitempty"` // hold data for pagination, filter, etc...
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
		// Id:          id, // id is gotten from header later
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
			errMap := ErrorDetail{
				"Field": e.Field(),
				"Tag":   e.Tag(),
				"Param": e.Param(),
				//errMap["Namespace"] = e.Namespace()
				//errMap["StructNamespace"] = e.StructNamespace()
				//errMap["StructField"] = e.StructField()
				//errMap["ActualTag"] = e.ActualTag()
				//errMap["Kind"] = e.Kind().String()
				//errMap["Type"] = e.Type().String()
				//errMap["Value"] = e.Value()
			}
			errorDetails = append(errorDetails, errMap)
		}
	}

	return &Response[T]{
		// Id:          id, // id is gotten from header later
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
		// Id:          id, // id is gotten from header later
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
