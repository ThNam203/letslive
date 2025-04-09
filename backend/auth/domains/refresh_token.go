package domains

import (
	"context"
	serviceresponse "sen1or/letslive/auth/responses"
	"time"

	"github.com/gofrs/uuid/v5"
)

type RefreshToken struct {
	Id        uuid.UUID  `json:"id" db:"id"`
	Token     string     `json:"token" db:"token"`
	ExpiresAt time.Time  `json:"expiresAt" db:"expires_at"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	RevokedAt *time.Time `json:"revokedAt" db:"revoked_at"`
	UserId    uuid.UUID  `json:"userId" db:"user_id"`
}

type RefreshTokenRepository interface {
	RevokeAllTokensOfUser(context.Context, uuid.UUID) *serviceresponse.ServiceErrorResponse

	Insert(context.Context, *RefreshToken) *serviceresponse.ServiceErrorResponse
	FindByValue(context.Context, string) (*RefreshToken, *serviceresponse.ServiceErrorResponse)
	Update(context.Context, *RefreshToken) *serviceresponse.ServiceErrorResponse
}
