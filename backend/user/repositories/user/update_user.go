package user

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/pkg/logger"
	"sen1or/letslive/user/response"

	"github.com/jackc/pgx/v5"
)

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
				// nil = skip (user didn't send this field)
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
