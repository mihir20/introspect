# Self Review

CLI tools to extract your completed work — Linear tickets and merged GitHub pull requests — within a configurable date range. Outputs to console table, JSON, and CSV.

Built with Go (standard library only, zero external dependencies).

## Tools

| Package | Description | API |
|---|---|---|
| `linear/` | Completed Linear issues assigned to you | [Linear GraphQL](https://linear.app/developers/graphql) |
| `pull_requests/` | Merged GitHub PRs authored by you | [GitHub GraphQL](https://docs.github.com/en/graphql) |

## Prerequisites

- Go 1.21+
- A [Linear API key](https://linear.app/settings) (for the Linear extractor)
- A [GitHub personal access token](https://github.com/settings/tokens) (for the PR extractor)

## Setup

1. Clone the repo and create a `.env` file at the root:

```bash
export LINEAR_API_KEY='lin_api_...'
export GITHUB_TOKEN='ghp_...'
```

2. Source it before running:

```bash
source .env
```

## Usage

```bash
# Run the Linear extractor (default)
make run

# Run the GitHub PR extractor
make run PKG=pull_requests

# Build all packages
make build-all

# Build and run a specific package
make build-run PKG=linear
```

## All Make Targets

| Command | Description |
|---|---|
| `make run PKG=<name>` | Run a package directly (default: `linear`) |
| `make build PKG=<name>` | Build binary to `bin/<name>` |
| `make build-run PKG=<name>` | Build then execute |
| `make build-all` | Build all packages |
| `make clean` | Remove `bin/`, JSON, and CSV output files |
| `make fmt` | Format all Go code |
| `make deps` | Tidy go modules |
| `make help` | Show available commands |

## Output

Each extractor produces:

1. **Console** — formatted table with summary statistics
2. **JSON** — full structured data (`*_completed_tickets.json` / `*_merged.json`)
3. **CSV** — tabular export (`*_completed_tickets.csv` / `*_merged.csv`)

## Configuration

- **Date range** — hardcoded constants at the top of each extractor's source file
- **Output filenames** — set in `main()` of each extractor
