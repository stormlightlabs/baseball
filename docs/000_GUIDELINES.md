You can think of this as a three-layer project:

1. **Raw data → Postgres** (ETL for Lahman + Retrosheet)
2. **Postgres → Go domain model** (queries, joins, views)
3. **Go → HTTP API** (versioned, documented, cached)

I’ll walk through a concrete, opinionated plan with tools + example code.

---

## 1. Understand the two data sources

**Lahman**:

* Season- and career-level stats: batting, pitching, fielding, teams, managers, etc., 1871–2024. ([SABR][1])
* Official documentation lists all tables and fields; the R `Lahman` package site has a nice schema diagram. ([Lahman Data Package][2])

**Retrosheet**:

* Game-level and play-by-play data, including event files and processed CSVs (daily logs and parsed play-by-play). ([Retrosheet][3])
* Chadwick tools (cwevent, cwgame, etc.) convert raw event files into nice CSVs. ([Chadwick][4])

**Opinionated split:**

* **Lahman** → "reference + aggregate stats" API (Players, Teams, Seasons, Leaderboards).
* **Retrosheet** → "games + play-by-play" API (Games, Game logs, Events, Plate appearances).

## 2. Choose a database schema (don’t reinvent the wheel)

I strongly recommend Postgres and reusing community schemas instead of designing every table from scratch.

### Lahman schema

Use an existing Postgres DDL for Lahman, e.g.:

* `Baseball-PostgreSQL` repo – has `lahman/ddl` with CREATE TABLE scripts tuned for Postgres. ([GitHub][5])
* There are similar schemas in other repos and in the Lahman MySQL scripts referenced by the R package docs. ([GitHub][6])

This immediately gives you:

* `people`, `batting`, `pitching`, `fielding`, `teams`, `parks`, etc.
* Well-typed foreign keys and indexes.

### Retrosheet schema

Same `Baseball-PostgreSQL` repo has `retrosheet/ddl` for event-level and game-level tables. ([GitHub][5])

If you prefer not to use event files directly, Retrosheet also publishes **parsed CSVs** for plays and game logs, which are easier to load into relational tables. ([Retrosheet][7])

### Joining Lahman & Retrosheet

Chadwick Bureau + Retrosheet provide consistent IDs:

* Player IDs, team IDs, ballpark codes, franchise IDs, etc. ([Retrosheet][3])

Use those to build:

* `player_id_map` (if you need Lahman `playerID` ↔ Retrosheet `retroID`)
* `team_franchise_map` (if you want modern franchise rollups)

## 3. ETL: getting data into Postgres with Go

### 3.1. Download scripts

Create a small Go command (e.g. `cmd/fetch-data`) that:

* Downloads Lahman CSV archive from SABR/Lahman. ([SABR][1])
* Downloads Retrosheet daily logs and/or parsed play-by-play CSVs. ([Retrosheet][8])

### 3.2. Load CSV into Postgres

For speed, use `COPY` instead of row-by-row inserts.

Pseudo-code for a generic loader:

```go
package ingest

import (
    "database/sql"
    "fmt"
    "os"

    _ "github.com/lib/pq"
)

func CopyCSV(db *sql.DB, table, csvPath string) error {
    f, err := os.Open(csvPath)
    if err != nil { return err }
    defer f.Close()

    // assumes header row matches column order in table
    copyStmt := fmt.Sprintf("COPY %s FROM STDIN WITH (FORMAT csv, HEADER true)", table)
    conn, err := db.Conn(context.Background())
    if err != nil { return err }
    defer conn.Close()

    // pq has CopyIn, pgx has CopyFrom. Use one or the other depending on driver.
    // With pgx:
    rawConn := conn.Raw(func(driverConn interface{}) error {
        pgxConn := driverConn.(*pgx.Conn)
        _, err := pgxConn.PgConn().CopyFrom(context.Background(), []byte(copyStmt), f)
        return err
    })

    return rawConn
}
```

You’d then write thin wrappers:

* `ingestLahman(db)` – loops over Lahman CSVs → `CopyCSV(db, "lahman.batting", "Batting.csv")`, etc.
* `ingestRetrosheetPlays(db)` – loops over parsed play CSVs and game log CSVs → `CopyCSV(db, "retro.plays", "allplays.csv")`, etc.

### 3.3. ETL pipeline structure

* `cmd/ingest-lahman/main.go` – creates Postgres schema (applying DDL), then loads all Lahman CSVs.
* `cmd/ingest-retrosheet/main.go` – applies Retrosheet DDL, loads Retrosheet CSVs.
* Optional: a `Makefile` or `Taskfile.yml` with tasks like `make db-setup`, `make ingest-all`.

---

## 4. Go API design

### 4.1. Tech choices

Opinionated set:

* **Router**: standard library
* **DB layer**: hand rolled
* **Config**: Viper or just env vars + a small config struct.
* **Migrations**: hand rolled

### 4.2. API surface (v1)

Think of two "modules": **reference** and **game/events**.

**Reference (Lahman)**:

* `GET /v1/players`

    * Filters: `name`, `debut_year`, `team`, `position`, `page`, `per_page`.
* `GET /v1/players/{player_id}`

    * Returns Lahman `People` row + aggregated career stats (batting, pitching, fielding).
* `GET /v1/players/{player_id}/seasons`

    * Year-by-year stats (batting & pitching).
