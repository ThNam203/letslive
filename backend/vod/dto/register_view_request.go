package dto

type RegisterViewRequestDTO struct {
	WatchedSeconds int64 `json:"watchedSeconds" validate:"required,gte=0"`
}
