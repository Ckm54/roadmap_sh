package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/ckm54/roadmap_sh/expense-tracker/internal/expense"
)

// ExpenseRunner is the part of expense.Service the CLI needs. Depending on
// an interface lets tests swap in a fake without touching the filesystem.
type ExpenseRunner interface {
	Add(opts expense.AddOptions) (expense.ExpenseEntity, error)
	List() ([]expense.ExpenseEntity, error)
	Summary(opts expense.SummaryOptions) (expense.Cents, error)
}

type App struct {
	svc      ExpenseRunner
	out      io.Writer
	commands []command
}

type command struct {
	name    string
	usage   string
	summary string
	run     func(args []string) error
}

func New(svc ExpenseRunner, out io.Writer) *App {
	if out == nil {
		out = io.Discard
	}
	a := &App{svc: svc, out: out}

	// Every new command is one line here plus one run<Name> method.
	a.commands = []command{
		{"add", "add --description <text> --amount <number>", "Add an expense", a.runAdd},
		{"list", "list", "List all expenses sorted by id", a.runList},
		{"summary", "summary [--month <1-12>] [--year <yyyy>]", "Show total expenses, optionally for a month or year", a.runSummary},
	}
	return a
}

// Run executes one command given as already-split arguments, e.g. os.Args[1:].
func (a *App) Run(args []string) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		a.printHelp()
		return nil
	}

	cmd, ok := a.lookup(strings.ToLower(args[0]))
	if !ok {
		return fmt.Errorf("unknown command %q; run 'expense-tracker help' for options", args[0])
	}

	err := cmd.run(args[1:])
	if errors.Is(err, errHelpShown) {
		return nil
	}
	return err
}

func (a *App) lookup(name string) (command, bool) {
	i := slices.IndexFunc(a.commands, func(c command) bool { return c.name == name })
	if i < 0 {
		return command{}, false
	}
	return a.commands[i], true
}

func (a *App) printHelp() {
	fmt.Fprintln(a.out, "Usage:\n  expense-tracker <command> [flags]\n\nCommands:")
	for _, c := range a.commands {
		fmt.Fprintf(a.out, "  %-10s %s\n", c.name, c.summary)
	}
}

// errHelpShown signals that a command printed its own help and should stop.
var errHelpShown = errors.New("help shown")

// flagSet returns a FlagSet that reports errors instead of exiting, so the
// caller decides how to print them.
func (a *App) flagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

// parse parses fs and rejects leftover positional arguments.
func (a *App) parse(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			cmd, _ := a.lookup(fs.Name())
			fmt.Fprintf(a.out, "%s\n\nUsage:\n  expense-tracker %s\n\nFlags:\n", cmd.summary, cmd.usage)
			fs.SetOutput(a.out)
			fs.PrintDefaults()
			return errHelpShown
		}
		return fmt.Errorf("%s: %w", fs.Name(), err)
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("%s: unexpected argument %q", fs.Name(), fs.Arg(0))
	}
	return nil
}
