package main

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// findToken returns the first token available from GITHUB_TOKEN, GH_TOKEN
// or the GitHub CLI, along with where it came from. It returns empty
// strings when there is none, and the tool then runs unauthenticated.
func findToken(getenv func(string) string, ghToken func() (string, error)) (token, source string) {
	for _, name := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if v := strings.TrimSpace(getenv(name)); v != "" {
			return v, name
		}
	}

	if ghToken != nil {
		if t, err := ghToken(); err == nil && t != "" {
			return t, "gh auth token"
		}
	}

	return "", ""
}

// ghAuthToken asks the GitHub CLI for its token, if gh is installed and
// logged in.
func ghAuthToken() (string, error) {
	path, err := exec.LookPath("gh")
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, path, "auth", "token").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
