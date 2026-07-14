package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <tool>",
	Short: "Search for tools on GitHub",
	Long: `Search GitHub for tools matching the query.

Searches GitHub releases for popular development tools. Shows matching
projects with repository links and latest release information.

Examples:
  styx search ripgrep     # Find ripgrep repositories
  styx search "dev tool"  # Search for specific tool type`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSearch,
}

type GitHubSearchResult struct {
	Items []GitHubRepo `json:"items"`
}

type GitHubRepo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	URL         string `json:"html_url"`
	Description string `json:"description"`
	Stars       int    `json:"stargazers_count"`
	UpdatedAt   string `json:"updated_at"`
	Language    string `json:"language"`
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")

	// Search GitHub for tool
	results, err := searchGitHub(query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Printf("No results found for: %s\n", query)
		return nil
	}

	// Display results
	fmt.Printf("Found %d results for '%s':\n\n", len(results), query)
	displaySearchResults(results)

	return nil
}

func searchGitHub(query string) ([]GitHubRepo, error) {
	// Build search query: look for releases and binaries
	searchQuery := fmt.Sprintf("%s stars:>50 language:Go OR language:Rust released:>2023-01-01", url.QueryEscape(query))

	// GitHub API endpoint
	endpoint := fmt.Sprintf("https://api.github.com/search/repositories?q=%s&sort=stars&order=desc&per_page=10", searchQuery)

	// Create request with timeout
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add User-Agent (required by GitHub API)
	req.Header.Set("User-Agent", "styx-cli/0.1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var result GitHubSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Filter results: prefer compiled binaries (Go/Rust), exclude forks
	var filtered []GitHubRepo
	for _, repo := range result.Items {
		if repo.Language == "Go" || repo.Language == "Rust" || repo.Language == "C" {
			filtered = append(filtered, repo)
		}
	}

	// Sort by stars (descending)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Stars > filtered[j].Stars
	})

	return filtered, nil
}

func displaySearchResults(repos []GitHubRepo) {
	for i, repo := range repos {
		if i >= 10 {
			break // Limit to 10 results
		}

		fmt.Printf("%d. %s\n", i+1, repo.FullName)
		if repo.Description != "" {
			fmt.Printf("   %s\n", repo.Description)
		}
		fmt.Printf("   ⭐ %d stars | Language: %s\n", repo.Stars, repo.Language)
		fmt.Printf("   📦 %s\n", repo.URL)

		// Parse updated date
		if t, err := time.Parse(time.RFC3339, repo.UpdatedAt); err == nil {
			fmt.Printf("   🕐 Updated: %s\n", t.Format("2006-01-02"))
		}
		fmt.Println()
	}

	fmt.Println("To add a tool from these results to your config:")
	fmt.Println("  1. Visit the repository URL")
	fmt.Println("  2. Review releases and documentation")
	fmt.Println("  3. Create a tool spec in the registry")
	fmt.Println("  4. Add to styx.toml: tool_name = \"version\"")
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
