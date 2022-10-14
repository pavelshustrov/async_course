package users

import (
	"context"
	"education.org/popug-tasks/internal/app/auth/services"
	"github.com/google/uuid"
)

//go:generate mockery --name=Repository
type Repository interface {
	Create(ctx context.Context, user *services.User) (*services.User, error)
	Update(ctx context.Context, user *services.User) (*services.User, error)
}

type service struct {
	repo Repository
}

func New(repo Repository) *service {
	return &service{
		repo: repo,
	}
}

// Register create new user in database.
// returns User with public_id.
func (s *service) Register(ctx context.Context, user *services.User) (*services.User, error) {
	user.PublicID = uuid.New().String()

	return s.repo.Create(ctx, user)
}

// Update existing user in database.
func (s *service) Update(ctx context.Context, user *services.User) (*services.User, error) {
	return s.repo.Update(ctx, user)
}
