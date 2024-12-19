package test

import (
	"sen1or/lets-live/user/dto"
	"sen1or/lets-live/user/handlers"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/mock"
)

func setupHandler() *handlers.UserHandler {
	handler := handlers.NewUserHandler(&MockUserController{})
	return handler
}

type MockUserController struct {
	mock.Mock
}

func (m *MockUserController) Create(mockBody dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, error) {
	args := m.Called(mockBody)
	err := args.Error(1)

	if err != nil {
		return nil, err
	}

	userId, _ := uuid.NewGen().NewV4()
	userStreamAPIKey, _ := uuid.NewGen().NewV4()

	return &dto.CreateUserResponseDTO{
		ID:           userId,
		Username:     mockBody.Username,
		Email:        mockBody.Email,
		IsOnline:     false,
		CreatedAt:    time.Now(),
		StreamAPIKey: userStreamAPIKey,
	}, nil
}

func (m *MockUserController) GetByID(id uuid.UUID) (*dto.GetUserResponseDTO, error) {
	return nil, nil
}
func (m *MockUserController) GetByEmail(email string) (*dto.GetUserResponseDTO, error) {
	return nil, nil
}
func (m *MockUserController) GetByStreamAPIKey(key uuid.UUID) (*dto.GetUserResponseDTO, error) {

	return nil, nil
}
func (m *MockUserController) GetStreamingUsers() ([]*dto.GetUserResponseDTO, error) {
	return nil, nil
}
func (m *MockUserController) Update(updateDTO dto.UpdateUserRequestDTO) (*dto.UpdateUserResponseDTO, error) {
	return nil, nil
}
func (m *MockUserController) Delete(userID uuid.UUID) error {
	return nil
}
