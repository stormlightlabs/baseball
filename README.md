# Baseball API

A comprehensive REST API for baseball statistics built with Go, serving data from the Lahman Baseball Database and Retrosheet.

## Quick Start

### Build

```bash
task build
```

### Run the Server

```bash
task server:start
```

The API will be available at <http://localhost:8080>, with interactive documentation at <http://localhost:8080/docs/>.

## Development

## Swagger/OpenAPI Documentation

This project uses [swaggo/swag](https://github.com/swaggo/swag) for API documentation generation.

### Generating Swagger Docs

Use the task command to generate swagger documentation:

```bash
task swagger:generate
```

This will:

1. Generate swagger docs from your API annotations
2. Automatically fix known compatibility issues

### Known Issues

#### LeftDelim/RightDelim Build Errors

When generating swagger docs, swag may generate `LeftDelim` and `RightDelim` fields in `internal/docs/docs.go` that are incompatible with the current version of the swag library, causing build failures:

```log
internal/docs/docs.go:1085:2: unknown field LeftDelim in struct literal of type "github.com/swaggo/swag".Spec
internal/docs/docs.go:1086:2: unknown field RightDelim in struct literal of type "github.com/swaggo/swag".Spec
```

## Available Tasks

Run `task --list` to see all available tasks.

## Attribution

This project uses data from:

- **Lahman Baseball Database**: The information used here was obtained free of charge from and is copyrighted by Sean Lahman.
[SABR Lahman Database](https://sabr.org/lahman-database/)
- **Retrosheet**: The information used here was obtained free of charge from and is copyrighted by Retrosheet.
[Retrosheet.org](https://www.retrosheet.org/)
