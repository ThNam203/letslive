package dto

type CreateNotificationRequestDTO struct {
	UserId      string  `json:"userId" validate:"required,uuid"`
	Type        string  `json:"type" validate:"required,lte=50"`
	Title       string  `json:"title" validate:"required,lte=200"`
	Message     string  `json:"message" validate:"required,lte=500"`
	ActionUrl   *string `json:"actionUrl,omitempty"`
	ActionLabel *string `json:"actionLabel,omitempty" validate:"omitempty,lte=100"`
	ReferenceId *string `json:"referenceId,omitempty" validate:"omitempty,uuid"`
}
