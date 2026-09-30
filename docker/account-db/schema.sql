-- Creates every table the account service needs. Mounted into Postgres's
-- /docker-entrypoint-initdb.d/ (see docker-compose.yaml), so it runs
-- automatically the first time the account_db volume is initialized.
CREATE TABLE IF NOT EXISTS accounts (
  id CHAR(27) PRIMARY KEY,
  name VARCHAR(64) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_accounts_email ON accounts (email);
