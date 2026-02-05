package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	githubGraphQLURL = "https://api.github.com/graphql"
	searchQuery      = "is:pr author:@me is:merged merged:2025-01-01..2026-02-28"
	startDateDisplay = "January 2025"
	endDateDisplay   = "February 2026"
)

// GraphQL request/response types

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphQLResponse struct {
	Data   Data           `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string        `json:"message"`
	Path    []interface{} `json:"path,omitempty"`
}

type Data struct {
	Search SearchResult `json:"search"`
}

type SearchResult struct {
	IssueCount int               `json:"issueCount"`
	Edges      []PullRequestEdge `json:"edges"`
	PageInfo   PageInfo          `json:"pageInfo"`
}

type PullRequestEdge struct {
	Node   PullRequest `json:"node"`
	Cursor string      `json:"cursor"`
}

type PageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor"`
}

type PullRequest struct {
	Number       int        `json:"number"`
	Title        string     `json:"title"`
	URL          string     `json:"url"`
	Body         string     `json:"body"`
	State        string     `json:"state"`
	MergedAt     *string    `json:"mergedAt"`
	CreatedAt    string     `json:"createdAt"`
	UpdatedAt    string     `json:"updatedAt"`
	Additions    int        `json:"additions"`
	Deletions    int        `json:"deletions"`
	ChangedFiles int        `json:"changedFiles"`
	HeadRefName  string     `json:"headRefName"`
	Repository   Repository `json:"repository"`
	Reviews      CountNode  `json:"reviews"`
	Comments     CountNode  `json:"comments"`
	Labels       Labels     `json:"labels"`
}

type Repository struct {
	Name  string          `json:"name"`
	Owner RepositoryOwner `json:"owner"`
}

type RepositoryOwner struct {
	Login string `json:"login"`
}

type CountNode struct {
	TotalCount int `json:"totalCount"`
}

type Labels struct {
	Nodes []Label `json:"nodes"`
}

type Label struct {
	Name string `json:"name"`
}

// GraphQL query for fetching merged pull requests

const mergedPRsQuery = `
query GetMergedPRs($queryString: String!, $first: Int!, $after: String) {
	search(query: $queryString, type: ISSUE, first: $first, after: $after) {
		issueCount
		edges {
			node {
				... on PullRequest {
					number
					title
					url
					body
					state
					mergedAt
					createdAt
					updatedAt
					additions
					deletions
					changedFiles
					headRefName
					repository {
						name
						owner {
							login
						}
					}
					reviews {
						totalCount
					}
					comments {
						totalCount
					}
					labels(first: 20) {
						nodes {
							name
						}
					}
				}
			}
			cursor
		}
		pageInfo {
			hasNextPage
			endCursor
		}
	}
}
`

// makeGraphQLRequest sends a GraphQL request to the GitHub API
func makeGraphQLRequest(token string, query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	requestBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", githubGraphQLURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "pull-requests-extractor")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var graphQLResp GraphQLResponse
	if err := json.Unmarshal(body, &graphQLResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", graphQLResp.Errors[0].Message)
	}

	return &graphQLResp, nil
}

// getMergedPullRequests fetches all merged PRs using cursor-based pagination
func getMergedPullRequests(token string) ([]PullRequest, error) {
	var allPRs []PullRequest
	var afterCursor *string

	fmt.Println("Fetching merged pull requests...")

	for {
		variables := map[string]interface{}{
			"queryString": searchQuery,
			"first":       100,
			"after":       afterCursor,
		}

		resp, err := makeGraphQLRequest(token, mergedPRsQuery, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch pull requests: %w", err)
		}

		for _, edge := range resp.Data.Search.Edges {
			allPRs = append(allPRs, edge.Node)
		}

		fmt.Printf("Fetched %d PRs (total: %d / %d)\n",
			len(resp.Data.Search.Edges), len(allPRs), resp.Data.Search.IssueCount)

		if !resp.Data.Search.PageInfo.HasNextPage {
			break
		}
		afterCursor = resp.Data.Search.PageInfo.EndCursor
	}

	return allPRs, nil
}

// Helper functions

func formatDate(dateStr *string) string {
	if dateStr == nil {
		return "N/A"
	}
	t, err := time.Parse(time.RFC3339, *dateStr)
	if err != nil {
		return *dateStr
	}
	return t.Format("2006-01-02 15:04")
}

func formatDateString(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("2006-01-02 15:04")
}

func repoFullName(repo Repository) string {
	return repo.Owner.Login + "/" + repo.Name
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// printPRsTable displays pull requests in a formatted console table
func printPRsTable(prs []PullRequest) {
	if len(prs) == 0 {
		fmt.Println("\nNo pull requests found.")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 135))
	fmt.Printf("%-30s %-7s %-42s %-25s %-18s %-10s\n",
		"Repo", "PR#", "Title", "Branch", "Merged At", "+/-")
	fmt.Println(strings.Repeat("=", 135))

	for _, pr := range prs {
		repo := truncate(repoFullName(pr.Repository), 30)
		title := truncate(pr.Title, 42)
		branch := truncate(pr.HeadRefName, 25)
		mergedAt := formatDate(pr.MergedAt)
		changes := fmt.Sprintf("+%d/-%d", pr.Additions, pr.Deletions)

		fmt.Printf("%-30s %-7d %-42s %-25s %-18s %-10s\n",
			repo, pr.Number, title, branch, mergedAt, changes)
	}

	fmt.Println(strings.Repeat("=", 135))
}

