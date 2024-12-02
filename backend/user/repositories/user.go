package repositories

import (
	"context"
	"errors"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/user/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	GetByID(uuid.UUID) (*domains.User, error)
	GetByName(string) (*domains.User, error)
	GetByEmail(string) (*domains.User, error)
	GetByAPIKey(uuid.UUID) (*domains.User, error)
	GetByFacebookID(string) (*domains.User, error)
	GetStreamingUsers() ([]domains.User, error)

	Create(domains.User) (*domains.User, error)
	Update(domains.User) (*domains.User, error)
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

func (r *postgresUserRepo) GetByID(userId uuid.UUID) (*domains.User, error) {
	rows, err := r.dbConn.Query(context.Background(), "select * from users where id = $1", userId.String())
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.User])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &user, nil
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
	rows, err := r.dbConn.Query(context.Background(), "select * from users where stream_api_key = $1", apiKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.User])
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
		"username": newUser.Username,
		"email":    newUser.Email,
	}

	rows, err := r.dbConn.Query(context.Background(), "insert into users (username, email) values (@username, @email) returning *", params)
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

	return &user, err
}

func (r *postgresUserRepo) Update(user domains.User) (*domains.User, error) {
	logger.Infof("UPDATE users SET username = %s, is_online = %v WHERE id = %s RETURNING *", user.Username, user.IsOnline, user.ID)
	rows, err := r.dbConn.Query(context.Background(), "UPDATE users SET username = $1, is_online = $2 WHERE id = $3 RETURNING *", user.Username, user.IsOnline, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updatedUser, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &updatedUser, err
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
