package account

import (
	"context"
	"errors"

	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type Service interface {
	PostAccount(ctx context.Context, name, email, password string) (*Account, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
	Authenticate(ctx context.Context, email, password string) (*Account, error)
	GetAccountsCount(ctx context.Context) (uint64, error)
}

// Account is the internal representation of an account. PasswordHash is
// deliberately never sent over gRPC (see server.go) or exposed via
// GraphQL — it only ever lives inside this service and its database.
type Account struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

type accountService struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &accountService{r}
}

func (s *accountService) PostAccount(ctx context.Context, name, email, password string) (*Account, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	a := &Account{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		ID:           ksuid.New().String(),
	}
	if err := s.repository.PutAccount(ctx, *a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *accountService) GetAccount(ctx context.Context, id string) (*Account, error) {
	return s.repository.GetAccountByID(ctx, id)
}

func (s *accountService) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	return s.repository.ListAccounts(ctx, skip, take)
}

func (s *accountService) Authenticate(ctx context.Context, email, password string) (*Account, error) {
	a, err := s.repository.GetAccountByEmail(ctx, email)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return a, nil
}

func (s *accountService) GetAccountsCount(ctx context.Context) (uint64, error) {
	return s.repository.CountAccounts(ctx)
}
