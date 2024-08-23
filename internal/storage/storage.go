package storage

import (
	"context"

	"github.com/shynn2/cmd-gram/internal/models"
)

type Storage interface {
	CreateUser(ctx context.Context, u *models.UserDTO) (int, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}
