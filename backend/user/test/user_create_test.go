package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sen1or/lets-live/user/dto"
	"sen1or/lets-live/user/handlers"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser_Success(t *testing.T) {
	handler := setupHandler()

	validBody := dto.CreateUserRequestDTO{
		Username: "test_user",
		Email:    "test@example.com",
	}

	reqBody, _ := json.Marshal(validBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	handler.CreateUser(res, req)

	result := res.Result()
	defer result.Body.Close()

	assert.Equal(t, http.StatusOK, result.StatusCode)

	var response dto.CreateUserResponseDTO
	err := json.NewDecoder(result.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, validBody.Username, response.Username)
	assert.Equal(t, validBody.Email, response.Email)
	assert.False(t, response.IsOnline)
	assert.NotEmpty(t, response.Id)
	assert.NotEmpty(t, response.StreamAPIKey)
}

func TestCreateUser_ValidationFailure(t *testing.T) {
	handler := setupHandler()

	invalidBody := `{ "username": "", "email": "" }`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	handler.CreateUser(res, req)

	result := res.Result()
	defer result.Body.Close()

	var resMsg string
	json.NewEncoder(res.Body).Encode(&resMsg)

	assert.Equal(t, http.StatusBadRequest, result.StatusCode)
}

func TestCreateUser_ControllerError(t *testing.T) {
	mockController := &MockUserController{}
	mockController.On("Create", nil).Return(nil, errors.New("failed to create user")).Once()
	handler := handlers.NewUserHandler(mockController)

	body := `{ "username": "user_test", "email": "test@gmail.com" }`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	handler.CreateUser(res, req)

	result := res.Result()
	defer result.Body.Close()

	var resMsg string
	json.NewEncoder(res.Body).Encode(&resMsg)

	t.Log(resMsg)
	assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
}

func TestCreateUser_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: `{
                "username": "test_user",
                "email": "test@example.com"
            }`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Validation Error",
			requestBody: `{
                "username": "",
                "email": ""
            }`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Validation Error",
			requestBody: `{
                "username": "test_user",
                "email": ""
            }`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Validation Error",
			requestBody: `{
                "username": "",
                "email": "email@gmail.com"
            }`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Validation Error",
			requestBody: `{
                "username": "test_user",
                "email": "emailgmail.com"
            }`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := setupHandler()

			req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()

			handler.CreateUser(res, req)

			result := res.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expectedStatus, result.StatusCode)
		})
	}
}
