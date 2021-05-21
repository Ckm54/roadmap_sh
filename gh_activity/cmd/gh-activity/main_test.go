package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
)

func pushEvent(repo, ref, before, head string) GitHubEvent {
	payload, _ := json.Marshal(PushPayload{Ref: ref, Before: before, Head: head})
	return GitHubEvent{Type: "PushEvent", Repo: RepoInfo{Name: repo}, Payload: payload}
}

func event(typ, repo, payload string) GitHubEvent {
	if payload == "" {
		payload = "{}"
	}
	return GitHubEvent{Type: typ, Repo: RepoInfo{Name: repo}, Payload: json.RawMessage(payload)}
}

func TestFormatEvents(t *testing.T) {
	tests := []struct {
		name   string
		events []GitHubEvent
		counts map[string]int
		want   []string
	}{
		{
			name:   "push with a known count",
			events: []GitHubEvent{pushEvent("o/r", "refs/heads/main", "a1", "b1")},
			counts: map[string]int{"a1...b1": 3},
			want:   []string{"Pushed 3 commits to main in o/r"},
		},
		{
			name:   "push without a count",
			events: []GitHubEvent{pushEvent("o/r", "refs/heads/main", "a1", "b1")},
			want:   []string{"Pushed to main in o/r"},
		},
		{
			name: "consecutive pushes to one branch are merged",
			events: []GitHubEvent{
				pushEvent("o/r", "refs/heads/main", "b", "c"),
				pushEvent("o/r", "refs/heads/main", "a", "b"),
			},
			counts: map[string]int{"b...c": 1, "a...b": 2},
			want:   []string{"Pushed 3 commits to main in o/r"},
		},
		{
			name: "merged pushes with a missing count show the push count",
			events: []GitHubEvent{
				pushEvent("o/r", "refs/heads/main", "b", "c"),
				pushEvent("o/r", "refs/heads/main", "a", "b"),
			},
			counts: map[string]int{"b...c": 1},
			want:   []string{"Pushed 2 times to main in o/r"},
		},
		{
			name: "pushes to different branches or split by other events stay separate",
			events: []GitHubEvent{
				pushEvent("o/r", "refs/heads/main", "c", "d"),
				pushEvent("o/r", "refs/heads/dev", "b", "c"),
				event("WatchEvent", "o/r", `{"action":"started"}`),
				pushEvent("o/r", "refs/heads/dev", "a", "b"),
			},
			counts: map[string]int{"c...d": 1, "b...c": 1, "a...b": 1},
			want: []string{
				"Pushed 1 commit to main in o/r",
				"Pushed 1 commit to dev in o/r",
				"Starred o/r",
				"Pushed 1 commit to dev in o/r",
			},
		},
		{
			name:   "tag push",
			events: []GitHubEvent{pushEvent("o/r", "refs/tags/v1.0.0", "a", "b")},
			want:   []string{"Pushed to tag v1.0.0 in o/r"},
		},
		{
			name:   "legacy payload with size",
			events: []GitHubEvent{event("PushEvent", "o/r", `{"size":2,"ref":"refs/heads/main"}`)},
			want:   []string{"Pushed 2 commits to main in o/r"},
		},
		{
			name: "other event types",
			events: []GitHubEvent{
				event("IssuesEvent", "o/r", `{"action":"opened"}`),
				event("IssuesEvent", "o/r", `{"action":"closed"}`),
				event("PullRequestEvent", "o/r", `{"action":"merged"}`),
				event("CreateEvent", "o/r", `{"ref_type":"branch","ref":"feat"}`),
				event("IssueCommentEvent", "o/r", ""),
				event("ForkEvent", "o/r", ""),
				event("ReleaseEvent", "o/r", ""),
			},
			want: []string{
				"Opened a new issue in o/r",
				"Closed an issue in o/r",
				"Merged a pull request in o/r",
				"Created a new branch 'feat' in o/r",
				"Commented on an issue in o/r",
				"Forked o/r",
				"Performed a ReleaseEvent activity in o/r",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, warnings := formatEvents(tt.events, tt.counts)
			if len(warnings) > 0 {
				t.Errorf("unexpected warnings: %q", warnings)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
			}
		})
	}
}

