# Linear Tickets Extractor (Go)

A Go script to extract all Linear tickets assigned to you that are marked as "Done" between January 2025 and February 2026.

## Features

- ✅ Fetches all completed issues assigned to you
- 📅 Filters by completion date range (January 2025 - February 2026)
- 📊 Exports to both JSON and CSV formats
- 📈 Displays summary statistics (by team, priority)
- 🔄 Handles pagination automatically
- 🎨 Pretty-printed table output

## Prerequisites

- Go 1.16 or higher
- A Linear API key

## Getting Your Linear API Key

1. Go to [Linear Settings](https://linear.app/settings)
2. Navigate to **API** > **Personal API Keys**
3. Click **Create New API Key**
4. Copy the generated key

## Installation

1. Save the script to a file named `linear_tickets_extractor.go`

2. Initialize a Go module (if not already done):
```bash
go mod init linear-extractor
```

## Usage

### Set your API key as an environment variable:

**Linux/macOS:**
```bash
export LINEAR_API_KEY='your_api_key_here'
```

**Windows (Command Prompt):**
```cmd
set LINEAR_API_KEY=your_api_key_here
```

**Windows (PowerShell):**
```powershell
$env:LINEAR_API_KEY="your_api_key_here"
```

### Run the script:

```bash
go run linear_tickets_extractor.go
```

Or build and run:

```bash
go build -o linear-extractor linear_tickets_extractor.go
./linear-extractor
```

## Output

The script generates three types of output:

### 1. Console Output
- A formatted table showing all completed tickets
- Summary statistics (total count, breakdown by team and priority)

### 2. JSON Export (`linear_completed_tickets.json`)
Complete issue data including:
- Issue ID and identifier
- Title and description
- Team, project, and cycle information
- Priority and estimate
- Labels
- Timestamps (created, updated, completed)
- Assignee details
- Direct URL to the issue

### 3. CSV Export (`linear_completed_tickets.csv`)
Tabular format with columns:
- Identifier
- Title
- URL
- Team
- State
- Priority
- Estimate
- Labels
- Project
- Cycle
- Created At
- Completed At
- Assignee

## Customization

### Modify Date Range

Edit the constants at the top of the script:

```go
const (
    startDate = "2025-01-01T00:00:00.000Z"  // Change start date
    endDate   = "2026-02-28T23:59:59.999Z"  // Change end date
)
```

### Change Output Filenames

Modify the function calls in `main()`:

```go
exportToJSON(issues, "my_custom_name.json")
exportToCSV(issues, "my_custom_name.csv")
```

### Filter by Additional Criteria

The script currently filters for:
- Issues assigned to you (`assignedIssues`)
- Completed state type (`state.type == "completed"`)
- Completion date within range

To add more filters, modify the GraphQL query's `filter` object.

## GraphQL Query Details

The script uses Linear's GraphQL API with the following query structure:

```graphql
query GetCompletedIssues($after: String, $startDate: DateTime!, $endDate: DateTime!) {
  viewer {
    assignedIssues(
      first: 100
      after: $after
      filter: {
        completedAt: { gte: $startDate, lte: $endDate }
      }
    ) {
      nodes {
        # Issue fields...
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
}
```

## Troubleshooting

### Error: LINEAR_API_KEY environment variable not set
Make sure you've exported your API key before running the script.

### Error: API request failed with status 401
Your API key is invalid or expired. Generate a new one from Linear settings.

### Error: GraphQL errors
Check the error message for details. Common issues:
- Invalid date format
- Network connectivity problems
- Rate limiting (unlikely with pagination)

### No issues found
Possible reasons:
- No issues were completed in the specified date range
- No issues are assigned to you
- Issues might be in a different state (not "completed" type)

## API Rate Limits

Linear's API has rate limits. This script:
- Fetches 100 issues per request
- Uses pagination to handle large datasets
- Should work fine for most use cases

If you hit rate limits, you can:
- Add delays between requests
- Reduce the page size (change `first: 100` to a smaller number)

## Dependencies

This script uses only Go standard library packages:
- `encoding/json` - JSON marshaling/unmarshaling
- `encoding/csv` - CSV file writing
- `net/http` - HTTP client
- `os` - Environment variables and file operations
- `fmt` - Formatted I/O
- `time` - Date/time parsing and formatting

No external dependencies required!

## Security Notes

- Never commit your API key to version control
- Store API keys securely (environment variables, secrets manager)
- API keys have full access to your Linear workspace
- Revoke keys when no longer needed

## License

MIT License - feel free to modify and use as needed.

## Resources

- [Linear API Documentation](https://linear.app/developers)
- [Linear GraphQL API](https://linear.app/developers/graphql)
- [Apollo Studio (API Explorer)](https://studio.apollographql.com/public/Linear-API/variant/current/home)
