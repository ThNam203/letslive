package user

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresUserRepo) GetStatusesByIds(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT id, status FROM users WHERE id = ANY($1::uuid[])
	`, userIds)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	defer rows.Close()

	statuses := make(map[uuid.UUID]domains.UserStatus, len(userIds))
	for rows.Next() {
		var id uuid.UUID
		var status domains.UserStatus
		if err := rows.Scan(&id, &status); err != nil {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_DATABASE_ISSUE,
				nil,
				nil,
				nil,
			)
		}
		statuses[id] = status
	}

	if err := rows.Err(); err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return statuses, nil
}
