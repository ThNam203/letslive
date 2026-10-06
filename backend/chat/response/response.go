package response

import (
	"errors"
	"net/http"

	"sen1or/letslive/chat/domains"
	sharedresponse "sen1or/letslive/shared/response"
)

type Meta = sharedresponse.Meta
type Response[T any] = sharedresponse.Response[T]
type ResponseTemplate = sharedresponse.ResponseTemplate

func NewResponseFromTemplate[T any](tpl ResponseTemplate, data *T, meta *Meta, errorDetails *sharedresponse.ErrorDetails) *Response[T] {
	return sharedresponse.NewResponseFromTemplate(tpl, data, meta, errorDetails)
}

// Codes, keys and messages are the ones the Node chat service used, so the web
// client's i18n keys and error handling keep working.
const (
	RES_ERR_INVALID_INPUT_CODE   = 20000
	RES_ERR_INVALID_PAYLOAD_CODE = 20001
	RES_ERR_UNAUTHORIZED_CODE    = 20005
	RES_ERR_FORBIDDEN_CODE       = 20008
	RES_ERR_ROUTE_NOT_FOUND_CODE = 20012
	RES_ERR_DATABASE_QUERY_CODE  = 20015
	RES_ERR_DATABASE_ISSUE_CODE  = 20016
	RES_ERR_INTERNAL_SERVER_CODE = 20017

	RES_ERR_ROOM_NOT_FOUND_CODE         = 50018
	RES_ERR_CONVERSATION_NOT_FOUND_CODE = 50019
	RES_ERR_DM_ALREADY_EXISTS_CODE      = 50020
	RES_ERR_NOT_PARTICIPANT_CODE        = 50021
	RES_ERR_INSUFFICIENT_ROLE_CODE      = 50022
	RES_ERR_DM_MESSAGE_NOT_FOUND_CODE   = 50023
	RES_ERR_CANNOT_MESSAGE_SELF_CODE    = 50024
	RES_ERR_TOO_MANY_PARTICIPANTS_CODE  = 50025
	RES_ERR_USER_SETUP_INCOMPLETE_CODE  = 50026

	RES_SUCC_OK_CODE = 100000
)

const (
	RES_ERR_INVALID_INPUT_KEY   = "res_err_invalid_input"
	RES_ERR_INVALID_PAYLOAD_KEY = "res_err_invalid_payload"
	RES_ERR_UNAUTHORIZED_KEY    = "res_err_unauthorized"
	RES_ERR_FORBIDDEN_KEY       = "res_err_forbidden"
	RES_ERR_ROUTE_NOT_FOUND_KEY = "res_err_route_not_found"
	RES_ERR_DATABASE_QUERY_KEY  = "res_err_database_query"
	RES_ERR_DATABASE_ISSUE_KEY  = "res_err_database_issue"
	RES_ERR_INTERNAL_SERVER_KEY = "res_err_internal_server"

	RES_ERR_ROOM_NOT_FOUND_KEY         = "res_err_room_not_found"
	RES_ERR_CONVERSATION_NOT_FOUND_KEY = "res_err_conversation_not_found"
	RES_ERR_DM_ALREADY_EXISTS_KEY      = "res_err_dm_already_exists"
	RES_ERR_NOT_PARTICIPANT_KEY        = "res_err_not_participant"
	RES_ERR_INSUFFICIENT_ROLE_KEY      = "res_err_insufficient_role"
	RES_ERR_DM_MESSAGE_NOT_FOUND_KEY   = "res_err_dm_message_not_found"
	RES_ERR_CANNOT_MESSAGE_SELF_KEY    = "res_err_cannot_message_self"
	RES_ERR_TOO_MANY_PARTICIPANTS_KEY  = "res_err_too_many_participants"
	RES_ERR_USER_SETUP_INCOMPLETE_KEY  = "res_err_user_setup_incomplete"

	RES_SUCC_OK_KEY = "res_succ_ok"
)

func tpl(success bool, status, code int, key, message string) ResponseTemplate {
	return ResponseTemplate{Success: success, StatusCode: status, Code: code, Key: key, Message: message}
}

