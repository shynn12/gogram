package storage

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shynn2/cmd-gram/internal/models"
)

type db struct {
	pool *pgxpool.Pool
}

func (d *db) CreateUser(ctx context.Context, u *models.UserDTO) (id int, err error) {
	err = d.pool.QueryRow(ctx, "Insert into users (email, encrypted_password) Values($1, $2) Returning id",
		u.Email,
		u.EncryptedPassword,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *db) DeleteUser(ctx context.Context, u *models.User) (id int, err error) {
	err = d.pool.QueryRow(ctx, "Delete from users where id=$1 Returning id",
		&u.ID,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *db) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := d.pool.QueryRow(ctx, "Select id, email, encrypted_password from users where email = $1", email).Scan(
		&u.ID,
		&u.Email,
		&u.EncryptedPassword,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (d *db) CreateMessage(ctx context.Context, msg *models.MessageDTO) (id int, err error) {
	err = d.pool.QueryRow(ctx, "Insert into messages (user_id, body, time) VALUES ($1, $2, $3) Returning message_id", msg.UserID, msg.Body, msg.Time).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, err
}

func New(pool *pgxpool.Pool) Storage {
	return &db{
		pool: pool,
	}
}