* `GET /v1/teams` – All franchises, with filters by league and year.
* `GET /v1/seasons/{year}/leaders`

    * Query params: `stat=HR&min_pa=502&type=batting`.

**Game / play-by-play (Retrosheet)**

* `GET /v1/games`

    * Filters: `date`, `home_team`, `away_team`, `season`, `park`.
* `GET /v1/games/{game_id}`

    * Basic box score info + summary lines.
* `GET /v1/games/{game_id}/events`

    * Plate appearance / play-by-play rows ordered by inning, sequence.
* `GET /v1/players/{player_id}/game-logs`

    * Game-by-game lines from Retrosheet logs.

You can back these endpoints with Postgres **views** that join Lahman and Retrosheet, so handlers remain simple.

### 4.3. Example: simple player handler in Go

Assume:

* `sqlc` generated package `db` with `GetPlayerByID(ctx, playerID string) (db.Player, error)`.

```go
package api

import (
    "database/sql"
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/yourorg/baseballapi/internal/db"
)

type Server struct {
    router *chi.Mux
    q      *db.Queries
}

func NewServer(dbConn *sql.DB) *Server {
    s := &Server{
        router: chi.NewRouter(),
        q:      db.New(dbConn),
    }
    s.routes()
    return s
}

func (s *Server) routes() {
    s.router.Get("/v1/players/{id}", s.handleGetPlayer)
    // add more routes here
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.router.ServeHTTP(w, r)
}

func (s *Server) handleGetPlayer(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    ctx := r.Context()

    p, err := s.q.GetPlayerByID(ctx, id)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(p)
}
```

`main.go` just wires DB + server:

```go
func main() {
    dbURL := os.Getenv("DATABASE_URL")
    dbConn, err := sql.Open("postgres", dbURL)
    if err != nil { log.Fatal(err) }

    s := api.NewServer(dbConn)
    log.Println("listening on :8080")
    http.ListenAndServe(":8080", s)
}
```

---

## 5. Exposing joined Lahman + Retrosheet data

Once both schemas are loaded, create **views** that the API reads from.

### Example: career batting summary view

```sql
CREATE VIEW stats_career_batting AS
SELECT
  b.playerid,
  SUM(b.ab)  AS ab,
  SUM(b.h)   AS h,
  SUM(b.hr)  AS hr,
  SUM(b.bb)  AS bb,
  SUM(b.so)  AS so,
  SUM(b.r)   AS r,
  SUM(b.rbi) AS rbi
FROM lahman_batting b
GROUP BY b.playerid;
```

You can also create a **crosswalk** view that links Retrosheet and Lahman IDs if needed, using ID mapping resources or manual curation based on community projects (Chadwick/Retrosheet mapping resources). ([Retrosheet][9])

Your Go queries then just hit these views as if they’re tables.

---

## 6. Documentation & versioning

* Put everything under `/v1/...` to give yourself room to evolve.
* Use **OpenAPI** (e.g. `kin-openapi` or `swaggo`) to generate a JSON spec and host interactive docs at `/docs`.
* Provide clear **attribution**:

    * Lahman: mention SABR + Lahman database, including copyright and free-use statement. ([SABR][1])
    * Retrosheet: their terms require explicit credit: "The information used here was obtained free of charge from and is copyrighted by Retrosheet." ([GitHub][10])

---

## 7. Suggested first milestone

If you want a pragmatic progression:

1. **M1 – Lahman only**

   * Use an existing Lahman Postgres schema.
   * Implement ETL from CSV → Postgres.
   * Expose `/v1/players`, `/v1/players/{id}`, `/v1/seasons/{year}/leaders`.

2. **M2 – Retrosheet game logs**

   * Load Retrosheet daily logs CSV into `games` / `game_logs` tables. ([Retrosheet][8])
   * Expose `/v1/games` and `/v1/games/{id}`.

3. **M3 – Play-by-play**

   * Use Chadwick (or parsed play CSVs) to populate `events` / `plays`. ([Chadwick][4])
   * Expose `/v1/games/{id}/events`.

4. **M4 – Joined endpoints**

   * Add endpoints that combine Lahman’s season stats with Retrosheet game logs (e.g. "show all games where a player hit 3+ homers").

---

If you tell me what kind of queries you personally care about (dashboards, per-pitch stuff, leaderboards, projections, etc.), I can help you design a very targeted v1 schema and the exact SQL/sqlc queries to back each endpoint.

[1]: https://sabr.org/lahman-database/ "Lahman Baseball Database"
[2]: https://lahman.r-forge.r-project.org/doc/ "Documentation of the Lahman package"
[3]: https://www.retrosheet.org/game.htm "Retrosheet Event Files"
[4]: https://chadwick.readthedocs.io/ "Chadwick: Software Tools for Game-Level Baseball Data ..."
[5]: https://github.com/davidbmitchell/Baseball-PostgreSQL "davidbmitchell/Baseball-PostgreSQL"
[6]: https://github.com/michaeljaltamirano/lahman-baseball-database-2016-postgresql "2016 Lahman Baseball Database for PostgreSQL"
[7]: https://retrosheet.org/downloads/plays.html "Parsed Play-by-Play Data"
[8]: https://www.retrosheet.org/downloads/csvdownloads.html "Daily Logs (CSV Files) Available for Download"
[9]: https://www.retrosheet.org/resources/resources1.html "Data Resources"
[10]: https://github.com/chadwickbureau/retrosheet "chadwickbureau/retrosheet: Enhanced version of ..."
