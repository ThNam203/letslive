package mapper

import (
	"sen1or/lets-live/user/domains"
	"sen1or/lets-live/user/dto"
)

func CreateUserRequestDTOToUser(dto dto.CreateUserRequestDTO) *domains.User {
	return &domains.User{
		Username: dto.Username,
		Email:    dto.Email,
	}
}

func UserToCreateUserResponseDTO(user domains.User) *dto.CreateUserResponseDTO {
	return &dto.CreateUserResponseDTO{
		Id:           user.Id,
		Username:     user.Username,
		Email:        user.Email,
		IsOnline:     user.IsOnline,
		CreatedAt:    user.CreatedAt,
		StreamAPIKey: user.StreamAPIKey,
	}
}

func UserToGetUserResponseDTO(user domains.User) *dto.GetUserResponseDTO {
	return &dto.GetUserResponseDTO{
		Id:           user.Id,
		Username:     user.Username,
		Email:        user.Email,
		IsOnline:     user.IsOnline,
		CreatedAt:    user.CreatedAt,
		StreamAPIKey: user.StreamAPIKey,
	}
}

func UserToUpdateUserResponseDTO(user domains.User) *dto.UpdateUserResponseDTO {
	return &dto.UpdateUserResponseDTO{
		Id:           user.Id,
		Username:     user.Username,
		Email:        user.Email,
		IsOnline:     user.IsOnline,
		CreatedAt:    user.CreatedAt,
		StreamAPIKey: user.StreamAPIKey,
	}
}
