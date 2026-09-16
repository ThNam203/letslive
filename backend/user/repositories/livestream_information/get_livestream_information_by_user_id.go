package livestream_information

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresLivestreamInformationRepo) GetByUserId(ctx context.Context, userId uuid.UUID) (*domains.LivestreamInformation, error) {
	rows, err := r.dbConn.Query(ctx, "select * from livestream_information where user_id = $1", userId.String())
	if err != nil {
		return nil, domains.ErrDatabaseQuery
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.LivestreamInformation])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domains.ErrUserNotFound
		}

		return nil, domains.ErrDatabaseIssue
	}

	return &user, nil
}
