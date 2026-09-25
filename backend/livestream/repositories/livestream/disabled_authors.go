package livestream

import (
	"context"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/livestream/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresLivestreamRepo) SetAuthorDisabled(ctx context.Context, userId uuid.UUID, disabled bool) error {
	query := `delete from disabled_authors where user_id = $1`
	if disabled {
		query = `insert into disabled_authors (user_id) values ($1) on conflict (user_id) do nothing`
	}

	if _, err := r.dbConn.Exec(ctx, query, userId); err != nil {
		logger.Errorf(ctx, "db exec error [setauthordisabled user=%s disabled=%t: %v]", userId, disabled, err)
		return domains.ErrDatabaseQuery
	}

	return nil
}

func (r *postgresLivestreamRepo) ReplaceDisabledAuthors(ctx context.Context, userIds []uuid.UUID) error {
	if userIds == nil {
		userIds = []uuid.UUID{}
	}

	tx, err := r.dbConn.Begin(ctx)
	if err != nil {
		logger.Errorf(ctx, "db begin error [replacedisabledauthors: %v]", err)
		return domains.ErrDatabaseQuery
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `delete from disabled_authors where user_id <> all($1::uuid[])`, userIds); err != nil {
		logger.Errorf(ctx, "db exec error [replacedisabledauthors delete: %v]", err)
		return domains.ErrDatabaseQuery
	}

	if _, err := tx.Exec(ctx, `insert into disabled_authors (user_id) select unnest($1::uuid[]) on conflict (user_id) do nothing`, userIds); err != nil {
		logger.Errorf(ctx, "db exec error [replacedisabledauthors insert: %v]", err)
		return domains.ErrDatabaseQuery
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Errorf(ctx, "db commit error [replacedisabledauthors: %v]", err)
		return domains.ErrDatabaseQuery
	}

	return nil
}
