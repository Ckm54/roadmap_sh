package cli

import (
	"errors"
	"fmt"
	"strings"
)

// commandWords lists the extra words each command accepts, for completion.
var commandWords = map[string]string{
	"add":              "--priority --due --tag",
	"update":           "--priority --due --tag --clear-tags",
	"delete":           "--force",
	"clear":            "--force",
	"list":             "todo in-progress done --priority --tag --sort --json",
	"search":           "--sort --json",
	"show":             "--json",
	"completion":       "bash zsh fish",
	"mark-todo":        "",
	"mark-in-progress": "",
	"mark-done":        "",
	"version":          "",
	"help":             "",
}

func (a *App) runCompletion(args []string) error {
	positional, err := a.parse(a.flagSet("completion"), args)
	if err != nil {
		return err
	}
	if len(positional) != 1 {
		return errors.New("usage: completion bash|zsh|fish")
	}

	switch positional[0] {
	case "bash":
		fmt.Fprint(a.out, a.bashCompletion())
	case "zsh":
		fmt.Fprint(a.out, "autoload -U +X bashcompinit && bashcompinit\n"+a.bashCompletion())
	case "fish":
		fmt.Fprint(a.out, a.fishCompletion())
	default:
		return fmt.Errorf("unsupported shell %q (want bash, zsh or fish)", positional[0])
	}
	return nil
}

func (a *App) commandNames() string {
	names := make([]string, len(a.commands))
	for i, c := range a.commands {
		names[i] = c.name
	}
	return strings.Join(names, " ")
}

func (a *App) bashCompletion() string {
	var cases strings.Builder
	for _, c := range a.commands {
		words := commandWords[c.name]
		if c.name == "help" {
			words = a.commandNames()
		}
		if words != "" {
			fmt.Fprintf(&cases, "    %s) COMPREPLY=($(compgen -W %q -- \"$cur\")) ;;\n", c.name, words)
		}
	}

	return fmt.Sprintf(`# bash completion for task-cli
# Load with: source <(task-cli completion bash)
_task_cli() {
  local cur="${COMP_WORDS[COMP_CWORD]}"
  local prev="${COMP_WORDS[COMP_CWORD-1]}"

  if [[ $COMP_CWORD -eq 1 ]]; then
    COMPREPLY=($(compgen -W %q -- "$cur"))
    return
  fi

  case "$prev" in
    -p|--priority) COMPREPLY=($(compgen -W "low medium high" -- "$cur")); return ;;
    --sort) COMPREPLY=($(compgen -W "id priority due created updated" -- "$cur")); return ;;
    --due) COMPREPLY=($(compgen -W "today tomorrow none" -- "$cur")); return ;;
    -t|--tag) return ;;
  esac

  case "${COMP_WORDS[1]}" in
%s  esac
}
complete -F _task_cli task-cli
`, a.commandNames(), cases.String())
}

func (a *App) fishCompletion() string {
	var b strings.Builder
	b.WriteString("# fish completion for task-cli\n")
	b.WriteString("# Load with: task-cli completion fish | source\n")
	b.WriteString("complete -c task-cli -f\n")

	for _, c := range a.commands {
		fmt.Fprintf(&b, "complete -c task-cli -n __fish_use_subcommand -a %s -d %q\n", c.name, c.summary)
	}

	b.WriteString(`complete -c task-cli -n '__fish_seen_subcommand_from list' -a 'todo in-progress done'
complete -c task-cli -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'
complete -c task-cli -n '__fish_seen_subcommand_from help' -a '`)
	b.WriteString(a.commandNames())
	b.WriteString(`'
complete -c task-cli -n '__fish_seen_subcommand_from add update list' -s p -l priority -xa 'low medium high' -d 'Priority'
complete -c task-cli -n '__fish_seen_subcommand_from add update list' -s t -l tag -x -d 'Tag'
complete -c task-cli -n '__fish_seen_subcommand_from add update' -l due -xa 'today tomorrow none' -d 'Due date'
complete -c task-cli -n '__fish_seen_subcommand_from update' -l clear-tags -d 'Remove all tags'
complete -c task-cli -n '__fish_seen_subcommand_from list search' -l sort -xa 'id priority due created updated' -d 'Sort field'
complete -c task-cli -n '__fish_seen_subcommand_from list search show' -l json -d 'JSON output'
complete -c task-cli -n '__fish_seen_subcommand_from delete clear' -s f -l force -d 'Skip confirmation'
`)
	return b.String()
}