// printSummary displays summary statistics about the pull requests
func printSummary(prs []PullRequest) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total merged PRs: %d\n", len(prs))
	fmt.Printf("Date range: %s - %s\n", startDateDisplay, endDateDisplay)

	if len(prs) > 0 {
		repos := make(map[string]int)
		totalAdditions := 0
		totalDeletions := 0

		for _, pr := range prs {
			repos[repoFullName(pr.Repository)]++
			totalAdditions += pr.Additions
			totalDeletions += pr.Deletions
		}

		fmt.Println("\nPRs by repository:")
		for repo, count := range repos {
			fmt.Printf("  %s: %d\n", repo, count)
		}

		fmt.Printf("\nTotal lines added:   +%d\n", totalAdditions)
		fmt.Printf("Total lines deleted: -%d\n", totalDeletions)
	}

	fmt.Println(strings.Repeat("=", 60))
}

// compactPR is a flattened representation for JSON export
type compactPR struct {
	Repository   string   `json:"repository"`
	Number       int      `json:"number"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	URL          string   `json:"url"`
	Branch       string   `json:"branch"`
	State        string   `json:"state"`
	MergedAt     string   `json:"mergedAt"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
	Additions    int      `json:"additions"`
	Deletions    int      `json:"deletions"`
	ChangedFiles int      `json:"changedFiles"`
	Reviews      int      `json:"reviews"`
	Comments     int      `json:"comments"`
	Labels       []string `json:"labels,omitempty"`
}

// exportToJSON exports pull requests to a JSON file
func exportToJSON(prs []PullRequest, filename string) error {
	compact := make([]compactPR, len(prs))
	for i, pr := range prs {
		labels := make([]string, len(pr.Labels.Nodes))
		for j, l := range pr.Labels.Nodes {
			labels[j] = l.Name
		}

		compact[i] = compactPR{
			Repository:   repoFullName(pr.Repository),
			Description:  pr.Body,
			Number:       pr.Number,
			Title:        pr.Title,
			URL:          pr.URL,
			Branch:       pr.HeadRefName,
			State:        pr.State,
			MergedAt:     formatDate(pr.MergedAt),
			CreatedAt:    formatDateString(pr.CreatedAt),
			UpdatedAt:    formatDateString(pr.UpdatedAt),
			Additions:    pr.Additions,
			Deletions:    pr.Deletions,
			ChangedFiles: pr.ChangedFiles,
			Reviews:      pr.Reviews.TotalCount,
			Comments:     pr.Comments.TotalCount,
			Labels:       labels,
		}
	}

	data, err := json.MarshalIndent(compact, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	fmt.Printf("✅ Exported %d pull requests to %s\n", len(prs), filename)
	return nil
}

// exportToCSV exports pull requests to a CSV file
func exportToCSV(prs []PullRequest, filename string) error {
	if len(prs) == 0 {
		fmt.Println("No pull requests to export")
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"Repository", "PR#", "Title", "URL", "Branch", "State",
		"Merged At", "Created At", "Updated At",
		"Additions", "Deletions", "Changed Files",
		"Reviews", "Comments", "Labels",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, pr := range prs {
		labels := make([]string, len(pr.Labels.Nodes))
		for i, l := range pr.Labels.Nodes {
			labels[i] = l.Name
		}
		labelsStr := strings.Join(labels, "; ")

		row := []string{
			repoFullName(pr.Repository),
			fmt.Sprintf("%d", pr.Number),
			pr.Title,
			pr.URL,
			pr.HeadRefName,
			pr.State,
			formatDate(pr.MergedAt),
			formatDateString(pr.CreatedAt),
			formatDateString(pr.UpdatedAt),
			fmt.Sprintf("%d", pr.Additions),
			fmt.Sprintf("%d", pr.Deletions),
			fmt.Sprintf("%d", pr.ChangedFiles),
			fmt.Sprintf("%d", pr.Reviews.TotalCount),
			fmt.Sprintf("%d", pr.Comments.TotalCount),
			labelsStr,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	fmt.Printf("✅ Exported %d pull requests to %s\n", len(prs), filename)
	return nil
}

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("GitHub Merged Pull Requests Extractor")
	fmt.Println(strings.Repeat("=", 60))

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("\n❌ Error: GITHUB_TOKEN environment variable not set!")
		fmt.Println("\nTo set your token:")
		fmt.Println("  1. Go to GitHub Settings > Developer settings > Personal access tokens")
		fmt.Println("  2. Create a new token with 'repo' scope")
		fmt.Println("  3. Set it as an environment variable:")
		fmt.Println("     export GITHUB_TOKEN='your_token_here'")
		os.Exit(1)
	}

	fmt.Printf("\n📅 Searching for merged PRs from %s to %s\n\n", startDateDisplay, endDateDisplay)

	prs, err := getMergedPullRequests(token)
	if err != nil {
		fmt.Printf("❌ Error fetching pull requests: %v\n", err)
		os.Exit(1)
	}

	printPRsTable(prs)
	printSummary(prs)

	if len(prs) > 0 {
		fmt.Println("\n📁 Exporting to files...")

		if err := exportToJSON(prs, "pull_requests_merged.json"); err != nil {
			fmt.Printf("❌ Error exporting JSON: %v\n", err)
		}

		if err := exportToCSV(prs, "pull_requests_merged.csv"); err != nil {
			fmt.Printf("❌ Error exporting CSV: %v\n", err)
		}

		fmt.Println("\n✨ Done! Check the output files for full details.")
	} else {
		fmt.Println("\nNo merged pull requests found in the specified date range.")
	}
}
