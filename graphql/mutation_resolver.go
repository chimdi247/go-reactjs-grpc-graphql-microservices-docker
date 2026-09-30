package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/akhilsharma90/go-graphql-microservice/order"
)

var (
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrUnauthenticated  = errors.New("you must be logged in to do that")
	ErrForbidden        = errors.New("you can only place orders for your own account")
)

type mutationResolver struct {
	server *Server
}

func (r *mutationResolver) CreateAccount(ctx context.Context, in AccountInput) (*Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	a, err := r.server.accountClient.PostAccount(ctx, in.Name, in.Email, in.Password)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &Account{
		ID:    a.ID,
		Name:  a.Name,
		Email: a.Email,
	}, nil
}

// Login verifies email+password against the account service and, on
// success, mints a JWT here at the gateway (see auth.go) — the account
// service itself never issues or even knows about tokens.
func (r *mutationResolver) Login(ctx context.Context, email string, password string) (*AuthPayload, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	a, err := r.server.accountClient.Authenticate(ctx, email, password)
	if err != nil {
		log.Println(err)
		return nil, errors.New("invalid email or password")
	}

	token, err := signToken(a.ID, a.Email)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &AuthPayload{
		Token: token,
		Account: &Account{
			ID:    a.ID,
			Name:  a.Name,
			Email: a.Email,
		},
	}, nil
}

func (r *mutationResolver) CreateProduct(ctx context.Context, in ProductInput) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	p, err := r.server.catalogClient.PostProduct(ctx, in.Name, in.Description, in.Price)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}, nil
}

func (r *mutationResolver) CreateOrder(ctx context.Context, in OrderInput) (*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Placing an order requires being logged in, and only for your own
	// account — the accountId in the input must match the authenticated
	// token's account.
	authedAccountID, ok := accountIDFromContext(ctx)
	if !ok {
		return nil, ErrUnauthenticated
	}
	if authedAccountID != in.AccountID {
		return nil, ErrForbidden
	}

	var products []order.OrderedProduct
	for _, p := range in.Products {
		if p.Quantity <= 0 {
			return nil, ErrInvalidParameter
		}
		products = append(products, order.OrderedProduct{
			ID:       p.ID,
			Quantity: uint32(p.Quantity),
		})
	}
	o, err := r.server.orderClient.PostOrder(ctx, in.AccountID, products)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &Order{
		ID:         o.ID,
		CreatedAt:  o.CreatedAt,
		TotalPrice: o.TotalPrice,
	}, nil
}
