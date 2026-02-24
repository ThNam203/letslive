package user

import (
	"context"
	"encoding/json"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresUserRepo) GetRecommendedPublic(ctx context.Context, excludeUserId *uuid.UUID, excludeUserIds []uuid.UUID, page, limit int) ([]dto.GetUserPublicResponseDTO, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT
			u.id, u.username, u.email, u.status, u.auth_provider, u.created_at, u.display_name, u.phone_number, u.bio, u.profile_picture, u.background_picture,
			l.title, l.description, l.thumbnail_url,
			(COUNT(f.follower_id))::int AS follower_count,
			CASE
				WHEN $2::uuid IS NULL THEN false
				WHEN EXISTS (SELECT 1 FROM followers f2 WHERE f2.follower_id = $2::uuid AND f2.user_id = u.id) THEN true
				ELSE false
			END AS is_following,
			COALESCE(
				jsonb_object_agg(usl.platform, usl.url) FILTER (WHERE usl.platform IS NOT NULL),
				'{}'::jsonb
			) AS social_links_json
		FROM users u
		LEFT JOIN livestream_information l ON u.id = l.user_id
		LEFT JOIN followers f ON u.id = f.user_id
		LEFT JOIN user_social_links usl ON usl.user_id = u.id
		WHERE ($1::uuid IS NULL OR u.id != $1)
		  AND (cardinality($3::uuid[]) = 0 OR NOT (u.id = ANY($3::uuid[])))
		GROUP BY u.id, u.username, u.email, u.status, u.auth_provider, u.created_at, u.display_name, u.phone_number, u.bio, u.profile_picture, u.background_picture, l.user_id, l.title, l.description, l.thumbnail_url
		ORDER BY (COUNT(f.follower_id)) DESC
		LIMIT $4 OFFSET $5
	`, excludeUserId, excludeUserId, excludeUserIds, limit, page*limit)
	if err != nil {
		logger.Errorf(ctx, "failed to get recommended public users: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[dto.GetUserPublicResponseDTO])
	if err != nil {
		logger.Errorf(ctx, "failed to collect recommended public users: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	for i := range users {
		if len(users[i].SocialLinksJSON) > 0 {
			var linksMap map[string]string
			if err := json.Unmarshal([]byte(users[i].SocialLinksJSON), &linksMap); err == nil {
				social := &dto.SocialMediaLinks{}
				for k, v := range linksMap {
					val := v
					switch k {
					case "facebook":
						social.Facebook = &val
					case "twitter":
						social.Twitter = &val
					case "instagram":
						social.Instagram = &val
					case "linkedin":
						social.LinkedIn = &val
					case "github":
						social.Github = &val
					case "tiktok":
						social.TikTok = &val
					case "youtube":
						social.Youtube = &val
					case "website":
						social.Website = &val
					}
				}
				users[i].SocialMediaLinks = social
			}
		}
	}

	return users, nil
}
