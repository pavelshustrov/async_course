package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"education.org/popug-tasks/internal/app/auth/services"
)

type authService interface {
	LoginWithCredentials(ctx context.Context, credentials *services.Credentials) (*services.Token, error)
	VerifyToken(ctx context.Context, token *services.Token) (*services.User, error)
}

type authHandler struct {
	auth authService
}

func New(service authService) *authHandler {
	return &authHandler{
		auth: service,
	}
}

func (b *authHandler) Login(w http.ResponseWriter, req *http.Request) {
	creds := &services.Credentials{}

	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to read body: %v", err)))
		return
	}

	err = json.Unmarshal(bytes, creds)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to read body: %v", err)))
		return
	}

	token, err := b.auth.LoginWithCredentials(context.Background(), creds)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("something failed: %v", err)))
		return
	}

	w.WriteHeader(http.StatusCreated)
	payload, _ := json.Marshal(token)
	_, _ = w.Write(payload)
}

func (b *authHandler) Verify(w http.ResponseWriter, req *http.Request) {
	token := &services.Token{}

	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to read body: %v", err)))
		return
	}

	err = json.Unmarshal(bytes, token)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(fmt.Sprintf("failed to read body: %v", err)))
		return
	}

	user, err := b.auth.VerifyToken(context.Background(), token)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("something failed: %v", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	payload, _ := json.Marshal(user)
	_, _ = w.Write(payload)
}
