package dto

// PushDataDTO is the deep-link data attached to a push. Only these keys are
// accepted; the mobile app routes taps on them.
type PushDataDTO struct {
	Type        string `json:"type" validate:"required,max=50"`
	ReferenceID string `json:"referenceId" validate:"omitempty,max=100"`
	ActionURL   string `json:"actionUrl" validate:"omitempty,max=2048"`
}

// SendPushInternalRequestDTO is the body of POST /v1/internal/push: one push
// delivered to each listed user. Used by peer services (chat) for pushes that
// are not stored notifications, such as direct messages.
type SendPushInternalRequestDTO struct {
	UserIds []string    `json:"userIds" validate:"required,min=1,max=100,dive,uuid"`
	Title   string      `json:"title" validate:"required,max=100"`
	Body    string      `json:"body" validate:"required,max=500"`
	Data    PushDataDTO `json:"data"`
}
