# Introspect

CLI tools to extract your completed work — Linear tickets and merged GitHub pull requests — within a configurable date range. Outputs to console table, JSON, and CSV.

## Tech Stack

- **Language:** Go 1.21+ (standard library only, zero external dependencies)
- **APIs:** Linear GraphQL, GitHub GraphQL
- **Build:** Make

## Project Structure

```
linear/
  linear_tickets_extractor.go   # Completed Linear issues extractor
pull_requests/
  pull_requests_extractor.go    # Merged GitHub PRs extractor
Makefile                        # Build/run/clean (supports PKG= targeting)
go.mod                          # Go module definition
.env                            # API keys (not committed, see .env.sample)
```

Each package is a self-contained single-file CLI following the same architecture. Generated output files (JSON, CSV) are gitignored.

## Build & Run Commands

| Command | Description |
|---|---|
| `make run` | Run default package (linear) |
| `make run PKG=pull_requests` | Run the GitHub PR extractor |
| `make build PKG=<name>` | Build binary to `bin/<name>` |
| `make build-run PKG=<name>` | Build then execute |
| `make build-all` | Build all packages |
| `make clean` | Remove `bin/`, JSON, and CSV output files |
| `make fmt` | Format all Go code (`go fmt ./...`) |
| `make deps` | Tidy go modules |

## Configuration

Each extractor reads its API key from an environment variable and has hardcoded date range constants:

| Package | Env Var | Date Constants | Output Filenames |
|---|---|---|---|
| `linear` | `LINEAR_API_KEY` (`linear/linear_tickets_extractor.go:517`) | `linear/linear_tickets_extractor.go:15-18` | Set in `main()` at `:545-549` |
| `pull_requests` | `GITHUB_TOKEN` (`pull_requests/pull_requests_extractor.go:452`) | `pull_requests/pull_requests_extractor.go:15-19` | Set in `main()` at `:477-481` |

## Key Entry Points

**Linear extractor** (`linear/linear_tickets_extractor.go`):
- `main()` at `:511` — orchestrates fetch, display, and export
- `getCompletedIssues()` at `:163` — paginated GraphQL data fetching
- `makeGraphQLRequest()` at `:115` — HTTP/GraphQL client

**PR extractor** (`pull_requests/pull_requests_extractor.go`):
- `main()` at `:447` — orchestrates fetch, display, and export
- `getMergedPullRequests()` at `:199` — paginated GraphQL data fetching
- `makeGraphQLRequest()` at `:150` — HTTP/GraphQL client

## Additional Documentation

When working on this codebase, consult these files for context:

- **[Architectural Patterns](.claude/docs/architectural_patterns.md)** — Cross-cutting patterns shared by both extractors: GraphQL client design, cursor pagination, data pipeline, multi-format export, error handling conventions
- **[README](README.md)** — User-facing setup, usage guide, and Make targets