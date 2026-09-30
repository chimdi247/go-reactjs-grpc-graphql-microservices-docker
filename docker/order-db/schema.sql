-- Creates every table the order service needs. Mounted into Postgres's
-- /docker-entrypoint-initdb.d/ (see docker-compose.yaml), so it runs
-- automatically the first time the order_db volume is initialized.
CREATE TABLE IF NOT EXISTS orders (
  id CHAR(27) PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL,
  account_id CHAR(27) NOT NULL,
  total_price MONEY NOT NULL
);

CREATE TABLE IF NOT EXISTS order_products (
  order_id CHAR(27) REFERENCES orders (id) ON DELETE CASCADE,
  product_id CHAR(27),
  quantity INT NOT NULL,
  PRIMARY KEY (product_id, order_id)
);

CREATE INDEX IF NOT EXISTS idx_orders_account_id ON orders (account_id);
