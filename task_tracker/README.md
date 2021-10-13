# Task Tracker CLI

A command-line task manager written in Go with no dependencies outside the standard library. Tasks are stored in a local JSON file.

This is a solution to the [Task Tracker](https://roadmap.sh/projects/task-tracker) project on roadmap.sh.

## Installation

```bash
git clone https://github.com/ckm54/roadmap_sh.git
cd roadmap_sh/task_tracker
make install        # installs task-cli into $GOBIN (or ~/go/bin)
```

Run `make build` instead to put the binary in `./bin/task-cli`.

Prebuilt binaries for Linux, macOS and Windows are attached to each [GitHub release](https://github.com/ckm54/roadmap_sh/releases).

## Usage

Run a single command:

```bash
task-cli add "Buy groceries"
task-cli list
```

Or run `task-cli` with no arguments for an interactive prompt:

```
$ task-cli
Task tracker. Type 'help' for commands, 'exit' to quit.
task-cli>> add "Buy groceries" -p high
Task added successfully (ID: 1)
task-cli>> exit
```

Quoting works the same way in both modes. `"Buy  milk"` keeps its double space, and quotes inside the text are kept.

### Adding and updating

```bash
task-cli add "Buy groceries"
task-cli add "Finish report" --priority high --due 2026-11-01 --tag work
task-cli add "Call mum" -p low --due tomorrow -t family,phone

task-cli update 1 "Buy groceries and cook dinner"   # new description
task-cli update 2 --due none --priority medium      # change fields only
task-cli update 3 --tag home --tag errands          # replace tags
task-cli update 3 --clear-tags
```

- **Priority**: `low`, `medium` (default) or `high`.
- **Due dates**: `YYYY-MM-DD`, `today` or `tomorrow`. Use `none` with `update` to remove one.
- **Tags**: lowercase. Repeat the flag or comma-separate the values.

Flags can come before or after the description. To use text that starts with `-`, put `--` before it: `task-cli add -- -5 degrees tomorrow`.

### Changing status

Every status command accepts several IDs:

```bash
task-cli mark-in-progress 1
task-cli mark-done 1 2 3
task-cli mark-todo 2
```

If any ID doesn't exist, nothing is changed.

### Listing and searching

```bash
task-cli list                       # all tasks
task-cli list done                  # todo | in-progress | done
task-cli list --priority high --tag work
task-cli list --sort due            # id | priority | due | created | updated
task-cli list --json                # machine-readable output
task-cli search groceries           # case-insensitive text search
task-cli show 1                     # every field of one task
```

```
ID  STATUS       PRIORITY  DUE         DESCRIPTION    TAGS
1   todo         high      2026-10-01  Buy groceries  home
2   in-progress  medium    -           Write tests    -

2 tasks: 1 todo, 1 in-progress, 0 done
```

In a terminal, output is coloured, and due dates in the past are shown in red. Set `NO_COLOR=1` to turn colour off.

### Deleting

```bash
task-cli delete 3            # asks for confirmation
task-cli delete 3 4 --force  # no prompt
task-cli clear               # delete all done tasks
```

### Help, version and completion

```bash
task-cli help
task-cli help update        # or: task-cli update --help
task-cli version

source <(task-cli completion bash)                           # add to ~/.bashrc
source <(task-cli completion zsh)                            # add to ~/.zshrc
task-cli completion fish > ~/.config/fish/completions/task-cli.fish
```

## Storage

Tasks are saved to the first of these locations that applies:

1. `$TASK_CLI_FILE`, if set
2. `$XDG_DATA_HOME/task-cli/tasks.json`
3. `~/.local/share/task-cli/tasks.json` on Linux, or the user config directory on macOS and Windows

Each save writes a temporary file and renames it into place, so a crash can't leave a half-written file. Task IDs are never reused, even after you delete the newest task.

Each task looks like this:

```json
{
  "id": 1,
  "description": "Buy groceries",
  "status": "todo",
  "priority": "high",
  "dueDate": "2026-10-20T00:00:00+03:00",
  "tags": ["home"],
  "createdAt": "2026-10-17T21:00:00+03:00",
  "updatedAt": "2026-10-17T21:00:00+03:00"
}
```

Files saved by earlier versions (a plain array using `title` and `created_at`) are still read correctly. The next save converts them to the current format.

## Development

| Command                 | Description                                  |
| ----------------------- | -------------------------------------------- |
| `make build`            | Compile the binary to `./bin/task-cli`       |
| `make run ARGS="list"`  | Build and run with arguments                 |
| `make test`             | Run all tests with the race detector         |
| `make cover`            | Open an HTML coverage report                 |
| `make vet` / `make fmt` | Vet / format the code                        |
| `make lint`             | Run golangci-lint                            |
| `make release-snapshot` | Build release archives locally (goreleaser)  |
| `make clean`            | Remove build artifacts                       |

GitHub Actions checks formatting and runs vet, the tests, the build and golangci-lint on every push and pull request that touches `task_tracker/`. Pushing a `v*` tag publishes release binaries with goreleaser.

### Project structure

```
cmd/task-cli/    Entry point: one-shot command or interactive prompt
internal/
  cli/           Argument parsing, commands, output and completion scripts
  task/          Task model and service (business rules)
  storage/       JSON file persistence
```

The service depends only on the `task.DataStore` interface, so another backend such as SQLite could be added without changing the CLI.
