package repositories

import (
	"sen1or/letslive/livestream/domains"
	livestreamrepo "sen1or/letslive/livestream/repositories/livestream"
	transcodejobrepo "sen1or/letslive/livestream/repositories/transcode_job"
	vodrepo "sen1or/letslive/livestream/repositories/vod"
	vodcommentrepo "sen1or/letslive/livestream/repositories/vod_comment"
	vodcommentlikerepo "sen1or/letslive/livestream/repositories/vod_comment_like"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewLivestreamRepository(conn *pgxpool.Pool) domains.LivestreamRepository {
	return livestreamrepo.NewLivestreamRepository(conn)
}

func NewVODRepository(conn *pgxpool.Pool) domains.VODRepository {
	return vodrepo.NewVODRepository(conn)
}

func NewVODCommentRepository(conn *pgxpool.Pool) domains.VODCommentRepository {
	return vodcommentrepo.NewVODCommentRepository(conn)
}

func NewVODCommentLikeRepository(conn *pgxpool.Pool) domains.VODCommentLikeRepository {
	return vodcommentlikerepo.NewVODCommentLikeRepository(conn)
}

func NewTranscodeJobRepository(conn *pgxpool.Pool) domains.TranscodeJobRepository {
	return transcodejobrepo.NewTranscodeJobRepository(conn)
}
