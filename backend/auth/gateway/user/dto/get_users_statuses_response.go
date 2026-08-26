package dto

type GetUsersStatusesResponseDTO struct {
	Statuses map[string]string `json:"statuses"`
}
