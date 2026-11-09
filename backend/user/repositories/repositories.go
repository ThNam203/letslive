package repositories

import (
	"sen1or/letslive/user/domains"
	devicerepo "sen1or/letslive/user/repositories/device"
	followerrepo "sen1or/letslive/user/repositories/follower"
	giftrepo "sen1or/letslive/user/repositories/gift"
	inventoryrepo "sen1or/letslive/user/repositories/inventory"
	livestreaminforepo "sen1or/letslive/user/repositories/livestream_information"
	notificationrepo "sen1or/letslive/user/repositories/notification"
	userrepo "sen1or/letslive/user/repositories/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewUserRepository(conn *pgxpool.Pool) domains.UserRepository {
	return userrepo.NewUserRepository(conn)
}

func NewFollowRepository(conn *pgxpool.Pool) domains.FollowRepository {
	return followerrepo.NewFollowRepository(conn)
}

func NewLivestreamInformationRepository(conn *pgxpool.Pool) domains.LivestreamInformationRepository {
	return livestreaminforepo.NewLivestreamInformationRepository(conn)
}

func NewNotificationRepository(conn *pgxpool.Pool) domains.NotificationRepository {
	return notificationrepo.NewNotificationRepository(conn)
}

func NewInventoryRepository(conn *pgxpool.Pool) domains.InventoryRepository {
	return inventoryrepo.NewInventoryRepository(conn)
}

func NewDeviceTokenRepository(conn *pgxpool.Pool) domains.DeviceTokenRepository {
	return devicerepo.NewDeviceTokenRepository(conn)
}

func NewGiftRepository(conn *pgxpool.Pool) domains.GiftRepository {
	return giftrepo.NewGiftRepository(conn)
}
