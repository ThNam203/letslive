package response

import (
	sharedresponse "sen1or/letslive/shared/response"
)

// Re-export shared types for use within finance service
type Meta = sharedresponse.Meta
type ErrorDetail = sharedresponse.ErrorDetail
type ErrorDetails = sharedresponse.ErrorDetails
type Response[T any] = sharedresponse.Response[T]
type ResponseTemplate = sharedresponse.ResponseTemplate
