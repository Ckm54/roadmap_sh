package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ckm54/task_tracker/internal/task"
)

type TaskRunner interface {
	Add(description string, opts task.AddOptions) (task.TaskEntity, error)
	Get(id int) (task.TaskEntity, error)
	Update(id int, patch task.Patch) error
	SetStatus(ids []int, status task.Status) error
	Delete(ids []int) error
	Clear(status task.Status) (int, error)
	List(opts task.ListOptions) ([]task.TaskEntity, error)
}

type Options struct {
	Out     io.Writer
	Err     io.Writer
	In      io.Reader
	Color   bool
	Version string
	// Now is used for due dates and overdue checks; defaults to time.Now.
	Now func() time.Time
}

type App struct {
	svc      TaskRunner
	out      io.Writer
	errOut   io.Writer
	in       *bufio.Reader
	color    bool
	version  string
	now      func() time.Time
	commands []command
}

type command struct {
	name    string
	usage   string
	summary string
	run     func(args []string) error
}

func New(svc TaskRunner, opts Options) *App {
	a := &App{
		svc:     svc,
		out:     opts.Out,
		errOut:  opts.Err,
		color:   opts.Color,
		version: opts.Version,
		now:     opts.Now,
	}
	if a.out == nil {
		a.out = io.Discard
	}
	if a.errOut == nil {
		a.errOut = a.out
	}
	if opts.In != nil {
		a.in = bufio.NewReader(opts.In)
	}
	if a.now == nil {
		a.now = time.Now
	}
	if a.version == "" {
		a.version = "dev"
	}

	a.commands = []command{
		{"add", "add <description> [-p priority] [--due date] [-t tag]...", "Add a new task", a.runAdd},
		{"update", "update <id> [description] [-p priority] [--due date|none] [-t tag]... [--clear-tags]", "Change a task's description, priority, due date or tags", a.runUpdate},
		{"mark-todo", "mark-todo <id>...", "Mark tasks as todo", a.statusRunner(task.StatusTodo)},
		{"mark-in-progress", "mark-in-progress <id>...", "Mark tasks as in-progress", a.statusRunner(task.StatusInProgress)},
		{"mark-done", "mark-done <id>...", "Mark tasks as done", a.statusRunner(task.StatusDone)},
		{"delete", "delete <id>... [-f]", "Delete tasks (asks for confirmation)", a.runDelete},
		{"clear", "clear [-f]", "Delete all done tasks (asks for confirmation)", a.runClear},
		{"list", "list [todo|in-progress|done] [-p priority] [-t tag] [--sort field] [--json]", "List tasks, optionally filtered", a.runList},
		{"search", "search <text> [--sort field] [--json]", "Find tasks whose description contains text", a.runSearch},
		{"show", "show <id> [--json]", "Show all details of one task", a.runShow},
		{"completion", "completion bash|zsh|fish", "Print a shell completion script", a.runCompletion},
		{"version", "version", "Print the version", a.runVersion},
		{"help", "help [command]", "Show help for all commands or one command", a.runHelp},
	}
	return a
}

// Run executes one command given as already-split arguments.
func (a *App) Run(args []string) error {
	if len(args) == 0 {
		a.printGlobalHelp()
		return nil
	}

	name := strings.ToLower(args[0])
	switch name {
	case "-h", "--help":
		name = "help"
	case "-v", "--version":
		name = "version"
	}

	cmd, ok := a.lookup(name)
	if !ok {
		if s := a.suggest(name); s != "" {
			return fmt.Errorf("unknown command %q; did you mean %q? Type 'help' for options", args[0], s)
		}
		return fmt.Errorf("unknown command %q; type 'help' for options", args[0])
	}

	err := cmd.run(args[1:])
	if errors.Is(err, errHelpShown) {
		return nil
	}
	return err
}

// ExecuteLine splits a line of REPL input like a shell would and runs it.
func (a *App) ExecuteLine(line string) error {
	args, err := splitArgs(line)
	if err != nil {
		return err
	}
	return a.Run(args)
}

const prompt = "task-cli>> "

// REPL reads commands from the input until EOF, "exit" or "quit".
func (a *App) REPL() error {
	if a.in == nil {
		return errors.New("no input available for interactive mode")
	}

	fmt.Fprintln(a.out, "Task tracker. Type 'help' for commands, 'exit' to quit.")
	for {
		fmt.Fprint(a.out, prompt)

		line, err := a.in.ReadString('\n')
		if err != nil && line == "" {
			if errors.Is(err, io.EOF) {
				fmt.Fprintln(a.out)
				return nil
			}
			return err
		}

		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "":
			continue
		case "exit", "quit":
			return nil
		}

		if err := a.ExecuteLine(trimmed); err != nil {
			fmt.Fprintf(a.errOut, "Error: %v\n", err)
		}
	}
}

func (a *App) lookup(name string) (command, bool) {
	i := slices.IndexFunc(a.commands, func(c command) bool { return c.name == name })
	if i < 0 {
		return command{}, false
	}
	return a.commands[i], true
}

// suggest returns the closest command name within a small edit distance.
func (a *App) suggest(name string) string {
	best, bestDist := "", 3
	for _, c := range a.commands {
		if d := editDistance(name, c.name); d < bestDist {
			best, bestDist = c.name, d
		}
	}
	return best
}

