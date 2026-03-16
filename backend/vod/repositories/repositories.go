package repositories

import (
	"sen1or/letslive/vod/domains"
	transcodejobrepo "sen1or/letslive/vod/repositories/transcode_job"
	vodrepo "sen1or/letslive/vod/repositories/vod"
	vodcommentrepo "sen1or/letslive/vod/repositories/vod_comment"
	vodcommentlikerepo "sen1or/letslive/vod/repositories/vod_comment_like"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
