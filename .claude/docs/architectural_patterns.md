# Architectural Patterns

Both extractors (`linear/`, `pull_requests/`) share the same architecture. When adding a new extractor or modifying an existing one, follow these conventions.

## Data Pipeline

Every extractor's `main()` follows the same sequential pipeline:

```
Validate env config → Fetch (paginated) → Display table → Print summary → Export (JSON + CSV)
```

- `linear/linear_tickets_extractor.go:511`
- `pull_requests/pull_requests_extractor.go:447`

Each step is a standalone function with explicit inputs/outputs. No global mutable state — all data flows through function parameters and return values.

## GraphQL Client Pattern

All API communication goes through a single `makeGraphQLRequest()` function per package that handles serialization, HTTP transport, and error parsing.

| | Linear | GitHub |
|---|---|---|
| Function | `linear/linear_tickets_extractor.go:115` | `pull_requests/pull_requests_extractor.go:150` |
| API URL constant | `:16` | `:16` |
| Auth header | `Authorization: <key>` (bare) | `Authorization: Bearer <token>` |
| HTTP timeout | 30 seconds | 30 seconds |

Both accept a raw query string + variables map and return a typed `*GraphQLResponse`. Both check HTTP status codes and GraphQL-level errors before returning.

## Typed GraphQL Response Mapping

GraphQL responses are deserialized into a hierarchy of Go structs that mirror the query shape:

- Linear types: `linear/linear_tickets_extractor.go:22-112`
- GitHub types: `pull_requests/pull_requests_extractor.go:24-97`

Pattern: each GraphQL object maps to its own struct with JSON tags. Nested connections use the `Nodes` array pattern (e.g., `AssignedIssues.Nodes`, `Labels.Nodes`).

## Cursor-Based Pagination

Both extractors implement Relay-style cursor pagination:

- Linear: `getCompletedIssues()` at `linear/linear_tickets_extractor.go:163`
- GitHub: `getMergedPullRequests()` at `pull_requests/pull_requests_extractor.go:199`

Shared pattern:
1. Initialize `afterCursor *string` as nil
2. Loop: build variables map with current cursor, call `makeGraphQLRequest()`
3. Append results to accumulator slice
4. Break when `pageInfo.hasNextPage` is false
5. Otherwise, set `afterCursor = pageInfo.endCursor` and continue

Both fetch 100 items per page and log progress during pagination.

## Two-Layer Data Structs

Each extractor defines two struct layers for the same data:

1. **API response structs** — match GraphQL shape exactly, include nested objects and pointer types for nullable fields
   - `Issue` at `linear/linear_tickets_extractor.go:53`
   - `PullRequest` at `pull_requests/pull_requests_extractor.go:59`

2. **Compact export structs** — flattened representations for JSON/CSV output, with formatted values
   - `compactIssue` at `linear/linear_tickets_extractor.go:299`
   - `compactPR` at `pull_requests/pull_requests_extractor.go:326`

The compact structs denormalize nested fields (e.g., `issue.Team.Name` becomes `Team string`) and convert pointers to formatted strings.

## Multi-Format Export

Both extractors output the same dataset in three formats:

| Format | Linear | GitHub |
|---|---|---|
| Console table | `printIssuesTable()` `:474` | `printPRsTable()` `:269` |
| Summary stats | `printSummary()` `:438` | `printSummary()` `:295` |
| JSON file | `exportToJSON()` `:315` | `exportToJSON()` `:346` |
| CSV file | `exportToCSV()` `:364` | `exportToCSV()` `:388` |

Console tables use fixed-width formatting with `fmt.Sprintf` and field truncation for readability. JSON uses `json.MarshalIndent` for pretty-printing. CSV uses the standard library `encoding/csv` writer.

## Optional Field Handling

Nullable GraphQL fields use Go pointers to distinguish missing values from zero values:

- `*string` for dates like `CompletedAt`, `MergedAt`, and pagination `EndCursor`
- `*float64` for `Estimate`
- `*Project`, `*Cycle` for optional relationships

Nil checks produce `"N/A"` fallbacks in display and export functions. Both packages provide parallel `formatDate(*string)` and `formatDateString(string)` helpers for pointer and value variants.

## Error Handling Convention

- All errors are wrapped with `fmt.Errorf("context: %w", err)` for stack tracing
- Both HTTP status codes and GraphQL-level errors are checked in `makeGraphQLRequest()`
- Environment validation happens early in `main()` with a helpful setup message before any API calls
- Export errors are logged but do not halt execution — each export runs independently

## Adding a New Extractor

To add a new extractor, follow the existing pattern:
1. Create a new directory (e.g., `jira/`) with a single `*_extractor.go` file
2. Define GraphQL/API types, a `makeGraphQLRequest()` or equivalent, and a paginated fetch function
3. Create compact export structs for JSON/CSV output
4. Wire up `main()` following the pipeline: validate → fetch → display → export
5. The Makefile will auto-discover it via `make build-all`; run it with `make run PKG=<name>`