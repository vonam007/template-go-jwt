# template-go-jwt

Minimal Go backend template with JWT authentication and small helpers.

Quick start

1. Copy `.env.example` to `.env` and edit values or set env vars.
2. Run:

```cmd
cd /d d:\GitD\template-go-jwt
go mod tidy
go run .
```

Endpoints

- GET /hello — public
- POST /login — expects JSON {"user_id": "..."}, returns {"token": "..."}
- GET /protected — requires Authorization: Bearer <token>

Docker / Postgres

This project includes a `docker-compose.yml` that starts a Postgres database and builds/runs the app.

To run with Docker Compose (requires Docker Engine and Docker Compose):

```bash
docker compose up --build
```

The compose file exposes the Postgres service as `db` on the app network and provides environment variables to the app so it can connect:

- DB_HOST=db
- DB_PORT=5432
- DB_USER=postgres
- DB_PASSWORD=postgres
- DB_NAME=template

The app reads these env vars via `internal/util/config.LoadDB()` and you can use the returned `PostgresDSN()` when wiring a DB driver (pgx, database/sql + lib/pq) in your storage layer.

Example quick test:

1. Start compose: `docker compose up --build`
2. In another terminal, login as a seeded user `alice` (admin):

```bash
curl -s -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"user_id":"alice"}' | jq
```

3. Use the token to access admin route:

```bash
TOKEN=<token-from-login>
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/admin/
```

Run migrations as a one-shot job (build once, run often)

The compose file includes a `migrate` service that uses the same image as `app` and runs with `RUN_MIGRATE=1`.

Workflow to avoid rebuilding every time:

1. Build the app image once:

```cmd
cd /d d:\GitD\template-go-jwt
docker compose build app
```

2. Start Postgres (detached):

```cmd
docker compose up -d db
```

3. Run migrations against the running DB using the already-built image:

```cmd
docker compose run --rm migrate
```

`migrate` will connect to the `db`, run pending migrations (it records applied migrations in the `migration_records` table) and then exit. There's no need to rebuild your app image to run migrations — only rebuild when you change the app binary.


Speed up Docker builds with vendoring

To avoid downloading dependencies inside the Docker build and make builds reproducible and faster, you can vendor your dependencies and commit the `vendor/` directory into the repository.

1. Generate vendor directory locally:

```cmd
cd /d d:\GitD\template-go-jwt
go mod vendor
```

2. Commit `vendor/` to your repo (we removed `vendor/` from `.gitignore` so this is allowed):

```cmd
git add vendor/
git commit -m "chore: vendor dependencies for faster builds"
```

When `vendor/` is present, the Dockerfile will build using `go build -mod=vendor` and avoid fetching modules from the network inside the container, reducing build time and improving reproducibility.

If you prefer not to commit `vendor/`, Docker will still build but it will download modules in the builder stage (slower).

Next steps

- Add a database layer in `internal/storage` or `internal/db` and wire repository logic.
- Add more auth claims, refresh tokens, and user management as needed.
