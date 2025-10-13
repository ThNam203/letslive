package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresUserRepo struct {
	dbConn *pgxpool.Pool
}

func NewUserRepository(conn *pgxpool.Pool) domains.UserRepository {
	return &postgresUserRepo{
		dbConn: conn,
	}
}

func (r *postgresUserRepo) GetPublicInfoById(ctx context.Context, userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT 
			u.id, u.username, u.email, u.created_at, u.display_name, u.phone_number, u.bio, u.profile_picture, u.background_picture,
			l.title, l.description, l.thumbnail_url, 
			COUNT(f.follower_id) AS follower_count,
			CASE 
    		    WHEN EXISTS (
    		        SELECT 1 FROM followers f2 
					WHERE f2.follower_id = $2 AND f2.user_id = u.id
    		    ) THEN true 
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
		WHERE u.id = $1
		GROUP BY u.id, l.user_id, l.title, l.description, l.thumbnail_url
	`, userId.String(), authenticatedUserId)
	if err != nil {
		logger.Errorf(ctx, "failed to query user full information: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[dto.GetUserPublicResponseDTO])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_USER_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		logger.Errorf(ctx, "failed to collect user full information: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	if len(user.SocialLinksJSON) > 0 {
		var linksMap map[string]string
		if err := json.Unmarshal([]byte(user.SocialLinksJSON), &linksMap); err == nil {
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
				case "github":
					social.Github = &val
				case "tiktok":
					social.LinkedIn = &val
				case "youtube":
					social.Youtube = &val
				case "website":
					social.Website = &val
				}
			}
			user.SocialMediaLinks = social
		}
	}

	return &user, nil
}

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
				case "github":
					social.Github = &val
				case "tiktok":
					social.LinkedIn = &val
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

func (r postgresUserRepo) GetAll(ctx context.Context, page int) ([]domains.User, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT id, username, email, status, created_at, display_name, phone_number, bio, profile_picture, background_picture
		FROM users
		OFFSET $1 LIMIT $2
	`, page*10, 10)

	if err != nil {
		logger.Errorf(ctx, "failed to get all users: %s", err)
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return users, nil
}

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

func (r *postgresUserRepo) GetByUsername(ctx context.Context, username string) (*domains.User, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, "select * from users where username = $1", username)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_USER_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return &user, nil
}

func (r *postgresUserRepo) GetByEmail(ctx context.Context, email string) (*domains.User, *response.Response[any]) {
	rows, err := r.dbConn.Query(ctx, "select * from users where email = $1", email)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_USER_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return &user, nil
}

func (r *postgresUserRepo) GetByAPIKey(ctx context.Context, apiKey uuid.UUID) (*domains.User, *response.Response[any]) {
	var user domains.User
	rows, err := r.dbConn.Query(ctx, `
		SELECT u.id, u.username, u.email, u.created_at, u.stream_api_key, u.display_name, u.phone_number, u.bio, u.profile_picture, u.background_picture, l.user_id, l.title, l.description, l.thumbnail_url 
		FROM users u
		JOIN livestream_information l ON u.id = l.user_id
		WHERE u.stream_api_key = $1
	`, apiKey)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	defer rows.Close()

	user, err = pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_USER_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return &user, nil
}

func (r *postgresUserRepo) Create(ctx context.Context, username string, email string, provider domains.AuthProvider) (*domains.User, *response.Response[any]) {
	params := pgx.NamedArgs{
		"username":      username,
		"email":         email,
		"auth_provider": provider,
	}

	row, err := r.dbConn.Query(ctx, "insert into users (username, email, auth_provider) values (@username, @email, @auth_provider) returning *", params)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}

	createdUser, err := pgx.CollectOneRow(row, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_USER_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	return &createdUser, nil
}

