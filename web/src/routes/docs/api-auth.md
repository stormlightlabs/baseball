# Authentication & API Key Endpoints

Browser-based OAuth (GitHub + Codeberg) and API key management live inside `AuthRoutes`.
These routes are required for the dashboard, CLI login, and programmatic API access.

## Summary

| Endpoint                           | Dataset | Highlights                                                                                              |
| ---------------------------------- | ------- | ------------------------------------------------------------------------------------------------------- |
| `GET /v1/auth/github`              | -       | Initiates GitHub OAuth; responds with a redirect to github.com including CSRF-protected `state`.        |
| `GET /v1/auth/github/callback`     | -       | Completes GitHub OAuth, creates/updates the local user, and issues the `session_token` cookie.          |
| `GET /v1/auth/codeberg`            | -       | Same as above but points at Codeberg OAuth endpoints.                                                   |
| `GET /v1/auth/codeberg/callback`   | -       | Codeberg callback handler mirroring the GitHub logic.                                                   |
| `POST /v1/auth/logout`             | -       | Revokes the active session token and clears the cookie.                                                 |
| `GET /v1/auth/me`                  | -       | Returns the authenticated `User` object injected by auth middleware (email, avatar, timestamps).        |
| `POST /v1/auth/keys`               | -       | Creates a new API key for the logged-in user; request includes optional `name` and `expires_at`.        |
| `GET /v1/auth/keys`                | -       | Lists all API keys owned by the current user along with metadata (created_at, revoked flag, etc.).      |
| `DELETE /v1/auth/keys/{id}`        | -       | Revokes a specific API key if it belongs to the caller.                                                 |

## Endpoint Notes

### OAuth Entrypoints

- `GET /v1/auth/github` and `GET /v1/auth/codeberg` are front-channel redirects. They set an `oauth_state` cookie (10-minute TTL) used to validate callbacks.
- Callback endpoints exchange the OAuth code for an access token, upsert a local user (email/name/avatar), update `last_login`, and create a session row via `tokenRepo`.
- Successful callbacks set a `session_token` HTTP-only cookie (24-hour TTL) and redirect to `/`.
- Mis-matched or missing `state` cookies produce `500` responses with an "invalid OAuth state" error.

### Session Utilities

- `POST /v1/auth/logout` attempts to delete the backing session record (if present) before clearing the cookie and responding with `{ "message": "logged out" }`.
- `GET /v1/auth/me` simply returns the `core.User` stored on the request context. If middleware did not inject a user, the endpoint returns `500 Unauthorized`.

### API Key Management

- `POST /v1/auth/keys` expects JSON such as:

```json
{
  "name": "prod-cluster",
  "expires_at": "2025-01-01T00:00:00Z"
}
```

Both fields are optional; the handler sets the response status to `201` and returns the API key metadata, the plaintext key (only once), and a warning string.

- `GET /v1/auth/keys` returns an array of key records (ID, name, created/expired timestamps, revoked flag).
- `DELETE /v1/auth/keys/{id}` validates ownership before revoking the key and returns `{ "message": "API key revoked" }`.
