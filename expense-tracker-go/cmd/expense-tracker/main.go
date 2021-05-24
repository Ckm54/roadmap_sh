package main

import (
	"fmt"
	"os"

	"github.com/ckm54/roadmap_sh/expense-tracker/internal/cli"
	"github.com/ckm54/roadmap_sh/expense-tracker/internal/expense"
	"github.com/ckm54/roadmap_sh/expense-tracker/internal/storage"
)

func main() {
	path, err := storage.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	app := cli.New(expense.NewService(storage.NewJSONStore(path)), os.Stdout)

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
