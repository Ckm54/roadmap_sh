# GitHub User Activity CLI

A small command-line tool that shows a GitHub user's recent public activity in the terminal. It calls the [GitHub Events API](https://docs.github.com/en/rest/activity/events#list-public-events-for-a-user) directly and uses only the Go standard library.

This is a solution to the [GitHub User Activity](https://roadmap.sh/projects/github-user-activity) project on roadmap.sh.

## Installation

Requires Go 1.27 or newer.

```bash
go install github.com/ckm54/roadmap_sh/gh_activity/cmd/gh-activity@latest
```

Or build from source:

```bash
git clone https://github.com/ckm54/roadmap_sh.git
cd roadmap_sh/gh_activity
make build          # binary in ./bin/gh-activity
make install        # or install into $GOBIN (or ~/go/bin)
```

## Usage

```bash
gh-activity <username>
```

Example:

```
$ gh-activity octocat
Welcome to gh activity
Output:
- Pushed 3 commits to main in octocat/hello-world
- Opened a new issue in octocat/hello-world
- Starred octocat/linguist
- Created a new branch 'feature/login' in octocat/hello-world
- Merged a pull request in octocat/hello-world
```

Events are listed newest first. GitHub returns at most the 30 most recent public events, from the last 90 days. Private activity is never shown, even when the token belongs to the user you look up.

## Authentication

A token is optional, but with one you get a much higher rate limit and commit counts for pushes. The tool uses the first token it finds:

1. the `GITHUB_TOKEN` environment variable
2. the `GH_TOKEN` environment variable
3. `gh auth token`, if the [GitHub CLI](https://cli.github.com) is installed and logged in

|                       | Without a token | With a token |
| --------------------- | --------------- | ------------ |
| Rate limit            | 60 requests/hour per IP | 5,000 requests/hour |
| Push commit counts    | No (`Pushed to main in <repo>`) | Yes (`Pushed 3 commits to main in <repo>`) |

The token doesn't need any scopes, because only public data is read. If GitHub rejects the token, the tool prints a warning and carries on without it.

```bash
GITHUB_TOKEN=ghp_xxx gh-activity octocat
```

## Supported events

| GitHub event        | Output                                                   |
| ------------------- | -------------------------------------------------------- |
| `PushEvent`         | `Pushed N commits to <branch> in <repo>` (see below)     |
| `IssuesEvent`       | `Opened a new issue in <repo>`, `Closed an issue in <repo>`, … |
| `PullRequestEvent`  | `Opened a pull request in <repo>`, `Merged a pull request in <repo>`, … |
| `IssueCommentEvent` | `Commented on an issue in <repo>`                        |
| `WatchEvent`        | `Starred <repo>`                                         |
| `CreateEvent`       | `Created a new <branch/tag> '<name>' in <repo>`          |
| `ForkEvent`         | `Forked <repo>`                                          |
| Anything else       | `Performed a <EventType> activity in <repo>`             |

If a known event's payload can't be parsed, a `[Warning]` line goes to stderr and the remaining events are still printed.

### Push events

GitHub no longer includes commit counts in the Events API's push payloads. It only sends the branch and the commit SHAs before and after the push. The tool handles pushes like this:

- **With a token**, it asks the [compare API](https://docs.github.com/en/rest/commits/commits#compare-two-commits) how many commits each push added, running up to 8 requests at a time.
- **Without a token**, it skips those extra requests, which would quickly use up the 60-an-hour limit. It shows `Pushed to main in <repo>` and prints a tip on stderr.
- **Consecutive pushes** to the same branch are merged into one line, e.g. `Pushed 8 commits to main in <repo>`. Without a token this reads `Pushed 8 times to main in <repo>`.
- **Pushes that create a branch**, and pushes whose compare request fails (for example, after the repository was deleted), are shown without a count.
- **Tag pushes** are shown as `Pushed to tag v1.0.0 in <repo>`.

## Errors

The tool prints a message to stderr and exits with status `1` when:

| Situation                       | Message                                                            |
| ------------------------------- | ------------------------------------------------------------------ |
| No username given               | `username required`                                                |
| User doesn't exist (HTTP 404)   | `Error: GitHub user '<name>' not found.`                           |
| Rate limited (HTTP 403/429)     | `Error: GitHub API rate limit exceeded. It resets at 15:04.` Without a token, it also suggests setting one. |
| Any other non-200 response      | `GitHub API returned unexpected status: <status>`                  |
| Network error or timeout (10 s) | `Error executing the request: ...`                                 |

If the user exists but has no recent public activity, it prints `No recent public events found for user: "<name>"` and exits with status `0`.

## Known limitations

- **No paging or filtering.** Only the first page of events (up to 30) is fetched, and there are no flags yet for event type or the number of events.
- **`CreateEvent` for repositories.** When a whole repository is created, GitHub sends no ref name, so the line reads `Created a new repository '' in <repo>`.

## Development

| Command                   | Description                            |
| ------------------------- | -------------------------------------- |
| `make build`              | Compile the binary to `./bin`          |
| `make run ARGS="octocat"` | Build and run with arguments           |
| `make install`            | Install into `$GOBIN`                  |
| `make test`               | Run tests with the race detector       |
| `make vet` / `make fmt`   | Vet / format the code                  |
| `make all`                | Format, vet, test and build            |
| `make clean`              | Remove build artifacts                 |
| `make help`               | List all targets                       |

The tests run against a fake GitHub server (`httptest`), so they need no network access or token.

### Project structure

```
cmd/gh-activity/
  main.go        Entry point: fetches events, counts push commits, prints output and errors
  github.go      Minimal GitHub REST client (events, compare API, error and rate-limit parsing)
  token.go       Finds a token from GITHUB_TOKEN, GH_TOKEN or the gh CLI
  format.go      Turns events into lines, including merging pushes
  main_test.go   Tests
```
