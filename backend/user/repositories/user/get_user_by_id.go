package user

import (
	"context"
	"encoding/json"
	"errors"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

func (r *postgresUserRepo) GetById(ctx context.Context, userId uuid.UUID) (*domains.User, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT
			u.id,
			u.username,
			u.email,
			u.status,
			u.created_at,
			u.display_name,
			u.auth_provider,
			u.stream_api_key,
			u.phone_number,
			u.bio,
			u.profile_picture,
			u.background_picture,
			(SELECT (COUNT(*))::int FROM followers f WHERE f.user_id = u.id) AS follower_count,
			l.title,
			l.description,
			l.thumbnail_url,
			COALESCE(
				jsonb_object_agg(usl.platform, usl.url) FILTER (WHERE usl.platform IS NOT NULL),
				'{}'::jsonb
			) AS social_links_json
		FROM users u
		LEFT JOIN livestream_information l ON u.id = l.user_id
		LEFT JOIN user_social_links usl ON usl.user_id = u.id
		WHERE u.id = $1
		GROUP BY
			u.id, u.username, u.email, u.status, u.created_at, u.display_name, u.auth_provider,
			u.stream_api_key, u.phone_number, u.bio, u.profile_picture, u.background_picture,
			l.user_id, l.title, l.description, l.thumbnail_url
	`, userId.String())
	if err != nil {
		logger.Errorf(ctx, "failed to query user full information: %s", err)
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_QUERY, nil, nil, nil)
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](response.RES_ERR_USER_NOT_FOUND, nil, nil, nil)
		}
		logger.Errorf(ctx, "failed to collect user full information: %s", err)
		return nil, response.NewResponseFromTemplate[any](response.RES_ERR_DATABASE_ISSUE, nil, nil, nil)
	}

	if len(user.SocialLinksJSON) > 0 {
		var linksMap map[string]string
		if err := json.Unmarshal([]byte(user.SocialLinksJSON), &linksMap); err == nil {
			social := &domains.SocialMediaLinks{}
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
			user.SocialMediaLinks = *social
		}
	}

	return &user, nil
}
