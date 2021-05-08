# Task Tracker CLI

A simple interactive command-line task manager. Tasks are persisted locally in a `tasks.json` file.

This is a project from [roadmap.sh](https://roadmap.sh/projects/task-tracker).

## Getting Started

```bash
make run
```

Other available targets:

| Command      | Description                            |
| ------------ | -------------------------------------- |
| `make build` | Compile the binary to `./bin/task-cli` |
| `make run`   | Build and run the app                  |
| `make test`  | Run all tests                          |
| `make clean` | Remove build artifacts                 |
| `make help`  | Show all available targets             |

This drops you into an interactive REPL:

```
task-cli>>
```

Type `exit` or `quit` to leave.

## Commands

### Add a task

```
task-cli>> add Buy groceries
task-cli>> add "Clean the car"
```

### List tasks

```
# All tasks
task-cli>> list

# Filtered by status
task-cli>> list todo
task-cli>> list in-progress
task-cli>> list done
```

### Update a task title

```
task-cli>> update 1 Go to the market
```

### Mark a task status

```
task-cli>> mark-in-progress 1
task-cli>> mark-done 1
```

### Delete a task

```
task-cli>> delete 1
```

### Help

```
task-cli>> help
```

## Task Fields

Each task has an `id`, `title`, `status` (`todo`, `in-progress`, `done`), and timestamps for `created_at` and `updated_at`.

## Project Structure

```
cmd/task-cli/       # Entry point
internal/
  cli/              # Command handlers and routing
  task/             # Service layer and task entity
  storage/          # JSON file persistence
  constants/        # Shared status constants
```

## Running Tests

From the project's root run:

```bash
go test ./...
```

OR

```bash
make test
```
