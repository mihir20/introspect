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
	linearAPIURL = "https://api.linear.app/graphql"
	startDate    = "2025-01-01T00:00:00.000Z"
	endDate      = "2026-02-28T23:59:59.999Z"
)

// GraphQL Response Structures
type GraphQLResponse struct {
	Data   Data    `json:"data"`
	Errors []Error `json:"errors,omitempty"`
}

type Error struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

type Data struct {
	Viewer Viewer `json:"viewer"`
}

type Viewer struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Email          string         `json:"email"`
	AssignedIssues AssignedIssues `json:"assignedIssues"`
}

type AssignedIssues struct {
	Nodes    []Issue  `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

type PageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor"`
}

type Issue struct {
	ID          string   `json:"id"`
	Identifier  string   `json:"identifier"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Priority    int      `json:"priority"`
	Estimate    *float64 `json:"estimate"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
	CompletedAt *string  `json:"completedAt"`
	State       State    `json:"state"`
	Team        Team     `json:"team"`
	Project     *Project `json:"project"`
	Cycle       *Cycle   `json:"cycle"`
	Labels      Labels   `json:"labels"`
	Assignee    User     `json:"assignee"`
}

type State struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Cycle struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
}

type Labels struct {
	Nodes []Label `json:"nodes"`
}

