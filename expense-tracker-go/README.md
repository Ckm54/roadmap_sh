# Expense Tracker CLI

A small command-line expense tracker written in Go. Expenses are stored in a local JSON file.

This is a solution to the [Expense Tracker](https://roadmap.sh/projects/expense-tracker) project on roadmap.sh.

## Installation

Requires Go 1.27 or newer.

```bash
git clone https://github.com/ckm54/roadmap_sh.git
cd roadmap_sh/expense-tracker-go
make build          # binary in ./bin/expense-tracker
make install        # or install into $GOBIN (or ~/go/bin)
```

## Usage

```
$ expense-tracker add --description "Lunch" --amount 20
Expense added successfully (ID: 1)

$ expense-tracker add --description "Dinner" --amount 10.50
Expense added successfully (ID: 2)

$ expense-tracker list
# ID  Date        Description  Amount
# 1   2026-10-31  Lunch        $20.00
# 2   2026-10-31  Dinner       $10.50

$ expense-tracker summary
Total expenses: $30.50

$ expense-tracker summary --month 10
Total expenses for October: $30.50

$ expense-tracker summary --year 2026
Total expenses for 2026: $30.50

$ expense-tracker summary --month 8 --year 2025
Total expenses for August 2025: $0.00
```

| Command   | Flags                                      | Notes                                                  |
| --------- | ------------------------------------------ | ------------------------------------------------------ |
| `add`     | `--description <text>` `--amount <number>` | Amount must be positive, with at most 2 decimal places |
| `list`    | none                                       | Sorted by ID                                           |
| `summary` | `--month <1-12>` `--year <yyyy>`           | `--month` alone means that month of the current year   |

Run `expense-tracker help` or `expense-tracker <command> --help` for details.

Amounts are stored as whole cents, not floating-point numbers, so totals are always exact.

## Data file

Expenses are saved to `~/.local/share/expense-cli/expenses.json` on Linux (or `$XDG_DATA_HOME/expense-cli/`), and to the user config directory on macOS and Windows. Set `EXPENSE_CLI_FILE` to use a different file:

```bash
EXPENSE_CLI_FILE=./expenses.json expense-tracker list
```

## Development

| Command                   | Description                      |
| ------------------------- | -------------------------------- |
| `make build`              | Compile the binary to `./bin`    |
| `make run ARGS="summary"` | Build and run with arguments     |
| `make test`               | Run tests with the race detector |
| `make cover`              | Open an HTML coverage report     |
| `make vet` / `make fmt`   | Vet / format the code            |
| `make lint`               | Run golangci-lint                |
| `make all`                | Format, vet, test and build      |
| `make help`               | List all targets                 |

```
cmd/expense-tracker/   Entry point: wires the pieces together
internal/cli/          Commands, flag parsing and output
internal/expense/      Expense rules: validation, money, totals
internal/storage/      JSON file storage
```
