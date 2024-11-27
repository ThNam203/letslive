package repositories

import (
	"context"
	"errors"
	"sen1or/lets-live/auth/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

type AuthRepository interface {
	GetByID(uuid.UUID) (*domains.Auth, error)
	GetByUserID(uuid.UUID) (*domains.Auth, error)
	GetByEmail(string) (*domains.Auth, error)

	Create(domains.Auth) (*domains.Auth, error)
	UpdatePasswordHash(domains.Auth) (*domains.Auth, error)
	UpdateVerify(domains.Auth) (*domains.Auth, error)
	Delete(uuid.UUID) error
}

type postgresAuthRepo struct {
	dbConn *pgx.Conn
}

func NewAuthRepository(conn *pgx.Conn) AuthRepository {
	return &postgresAuthRepo{
		dbConn: conn,
	}
}

func (r *postgresAuthRepo) GetByID(authID uuid.UUID) (*domains.Auth, error) {
	rows, err := r.dbConn.Query(context.Background(), "select * from auths where id = $1", authID.String())
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *postgresAuthRepo) GetByUserID(userID uuid.UUID) (*domains.Auth, error) {
	rows, err := r.dbConn.Query(context.Background(), "select * from auths where user_id = $1", userID.String())
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *postgresAuthRepo) GetByEmail(email string) (*domains.Auth, error) {
	rows, err := r.dbConn.Query(context.Background(), "select * from auths where email = $1", email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *postgresAuthRepo) Create(newAuth domains.Auth) (*domains.Auth, error) {
	params := pgx.NamedArgs{
		"email":         newAuth.Email,
		"password_hash": newAuth.PasswordHash,
		"is_verified":   newAuth.IsVerified,
		"user_id":       newAuth.UserID,
	}

	rows, err := r.dbConn.Query(context.Background(), "insert into auths (email, password_hash, is_verified, user_id) values (@email, @password_hash, @is_verified, @user_id) RETURNING *", params)
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &user, err
}

func (r *postgresAuthRepo) UpdatePasswordHash(user domains.Auth) (*domains.Auth, error) {
	rows, err := r.dbConn.Query(context.Background(), "UPDATE auths SET password_hash = $1 WHERE id = $3 RETURNING *", user.PasswordHash, user.ID)

	updatedAuth, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &updatedAuth, err
}

func (r *postgresAuthRepo) UpdateVerify(user domains.Auth) (*domains.Auth, error) {
	rows, err := r.dbConn.Query(context.Background(), "UPDATE auths SET is_verified = $2 WHERE id = $3 RETURNING *", user.IsVerified, user.ID)

	updatedAuth, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domains.Auth])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &updatedAuth, err
}

func (r *postgresAuthRepo) Delete(userID uuid.UUID) error {
	_, err := r.dbConn.Exec(context.Background(), "DELETE FROM auths WHERE id = $1", userID.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRecordNotFound
		}
	}

	return err
}
