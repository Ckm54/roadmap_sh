package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type PushPayload struct {
	Size   int    `json:"size"` // no longer sent by the Events API, kept for older events
	Ref    string `json:"ref"`
	Before string `json:"before"`
	Head   string `json:"head"`
}

// compareKey identifies a push's commit range for the compare API.
func (p PushPayload) compareKey() string {
	return p.Before + "..." + p.Head
}

// canCompare reports whether the compare API can count this push's commits.
// A push that creates a branch has an all-zero "before" SHA.
func (p PushPayload) canCompare() bool {
	return p.Size == 0 && p.Head != "" && strings.Trim(p.Before, "0") != ""
}

// target describes where a push went, e.g. "main" or "tag v1.0.0".
func (p PushPayload) target() string {
	if tag, ok := strings.CutPrefix(p.Ref, "refs/tags/"); ok {
		return "tag " + tag
	}
	return strings.TrimPrefix(p.Ref, "refs/heads/")
}

type ActionPayload struct {
	Action string `json:"action"` // used by issues, PRs and watch events
}

type CreatePayload struct {
	RefType string `json:"ref_type"` // branch or tag
	Ref     string `json:"ref"`      // name of the branch or tag
}

// pushGroup merges consecutive pushes to the same repo and branch into one line.
type pushGroup struct {
	repo, target string
	pushes       int
	commits      int
	allCounted   bool
}

func (g *pushGroup) String() string {
	where := g.repo
	if g.target != "" {
		where = g.target + " in " + g.repo
	}

	switch {
	case g.allCounted:
		return fmt.Sprintf("Pushed %d %s to %s", g.commits, plural(g.commits, "commit"), where)
	case g.pushes == 1:
		return fmt.Sprintf("Pushed to %s", where)
	default:
		return fmt.Sprintf("Pushed %d times to %s", g.pushes, where)
	}
}

// formatEvents turns events into output lines. counts maps a push's
// compareKey to its commit count; pushes without a count are shown without
// one. Events whose payload can't be parsed are reported in warnings.
func formatEvents(events []GitHubEvent, counts map[string]int) (lines, warnings []string) {
	var group *pushGroup
	flush := func() {
		if group != nil {
			lines = append(lines, group.String())
			group = nil
		}
	}

	for _, event := range events {
		if event.Type == "PushEvent" {
			var p PushPayload
			if err := json.Unmarshal(event.Payload, &p); err != nil {
				warnings = append(warnings, fmt.Sprintf("Failed to parse %s: invalid push payload structure: %v", event.Type, err))
				continue
			}

			commits, counted := p.Size, p.Size > 0
			if c, ok := counts[p.compareKey()]; ok && !counted {
				commits, counted = c, true
			}

			if group != nil && group.repo == event.Repo.Name && group.target == p.target() {
				group.pushes++
				group.commits += commits
				group.allCounted = group.allCounted && counted
				continue
			}

			flush()
			group = &pushGroup{repo: event.Repo.Name, target: p.target(), pushes: 1, commits: commits, allCounted: counted}
			continue
		}

		flush()
		line, err := formatEvent(event)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Failed to parse %s: %v", event.Type, err))
			continue
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	flush()

	return lines, warnings
}

func formatEvent(event GitHubEvent) (string, error) {
	repoName := event.Repo.Name

	switch event.Type {
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
		return fmt.Sprintf("Performed a %s activity in %s", event.Type, repoName), nil
	}
}

func capitalize(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if size == 0 {
		return ""
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
