# Chirpy

Chirpy is a small REST API and static file server: user accounts with Argon2-hashed passwords, JWT access tokens, refresh tokens, short messages (“chirps”) stored in PostgreSQL, and optional webhooks for a premium tier (`is_chirpy_red`). It fits the Boot.dev-style Chirpy project shape (Go standard library HTTP, `sqlc`, `pq`).

## Requirements

- [Go](https://go.dev/dl/) **1.25+** (see `go.mod`)
- [PostgreSQL](https://www.postgresql.org/) with a database you can connect to via `DB_URL`

Optional tooling used in development:

- [sqlc](https://docs.sqlc.dev/) to regenerate typed queries after changing SQL under `sql/queries/` (configuration in `sqlc.yaml`)

## Configuration

Environment variables are read at startup (`github.com/joho/godotenv` loads a local `.env` if present):

| Variable     | Purpose |
|-------------|---------|
| `DB_URL`    | PostgreSQL connection string (passed to `database/sql`). |
| `PLATFORM`  | When set to `dev`, `POST /admin/reset` is allowed (deletes users and resets the file-server hit counter). Any other value returns `403 Forbidden` on that endpoint. |
| `SECRET`    | HMAC key for signing and verifying JWT access tokens (login, chirp auth, etc.). |
| `POLKA_KEY` | Shared secret for `Authorization: ApiKey <key>` on the Polka webhook (must match incoming requests). |

## Database

Schema migrations live under `sql/schema/` (GNU-style `goose` comments). Apply them with your migration tool of choice against the target database before running the server. Tables include `users`, `chirps`, and `refresh_tokens`, with constraints linking chirps to users.

After changing `.sql` query files under `sql/queries/`, run:

```bash
sqlc generate
```

Regenerated Go code is written to `internal/database/` (do not edit `*.sql.go` by hand).

## Run

From the repository root:

```bash
export DB_URL="postgres://user:pass@localhost:5432/chirpy?sslmode=disable"
export PLATFORM="dev"
export SECRET="your-jwt-signing-secret"
export POLKA_KEY="your-polka-webhook-secret"
go run .
```

The server listens on **port 8080**.

## HTTP surface

Unless noted, JSON request bodies should set `Content-Type: application/json`. Error responses use `{"error":"<message>"}`.

### Static files and health

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/app/*` | Files from the repo root (`index.html`, etc.). Each request increments an internal hit counter. |
| `GET` | `/api/healthz` | Returns `200` with body `OK` (plain text). |
| `GET` | `/admin/metrics` | HTML page reporting how many hits `/app/` has served. |
| `POST` | `/admin/reset` | **Dev only** (`PLATFORM=dev`): deletes all users and resets the hit counter. |

### Users and auth

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/users` | — | Create user (`email`, `password`). Returns user JSON without tokens. Status `201` on success. |
| `PUT` | `/api/users` | Bearer JWT | Update `email` and `password` for the authenticated user. |
| `POST` | `/api/login` | — | Login with `email` / `password`. Returns user plus `token` (JWT, ~1h) and `refresh_token`. |
| `POST` | `/api/refresh` | Bearer refresh token | Returns a new access JWT in `{"token":"<jwt>"}` if refresh token exists, is unrevoked, and not expired. |
| `POST` | `/api/revoke` | Bearer refresh token | Revokes the refresh token (`204 No Content` on success). |

### Chirps

Creating and deleting chirps require a valid access JWT (`Authorization: Bearer <jwt>`).

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/chirps` | Body: `{"body":"<text>"}`. Body max length **140 bytes** UTF-8 (Go `len` on string). Forbidden words (`kerfuffle`, `sharbert`, `fornax`) are replaced with `****`; matching is whole “words” split on ASCII spaces. Returns the created chirp. |
| `GET` | `/api/chirps` | Lists chirps ascending by `created_at` by default. Query: `author_id=<uuid>` to filter by user; `sort=desc` for newest first. |
| `GET` | `/api/chirps/{chirpID}` | Single chirp by UUID path segment. |
| `DELETE` | `/api/chirps/{chirpID}` | Deletes if the chirp belongs to the JWT subject; otherwise `403`. |

### Webhooks

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/polka/webhooks` | `Authorization: ApiKey <POLKA_KEY>` | Body: JSON with `event` and `data.user_id`. If `event` is `user.upgraded`, marks that user as Chirpy Red. Other events succeed with no body (`204`). |

## Testing

Tests cover JWT creation/validation, Bearer and API-key parsing, Argon2 password hashing, refresh token formatting, chirp sanitization and length rules, and the health handler.

```bash
go test ./...
```

Packages under `internal/database` are generated from SQL and rely on integration tests externally if you add them; unit tests avoid requiring a running database.

## Project layout

| Path | Role |
|------|------|
| `main.go` | Server wiring, routing, static `/app/`, admin routes. |
| `internal/auth/` | JWT, Argon2id passwords, refresh token bytes, header helpers. |
| `internal/handlers/` | User, chirp, and webhook HTTP handlers and JSON helpers. |
| `internal/database/` | `sqlc`-generated queries and models. |
| `sql/schema/` | Table definitions. |
| `sql/queries/` | `sqlc` query definitions. |
