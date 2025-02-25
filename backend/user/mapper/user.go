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

func UserToGetUserResponseDTO(user domains.User) *dto.GetUserResponseDTO {
	return &dto.GetUserResponseDTO{
		Id:                user.Id,
		Username:          user.Username,
		Email:             user.Email,
		IsOnline:          user.IsOnline,
		IsVerified:        user.IsVerified,
		CreatedAt:         user.CreatedAt,
		PhoneNumber:       user.PhoneNumber,
		Bio:               user.Bio,
		DisplayName:       user.DisplayName,
		ProfilePicture:    user.ProfilePicture,
		BackgroundPicture: user.BackgroundPicture,
		VODs:              []dto.GetLivestreamResponseDTO{},
	}
}

func UpdateUserRequestDTOToUser(dto dto.UpdateUserRequestDTO) domains.User {
	return domains.User{
		Id:          dto.Id,
		IsOnline:    *dto.IsOnline,
		PhoneNumber: dto.PhoneNumber,
		DisplayName: dto.DisplayName,
		Bio:         dto.Bio,
	}
}
