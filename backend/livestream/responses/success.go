package serviceresponse

type ServiceSuccessResponse struct {
	StatusCode int
	Message    string
	Data       any
}

func NewServiceSuccessResponse(statusCode int, message string, data any) *ServiceSuccessResponse {
	return &ServiceSuccessResponse{
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}
}

var ()
