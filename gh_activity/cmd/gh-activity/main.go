package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// maxConcurrentCompares caps parallel compare API calls when counting commits.
const maxConcurrentCompares = 8

type app struct {
	out, errOut io.Writer
	baseURL     string
	token       string
	tokenSource string
}

func main() {
	token, source := findToken(os.Getenv, ghAuthToken)
	a := &app{
		out:         os.Stdout,
		errOut:      os.Stderr,
		baseURL:     defaultBaseURL,
		token:       token,
		tokenSource: source,
	}
	os.Exit(a.run(os.Args[1:]))
}

func (a *app) run(args []string) int {
	fmt.Fprintln(a.out, "Welcome to gh activity")

	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		fmt.Fprintln(a.errOut, "username required")
		return 1
	}
	username := strings.TrimSpace(args[0])

	ctx := context.Background()
	client := NewClient(a.baseURL, a.token)

	events, err := client.UserEvents(ctx, username)

	// A bad token shouldn't stop public data from loading: warn and retry
	// without it.
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusUnauthorized && a.token != "" {
		fmt.Fprintf(a.errOut, "[Warning] GitHub rejected the token from %s; continuing without it.\n", a.tokenSource)
		a.token, a.tokenSource = "", ""
		client = NewClient(a.baseURL, "")
		events, err = client.UserEvents(ctx, username)
	}

	if err != nil {
		a.printError(username, err)
		return 1
	}

	if len(events) == 0 {
		fmt.Fprintf(a.out, "No recent public events found for user: %q\n", username)
		return 0
	}

	counts, pushesWithoutCount := a.commitCounts(ctx, client, events)

	lines, warnings := formatEvents(events, counts)
	for _, w := range warnings {
		fmt.Fprintf(a.errOut, "[Warning] %s\n", w)
	}

	fmt.Fprintln(a.out, "Output:")
	for _, line := range lines {
		fmt.Fprintf(a.out, "- %s\n", line)
	}

	if pushesWithoutCount {
		fmt.Fprintln(a.errOut, "Tip: set GITHUB_TOKEN or run 'gh auth login' to see commit counts for pushes.")
	}
	return 0
}

type compareRequest struct {
	repo string
	push PushPayload
}

// commitCounts looks up commit counts for push events with the compare API.
// It only does so with a token, since each push costs one request and the
// unauthenticated limit is 60 an hour. Failed lookups are skipped, and those
// pushes are shown without a count. The second result reports whether there
// were pushes that went uncounted only because there was no token.
func (a *app) commitCounts(ctx context.Context, client *Client, events []GitHubEvent) (map[string]int, bool) {
	var requests []compareRequest
	seen := map[string]bool{}
	for _, e := range events {
		if e.Type != "PushEvent" {
			continue
		}
		var p PushPayload
		if json.Unmarshal(e.Payload, &p) != nil || !p.canCompare() || seen[p.compareKey()] {
			continue
		}
		seen[p.compareKey()] = true
		requests = append(requests, compareRequest{repo: e.Repo.Name, push: p})
	}

	counts := map[string]int{}
	if len(requests) == 0 {
		return counts, false
	}
	if a.token == "" {
		return counts, true
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		sem = make(chan struct{}, maxConcurrentCompares)
	)
	for _, r := range requests {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			n, err := client.CommitCount(ctx, r.repo, r.push.Before, r.push.Head)
			if err != nil {
				return
			}
			mu.Lock()
			counts[r.push.compareKey()] = n
			mu.Unlock()
		})
	}
	wg.Wait()

	return counts, false
}

func (a *app) printError(username string, err error) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		fmt.Fprintf(a.errOut, "Error executing the request: %v\n", err)
		return
	}

	switch {
	case apiErr.StatusCode == http.StatusNotFound:
		fmt.Fprintf(a.errOut, "Error: GitHub user '%s' not found.\n", username)
	case apiErr.RateLimited:
		msg := "Error: GitHub API rate limit exceeded."
		if !apiErr.Reset.IsZero() {
			msg += fmt.Sprintf(" It resets at %s.", apiErr.Reset.Local().Format("15:04"))
		}
		if a.token == "" {
			msg += " Set GITHUB_TOKEN or run 'gh auth login' for a higher limit."
		}
		fmt.Fprintln(a.errOut, msg)
	default:
		fmt.Fprintf(a.errOut, "GitHub API returned unexpected status: %s\n", apiErr.Status)
	}
}
