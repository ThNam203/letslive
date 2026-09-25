package user

import (
	"context"
	"sen1or/letslive/user/domains"

	"github.com/gofrs/uuid/v5"
)

func (r *postgresUserRepo) GetStatusesByIds(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, error) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT id, status FROM users WHERE id = ANY($1::uuid[])
	`, userIds)
	if err != nil {
		return nil, domains.ErrDatabaseQuery
	}
	defer rows.Close()

	statuses := make(map[uuid.UUID]domains.UserStatus, len(userIds))
	for rows.Next() {
		var id uuid.UUID
		var status domains.UserStatus
		if err := rows.Scan(&id, &status); err != nil {
			return nil, domains.ErrDatabaseIssue
		}
		statuses[id] = status
	}

	if err := rows.Err(); err != nil {
		return nil, domains.ErrDatabaseIssue
	}

	return statuses, nil
}
