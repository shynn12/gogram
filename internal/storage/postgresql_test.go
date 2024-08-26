package storage

import (
	"context"
	"log"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shynn2/cmd-gram/internal/models"
)

var TestDB *db
var DatabaseURL = "postgres://postgres:psql@localhost:5432/cmdgram"

func TestMain(m *testing.M) {
	pool, err := pgxpool.Connect(context.Background(), DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	if err = pool.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}
	TestDB = &db{
		pool: pool,
	}

	m.Run()
}

func Test_db_User(t *testing.T) {
	u := &models.UserDTO{
		Email:             "123@12223",
		EncryptedPassword: "131EWDSAD1E21ASD",
	}
	id, err := TestDB.CreateUser(context.Background(), u)
	if err != nil {
		t.Error(err)
		return
	}
	u1, err := TestDB.FindByEmail(context.Background(), "123@12223")

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(u1)

	if u1.ID != id {
		t.Errorf("FindByEmail() = %v, want %v", u1.ID, id)
		return
	}

	id, err = TestDB.DeleteUser(context.Background(), u1)
	if err != nil {
		t.Error(err)
		return
	}
	if id != u1.ID {
		t.Errorf("DeleteUser() = %v, want %v", u1.ID, id)
	}
}

func Test_db_Chat(t *testing.T) {
	_, _ = TestDB.CreateUser(context.Background(), &models.UserDTO{Email: "d@d", EncryptedPassword: "123321"})
	_, _ = TestDB.CreateUser(context.Background(), &models.UserDTO{Email: "e@e", EncryptedPassword: "123321"})
	u, _ := TestDB.FindByEmail(context.Background(), "d@d")
	u2, _ := TestDB.FindByEmail(context.Background(), "e@e")

	_, err := TestDB.CreateChat(context.Background(), []*models.User{u, u2})
	if err != nil {
		t.Fatal(err)
	}

	c1, err := TestDB.GetAllChats(context.Background(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	c2, err := TestDB.GetAllChats(context.Background(), u2.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(c1, c2)

	if slices.Equal(c1, c2) {
		t.Fatal("chats of two users are not equal")
	}
}

func Test_db_Message(t *testing.T) {
	_, _ = TestDB.CreateUser(context.Background(), &models.UserDTO{Email: "d@d", EncryptedPassword: "123321"})
	_, _ = TestDB.CreateUser(context.Background(), &models.UserDTO{Email: "e@e", EncryptedPassword: "123321"})
	u, _ := TestDB.FindByEmail(context.Background(), "d@d")
	u2, _ := TestDB.FindByEmail(context.Background(), "e@e")

	cid, err := TestDB.CreateChat(context.Background(), []*models.User{u, u2})
	if err != nil {
		t.Fatal(err)
	}

	id, err := TestDB.CreateMessage(context.Background(), &models.MessageDTO{UserID: u.ID, Body: "hello", ChatID: cid, Time: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	msgs, err := TestDB.GetAllMessages(context.Background(), cid)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(msgs)
	for _, v := range msgs {
		if v.ID == id {
			return
		}
	}
	t.Errorf("no message witj id: %d", id)
}
