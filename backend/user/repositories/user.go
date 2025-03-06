package repositories

import (
	"context"
	"errors"
	"fmt"
	"sen1or/lets-live/user/domains"
	servererrors "sen1or/lets-live/user/errors"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetById(uuid.UUID) (*domains.User, *servererrors.ServerError)
	Query(isOnline, username string, page int) ([]*domains.User, *servererrors.ServerError)
	GetByName(string) (*domains.User, *servererrors.ServerError)
	GetByEmail(string) (*domains.User, *servererrors.ServerError)
	GetByAPIKey(uuid.UUID) (*domains.User, *servererrors.ServerError)

	Create(username, email string, isVerified bool) *servererrors.ServerError
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

func (r *postgresUserRepo) GetById(userId uuid.UUID) (*domains.User, *servererrors.ServerError) {
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT u.id, u.username, u.email, u.is_verified, u.is_online, u.is_active, u.created_at, u.stream_api_key, u.display_name, u.phone_number, u.bio, u.profile_picture, u.background_picture, l.user_id, l.title, l.description, l.thumbnail_url 
		FROM users u
		JOIN livestream_information l ON u.id = l.user_id
		WHERE u.id = $1
	`, userId.String())
	if err != nil {
		return nil, servererrors.ErrDatabaseQuery
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.User])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, servererrors.ErrUserNotFound
		}

		return nil, servererrors.ErrDatabaseIssue
	}

	return &user, nil
}

func (r *postgresUserRepo) Query(onlineStatus string, username string, page int) ([]*domains.User, *servererrors.ServerError) {
	whereConditions := []string{}
	args := []any{}
	argIndex := 1

	if len(onlineStatus) > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("is_online = $%d", argIndex))
		args = append(args, onlineStatus)
		argIndex++
	}

	if len(username) > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("SOUNDEX(username) = SOUNDEX($%d)", argIndex))
		args = append(args, username)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = " WHERE " + strings.Join(whereConditions, " AND ")
	}

	args = append(args, page*20, 20)
	query := fmt.Sprintf("SELECT * FROM users %s OFFSET $%d LIMIT $%d", whereClause, argIndex, argIndex+1)

	rows, err := r.dbConn.Query(context.Background(), query, args...)
	if err != nil {
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
		SELECT u.id, u.username, u.email, u.is_verified, u.is_online, u.is_active, u.created_at, u.stream_api_key, u.display_name, u.phone_number, u.bio, u.profile_picture, u.background_picture, l.user_id, l.title, l.description, l.thumbnail_url 
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

func (r *postgresUserRepo) Create(username, email string, isVerified bool) *servererrors.ServerError {
	params := pgx.NamedArgs{
		"username":    username,
		"email":       email,
		"is_verified": isVerified,
	}

	result, err := r.dbConn.Exec(context.Background(), "insert into users (username, email, is_verified) values (@username, @email, @is_verified) returning *", params)
	if err != nil {
		return servererrors.ErrDatabaseQuery
	}

	if result.RowsAffected() == 0 {
		return servererrors.ErrDatabaseIssue
	}

	return nil
}

func (r *postgresUserRepo) Update(user domains.User) (*domains.User, *servererrors.ServerError) {
	params := pgx.NamedArgs{
		"id":           user.Id,
		"is_online":    user.IsOnline,
		"is_active":    user.IsActive,
		"display_name": user.DisplayName,
		"phone_number": user.PhoneNumber,
		"bio":          user.Bio,
	}

	rows, err := r.dbConn.Query(
		context.Background(), `
		UPDATE users 
		SET is_online = @is_online, is_active = @is_active, display_name = @display_name, phone_number = @phone_number, bio = @bio 
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