func (r *postgresUserRepo) Update(ctx context.Context, user dto.UpdateUserRequestDTO) (*domains.User, *response.Response[any]) {
	tx, err := r.dbConn.Begin(ctx)
	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE, nil, nil, nil,
		)
	}
	defer tx.Rollback(ctx)

	params := pgx.NamedArgs{
		"id":           user.Id,
		"status":       user.Status,
		"display_name": user.DisplayName,
		"phone_number": user.PhoneNumber,
		"bio":          user.Bio,
	}

	rows, err := tx.Query(
		ctx, `
		UPDATE users 
		SET display_name = @display_name, phone_number = @phone_number, bio = @bio, status = @status
		WHERE id = @id 
		RETURNING *
	`, params)

	if err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	}
	defer rows.Close()

	updatedUser, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, response.NewResponseFromTemplate[any](
				response.RES_ERR_USER_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		logger.Errorf(ctx, "database issue when update profile: %s", err.Error())
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE,
			nil,
			nil,
			nil,
		)
	}

	logger.Debugf(ctx, "social media links: %+v", user.SocialMediaLinks)

	if user.SocialMediaLinks != nil {
		links := user.SocialMediaLinks

		platforms := map[string]*string{
			"facebook":  links.Facebook,
			"twitter":   links.Twitter,
			"instagram": links.Instagram,
			"github":    links.Github,
			"linkedin":  links.LinkedIn,
			"youtube":   links.Youtube,
			"website":   links.Website,
		}

		for platform, url := range platforms {
			if url == nil {
				// nil = skip (user didn’t send this field)
				continue
			}

			if *url == "" {
				// empty string = remove link
				_, err := tx.Exec(ctx, `
				DELETE FROM user_social_links
				WHERE user_id = $1 AND platform = $2
			`, user.Id, platform)
				if err != nil {
					logger.Errorf(ctx, "failed to delete social link %s: %v", platform, err)
					return nil, response.NewResponseFromTemplate[any](
						response.RES_ERR_DATABASE_QUERY, nil, nil, nil,
					)
				}
				continue
			}

			// non-empty string = upsert link
			_, err := tx.Exec(ctx, `
			INSERT INTO user_social_links (user_id, platform, url)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, platform)
			DO UPDATE SET
				url = EXCLUDED.url,
				updated_at = NOW()
		`, user.Id, platform, *url)
			if err != nil {
				logger.Errorf(ctx, "failed to upsert social link %s: %v", platform, err)
				return nil, response.NewResponseFromTemplate[any](
					response.RES_ERR_DATABASE_QUERY, nil, nil, nil,
				)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_ISSUE, nil, nil, nil,
		)
	}

	return &updatedUser, nil
}

func (r *postgresUserRepo) UpdateStreamAPIKey(ctx context.Context, userId uuid.UUID, newKey string) *response.Response[any] {
	result, err := r.dbConn.Exec(ctx, "UPDATE users SET stream_api_key = $1 WHERE id = $2", newKey, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.NewResponseFromTemplate[any](
				response.RES_ERR_USER_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	} else if result.RowsAffected() == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_USER_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}

	return nil
}

func (r *postgresUserRepo) UpdateProfilePicture(ctx context.Context, userId uuid.UUID, path string) *response.Response[any] {
	result, err := r.dbConn.Exec(ctx, "UPDATE users SET profile_picture = $1 WHERE id = $2", path, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.NewResponseFromTemplate[any](
				response.RES_ERR_USER_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	} else if result.RowsAffected() == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_USER_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}

	return nil
}

func (r *postgresUserRepo) UpdateBackgroundPicture(ctx context.Context, userId uuid.UUID, path string) *response.Response[any] {
	result, err := r.dbConn.Exec(ctx, "UPDATE users SET background_picture = $1 WHERE id = $2", path, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response.NewResponseFromTemplate[any](
				response.RES_ERR_USER_NOT_FOUND,
				nil,
				nil,
				nil,
			)
		}

		return response.NewResponseFromTemplate[any](
			response.RES_ERR_DATABASE_QUERY,
			nil,
			nil,
			nil,
		)
	} else if result.RowsAffected() == 0 {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_USER_NOT_FOUND,
			nil,
			nil,
			nil,
		)
	}

	return nil
}