type Label struct {
	Name string `json:"name"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GraphQL Request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// makeGraphQLRequest sends a GraphQL request to the Linear API
func makeGraphQLRequest(apiKey string, query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	requestBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", linearAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiKey)

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
		return nil, fmt.Errorf("GraphQL errors: %v", graphQLResp.Errors)
	}

	return &graphQLResp, nil
}

// getCompletedIssues fetches all completed issues assigned to the authenticated user
func getCompletedIssues(apiKey string) ([]Issue, error) {
	query := `
	query GetCompletedIssues($after: String, $startDate: DateTimeOrDuration!, $endDate: DateTimeOrDuration!) {
		viewer {
			id
			name
			email
			assignedIssues(
				first: 100
				after: $after
				includeArchived: true
				filter: {
					completedAt: { gte: $startDate, lte: $endDate }
				}
			) {
				nodes {
					id
					identifier
					title
					description
					url
					priority
					estimate
					createdAt
					updatedAt
					completedAt
					state {
						id
						name
						type
					}
					team {
						id
						name
						key
					}
					project {
						id
						name
					}
					cycle {
						number
						name
					}
					labels {
						nodes {
							name
						}
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	}
	`

	var allIssues []Issue
	var afterCursor *string

	fmt.Println("Fetching completed issues...")

	for {
		variables := map[string]interface{}{
			"startDate": startDate,
			"endDate":   endDate,
			"after":     afterCursor,
		}

		resp, err := makeGraphQLRequest(apiKey, query, variables)
		if err != nil {
			return nil, err
		}

		issues := resp.Data.Viewer.AssignedIssues.Nodes
		allIssues = append(allIssues, issues...)

		fmt.Printf("Fetched %d issues (total: %d)\n", len(issues), len(allIssues))

		pageInfo := resp.Data.Viewer.AssignedIssues.PageInfo
		if !pageInfo.HasNextPage {
			break
		}
		afterCursor = pageInfo.EndCursor
	}

	// Filter for only completed state types
	var doneIssues []Issue
	for _, issue := range allIssues {
		if issue.State.Type == "completed" {
			doneIssues = append(doneIssues, issue)
		}
	}

	return doneIssues, nil
}

// formatPriority converts priority number to human-readable string
func formatPriority(priority int) string {
	priorityMap := map[int]string{
		0: "No priority",
		1: "Urgent",
		2: "High",
		3: "Medium",
		4: "Low",
	}
	if val, ok := priorityMap[priority]; ok {
		return val
	}
	return "Unknown"
}

// formatDate formats ISO date string to readable format
func formatDate(dateStr *string) string {
	if dateStr == nil {
		return "N/A"
	}
	t, err := time.Parse(time.RFC3339, *dateStr)
	if err != nil {
		return *dateStr
	}
	return t.Format("2006-01-02 15:04:05")
}

// formatDateString formats ISO date string to readable format (non-pointer version)
func formatDateString(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("2006-01-02 15:04:05")
}

// compactIssue is a flattened, minimal representation for JSON export
type compactIssue struct {
	Identifier  string   `json:"identifier"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Team        string   `json:"team"`
	Priority    string   `json:"priority"`
	Estimate    string   `json:"estimate,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Project     string   `json:"project,omitempty"`
	Cycle       string   `json:"cycle,omitempty"`
	CreatedAt   string   `json:"createdAt"`
	CompletedAt string   `json:"completedAt"`
}

// exportToJSON exports issues to a compact JSON file
func exportToJSON(issues []Issue, filename string) error {
	compact := make([]compactIssue, len(issues))
	for i, issue := range issues {
		labels := make([]string, len(issue.Labels.Nodes))
		for j, l := range issue.Labels.Nodes {
			labels[j] = l.Name
		}

		var project, cycle, estimate string
		if issue.Project != nil {
			project = issue.Project.Name
		}
		if issue.Cycle != nil {
			cycle = issue.Cycle.Name
		}
		if issue.Estimate != nil {
			estimate = fmt.Sprintf("%.0f", *issue.Estimate)
		}

		compact[i] = compactIssue{
			Identifier:  issue.Identifier,
			Title:       issue.Title,
			Description: issue.Description,
			URL:         issue.URL,
			Team:        issue.Team.Name,
			Priority:    formatPriority(issue.Priority),
			Estimate:    estimate,
			Labels:      labels,
			Project:     project,
			Cycle:       cycle,
			CreatedAt:   formatDateString(issue.CreatedAt),
			CompletedAt: formatDate(issue.CompletedAt),
		}
	}

	data, err := json.MarshalIndent(compact, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	fmt.Printf("\n✅ Exported %d issues to %s\n", len(issues), filename)
	return nil
}

// exportToCSV exports issues to CSV file
func exportToCSV(issues []Issue, filename string) error {
	if len(issues) == 0 {
		fmt.Println("No issues to export")
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"Identifier", "Title", "URL", "Team", "State", "Priority",
		"Estimate", "Labels", "Project", "Cycle", "Created At",
		"Completed At", "Assignee",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, issue := range issues {
		labels := []string{}
		for _, label := range issue.Labels.Nodes {
			labels = append(labels, label.Name)
		}
		labelsStr := strings.Join(labels, ", ")

		project := "N/A"
		if issue.Project != nil {
			project = issue.Project.Name
		}

		cycle := "N/A"
		if issue.Cycle != nil {
			cycle = issue.Cycle.Name
		}

		estimate := "N/A"
		if issue.Estimate != nil {
			estimate = fmt.Sprintf("%.0f", *issue.Estimate)
		}

		row := []string{
			issue.Identifier,
			issue.Title,
			issue.URL,
			issue.Team.Name,
			issue.State.Name,
			formatPriority(issue.Priority),
			estimate,
			labelsStr,
			project,
			cycle,
			formatDateString(issue.CreatedAt),
			formatDate(issue.CompletedAt),
			issue.Assignee.Name,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	fmt.Printf("✅ Exported %d issues to %s\n", len(issues), filename)
	return nil
}

// printSummary prints a summary of the issues
func printSummary(issues []Issue) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("SUMMARY")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total completed issues: %d\n", len(issues))
	fmt.Println("Date range: January 2025 - February 2026")

	if len(issues) > 0 {
		// Group by team
		teams := make(map[string]int)
		for _, issue := range issues {
			teams[issue.Team.Name]++
		}

		fmt.Println("\nIssues by team:")
		for team, count := range teams {
			fmt.Printf("  %s: %d\n", team, count)
		}

		// Group by priority
		priorities := make(map[string]int)
		for _, issue := range issues {
			priority := formatPriority(issue.Priority)
			priorities[priority]++
		}

		fmt.Println("\nIssues by priority:")
		for priority, count := range priorities {
			fmt.Printf("  %s: %d\n", priority, count)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

// printIssuesTable prints issues in a formatted table
func printIssuesTable(issues []Issue) {
	if len(issues) == 0 {
		fmt.Println("\nNo issues found.")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 120))
	fmt.Printf("%-15s %-50s %-20s %-20s\n", "ID", "Title", "Team", "Completed")
	fmt.Println(strings.Repeat("=", 120))

	for _, issue := range issues {
		identifier := issue.Identifier
		if len(identifier) > 15 {
			identifier = identifier[:15]
		}

		title := issue.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}

		team := issue.Team.Name
		if len(team) > 20 {
			team = team[:20]
		}

		completed := formatDate(issue.CompletedAt)
		if len(completed) > 20 {
			completed = completed[:20]
		}

		fmt.Printf("%-15s %-50s %-20s %-20s\n", identifier, title, team, completed)
	}

	fmt.Println(strings.Repeat("=", 120))
}

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Linear Completed Tickets Extractor")
	fmt.Println(strings.Repeat("=", 60))

	// Check for API key
	apiKey := os.Getenv("LINEAR_API_KEY")
	if apiKey == "" {
		fmt.Println("\n❌ Error: LINEAR_API_KEY environment variable not set!")
		fmt.Println("\nTo set your API key:")
		fmt.Println("  1. Go to Linear Settings > API > Personal API Keys")
		fmt.Println("  2. Create a new API key")
		fmt.Println("  3. Set it as an environment variable:")
		fmt.Println("     export LINEAR_API_KEY='your_api_key_here'")
		os.Exit(1)
	}

	fmt.Printf("\n📅 Searching for completed tickets from %s to %s\n\n", startDate, endDate)

	// Fetch issues
	issues, err := getCompletedIssues(apiKey)
	if err != nil {
		fmt.Printf("❌ Error fetching issues: %v\n", err)
		os.Exit(1)
	}

	// Print results
	printIssuesTable(issues)
	printSummary(issues)

	// Export to files
	if len(issues) > 0 {
		fmt.Println("\n📁 Exporting to files...")

		if err := exportToJSON(issues, "linear_completed_tickets.json"); err != nil {
			fmt.Printf("❌ Error exporting JSON: %v\n", err)
		}

		if err := exportToCSV(issues, "linear_completed_tickets.csv"); err != nil {
			fmt.Printf("❌ Error exporting CSV: %v\n", err)
		}

		fmt.Println("\n✨ Done! Check the output files for full details.")
	} else {
		fmt.Println("\nNo completed issues found in the specified date range.")
	}
}
