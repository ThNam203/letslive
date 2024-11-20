package mapper

import (
	"sen1or/lets-live/auth/domains"
	"sen1or/lets-live/auth/dto"
)

func AuthToSignUpResponseDTO(createdAuth domains.Auth) *dto.SignUpResponseDTO {
	return &dto.SignUpResponseDTO{
		ID:         createdAuth.ID,
		UserID:     createdAuth.UserID,
		Email:      createdAuth.Email,
		IsVerified: createdAuth.IsVerified,
	}
}
