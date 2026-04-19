# Apps

## Web App

The web app runs as a static site and calls API endpoints under `/v1`.

- Home page surfaces search + featured queries.
- Docs pages provide API navigation and implementation notes.
- Account page (`/account`) handles OAuth sign-in and API key management.

## Mobile App

The Flutter app consumes the same `/v1` endpoints:

- Home/meta health
- Players, teams, games
- Leaders and seasons

Keep path usage consistent with web/API clients (`/v1/...`).

## Authentication

- OAuth entrypoints:
  - `GET /v1/auth/github`
  - `GET /v1/auth/codeberg`
- Session + API key endpoints:
  - `GET /v1/auth/me`
  - `POST /v1/auth/keys`
  - `GET /v1/auth/keys`
  - `DELETE /v1/auth/keys/{id}`

For API contracts and schemas, use Swagger at [/explorer](/explorer).
