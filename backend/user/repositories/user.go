package repositories

import (
	"context"
	"errors"
	"sen1or/lets-live/user/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetById(uuid.UUID) (*domains.User, error)
	GetAll() ([]*domains.User, error)
	GetByName(string) (*domains.User, error)
	GetByEmail(string) (*domains.User, error)
	GetByAPIKey(uuid.UUID) (*domains.User, error)
	GetByFacebookID(string) (*domains.User, error)
	GetStreamingUsers() ([]domains.User, error)

	Create(domains.User) (*domains.User, error)
	Update(domains.User) (*domains.User, error)
	UpdateUserVerified(userId uuid.UUID) error
	UpdateStreamAPIKey(userId uuid.UUID, newKey string) error
	UpdateProfilePicture(uuid.UUID, string) error
	UpdateBackgroundPicture(uuid.UUID, string) error
	Delete(uuid.UUID) error
}

type postgresUserRepo struct {
	dbConn *pgxpool.Pool
}

func NewUserRepository(conn *pgxpool.Pool) UserRepository {
	return &postgresUserRepo{
		dbConn: conn,
	}
}

func (r *postgresUserRepo) GetById(userId uuid.UUID) (*domains.User, error) {
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT u.id, u.username, u.email, u.is_verified, u.is_online, u.is_active, u.created_at, u.stream_api_key, u.display_name, u.phone_number, u.bio, u.profile_picture, u.background_picture, l.user_id, l.title, l.description, l.thumbnail_url 
		FROM users u
		LEFT JOIN livestream_information l ON u.id = l.user_id
		where u.id = $1
	`, userId.String())
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.User])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (r *postgresUserRepo) GetAll() ([]*domains.User, error) {
	// TODO: pagination
	rows, err := r.dbConn.Query(context.Background(), "select * from users limit 100")
	if err != nil {
		return nil, err
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domains.User])

	if err != nil {
		return nil, err
	}

	var returnUsers = []*domains.User{}
	for _, u := range users {
		returnUsers = append(returnUsers, &u)
	}

	return returnUsers, nil
}

func (r *postgresUserRepo) GetByName(username string) (*domains.User, error) {
	rows, err := r.dbConn.Query(context.Background(), "select * from users where username = $1", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &user, nil
}
func (r *postgresUserRepo) GetByEmail(email string) (*domains.User, error) {
	rows, err := r.dbConn.Query(context.Background(), "select * from users where email = $1", email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &user, nil
}

// TODO: revise the oauth2 login
func (r *postgresUserRepo) GetByFacebookID(facebookID string) (*domains.User, error) {
	rows, err := r.dbConn.Query(context.Background(), "select * from users where facebook_id = $1", facebookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &user, nil

}

func (r *postgresUserRepo) GetByAPIKey(apiKey uuid.UUID) (*domains.User, error) {
	var user domains.User
	rows, err := r.dbConn.Query(context.Background(), `
		SELECT u.id, u.username, u.email, u.is_verified, u.is_online, u.is_active, u.created_at, u.stream_api_key, u.display_name, u.phone_number, u.bio, u.profile_picture, u.background_picture, l.user_id, l.title, l.description, l.thumbnail_url 
		FROM users u
		LEFT JOIN livestream_information l ON u.id = l.user_id
		where u.stream_api_key = $1
	`, apiKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err = pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.User])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRecordNotFound
	}

	return &user, nil
}

func (r *postgresUserRepo) GetStreamingUsers() ([]domains.User, error) {
	rows, err := r.dbConn.Query(context.Background(), "select * from users where is_online = $1", true)
	defer rows.Close()

	streamingUsers, err := pgx.CollectRows(rows, pgx.RowToStructByName[domains.User])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domains.User{}, nil
		}

		return nil, err
	}

	return streamingUsers, nil
}

func (r *postgresUserRepo) Create(newUser domains.User) (*domains.User, error) {
	params := pgx.NamedArgs{
		"username":    newUser.Username,
		"email":       newUser.Email,
		"is_verified": newUser.IsVerified,
	}

	rows, err := r.dbConn.Query(context.Background(), "insert into users (username, email, is_verified) values (@username, @email, @is_verified) returning *", params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &user, err
}

func (r *postgresUserRepo) Update(user domains.User) (*domains.User, error) {
	params := pgx.NamedArgs{
		"id":           user.Id,
		"is_online":    user.IsOnline,
		"is_active":    user.IsActive,
		"display_name": user.DisplayName,
		"phone_number": user.PhoneNumber,
		"bio":          user.Bio,
	}

	rows, err := r.dbConn.Query(context.Background(), "UPDATE users SET is_online = @is_online, is_active = @is_active, display_name = @display_name, phone_number = @phone_number, bio = @bio WHERE id = @id RETURNING *", params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updatedUser, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &updatedUser, err
}

func (r *postgresUserRepo) UpdateStreamAPIKey(userId uuid.UUID, newKey string) error {
	_, err := r.dbConn.Query(context.Background(), "UPDATE users SET stream_api_key = $1 WHERE id = $2", newKey, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}

func (r *postgresUserRepo) UpdateProfilePicture(userId uuid.UUID, path string) error {
	_, err := r.dbConn.Exec(context.Background(), "UPDATE users SET profile_picture = $1 WHERE id = $2", path, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}

func (r *postgresUserRepo) UpdateBackgroundPicture(userId uuid.UUID, path string) error {
	_, err := r.dbConn.Exec(context.Background(), "UPDATE users SET background_picture = $1 WHERE id = $2", path, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRecordNotFound
		}
		return err
	}

	return nil
}

func (r *postgresUserRepo) Delete(userID uuid.UUID) error {
	_, err := r.dbConn.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRecordNotFound
		}

		return err // lol
	}

	return err
}

func (r *postgresUserRepo) UpdateUserVerified(userId uuid.UUID) error {
	_, err := r.dbConn.Exec(context.Background(), "UPDATE users SET is_verified = $1 WHERE id = $2", true, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRecordNotFound
		}

		return err // lol
	}

	return err

}
