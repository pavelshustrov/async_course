package users

import (
	"context"

	"education.org/popug-tasks/internal/app/auth/services"
)

//go:generate mockery --name=Repository
type Repository interface {
	Login(ctx context.Context, credentials *services.Credentials) (*services.Token, error)
	Verify(ctx context.Context, token *services.Token) (*services.User, error)
}

type service struct {
	repo Repository
}

func New(repo Repository) *service {
	return &service{
		repo: repo,
	}
}

// LoginWithCredentials verifies credentials and creates token for public_id.
func (s *service) LoginWithCredentials(ctx context.Context, credentials *services.Credentials) (*services.Token, error) {
	return s.repo.Login(ctx, credentials)
}

// VerifyToken token in database.
func (s *service) VerifyToken(ctx context.Context, token *services.Token) (*services.User, error) {
	return s.repo.Verify(ctx, token)
}
