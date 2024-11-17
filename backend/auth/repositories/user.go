package repositories

import (
	"context"
	"sen1or/lets-live/auth/domains"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	GetByID(uuid.UUID) (*domains.User, error)
	GetByEmail(string) (*domains.User, error)

	Create(*domains.User) error
	Update(*domains.User) error
	Delete(uuid.UUID) error
}

type postgresUserRepo struct {
	dbConn *pgx.Conn
}

func NewUserRepository(conn *pgx.Conn) UserRepository {
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
		return nil, err
	}

	return &user, nil
}

func (r *postgresUserRepo) Create(newUser *domains.User) error {
	params := pgx.NamedArgs{
		"email":         newUser.Email,
		"password_hash": newUser.PasswordHash,
		"is_verified":   newUser.IsVerified,
	}

	_, err := r.dbConn.Exec(context.Background(), "insert into users (email, password_hash, is_verified) values (@email, @password_hash, @is_verified)", params)

	return err
}

func (r *postgresUserRepo) Update(user *domains.User) error {
	_, err := r.dbConn.Exec(context.Background(), "UPDATE users SET password_hash= $1, is_verified = $2 WHERE id = $3", user.PasswordHash, user.IsVerified, user.ID)
	return err
}

func (r *postgresUserRepo) Delete(userID uuid.UUID) error {
	_, err := r.dbConn.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID.String())
	return err
}
