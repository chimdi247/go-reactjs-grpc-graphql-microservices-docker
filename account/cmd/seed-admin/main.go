// seed-admin creates the default admin login directly in the accounts
// table so there's something to log in with right after
// `docker compose up`.
//
// Why a small Go program instead of a static .sql file: the password
// must be bcrypt-hashed, and a bcrypt hash embeds a random salt — there
// is no way to write a correct, verifiable hash of "password123" as a
// plain string without actually running bcrypt. This runs the exact
// same golang.org/x/crypto/bcrypt the account service itself uses to
// verify logins, then does a plain SQL INSERT — so the *mechanism* is
// still SQL, just parameterized with a hash computed for real instead
// of a fabricated one that would silently make login impossible.
//
// Idempotent (ON CONFLICT (email) DO NOTHING): safe to run on every
// `docker compose up`.
package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/ksuid"
	"github.com/tinrab/retry"
	"golang.org/x/crypto/bcrypt"
)

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	email := getenv("ADMIN_EMAIL", "admin@example.com")
	name := getenv("ADMIN_NAME", "Admin")
	password := getenv("ADMIN_PASSWORD", "password123")

	var db *sql.DB
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		db, err = sql.Open("postgres", databaseURL)
		if err != nil {
			log.Println(err)
			return err
		}
		if err = db.Ping(); err != nil {
			log.Println(err)
		}
		return err
	})
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Wait for the accounts table to exist (created by
	// docker/account-db/schema.sql on first Postgres init — this
	// program can start racing against that on a fresh volume).
	retry.ForeverSleep(2*time.Second, func(_ int) error {
		_, err := db.ExecContext(ctx, "SELECT 1 FROM accounts LIMIT 1")
		if err != nil {
			log.Println("waiting for accounts table:", err)
		}
		return err
	})

	var existingID string
	err := db.QueryRowContext(ctx, "SELECT id FROM accounts WHERE email = $1", email).Scan(&existingID)
	if err == nil {
		log.Printf("admin account already exists (id=%s), skipping seed", existingID)
		return
	}
	if err != sql.ErrNoRows {
		log.Fatalf("failed to check for existing admin account: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash admin password: %v", err)
	}

	id := ksuid.New().String()
	_, err = db.ExecContext(
		ctx,
		"INSERT INTO accounts (id, name, email, password_hash) VALUES ($1, $2, $3, $4) ON CONFLICT (email) DO NOTHING",
		id, name, email, string(hash),
	)
	if err != nil {
		log.Fatalf("failed to seed admin account: %v", err)
	}

	log.Printf("seeded admin account: %s", email)
}
