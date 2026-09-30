# Local development — single `docker compose` stack

Everything — the app **and** the full observability stack — runs from one
`docker-compose.yaml` at the repo root.

```bash
docker compose up -d --build
```

First boot takes a few minutes: each Go service resolves its
dependencies via `go mod tidy` *at build time* (see "Why go.mod changed"
below), plus Elasticsearch/Postgres init and JVM-free Go binaries
starting up. Then:

| What | URL | Notes |
|---|---|---|
| Frontend | http://localhost:5173 | browse catalog, sign up, log in, checkout, view orders |
| Admin login | — | **admin@example.com / password123** (seeded automatically, bcrypt-hashed) |
| GraphQL API | http://localhost:8080/graphql | playground at `/playground`, health at `/health` |
| Grafana | http://localhost:3001 | **admin / admin** — Prometheus, Loki, Tempo pre-wired; one dashboard auto-provisioned |
| Prometheus | http://localhost:9090 | metrics + alert rules |
| Alertmanager | http://localhost:9093 | firing alerts, routed to Mailpit |
| Tempo | http://localhost:3200 | trace query API |
| Loki | http://localhost:3100 | log query API |
| Mailpit (mail inbox) | http://localhost:8025 | alert emails land here |
| account_db (Postgres) | localhost:5432 | db/user/pass: `akhil`/`akhil`/`123456` |
| order_db (Postgres) | localhost:5433 | db/user/pass: `akhil`/`akhil`/`123456` |
| catalog_db (Elasticsearch) | http://localhost:9200 | product catalog index |
| cAdvisor | http://localhost:8081 | raw per-container metrics UI |
| node-exporter | http://localhost:9100/metrics | raw host metrics |

## What was missing and had to be built

The original reference project had **no authentication at all** (no
password field, no login) and **no frontend**. Wiring up a real login
flow meant:

- **`accounts` table gained `email` (unique) and `password_hash`
  columns.** `account.proto` gained an `Authenticate` RPC (bcrypt
  verification, entirely inside the account service — the password hash
  never leaves it) and a `GetAccountsCount` RPC (backs the "total users"
  business KPI).
- **JWTs are minted and verified at the GraphQL gateway** (`graphql/auth.go`),
  not by the account service — account only ever answers "is this
  email+password valid," which keeps token concerns at the edge where
  the rest of the API already lives.
- **New GraphQL schema**: `login(email, password): AuthPayload!`,
  `me: Account`, `totalAccounts: Int!`, and `AccountInput`/`Account` now
  carry `email`/`password`. `createOrder` now requires a valid token
  whose account matches the order's `accountId`.
- **`account.pb.go` was hand-edited, not regenerated** — the new fields
  and RPCs follow the exact existing (old-style, struct-tag/reflection
  based) codegen pattern in that file. This is safe because Marshal/
  Unmarshal go entirely through Go struct tags, not the compiled
  descriptor bytes — see the comment block at the top of that file.
  `graphql/generated.go`, by contrast, is regenerated at **Docker build
  time** via the project's own existing `go:generate` directive (gqlgen's
  codegen is far too complex to hand-edit safely) — see
  `docker/graphql.app.dockerfile`.

## Why go.mod / Dockerfiles changed

The project vendored its dependencies (`vendor/`, `-mod vendor` builds)
against Go 1.13 and 2019-era libraries. Adding bcrypt, JWT, and
OpenTelemetry meant either hand-vendoring new packages (fragile, and
their transitive dependency trees would need vendoring too) or letting
the real Go toolchain resolve everything properly. Every Dockerfile now
runs `go mod tidy` at build time (`GOFLAGS=-mod=mod`) instead of building
from a committed `vendor/` folder, and `go.mod` was bumped to `go 1.22`.
This needs network access during `docker build` — that's normal and
already true for pulling base images.

## Admin login

Seeded by `account/cmd/seed-admin/main.go`, a tiny Go program (not a
static `.sql` file) run as a one-off `seed-admin` service. It computes a
**real** bcrypt hash of the password using the same
`golang.org/x/crypto/bcrypt` the account service itself verifies logins
with, then does a plain `INSERT ... ON CONFLICT (email) DO NOTHING`. A
hardcoded hash string would either be wrong (bcrypt embeds a random
salt) or need me to fabricate one — this runs the real algorithm instead.

```
email:    admin@example.com
password: password123
```

Override via `ADMIN_EMAIL` / `ADMIN_NAME` / `ADMIN_PASSWORD` on the
`seed-admin` service in `docker-compose.yaml`.

## What's wired up

