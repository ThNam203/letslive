package services

import (
	"context"
	"os"
	"time"

	"sen1or/letslive/admin/domains"
	"sen1or/letslive/admin/response"
	"sen1or/letslive/admin/types"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const AccessTokenMaxAge = 24 * time.Hour

type AuthService struct {
	repo domains.AdminAccountRepository
}

func NewAuthService(repo domains.AdminAccountRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Login returns the same RES_ERR_INVALID_CREDENTIALS for an unknown email and a wrong
// password on purpose — distinguishing the two would let a caller enumerate admin emails.
func (s *AuthService) Login(ctx context.Context, email, password string) (*domains.AdminAccount, *response.Response[any]) {
	account, errResp := s.repo.GetByEmail(ctx, email)
	if errResp != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_CREDENTIALS, nil, nil, nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_INVALID_CREDENTIALS, nil, nil, nil)
	}

	return account, nil
}

func (s *AuthService) GenerateAccessToken(adminId string) (string, time.Time, *response.Response[any]) {
	expiresAt := time.Now().Add(AccessTokenMaxAge)
	claims := types.AdminClaims{
		AdminId: adminId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := unsignedToken.SignedString([]byte(os.Getenv("ADMIN_JWT_SECRET")))
	if err != nil {
		return "", time.Time{}, response.NewResponseFromTemplate[any](response.RES_ERR_INTERNAL_SERVER, nil, nil, nil)
	}

	return signedToken, expiresAt, nil
}

func (s *AuthService) GetByID(ctx context.Context, adminIdStr string) (*domains.AdminAccount, *response.Response[any]) {
	adminId, err := uuid.FromString(adminIdStr)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_UNAUTHORIZED, nil, nil, nil)
	}
	return s.repo.GetByID(ctx, adminId)
}
