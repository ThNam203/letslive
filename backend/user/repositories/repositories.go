package repositories

import (
	"sen1or/letslive/user/domains"
	followerrepo "sen1or/letslive/user/repositories/follower"
	livestreaminforepo "sen1or/letslive/user/repositories/livestream_information"
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
