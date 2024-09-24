package storage

import (
	"context"

	"github.com/shynn2/cmd-gram/internal/models"
)

type Storage interface {
	CreateUser(ctx context.Context, u *models.UserDTO) (int, error)
	DeleteUser(ctx context.Context, u *models.User) (id int, err error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	CreateChat(ctx context.Context, u []*models.User) (id int, err error)
	GetAllChats(ctx context.Context, uid int) (*models.Chats, error)
	CreateMessage(ctx context.Context, msg *models.MessageDTO) (id int, err error)
	GetAllMessages(ctx context.Context, cid int) ([]models.Message, error)
}
