package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type GitHubEvent struct {
	Type    string          `json:"type"`
	Repo    RepoInfo        `json:"repo"`
	Payload json.RawMessage `json:"payload"`
}

type RepoInfo struct {
	Name string `json:"name"`
}

type PushPayload struct {
	Size int `json:"size"`
}

type ActionPayload struct {
	Action string `json:"action"` // used by issues, PRs and watch events
}

type CreatePayload struct {
	RefType string `json:"ref_type"` // branch or tag
	Ref     string `json:"ref"`      // name of the branch or tag
}

func main() {
	fmt.Println("Welcome to gh activity")

	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "username required\n")
		os.Exit(1)
	}

	gh_username := strings.TrimSpace(args[0])

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	gh_url := fmt.Sprintf("https://api.github.com/users/%s/events", gh_username)
	req, err := http.NewRequest(http.MethodGet, gh_url, nil)
	if err != nil {
		log.Fatalf("Error creating the request: %v", err)
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Go-GitHub-CLI")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error executing the request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			fmt.Fprintf(os.Stderr, "Error: GitHub user '%s' not found.\n", gh_username)
		case http.StatusForbidden:
			fmt.Fprintln(os.Stderr, "Error: Rate limit exceeded. Try again later or use an API token.")
		default:
			fmt.Fprintf(os.Stderr, "GitHub API returned unexpected status: %s\n", resp.Status)
		}
		os.Exit(1)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response payload: %v", err)
	}

	var events []GitHubEvent
	if err := json.Unmarshal(bodyBytes, &events); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	if len(events) == 0 {
		fmt.Printf("No recent public events found for user: %q\n", gh_username)
		return
	}

	fmt.Println("Output:")
	for _, event := range events {
		formattedText, err := formatEvent(event)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[Warning] Failed to parse %s: %v\n", event.Type, err)
			continue
		}

		if formattedText != "" {
			fmt.Printf("- %s\n", formattedText)
		}
	}

}

func formatEvent(event GitHubEvent) (string, error) {
	repoName := event.Repo.Name

	switch event.Type {
	case "PushEvent":
		var p PushPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return "", fmt.Errorf("invalid push payload structure: %w", err)
		}
		commitStr := "commit"
		if p.Size != 1 {
			commitStr = "commits"
		}
		return fmt.Sprintf("Pushed %d %s to %s", p.Size, commitStr, repoName), nil

	case "IssuesEvent":
		var p ActionPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return "", fmt.Errorf("invalid issues payload structure: %w", err)
		}
		if p.Action == "opened" {
			return fmt.Sprintf("Opened a new issue in %s", repoName), nil
		}
		return fmt.Sprintf("%s an issue in %s", capitalize(p.Action), repoName), nil

	case "WatchEvent":
		var p ActionPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return "", fmt.Errorf("invalid watch payload structure: %w", err)
		}
		if p.Action == "started" {
			return fmt.Sprintf("Starred %s", repoName), nil
		}
		return fmt.Sprintf("Watched %s", repoName), nil

	case "PullRequestEvent":
		var p ActionPayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return "", fmt.Errorf("invalid PR payload structure: %w", err)
		}
		return fmt.Sprintf("%s a pull request in %s", capitalize(p.Action), repoName), nil

	case "CreateEvent":
		var p CreatePayload
		if err := json.Unmarshal(event.Payload, &p); err != nil {
			return "", fmt.Errorf("invalid create payload structure: %w", err)
		}
		return fmt.Sprintf("Created a new %s '%s' in %s", p.RefType, p.Ref, repoName), nil

	case "IssueCommentEvent":
		return fmt.Sprintf("Commented on an issue in %s", repoName), nil

	case "ForkEvent":
		return fmt.Sprintf("Forked %s", repoName), nil

	default:
		// Return empty for unhandled types instead of throwing an error
		return fmt.Sprintf("Performed a %s activity in %s", event.Type, repoName), nil
	}
}

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return string(s[0]-32) + s[1:]
}