- **Tables**: `docker/account-db/schema.sql` and `docker/order-db/schema.sql`
  are mounted into each Postgres container's `/docker-entrypoint-initdb.d/`
  and run automatically on first volume init.
- **Traces**: every Go service wraps its gRPC server/client with
  `otelgrpc` (traces every call, no per-method code) and exports OTLP to
  `otel-collector`, which forwards to `tempo`. The GraphQL gateway's HTTP
  handler is wrapped with `otelhttp` the same way.
- **Metrics**: each service exposes `/metrics` (OTel metrics SDK →
  Prometheus exposition) and `/health` on port 8081 (8080 for the
  gateway's `/health`, since 8080 is already its HTTP port). A shared
  `pkg/telemetry` package provides a gRPC unary interceptor and an HTTP
  middleware that record `app_requests_total{method,status}` and
  `app_request_duration_milliseconds{method}` — explicit, known metric
  names the dashboard queries with confidence, rather than guessing at a
  contrib library's internal metric naming.
- **Business KPI**: `account_total_users` is an OTel observable gauge in
  the account service (`SELECT COUNT(*) FROM accounts`, refreshed on
  every Prometheus scrape) — also exposed as the `totalAccounts` GraphQL
  query, which the frontend's Admin page shows directly.
- **Logs**: unchanged — services still log to stdout; `promtail` tails
  every container's stdout via the Docker socket and ships it to `loki`.
- **Dashboard**: one auto-provisioned dashboard, **"ShopSphere — All
  Services Dashboard,"** covering uptime, request rate, error rates
  (split by gRPC status vs HTTP status, since account/catalog/order and
  the gateway use different status conventions), P95/P99 latency,
  availability SLO, business KPI, and host + per-container CPU/memory.
- **Alerting — Alertmanager**, as requested: Prometheus rules in
  `observability/prometheus/alerts.yml` fire when **host or container
  CPU/memory exceeds 50%**, routed through `alertmanager`, which emails
  via Mailpit (`smtp_smarthost: mailpit:1025` in `alertmanager.yml`) so
  there's somewhere to actually see the alert land — no real mail server
  needed. Check http://localhost:9093 or http://localhost:8025.

## Frontend

React + Vite + hand-written shadcn/ui-style components (Button, Card,
Input, Label, Badge, Dialog, Separator — the actual shadcn CLI needs npm
registry access to scaffold, which wasn't available while building this,
so these are written directly following shadcn's own conventions:
Radix primitives + `class-variance-authority` + a `cn()` `tailwind-merge`
helper). A small hand-rolled `fetch`-based GraphQL client
(`src/lib/api.js`) talks directly to the gateway — no codegen, nothing
to keep in sync beyond the query strings themselves.

Pages: Catalog (browse + add to cart), Cart (checkout → `createOrder`),
Orders (past orders), Login/Signup, and Admin (total users stat +
Grafana link).

## Corrections made along the way

- **`UserRepository`-equivalent auth didn't exist at all** — see "What
  was missing" above.
- **`account.proto`/`account.pb.go`**: added `email`, `password`,
  `Authenticate`, `GetAccountsCount` (see above).
- **Vendoring vs `go mod tidy`**: switched build strategy (see above) —
  otherwise the new dependencies simply couldn't be added.
- **GraphQL gateway had no CORS handling** — added `withCORS` in
  `graphql/main.go` so the frontend (a different origin/port) can call it.
- **Consolidated Dockerfiles** into `docker/` (as requested) — the
  originals lived next to each service; build contexts were updated
  accordingly (all now build from the repo root).

## Troubleshooting

- **First `docker compose up --build` is slow / needs network**: this is
  expected — `go mod tidy` and `go generate` (gqlgen) both run at image
  build time now instead of using a committed `vendor/` folder. Subsequent
  builds are cached by Docker layer caching as usual.
- **Fresh start**: `docker compose down -v` wipes all volumes (Postgres,
  Elasticsearch, Grafana, Prometheus/Loki/Tempo storage) — next `up`
  re-seeds everything.
- **cAdvisor / node-exporter show partial data on macOS or Windows**:
  Docker Desktop runs containers inside a VM, so host-level metrics are
  for that VM, not your physical machine — a Docker Desktop limitation.
- **`account.pb.go` reflection introspection**: the four new message
  types' `Descriptor()` methods point at placeholder indices into the
  old compiled descriptor blob (there's no way to regenerate that blob
  by hand). This has no effect on actual RPC calls (which go through
  struct-tag reflection, not the descriptor), but `grpcurl -list` / server
  reflection on the new RPCs may look slightly off cosmetically.
