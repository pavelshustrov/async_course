package users

import (
	"context"
	"encoding/json"
	"fmt"

	"education.org/popug-tasks/internal/app/account/services"
)

type usersRepo interface {
	Create(ctx context.Context, user *services.User) (*services.User, error)
	Update(ctx context.Context, user *services.User) (*services.User, error)
	UpdateToken(ctx context.Context, user *services.User) (*services.User, error)
}

type usersHandler struct {
	repo usersRepo
}

func New(repo usersRepo) *usersHandler {
	return &usersHandler{
		repo: repo,
	}
}

func (uh *usersHandler) Updated() func(string) error {
	return func(s string) error {
		var (
			user services.User
		)

		err := json.Unmarshal([]byte(s), &user)
		if err != nil {
			return err
		}

		_, err = uh.repo.Update(context.Background(), &user)
		if err != nil {
			return err
		}

		fmt.Println("user.updated", s)
		return nil
	}
}

func (uh *usersHandler) Created() func(string) error {
	return func(s string) error {
		var (
			user services.User
		)

		err := json.Unmarshal([]byte(s), &user)
		if err != nil {
			return err
		}

		_, err = uh.repo.Create(context.Background(), &user)
		if err != nil {
			return err
		}

		fmt.Println("user.created", s)
		return nil
	}
}

func (uh *usersHandler) Auth() func(string) error {
	return func(payload string) error {
		var (
			user services.User
		)

		err := json.Unmarshal([]byte(payload), &user)
		if err != nil {
			return err
		}

		_, err = uh.repo.UpdateToken(context.Background(), &user)
		if err != nil {
			return err
		}

		fmt.Println("user.auth", payload)
		return nil
	}
}
