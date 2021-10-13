package main

import (
	"fmt"
	"os"

	"github.com/ckm54/task_tracker/internal/cli"
	"github.com/ckm54/task_tracker/internal/storage"
	"github.com/ckm54/task_tracker/internal/task"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	path, err := storage.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	app := cli.New(task.NewService(storage.NewJSONStore(path)), cli.Options{
		Out:     os.Stdout,
		Err:     os.Stderr,
		In:      os.Stdin,
		Color:   useColor(),
		Version: version,
	})

	if len(os.Args) > 1 {
		if err := app.Run(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := app.REPL(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// useColor follows the NO_COLOR convention (https://no-color.org) and only
// colours output when stdout is a terminal.
func useColor() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok || os.Getenv("TERM") == "dumb" {
		return false
	}
	info, err := os.Stdout.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
