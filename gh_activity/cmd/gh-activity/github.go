package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.github.com"

type GitHubEvent struct {
	Type    string          `json:"type"`
	Repo    RepoInfo        `json:"repo"`
	Payload json.RawMessage `json:"payload"`
}

type RepoInfo struct {
	Name string `json:"name"`
}

// Client is a minimal GitHub REST API client. An empty token makes
// unauthenticated requests (60 per hour instead of 5,000).
type Client struct {
	http    *http.Client
	baseURL string
	token   string
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		http:    &http.Client{Timeout: 10 * time.Second},
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
	}
}

// APIError is returned for any non-200 response.
type APIError struct {
	StatusCode  int
	Status      string
	Message     string
	RateLimited bool
	Reset       time.Time
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("GitHub API returned %s: %s", e.Status, e.Message)
	}
	return fmt.Sprintf("GitHub API returned %s", e.Status)
}

// UserEvents returns a user's recent public events. It uses the /public
// endpoint because /events also includes private activity when the token
// belongs to that user.
func (c *Client) UserEvents(ctx context.Context, username string) ([]GitHubEvent, error) {
	var events []GitHubEvent
	err := c.get(ctx, "/users/"+url.PathEscape(username)+"/events/public", &events)
	return events, err
}

// CommitCount returns how many commits head has that base doesn't, using
// the compare API. repo is "owner/name".
func (c *Client) CommitCount(ctx context.Context, repo, base, head string) (int, error) {
	var result struct {
		TotalCommits int `json:"total_commits"`
	}
	path := fmt.Sprintf("/repos/%s/compare/%s...%s", repo, url.PathEscape(base), url.PathEscape(head))
	if err := c.get(ctx, path, &result); err != nil {
		return 0, err
	}
	return result.TotalCommits, nil
}

func (c *Client) get(ctx context.Context, path string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating the request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "gh-activity")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return newAPIError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}
	return nil
}

func newAPIError(resp *http.Response) *APIError {
	apiErr := &APIError{StatusCode: resp.StatusCode, Status: resp.Status}

	var body struct {
		Message string `json:"message"`
	}
	if data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10)); err == nil {
		_ = json.Unmarshal(data, &body)
	}
	apiErr.Message = body.Message

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		apiErr.RateLimited = resp.Header.Get("X-RateLimit-Remaining") == "0" ||
			strings.Contains(strings.ToLower(body.Message), "rate limit")
	}
	if reset, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil {
		apiErr.Reset = time.Unix(reset, 0)
	}

	return apiErr
}
