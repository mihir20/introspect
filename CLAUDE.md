# Linear Completed Tickets Extractor

CLI tool that exports all completed Linear issues assigned to the authenticated user within a configurable date range. Outputs to console table, JSON, and CSV.

## Tech Stack

- **Language:** Go 1.21+ (standard library only, zero external dependencies)
- **API:** Linear GraphQL API (`https://api.linear.app/graphql`)
- **Build:** Make

## Project Structure

```
linear/
  linear_tickets_extractor.go   # Entire application (single-file CLI)
  README.md                     # Detailed feature docs and usage guide
Makefile                        # Build/run/clean commands
go.mod                          # Go module definition
.env                            # LINEAR_API_KEY (not committed)
```

Generated output files (gitignored): `linear_completed_tickets.json`, `linear_completed_tickets.csv`

## Build & Run Commands

| Command | Description |
|---|---|
| `make run` | Run the linear extractor directly |
| `make build` | Build binary to `bin/linear` |
| `make build-run` | Build then execute |
| `make build-all` | Build all packages in the project |
| `make clean` | Remove `bin/`, JSON, and CSV output files |
| `make fmt` | Format all Go code (`go fmt ./...`) |
| `make deps` | Tidy go modules |

Target a specific package: `make run PKG=<name>` (defaults to `linear`).

## Configuration

- **`LINEAR_API_KEY`** env var required at runtime (`linear/linear_tickets_extractor.go:471`)
- Date range is hardcoded as constants (`linear/linear_tickets_extractor.go:15-18`)
- Output filenames are hardcoded in `main()` (`linear/linear_tickets_extractor.go:499-503`)

## Key Entry Points

- `main()` at `linear/linear_tickets_extractor.go:465` — orchestrates fetch, display, and export
- `getCompletedIssues()` at `linear/linear_tickets_extractor.go:163` — paginated GraphQL data fetching
- `makeGraphQLRequest()` at `linear/linear_tickets_extractor.go:115` — low-level HTTP/GraphQL client

## Additional Documentation

When working on this codebase, consult these files for context:

- **[Architectural Patterns](.claude/docs/architectural_patterns.md)** — GraphQL client design, pagination, data flow, error handling conventions
- **[Feature Docs & Usage](linear/README.md)** — Comprehensive usage guide, customization instructions, troubleshooting, API rate limits
