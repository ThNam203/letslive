package repositories

import (
	"context"
	"errors"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/user/domains"
	"sen1or/lets-live/user/dto"
	servererrors "sen1or/lets-live/user/errors"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetFullInfoById(uuid.UUID) (*domains.User, *servererrors.ServerError)
	GetById(userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserResponseDTO, *servererrors.ServerError)
	GetAll(page int) ([]domains.User, *servererrors.ServerError)
	SearchUserByUsername(username string) ([]*domains.User, *servererrors.ServerError)
	GetByName(string) (*domains.User, *servererrors.ServerError)
	GetByEmail(string) (*domains.User, *servererrors.ServerError)
	GetByAPIKey(uuid.UUID) (*domains.User, *servererrors.ServerError)

	Create(username, email string, isVerified bool, authProvider domains.AuthProvider) (*domains.User, *servererrors.ServerError)
	Update(domains.User) (*domains.User, *servererrors.ServerError)
	UpdateUserVerified(userId uuid.UUID) *servererrors.ServerError
	UpdateStreamAPIKey(userId uuid.UUID, newKey string) *servererrors.ServerError
	UpdateProfilePicture(uuid.UUID, string) *servererrors.ServerError
	UpdateBackgroundPicture(uuid.UUID, string) *servererrors.ServerError
	Delete(uuid.UUID) *servererrors.ServerError
}

type postgresUserRepo struct {
	dbConn *pgxpool.Pool
}

func NewUserRepository(conn *pgxpool.Pool) UserRepository {
	return &postgresUserRepo{
		dbConn: conn,
	}
}

// TODO: holy shesh i need to redo the whole database queries
// the authenticatedUserId is used for checking if the caller is following the userId
// the authenticatedUserId can be null if for INTERNAL USE
func (r *postgresUserRepo) GetById(userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserResponseDTO, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT 
			u.id, u.username, u.email, u.is_verified, u.created_at, u.display_name, u.phone_number, u.bio, u.profile_picture, u.background_picture, 
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
		GROUP BY u.id, l.user_id, l.title, l.description, l.thumbnail_url;
	`, userId.String(), authenticatedUserId)
	if err != nil {
		logger.Errorf("failed to query user full information: %s", err)
		return nil, servererrors.ErrDatabaseQuery
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[dto.GetUserResponseDTO])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrUserNotFound
		}

		logger.Errorf("failed to collect user full information: %s", err)
		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresUserRepo) GetFullInfoById(userId uuid.UUID) (*domains.User, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT 
			u.id, u.username, u.email, u.is_verified, u.created_at, u.display_name, u.auth_provider, u.stream_api_key, u.phone_number, u.bio, u.profile_picture, u.background_picture, l.user_id, l.title, l.description, l.thumbnail_url
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

func (r postgresUserRepo) GetAll(page int) ([]domains.User, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), `
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

func (r *postgresUserRepo) SearchUserByUsername(query string) ([]*domains.User, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT *
		FROM users
		WHERE username % $1 OR levenshtein(username, $1) <= 5
		ORDER BY similarity(username, $1) DESC, levenshtein(username, $1) ASC
		LIMIT 10
	`, query)
	if err != nil {
		logger.Errorf("failed to search for users: %s", err)
		return nil, servererrors.ErrDatabaseQuery
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		return nil, servererrors.ErrDatabaseIssue
	}

	var returnUsers []*domains.User
	for _, u := range users {
		returnUsers = append(returnUsers, &u)
	}

	return returnUsers, nil
}

func (r *postgresUserRepo) GetByName(username string) (*domains.User, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), "select * from users where username = $1", username)
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

func (r *postgresUserRepo) GetByEmail(email string) (*domains.User, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), "select * from users where email = $1", email)
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

func (r *postgresUserRepo) GetByAPIKey(apiKey uuid.UUID) (*domains.User, *servererrors.ServerError) {
	var user domains.User
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT u.id, u.username, u.email, u.is_verified, u.created_at, u.stream_api_key, u.display_name, u.phone_number, u.bio, u.profile_picture, u.background_picture, l.user_id, l.title, l.description, l.thumbnail_url 
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

func (r *postgresUserRepo) Create(username string, email string, isVerified bool, provider domains.AuthProvider) (*domains.User, *servererrors.ServerError) {
	params := pgx.NamedArgs{
		"username":      username,
		"email":         email,
		"is_verified":   isVerified,
		"auth_provider": provider,
	}

	row, err := r.dbConn.Query(context.Background(), "insert into users (username, email, is_verified, auth_provider) values (@username, @email, @is_verified, @auth_provider) returning *", params)
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

func (r *postgresUserRepo) Update(user domains.User) (*domains.User, *servererrors.ServerError) {
	params := pgx.NamedArgs{
		"id":           user.Id,
		"display_name": user.DisplayName,
		"phone_number": user.PhoneNumber,
		"bio":          user.Bio,
	}

	rows, err := r.dbConn.Query(
		context.Background(), `
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

func (r *postgresUserRepo) UpdateStreamAPIKey(userId uuid.UUID, newKey string) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "UPDATE users SET stream_api_key = $1 WHERE id = $2", newKey, userId)
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

func (r *postgresUserRepo) UpdateProfilePicture(userId uuid.UUID, path string) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "UPDATE users SET profile_picture = $1 WHERE id = $2", path, userId)
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

func (r *postgresUserRepo) UpdateBackgroundPicture(userId uuid.UUID, path string) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "UPDATE users SET background_picture = $1 WHERE id = $2", path, userId)
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

func (r *postgresUserRepo) Delete(userID uuid.UUID) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID.String())
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

func (r *postgresUserRepo) UpdateUserVerified(userId uuid.UUID) *servererrors.ServerError {
	result, err := r.dbConn.Exec(context.Background(), "UPDATE users SET is_verified = $1 WHERE id = $2", true, userId)
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
