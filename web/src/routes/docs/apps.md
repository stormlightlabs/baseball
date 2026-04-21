# Apps

## Web App

The web app runs as a static site (deployed to a CDN) and calls API endpoints under `/v1`.

- Home page surfaces search + featured queries.
- Docs pages provide API navigation and implementation notes.

## Mobile App

The Flutter app consumes the same `/v1` endpoints:

- Home/meta health
- Players, teams, games
- Leaders and seasons

Keep path usage consistent with web/API clients (`/v1/...`).

For detailed API contracts and schemas, checkout out our Swagger docs at [/explorer](/explorer).
