# Chirpy

Chirpy is a lightweight microblogging API written in Go. It exposes a small REST API for creating users, logging in, managing short “chirps,” and handling token-based auth with refresh tokens. It also includes a minimal static frontend (served at `/app/`) and a small admin metrics page.

## What this project does

- Serves a simple HTTP API for users and chirps
- Uses Postgres for persistence and `sqlc`-generated queries
- Uses JWT access tokens plus refresh tokens for authentication
- Supports a “Chirpy Red” upgrade via a webhook
- Serves static assets from the project root at `/app/`

## Requirements

- Go 1.20+ (or any version compatible with `go.mod`)
- PostgreSQL 13+ (or any version compatible with lib/pq)

## Configuration

Chirpy uses environment variables to configure runtime behavior:

- `DB_URL` (required): Postgres connection string
- `PLATFORM` (required): Used to gate admin reset (`dev` enables `/admin/reset`)
- `JWT_SECRET` (optional but recommended): Secret used to sign JWTs
- `POLKA_KEY` (required): API key used to validate the Polka webhook

Example:

```bash
export DB_URL="postgres://user:pass@localhost:5432/chirpy?sslmode=disable"
export PLATFORM="dev"
export JWT_SECRET="your-jwt-secret"
export POLKA_KEY="your-polka-api-key"
```

## Database setup

The migration files are Goose migrations and should be applied with Goose:

```bash
goose -dir sql/schema postgres "$DB_URL" up
```

## Install and run

```bash
go mod download
go run .
```

By default the server listens on `:8080`.

## Routes

All API responses are JSON unless otherwise noted. When authentication is required, include `Authorization: Bearer <token>`.

### Health and admin

- `GET /api/healthz`
  - Returns `200 OK` with `OK` text.

- `GET /admin/metrics`
  - Returns an HTML page showing the total number of `/app/` file server hits.

- `POST /admin/reset`
  - Resets the file server hit counter and truncates users.
  - Only works when `PLATFORM=dev`.

### Static frontend

- `GET /app/`
  - Serves static files from the project root (e.g., `index.html`).

### Users and auth

- `POST /api/users`
  - Create a user.
  - Body:
    ```json
    {
      "email": "user@example.com",
      "password": "plaintext-password"
    }
    ```
  - Response: user object with `id`, `email`, `created_at`, `updated_at`, `is_chirpy_red`.

- `PUT /api/users`
  - Update the authenticated user.
  - Auth: `Authorization: Bearer <access_token>`
  - Body:
    ```json
    {
      "email": "new@example.com",
      "password": "new-password"
    }
    ```
  - Response: updated user object.

- `POST /api/login`
  - Login and return tokens.
  - Body:
    ```json
    {
      "email": "user@example.com",
      "password": "plaintext-password"
    }
    ```
  - Response:
    ```json
    {
      "id": "...",
      "email": "user@example.com",
      "created_at": "...",
      "updated_at": "...",
      "is_chirpy_red": false,
      "token": "<access_jwt>",
      "refresh_token": "<refresh_token>"
    }
    ```

- `POST /api/refresh`
  - Exchange a refresh token for a new access token.
  - Auth: `Authorization: Bearer <refresh_token>`
  - Response:
    ```json
    { "token": "<new_access_jwt>" }
    ```

- `POST /api/revoke`
  - Revoke a refresh token.
  - Auth: `Authorization: Bearer <refresh_token>`
  - Response: `204 No Content`

### Chirps

- `POST /api/chirps`
  - Create a chirp for the authenticated user.
  - Auth: `Authorization: Bearer <access_token>`
  - Body:
    ```json
    { "body": "hello chirpy" }
    ```
  - Notes:
    - Max length: 140 characters
    - Filters banned words: `kerfuffle`, `sharbert`, `fornax` (case-insensitive), replacing them with `****`
  - Response: chirp object with `id`, `user_id`, `body`, `created_at`, `updated_at`.

- `GET /api/chirps`
  - List chirps.
  - Query params:
    - `author_id` (optional UUID): filter by author
    - `sort` (optional): `asc` (default) or `desc`
  - Response: array of chirp objects.

- `GET /api/chirps/{chirpID}`
  - Retrieve a single chirp by ID.
  - Response: chirp object.

- `DELETE /api/chirps/{chirpID}`
  - Delete a chirp by ID.
  - Auth: `Authorization: Bearer <access_token>`
  - Only the chirp author can delete.
  - Response: `204 No Content`

### Polka webhook

- `POST /api/polka/webhooks`
  - Marks a user as “Chirpy Red.”
  - Auth: `Authorization: ApiKey <POLKA_KEY>`
  - Body:
    ```json
    {
      "event": "user.upgraded",
      "data": { "user_id": "<uuid>" }
    }
    ```
  - If the `event` is not `user.upgraded`, the handler responds with `204 No Content` and does nothing.

## Example usage

Create a user, log in, and post a chirp:

```bash
curl -s -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret"}'

login=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret"}')

token=$(echo "$login" | jq -r '.token')

curl -s -X POST http://localhost:8080/api/chirps \
  -H "Authorization: Bearer '$token'" \
  -H "Content-Type: application/json" \
  -d '{"body":"hello chirpy"}'
```
