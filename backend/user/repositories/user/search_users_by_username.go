package user

import (
	"context"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresUserRepo) SearchUsersByUsername(ctx context.Context, query string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT
		    u.id,
		    u.username,
		    u.email,
		    u.created_at,
		    u.display_name,
		    u.phone_number,
		    u.bio,
		    u.profile_picture,
		    u.background_picture,
			l.user_id,
			l.title, 
			l.description, 
			l.thumbnail_url, 
		    COUNT(f.follower_id) AS follower_count,
		    CASE
		        WHEN EXISTS (
		            SELECT 1 FROM followers f2 WHERE f2.follower_id = $2 AND f2.user_id = u.id
		        ) THEN true
		        ELSE false
		    END AS is_following
		FROM
		    users u
		LEFT JOIN
		    livestream_information l ON u.id = l.user_id
		LEFT JOIN
		    followers f ON u.id = f.user_id
		WHERE 
		    u.username % $1 OR levenshtein(u.username, $1) <= 5
		GROUP BY 
		    u.id,
		    l.user_id,
		    l.title,
		    l.description,
		    l.thumbnail_url
		ORDER BY
		    similarity(u.username, $1) DESC,
		    levenshtein(u.username, $1) ASC
		LIMIT 10;
	`, query, authenticatedUserId)
	if err != nil {
		logger.Errorf(ctx, "failed to search for users: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[dto.GetUserPublicResponseDTO])
	if err != nil {
		logger.Errorf(ctx, "failed to collect rows: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return users, nil
}
