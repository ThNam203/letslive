package dto

import "github.com/gofrs/uuid/v5"

// GetUsersBatchInternalRequestDTO is the contract for the internal id -> identity
// lookup. Services that render another user's name must resolve it here rather
// than trusting a name supplied by the caller.
type GetUsersBatchInternalRequestDTO struct {
	Ids []string `json:"ids" validate:"required,min=1,max=100,dive,uuid"`
}

// UserIdentityInternalResponseDTO is the minimal identity of a user: only what a
// peer service needs to label them in its own UI.
type UserIdentityInternalResponseDTO struct {
	Id             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	ProfilePicture *string   `json:"profilePicture"`
}