var (
	RES_SUCC_OK      = tpl(true, http.StatusOK, RES_SUCC_OK_CODE, RES_SUCC_OK_KEY, "")
	RES_SUCC_CREATED = tpl(true, http.StatusCreated, RES_SUCC_OK_CODE, RES_SUCC_OK_KEY, "")

	RES_ERR_INVALID_INPUT   = tpl(false, http.StatusBadRequest, RES_ERR_INVALID_INPUT_CODE, RES_ERR_INVALID_INPUT_KEY, "Input invalid.")
	RES_ERR_INVALID_PAYLOAD = tpl(false, http.StatusBadRequest, RES_ERR_INVALID_PAYLOAD_CODE, RES_ERR_INVALID_PAYLOAD_KEY, "Payload invalid.")
	RES_ERR_UNAUTHORIZED    = tpl(false, http.StatusUnauthorized, RES_ERR_UNAUTHORIZED_CODE, RES_ERR_UNAUTHORIZED_KEY, "Unauthorized.")
	RES_ERR_FORBIDDEN       = tpl(false, http.StatusForbidden, RES_ERR_FORBIDDEN_CODE, RES_ERR_FORBIDDEN_KEY, "Forbidden.")
	RES_ERR_ROUTE_NOT_FOUND = tpl(false, http.StatusNotFound, RES_ERR_ROUTE_NOT_FOUND_CODE, RES_ERR_ROUTE_NOT_FOUND_KEY, "Route not found.")
	RES_ERR_DATABASE_QUERY  = tpl(false, http.StatusInternalServerError, RES_ERR_DATABASE_QUERY_CODE, RES_ERR_DATABASE_QUERY_KEY, "Database query failed.")
	RES_ERR_DATABASE_ISSUE  = tpl(false, http.StatusInternalServerError, RES_ERR_DATABASE_ISSUE_CODE, RES_ERR_DATABASE_ISSUE_KEY, "Database issue.")
	RES_ERR_INTERNAL_SERVER = tpl(false, http.StatusInternalServerError, RES_ERR_INTERNAL_SERVER_CODE, RES_ERR_INTERNAL_SERVER_KEY, "Internal server error.")

	RES_ERR_ROOM_NOT_FOUND         = tpl(false, http.StatusNotFound, RES_ERR_ROOM_NOT_FOUND_CODE, RES_ERR_ROOM_NOT_FOUND_KEY, "Room not found")
	RES_ERR_CONVERSATION_NOT_FOUND = tpl(false, http.StatusNotFound, RES_ERR_CONVERSATION_NOT_FOUND_CODE, RES_ERR_CONVERSATION_NOT_FOUND_KEY, "Conversation not found")
	RES_ERR_DM_ALREADY_EXISTS      = tpl(false, http.StatusConflict, RES_ERR_DM_ALREADY_EXISTS_CODE, RES_ERR_DM_ALREADY_EXISTS_KEY, "DM conversation already exists")
	RES_ERR_NOT_PARTICIPANT        = tpl(false, http.StatusForbidden, RES_ERR_NOT_PARTICIPANT_CODE, RES_ERR_NOT_PARTICIPANT_KEY, "You are not a participant of this conversation")
	RES_ERR_INSUFFICIENT_ROLE      = tpl(false, http.StatusForbidden, RES_ERR_INSUFFICIENT_ROLE_CODE, RES_ERR_INSUFFICIENT_ROLE_KEY, "Insufficient permissions for this action")
	RES_ERR_DM_MESSAGE_NOT_FOUND   = tpl(false, http.StatusNotFound, RES_ERR_DM_MESSAGE_NOT_FOUND_CODE, RES_ERR_DM_MESSAGE_NOT_FOUND_KEY, "Message not found")
	RES_ERR_CANNOT_MESSAGE_SELF    = tpl(false, http.StatusBadRequest, RES_ERR_CANNOT_MESSAGE_SELF_CODE, RES_ERR_CANNOT_MESSAGE_SELF_KEY, "Cannot create a conversation with yourself")
	RES_ERR_TOO_MANY_PARTICIPANTS  = tpl(false, http.StatusBadRequest, RES_ERR_TOO_MANY_PARTICIPANTS_CODE, RES_ERR_TOO_MANY_PARTICIPANTS_KEY, "Too many participants")
	RES_ERR_USER_SETUP_INCOMPLETE  = tpl(false, http.StatusBadRequest, RES_ERR_USER_SETUP_INCOMPLETE_CODE, RES_ERR_USER_SETUP_INCOMPLETE_KEY, "User has not finished account setup")
)

// FromError maps a domain error to the response envelope. Unknown errors are
// reported as a generic internal error without leaking their text.
func FromError(err error) *Response[any] {
	template := RES_ERR_INTERNAL_SERVER
	switch {
	case errors.Is(err, domains.ErrInvalidInput), errors.Is(err, domains.ErrAlreadyExists):
		template = RES_ERR_INVALID_INPUT
	case errors.Is(err, domains.ErrRoomNotFound):
		template = RES_ERR_ROOM_NOT_FOUND
	case errors.Is(err, domains.ErrForbidden):
		template = RES_ERR_FORBIDDEN
	case errors.Is(err, domains.ErrDatabaseIssue):
		template = RES_ERR_DATABASE_ISSUE
	case errors.Is(err, domains.ErrConversationNotFound):
		template = RES_ERR_CONVERSATION_NOT_FOUND
	case errors.Is(err, domains.ErrNotParticipant):
		template = RES_ERR_NOT_PARTICIPANT
	case errors.Is(err, domains.ErrInsufficientRole):
		template = RES_ERR_INSUFFICIENT_ROLE
	case errors.Is(err, domains.ErrDmMessageNotFound):
		template = RES_ERR_DM_MESSAGE_NOT_FOUND
	case errors.Is(err, domains.ErrCannotMessageSelf):
		template = RES_ERR_CANNOT_MESSAGE_SELF
	case errors.Is(err, domains.ErrTooManyParticipants):
		template = RES_ERR_TOO_MANY_PARTICIPANTS
	case errors.Is(err, domains.ErrUserSetupIncomplete):
		template = RES_ERR_USER_SETUP_INCOMPLETE
	}
	return NewResponseFromTemplate[any](template, nil, nil, nil)
}
