# roadmap.sh Solutions

My solutions to [roadmap.sh](https://roadmap.sh/projects) project challenges.

## Projects

| Project                       | roadmap.sh challenge                                                     | Description                                                                         |
| ----------------------------- | ------------------------------------------------------------------------ | ----------------------------------------------------------------------------------- |
| [task_tracker](task_tracker/) | [Task Tracker](https://roadmap.sh/projects/task-tracker)                 | CLI for managing tasks, with statuses, priorities, due dates, tags and JSON storage |
| [gh_activity](gh_activity/)   | [GitHub User Activity](https://roadmap.sh/projects/github-user-activity) | CLI that shows a GitHub user's recent public activity using the GitHub API          |
| [expense-tracker-go](expense-tracker-go/) | [Expense Tracker](https://roadmap.sh/projects/expense-tracker) | CLI for recording expenses and totalling them by month or year, with JSON storage |

## Getting started

You need [Go](https://go.dev/dl/) 1.27 or newer, plus `make` if you want to use the Makefiles.

```bash
git clone https://github.com/ckm54/roadmap_sh.git
cd roadmap_sh/<project>
make build      # binary goes in ./bin
make test
make help       # list all targets
```

See each project's README for usage details.

## Repository layout

```
.
├── .github/workflows/   CI and release workflows
├── task_tracker/        Task Tracker CLI (task-cli)
├── gh_activity/         GitHub User Activity CLI (gh-activity)
└── expense-tracker-go/  Expense Tracker CLI (expense-tracker)
```

## CI and releases

[WIP]