func editDistance(s, t string) int {
	prev := make([]int, len(t)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(s); i++ {
		cur := make([]int, len(t)+1)
		cur[0] = i
		for j := 1; j <= len(t); j++ {
			cost := 1
			if s[i-1] == t[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(t)]
}

func (a *App) printGlobalHelp() {
	fmt.Fprintln(a.out, "Task tracker is a CLI tool to help you manage your tasks.")
	fmt.Fprintln(a.out, "\nUsage:")
	fmt.Fprintln(a.out, "  task-cli <command> [arguments]")
	fmt.Fprintln(a.out, "  task-cli                         (no arguments starts interactive mode)")
	fmt.Fprintln(a.out, "\nCommands:")

	width := 0
	for _, c := range a.commands {
		width = max(width, len(c.name))
	}
	for _, c := range a.commands {
		fmt.Fprintf(a.out, "  %-*s  %s\n", width, c.name, c.summary)
	}

	fmt.Fprintln(a.out, "\nRun 'task-cli help <command>' for details on a command.")
}

func (a *App) runHelp(args []string) error {
	if len(args) == 0 {
		a.printGlobalHelp()
		return nil
	}
	cmd, ok := a.lookup(strings.ToLower(args[0]))
	if !ok {
		return fmt.Errorf("unknown command %q", args[0])
	}
	return cmd.run([]string{"--help"})
}

func (a *App) runVersion(args []string) error {
	if _, err := a.parse(a.flagSet("version"), args); err != nil {
		return err
	}
	fmt.Fprintf(a.out, "task-cli %s\n", a.version)
	return nil
}

// errHelpShown signals that a command printed its own help and should stop.
var errHelpShown = errors.New("help shown")

func (a *App) flagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

// parse parses flags that may appear anywhere among the positional arguments,
// so "add Buy milk -p high" works as well as "add -p high Buy milk". A bare
// "--" ends flag parsing.
func (a *App) parse(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for {
		if err := fs.Parse(args); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				a.printCommandHelp(fs)
				return nil, errHelpShown
			}
			return nil, fmt.Errorf("%s: %w", fs.Name(), err)
		}

		rest := fs.Args()
		if len(rest) == 0 {
			return positional, nil
		}
		if consumed := len(args) - len(rest); consumed > 0 && args[consumed-1] == "--" {
			return append(positional, rest...), nil
		}
		positional = append(positional, rest[0])
		args = rest[1:]
	}
}

func (a *App) printCommandHelp(fs *flag.FlagSet) {
	cmd, _ := a.lookup(fs.Name())
	fmt.Fprintf(a.out, "%s\n\nUsage:\n  task-cli %s\n", cmd.summary, cmd.usage)

	// Shorthands are registered as separate flags whose usage reads
	// "shorthand for --long"; fold them into the long flag's line.
	shorthands := map[string]string{}
	fs.VisitAll(func(f *flag.Flag) {
		if long, ok := strings.CutPrefix(f.Usage, shorthandPrefix); ok {
			shorthands[long] = f.Name
		}
	})

	var lines [][2]string
	fs.VisitAll(func(f *flag.Flag) {
		if strings.HasPrefix(f.Usage, shorthandPrefix) {
			return
		}
		name := "    --" + f.Name
		if short, ok := shorthands[f.Name]; ok {
			name = "-" + short + ", --" + f.Name
		}
		valueName, usage := flag.UnquoteUsage(f)
		if valueName != "" {
			name += " " + valueName
		}
		if f.DefValue != "" && f.DefValue != "false" {
			usage += fmt.Sprintf(" (default %s)", f.DefValue)
		}
		lines = append(lines, [2]string{name, usage})
	})

	if len(lines) == 0 {
		return
	}

	width := 0
	for _, l := range lines {
		width = max(width, len(l[0]))
	}
	fmt.Fprintln(a.out, "\nFlags:")
	for _, l := range lines {
		fmt.Fprintf(a.out, "  %-*s  %s\n", width, l[0], l[1])
	}
}

const shorthandPrefix = "shorthand for --"

// confirm asks a yes/no question.
// returns an error with no input attached or at EOF.
func (a *App) confirm(question string) (bool, error) {
	if a.in == nil {
		return false, errors.New("confirmation required; re-run with --force")
	}

	fmt.Fprintf(a.out, "%s [y/N]: ", question)
	line, err := a.in.ReadString('\n')
	if err != nil && line == "" {
		fmt.Fprintln(a.out)
		return false, errors.New("confirmation required; re-run with --force")
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	}
	return false, nil
}

func parseID(s string) (int, error) {
	id, err := strconv.Atoi(s)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid ID %q: must be a positive number", s)
	}
	return id, nil
}

func parseIDs(args []string) ([]int, error) {
	if len(args) == 0 {
		return nil, errors.New("at least one task ID is required")
	}

	var ids []int
	for _, arg := range args {
		id, err := parseID(arg)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func joinIDs(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(id)
	}
	return strings.Join(parts, ", ")
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// joinText turns positional words into one trimmed piece of text.
func joinText(args []string) string {
	return strings.TrimSpace(strings.Join(args, " "))
}

// splitArgs splits a line into arguments the way a POSIX shell would for the
// simple cases: whitespace separates arguments, single quotes are literal,
// double quotes allow backslash escapes, and a backslash outside quotes
// escapes the next character.
func splitArgs(line string) ([]string, error) {
	var (
		args    []string
		current strings.Builder
		inArg   bool
		quote   rune
		escaped bool
	)

	for _, r := range line {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case quote == '\'':
			if r == '\'' {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case quote == '"':
			switch r {
			case '"':
				quote = 0
			case '\\':
				escaped = true
			default:
				current.WriteRune(r)
			}
		case r == '\\':
			escaped, inArg = true, true
		case r == '\'' || r == '"':
			quote, inArg = r, true
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			if inArg {
				args = append(args, current.String())
				current.Reset()
				inArg = false
			}
		default:
			current.WriteRune(r)
			inArg = true
		}
	}

	if quote != 0 {
		return nil, fmt.Errorf("unterminated %c quote", quote)
	}
	if escaped {
		current.WriteRune('\\')
	}
	if inArg {
		args = append(args, current.String())
	}
	return args, nil
}
