# Architectural Patterns

## GraphQL Client Pattern

All API communication goes through a single generic request function that handles serialization, HTTP transport, and error parsing.

- **Request function:** `makeGraphQLRequest()` at `linear/linear_tickets_extractor.go:115`
- Accepts raw query string + variables map, returns typed `*GraphQLResponse`
- Uses `GraphQLRequest` struct for serialization (`linear/linear_tickets_extractor.go:109-112`)
- HTTP client has a 30-second timeout (`linear/linear_tickets_extractor.go:134`)
- Auth via `Authorization` header (bare API key, no `Bearer` prefix) (`linear/linear_tickets_extractor.go:132`)

## Typed GraphQL Response Mapping

GraphQL responses are deserialized into a hierarchy of Go structs that mirror the query shape. All response types are defined at `linear/linear_tickets_extractor.go:22-112`.

Pattern: each GraphQL object maps to its own struct with JSON tags. Nested connections use the `Nodes` array pattern (e.g., `AssignedIssues.Nodes`, `Labels.Nodes`).

## Cursor-Based Pagination

Pagination follows the Relay connection spec used by Linear's API.

- Implemented in `getCompletedIssues()` at `linear/linear_tickets_extractor.go:226-253`
- Fetches 100 items per page (`first: 100` in the GraphQL query)
- Loops until `pageInfo.hasNextPage` is false
- Passes `endCursor` as the `$after` variable for the next page
- Results are accumulated into a single `allIssues` slice

## Two-Phase Filtering

Data filtering happens at two levels:

1. **Server-side (GraphQL):** Date range filter via `completedAt: { gte: $startDate, lte: $endDate }` (`linear/linear_tickets_extractor.go:174`)
2. **Client-side (Go):** State type filter `issue.State.Type == "completed"` (`linear/linear_tickets_extractor.go:256-261`)

This is necessary because the GraphQL date filter alone may include issues in non-"completed" states.

## Optional Field Handling

Optional/nullable fields use Go pointers to distinguish missing values from zero values:

- `*string` for `CompletedAt`, `EndCursor` (`linear/linear_tickets_extractor.go:50,63`)
- `*float64` for `Estimate` (`linear/linear_tickets_extractor.go:60`)
- `*Project`, `*Cycle` for optional relationships (`linear/linear_tickets_extractor.go:66-67`)

Nil checks produce `"N/A"` fallbacks in display and export functions (e.g., `formatDate()` at `linear/linear_tickets_extractor.go:282-291`).

## Error Handling Convention

- All errors are wrapped with `fmt.Errorf("context: %w", err)` for stack tracing
- Both HTTP status codes and GraphQL-level errors are checked (`linear/linear_tickets_extractor.go:146-157`)
- Environment validation happens early in `main()` before any API calls (`linear/linear_tickets_extractor.go:471-480`)
- Export errors are logged but do not halt execution — each export runs independently (`linear/linear_tickets_extractor.go:499-505`)

## Multi-Format Export

The same dataset is exported in three formats from `main()`:

1. **Console table** via `printIssuesTable()` (`linear/linear_tickets_extractor.go:428`) — fixed-width columns, truncated for readability
2. **JSON file** via `exportToJSON()` (`linear/linear_tickets_extractor.go:303`) — pretty-printed with `json.MarshalIndent`
3. **CSV file** via `exportToCSV()` (`linear/linear_tickets_extractor.go:318`) — standard library `encoding/csv` writer

## Functional Decomposition

The application follows a pipeline architecture in `main()`:

```
Validate config -> Fetch (paginated) -> Filter -> Display -> Export (JSON + CSV)
```

Each step is a standalone function with explicit inputs/outputs. No global mutable state — all data flows through function parameters and return values.

## Data Formatting Helpers

Reusable formatters handle presentation concerns:

- `formatPriority()` (`linear/linear_tickets_extractor.go:267`) — maps integer priority to label via lookup map
- `formatDate()` / `formatDateString()` (`linear/linear_tickets_extractor.go:282,294`) — pointer/value variants for ISO 8601 to human-readable conversion