package account

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("account not found")

type Repository interface {
	Close()
	PutAccount(ctx context.Context, a Account) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	ListAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
	CountAccounts(ctx context.Context) (uint64, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &postgresRepository{db}, nil
}

func (r *postgresRepository) Close() {
	r.db.Close()
}

func (r *postgresRepository) Ping() error {
	return r.db.Ping()
}

func (r *postgresRepository) PutAccount(ctx context.Context, a Account) error {
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO accounts(id, name, email, password_hash) VALUES($1, $2, $3, $4)",
		a.ID, a.Name, a.Email, a.PasswordHash,
	)
	return err
}

func (r *postgresRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, name, email, password_hash FROM accounts WHERE id = $1", id)
	a := &Account{}
	if err := row.Scan(&a.ID, &a.Name, &a.Email, &a.PasswordHash); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

func (r *postgresRepository) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, name, email, password_hash FROM accounts WHERE email = $1", email)
	a := &Account{}
	if err := row.Scan(&a.ID, &a.Name, &a.Email, &a.PasswordHash); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

func (r *postgresRepository) ListAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, name, email, password_hash FROM accounts ORDER BY id DESC OFFSET $1 LIMIT $2",
		skip,
		take,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		a := &Account{}
		if err = rows.Scan(&a.ID, &a.Name, &a.Email, &a.PasswordHash); err == nil {
			accounts = append(accounts, *a)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *postgresRepository) CountAccounts(ctx context.Context) (uint64, error) {
	row := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM accounts")
	var count uint64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
