package services

import (
	"context"

	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"

	"github.com/gofrs/uuid/v5"
)

type DeviceService struct {
	deviceRepo domains.DeviceTokenRepository
}

func NewDeviceService(deviceRepo domains.DeviceTokenRepository) *DeviceService {
	return &DeviceService{deviceRepo: deviceRepo}
}

// Register stores the token for userID, taking it over from any previous owner.
func (s *DeviceService) Register(ctx context.Context, userID uuid.UUID, req dto.RegisterDeviceRequestDTO) (*domains.DeviceToken, error) {
	return s.deviceRepo.Upsert(ctx, domains.DeviceToken{
		UserId:   userID,
		Token:    req.Token,
		Platform: req.Platform,
		Locale:   req.Locale,
	})
}

// Unregister removes the caller's own token. Unknown or foreign tokens are a
// no-op so the endpoint never reveals whether a token exists.
func (s *DeviceService) Unregister(ctx context.Context, userID uuid.UUID, token string) error {
	if token == "" || len(token) > dto.MaxDeviceTokenLength {
		return domains.ErrInvalidInput
	}
	return s.deviceRepo.DeleteByUserAndToken(ctx, userID, token)
}
