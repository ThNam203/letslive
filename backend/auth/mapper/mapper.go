package mapper

import (
	"sen1or/lets-live/auth/domains"
	"sen1or/lets-live/auth/dto"
)

func AuthToSignUpResponseDTO(createdAuth domains.Auth) *dto.SignUpResponseDTO {
	return &dto.SignUpResponseDTO{
		Id:         createdAuth.Id,
		UserId:     createdAuth.UserId,
		Email:      createdAuth.Email,
		IsVerified: createdAuth.IsVerified,
	}
}
