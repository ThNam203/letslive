package mapper

import (
	"sen1or/lets-live/user/domains"
	"sen1or/lets-live/user/dto"
)

func CreateUserRequestDTOToUser(dto dto.CreateUserRequestDTO) *domains.User {
	return &domains.User{
		Username:   dto.Username,
		Email:      dto.Email,
		IsVerified: dto.IsVerified,
	}
}

func UserToGetUserPublicResponseDTO(user domains.User) *dto.GetUserPublicResponseDTO {
	return &dto.GetUserPublicResponseDTO{
		Id:                user.Id,
		Username:          user.Username,
		Email:             user.Email,
		IsVerified:        user.IsVerified,
		CreatedAt:         user.CreatedAt,
		PhoneNumber:       user.PhoneNumber,
		Bio:               user.Bio,
		DisplayName:       user.DisplayName,
		ProfilePicture:    user.ProfilePicture,
		BackgroundPicture: user.BackgroundPicture,
	}
}

func UpdateUserRequestDTOToUser(dto dto.UpdateUserRequestDTO) domains.User {
	return domains.User{
		Id:          dto.Id,
		PhoneNumber: dto.PhoneNumber,
		DisplayName: dto.DisplayName,
		Bio:         dto.Bio,
	}
}
