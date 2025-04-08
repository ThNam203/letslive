package repositories

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	servererrors "sen1or/letslive/user/errors"
	"sen1or/letslive/user/pkg/logger"

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

func (r *postgresUserRepo) GetPublicInfoById(ctx context.Context, userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT 
			u.id, u.username, u.email, u.created_at, u.display_name, u.phone_number, u.bio, u.profile_picture, u.background_picture, 
			l.user_id, l.title, l.description, l.thumbnail_url, 
			COUNT(f.follower_id) AS follower_count,
			CASE 
    		    WHEN EXISTS (
    		        SELECT 1 FROM followers f2 
					WHERE f2.follower_id = $2 AND f2.user_id = u.id
    		    ) THEN true 
    		    ELSE false 
    		END AS is_following
		FROM users u
		LEFT JOIN livestream_information l ON u.id = l.user_id
		LEFT JOIN followers f ON u.id = f.user_id
		WHERE u.id = $1
		GROUP BY u.id, l.user_id, l.title, l.description, l.thumbnail_url
	`, userId.String(), authenticatedUserId)
	if err != nil {
		logger.Errorf("failed to query user full information: %s", err)
		return nil, servererrors.ErrDatabaseQuery
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[dto.GetUserPublicResponseDTO])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrUserNotFound
		}

		logger.Errorf("failed to collect user full information: %s", err)
		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresUserRepo) GetById(ctx context.Context, userId uuid.UUID) (*domains.User, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT 
			u.id, u.username, u.email, u.created_at, u.display_name, u.auth_provider, u.stream_api_key, u.phone_number, u.bio, u.profile_picture, u.background_picture, l.user_id, l.title, l.description, l.thumbnail_url
		FROM users u
		LEFT JOIN livestream_information l ON u.id = l.user_id
		WHERE u.id = $1
	`, userId.String())
	if err != nil {
		logger.Errorf("failed to query user full information: %s", err)
		return nil, servererrors.ErrDatabaseQuery
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.User])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrUserNotFound
		}

		logger.Errorf("failed to collect user full information: %s", err)
		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r postgresUserRepo) GetAll(ctx context.Context, page int) ([]domains.User, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(ctx, `
		SELECT *
		FROM users
		OFFSET $1 LIMIT $2
	`, page*10, 10)

	if err != nil {
		logger.Errorf("failed to get all users: %s", err)
		return nil, servererrors.ErrDatabaseQuery
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		return nil, servererrors.ErrDatabaseIssue
	}

	return users, nil
}

func (r *postgresUserRepo) SearchUsersByUsername(ctx context.Context, query string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *servererrors.ServerError) {
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
		logger.Errorf("failed to search for users: %s", err)
		return nil, servererrors.ErrDatabaseQuery
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[dto.GetUserPublicResponseDTO])
	if err != nil {
		logger.Errorf("failed to collect rows: %s", err)
		return nil, servererrors.ErrDatabaseIssue
	}

	return users, nil
}

func (r *postgresUserRepo) GetByUsername(ctx context.Context, username string) (*domains.User, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(ctx, "select * from users where username = $1", username)
	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrUserNotFound
		}

		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresUserRepo) GetByEmail(ctx context.Context, email string) (*domains.User, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(ctx, "select * from users where email = $1", email)
	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrUserNotFound
		}
		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresUserRepo) GetByAPIKey(ctx context.Context, apiKey uuid.UUID) (*domains.User, *servererrors.ServerError) {
	var user domains.User
	rows, err := r.dbConn.Query(ctx, `
		SELECT u.id, u.username, u.email, u.created_at, u.stream_api_key, u.display_name, u.phone_number, u.bio, u.profile_picture, u.background_picture, l.user_id, l.title, l.description, l.thumbnail_url 
		FROM users u
		JOIN livestream_information l ON u.id = l.user_id
		WHERE u.stream_api_key = $1
	`, apiKey)
	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	user, err = pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrUserNotFound
		}

		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresUserRepo) Create(ctx context.Context, username string, email string, provider domains.AuthProvider) (*domains.User, *servererrors.ServerError) {
	params := pgx.NamedArgs{
		"username":      username,
		"email":         email,
		"auth_provider": provider,
	}

	row, err := r.dbConn.Query(ctx, "insert into users (username, email, auth_provider) values (@username, @email,  @auth_provider) returning *", params)
	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}

	createdUser, err := pgx.CollectOneRow(row, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrUserNotFound
		}

		return nil, servererrors.ErrDatabaseIssue
	}

	return &createdUser, nil
}

func (r *postgresUserRepo) Update(ctx context.Context, user dto.UpdateUserRequestDTO) (*domains.User, *servererrors.ServerError) {
	params := pgx.NamedArgs{
		"id":           user.Id,
		"display_name": user.DisplayName,
		"phone_number": user.PhoneNumber,
		"bio":          user.Bio,
	}

	rows, err := r.dbConn.Query(
		ctx, `
		UPDATE users 
		SET display_name = @display_name, phone_number = @phone_number, bio = @bio 
		WHERE id = @id 
		RETURNING *
	`, params)

	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}
	defer rows.Close()

	updatedUser, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrUserNotFound
		}

		return nil, servererrors.ErrDatabaseIssue
	}

	return &updatedUser, nil
}

func (r *postgresUserRepo) UpdateStreamAPIKey(ctx context.Context, userId uuid.UUID, newKey string) *servererrors.ServerError {
	result, err := r.dbConn.Exec(ctx, "UPDATE users SET stream_api_key = $1 WHERE id = $2", newKey, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return servererrors.ErrUserNotFound
		}

		return servererrors.ErrDatabaseQuery
	} else if result.RowsAffected() == 0 {
		return servererrors.ErrUserNotFound
	}

	return nil
}

func (r *postgresUserRepo) UpdateProfilePicture(ctx context.Context, userId uuid.UUID, path string) *servererrors.ServerError {
	result, err := r.dbConn.Exec(ctx, "UPDATE users SET profile_picture = $1 WHERE id = $2", path, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return servererrors.ErrUserNotFound
		}

		return servererrors.ErrDatabaseQuery
	} else if result.RowsAffected() == 0 {
		return servererrors.ErrUserNotFound
	}

	return nil
}

func (r *postgresUserRepo) UpdateBackgroundPicture(ctx context.Context, userId uuid.UUID, path string) *servererrors.ServerError {
	result, err := r.dbConn.Exec(ctx, "UPDATE users SET background_picture = $1 WHERE id = $2", path, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return servererrors.ErrUserNotFound
		}

		return servererrors.ErrDatabaseQuery
	} else if result.RowsAffected() == 0 {
		return servererrors.ErrUserNotFound
	}

	return nil
}
