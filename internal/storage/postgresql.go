package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shynn2/cmd-gram/internal/models"
)

type db struct {
	pool *pgxpool.Pool
}

func (d *db) CreateUser(ctx context.Context, u *models.UserDTO) (id int, err error) {
	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{})
	defer tx.Rollback(ctx)

	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(ctx, "Insert into users (email, encrypted_password) Values($1, $2) Returning id",
		u.Email,
		u.EncryptedPassword,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *db) DeleteUser(ctx context.Context, u *models.User) (id int, err error) {
	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{})
	defer tx.Rollback(ctx)

	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(ctx, "Delete from users where id=$1 Returning id",
		&u.ID,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *db) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}

	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{})
	defer tx.Rollback(ctx)

	if err != nil {
		return nil, err
	}

	err = tx.QueryRow(ctx, "Select id, email, encrypted_password from users where email = $1", email).Scan(
		&u.ID,
		&u.Email,
		&u.EncryptedPassword,
	)
	if err != nil {
		return nil, err
	}

	tx.Commit(ctx)
	return u, nil
}

func (d *db) CreateChat(ctx context.Context, u []*models.User) (id int, err error) {
	if len(u) < 2 {
		return 0, fmt.Errorf("cant create chat with less than 2 users")
	}
	tx, err := d.pool.Begin(ctx)
	defer tx.Rollback(ctx)

	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(ctx, "insert into chats(name) values($1) returning id;", fmt.Sprintf("%s-%s", u[0].Email, u[1].Email)).Scan(&id)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx, "insert into party (user_id, chat_id) values($1, $2)", u[0].ID, id)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx, "insert into party (user_id, chat_id) values($1, $2)", u[1].ID, id)
	if err != nil {
		return 0, err
	}
	tx.Commit(ctx)

	return id, nil
}

func (d *db) GetAllChats(ctx context.Context, u *models.User) ([]*models.Chat, error) {
	chats := []*models.Chat{}
	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{})
	defer tx.Rollback(ctx)

	if err != nil {
		return chats, err
	}

	rows, err := tx.Query(ctx, "Select chat_id from party where user_id = $1", u.ID)
	if err != nil {
		return chats, err
	}
	chat := &models.Chat{}
	var chatId int
	chatsID := []int{}
	for rows.Next() {
		err = rows.Scan(&chatId)
		if err != nil {
			return nil, err
		}
		chatsID = append(chatsID, chatId)
	}
	rows.Close()

	for _, v := range chatsID {
		err = tx.QueryRow(ctx, "Select id, name from chats where id = $1", v).Scan(&chat.ID, &chat.Name)
		if err != nil {
			return nil, err
		}

		chats = append(chats, chat)
	}
	tx.Commit(ctx)
	return chats, err
}

func (d *db) CreateMessage(ctx context.Context, msg *models.MessageDTO) (id int, err error) {
	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{})
	defer tx.Rollback(ctx)

	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(ctx, "Insert into messages (user_id, body, chat_id, time) VALUES ($1, $2, $3, $4) Returning message_id", msg.UserID, msg.Body, msg.ChatID, msg.Time).Scan(&id)
	if err != nil {
		return 0, err
	}

	tx.Commit(ctx)

	return id, err
}

func (d *db) GetAllMessages(ctx context.Context, c *models.Chat) (id int, err error) {
	return id, err
}

func New(pool *pgxpool.Pool) Storage {
	return &db{
		pool: pool,
	}
}
