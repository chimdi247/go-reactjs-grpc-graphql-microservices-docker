package account

import (
	"context"

	"github.com/akhilsharma90/go-graphql-microservice/account/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.AccountServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(
		url,
		grpc.WithInsecure(),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}
	c := pb.NewAccountServiceClient(conn)
	return &Client{conn, c}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostAccount(ctx context.Context, name, email, password string) (*Account, error) {
	r, err := c.service.PostAccount(
		ctx,
		&pb.PostAccountRequest{Name: name, Email: email, Password: password},
	)
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:    r.Account.Id,
		Name:  r.Account.Name,
		Email: r.Account.Email,
	}, nil
}

func (c *Client) GetAccount(ctx context.Context, id string) (*Account, error) {
	r, err := c.service.GetAccount(
		ctx,
		&pb.GetAccountRequest{Id: id},
	)
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:    r.Account.Id,
		Name:  r.Account.Name,
		Email: r.Account.Email,
	}, nil
}

func (c *Client) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	r, err := c.service.GetAccounts(
		ctx,
		&pb.GetAccountsRequest{
			Skip: skip,
			Take: take,
		},
	)
	if err != nil {
		return nil, err
	}
	accounts := []Account{}
	for _, a := range r.Accounts {
		accounts = append(accounts, Account{
			ID:    a.Id,
			Name:  a.Name,
			Email: a.Email,
		})
	}
	return accounts, nil
}

// Authenticate calls the account service to verify email+password. The
// caller (GraphQL gateway) never sees a password hash — only the account
// on success, or an error on failure.
func (c *Client) Authenticate(ctx context.Context, email, password string) (*Account, error) {
	r, err := c.service.Authenticate(
		ctx,
		&pb.AuthenticateRequest{Email: email, Password: password},
	)
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:    r.Account.Id,
		Name:  r.Account.Name,
		Email: r.Account.Email,
	}, nil
}

func (c *Client) GetAccountsCount(ctx context.Context) (uint64, error) {
	r, err := c.service.GetAccountsCount(ctx, &pb.GetAccountsCountRequest{})
	if err != nil {
		return 0, err
	}
	return r.Count, nil
}