func TestFormatEvents_BadPayloadWarns(t *testing.T) {
	lines, warnings := formatEvents([]GitHubEvent{
		event("IssuesEvent", "o/r", `"not an object"`),
		event("PushEvent", "o/r", `[]`),
		event("ForkEvent", "o/r", ""),
	}, nil)

	if !slices.Equal(lines, []string{"Forked o/r"}) {
		t.Errorf("lines = %q", lines)
	}
	if len(warnings) != 2 {
		t.Errorf("warnings = %q, want 2", warnings)
	}
}

func TestPushPayload_CanCompare(t *testing.T) {
	if !(PushPayload{Before: "abc", Head: "def"}).canCompare() {
		t.Error("normal push should be comparable")
	}
	if (PushPayload{Before: "0000000000000000000000000000000000000000", Head: "def"}).canCompare() {
		t.Error("a push that creates a branch has nothing to compare against")
	}
	if (PushPayload{Size: 2, Before: "abc", Head: "def"}).canCompare() {
		t.Error("a payload that already has a size doesn't need comparing")
	}
}

func TestCapitalize(t *testing.T) {
	for in, want := range map[string]string{"": "", "closed": "Closed", "Merged": "Merged", "édité": "Édité"} {
		if got := capitalize(in); got != want {
			t.Errorf("capitalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFindToken(t *testing.T) {
	env := func(vars map[string]string) func(string) string {
		return func(k string) string { return vars[k] }
	}
	gh := func(token string, err error) func() (string, error) {
		return func() (string, error) { return token, err }
	}

	tests := []struct {
		name       string
		vars       map[string]string
		ghToken    func() (string, error)
		wantToken  string
		wantSource string
	}{
		{"GITHUB_TOKEN first", map[string]string{"GITHUB_TOKEN": "a", "GH_TOKEN": "b"}, gh("c", nil), "a", "GITHUB_TOKEN"},
		{"then GH_TOKEN", map[string]string{"GH_TOKEN": " b "}, gh("c", nil), "b", "GH_TOKEN"},
		{"then the gh CLI", nil, gh("c", nil), "c", "gh auth token"},
		{"gh not logged in", nil, gh("", errors.New("not logged in")), "", ""},
		{"no gh at all", nil, nil, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, source := findToken(env(tt.vars), tt.ghToken)
			if token != tt.wantToken || source != tt.wantSource {
				t.Errorf("findToken() = %q, %q; want %q, %q", token, source, tt.wantToken, tt.wantSource)
			}
		})
	}
}

// fakeGitHub serves canned events and compare results, and records auth.
type fakeGitHub struct {
	events       string
	eventsStatus int
	validToken   string
	compares     atomic.Int32
	rateLimited  bool
}

func (f *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if auth != "" && auth != "Bearer "+f.validToken {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"message":"Bad credentials"}`)
		return
	}

	switch {
	case r.URL.Path == "/users/octocat/events/public":
		if f.rateLimited {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", "1893456000")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"message":"API rate limit exceeded"}`)
			return
		}
		if f.eventsStatus != 0 {
			w.WriteHeader(f.eventsStatus)
			return
		}
		fmt.Fprint(w, f.events)

	case strings.HasPrefix(r.URL.Path, "/repos/octocat/hello/compare/"):
		f.compares.Add(1)
		if auth == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		fmt.Fprint(w, `{"total_commits": 2}`)

	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"Not Found"}`)
	}
}

const twoPushes = `[
 {"type":"PushEvent","repo":{"name":"octocat/hello"},"payload":{"ref":"refs/heads/main","before":"b","head":"c"}},
 {"type":"PushEvent","repo":{"name":"octocat/hello"},"payload":{"ref":"refs/heads/main","before":"a","head":"b"}},
 {"type":"WatchEvent","repo":{"name":"octocat/linguist"},"payload":{"action":"started"}}
]`

func runApp(t *testing.T, fake *fakeGitHub, token string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)

	var out, errOut bytes.Buffer
	a := &app{out: &out, errOut: &errOut, baseURL: srv.URL, token: token, tokenSource: "GITHUB_TOKEN"}
	code = a.run(args)
	return code, out.String(), errOut.String()
}

func TestRun_WithTokenCountsCommits(t *testing.T) {
	fake := &fakeGitHub{events: twoPushes, validToken: "secret"}
	code, out, errOut := runApp(t, fake, "secret", "octocat")

	want := "Welcome to gh activity\nOutput:\n- Pushed 4 commits to main in octocat/hello\n- Starred octocat/linguist\n"
	if code != 0 || out != want || errOut != "" {
		t.Errorf("code=%d\nstdout:\n%s\nstderr:\n%s", code, out, errOut)
	}
	if n := fake.compares.Load(); n != 2 {
		t.Errorf("compare API called %d times, want 2", n)
	}
}

func TestRun_WithoutTokenSkipsCompareAndShowsTip(t *testing.T) {
	fake := &fakeGitHub{events: twoPushes}
	code, out, errOut := runApp(t, fake, "", "octocat")

	if code != 0 || !strings.Contains(out, "- Pushed 2 times to main in octocat/hello\n") {
		t.Errorf("code=%d stdout:\n%s", code, out)
	}
	if !strings.Contains(errOut, "Tip: set GITHUB_TOKEN") {
		t.Errorf("stderr = %q, want a tip", errOut)
	}
	if n := fake.compares.Load(); n != 0 {
		t.Errorf("compare API called %d times without a token, want 0", n)
	}
}

func TestRun_BadTokenFallsBackToAnonymous(t *testing.T) {
	fake := &fakeGitHub{events: twoPushes, validToken: "secret"}
	code, out, errOut := runApp(t, fake, "wrong", "octocat")

	if code != 0 || !strings.Contains(out, "Starred octocat/linguist") {
		t.Errorf("code=%d stdout:\n%s", code, out)
	}
	if !strings.Contains(errOut, "rejected the token from GITHUB_TOKEN") {
		t.Errorf("stderr = %q, want a warning", errOut)
	}
}

func TestRun_Errors(t *testing.T) {
	tests := []struct {
		name    string
		fake    *fakeGitHub
		token   string
		args    []string
		wantErr string
	}{
		{"no username", &fakeGitHub{}, "", nil, "username required"},
		{"blank username", &fakeGitHub{}, "", []string{"  "}, "username required"},
		{"unknown user", &fakeGitHub{}, "", []string{"nobody"}, "Error: GitHub user 'nobody' not found."},
		{"rate limited without token", &fakeGitHub{rateLimited: true}, "", []string{"octocat"}, "rate limit exceeded. It resets at"},
		{"rate limit suggests a token", &fakeGitHub{rateLimited: true}, "", []string{"octocat"}, "Set GITHUB_TOKEN or run 'gh auth login'"},
		{"server error", &fakeGitHub{eventsStatus: http.StatusBadGateway}, "", []string{"octocat"}, "unexpected status: 502 Bad Gateway"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, errOut := runApp(t, tt.fake, tt.token, tt.args...)
			if code != 1 || !strings.Contains(errOut, tt.wantErr) {
				t.Errorf("code=%d stderr=%q, want 1 and %q", code, errOut, tt.wantErr)
			}
		})
	}

	t.Run("rate limited with token doesn't suggest one", func(t *testing.T) {
		_, _, errOut := runApp(t, &fakeGitHub{rateLimited: true, validToken: "s"}, "s", "octocat")
		if strings.Contains(errOut, "Set GITHUB_TOKEN") {
			t.Errorf("stderr = %q", errOut)
		}
	})
}

func TestRun_NoEvents(t *testing.T) {
	code, out, _ := runApp(t, &fakeGitHub{events: "[]"}, "", "octocat")
	if code != 0 || !strings.Contains(out, `No recent public events found for user: "octocat"`) {
		t.Errorf("code=%d stdout=%q", code, out)
	}
}
