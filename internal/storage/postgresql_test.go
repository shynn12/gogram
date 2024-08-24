package storage

import (
	"context"
	"log"
	"testing"

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

func Test_db_CreateUser(t *testing.T) {
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
