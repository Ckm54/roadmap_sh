package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/ckm54/task_tracker/internal/task"
)

// ANSI SGR codes.
const (
	styleBold   = "1"
	styleDim    = "2"
	styleRed    = "31"
	styleGreen  = "32"
	styleYellow = "33"
	styleCyan   = "36"
)

func (a *App) paint(s string, styles ...string) string {
	if !a.color || len(styles) == 0 {
		return s
	}
	return "\x1b[" + strings.Join(styles, ";") + "m" + s + "\x1b[0m"
}

func statusStyle(s task.Status) []string {
	switch s {
	case task.StatusDone:
		return []string{styleGreen}
	case task.StatusInProgress:
		return []string{styleCyan}
	default:
		return []string{styleYellow}
	}
}

func priorityStyle(p task.Priority) []string {
	switch p {
	case task.PriorityHigh:
		return []string{styleRed, styleBold}
	case task.PriorityLow:
		return []string{styleDim}
	}
	return nil
}

type cell struct {
	text   string
	styles []string
}

// printTable pads cells by their visible width first and colours them
// afterwards - escape codes never throw the columns out of line.
func (a *App) printTable(tasks []task.TaskEntity) {
	now := a.now()
	header := []string{"ID", "STATUS", "PRIORITY", "DUE", "DESCRIPTION", "TAGS"}

	rows := make([][]cell, 0, len(tasks))
	counts := map[task.Status]int{}
	for _, t := range tasks {
		counts[t.Status]++

		due := cell{text: "-", styles: []string{styleDim}}
		if t.Due != nil {
			due = cell{text: t.Due.Format(task.DateLayout)}
			if t.Overdue(now) {
				due.styles = []string{styleRed}
			}
		}

		description := cell{text: t.Description}
		if t.Status == task.StatusDone {
			description.styles = []string{styleDim}
		}

		tags := cell{text: "-", styles: []string{styleDim}}
		if len(t.Tags) > 0 {
			tags = cell{text: strings.Join(t.Tags, ", ")}
		}

		rows = append(rows, []cell{
			{text: fmt.Sprint(t.ID)},
			{text: string(t.Status), styles: statusStyle(t.Status)},
			{text: string(t.Priority), styles: priorityStyle(t.Priority)},
			due,
			description,
			tags,
		})
	}

	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, c := range row {
			widths[i] = max(widths[i], utf8.RuneCountInString(c.text))
		}
	}

	writeRow := func(row []cell) {
		var b strings.Builder
		for i, c := range row {
			b.WriteString(a.paint(c.text, c.styles...))
			if i < len(row)-1 {
				b.WriteString(strings.Repeat(" ", widths[i]-utf8.RuneCountInString(c.text)+2))
			}
		}
		fmt.Fprintln(a.out, b.String())
	}

	headerRow := make([]cell, len(header))
	for i, h := range header {
		headerRow[i] = cell{text: h, styles: []string{styleBold}}
	}
	writeRow(headerRow)
	for _, row := range rows {
		writeRow(row)
	}

	fmt.Fprintf(a.out, "\n%s: %d todo, %d in-progress, %d done\n",
		plural(len(tasks), "task"),
		counts[task.StatusTodo], counts[task.StatusInProgress], counts[task.StatusDone])
}

func (a *App) printDetails(t task.TaskEntity) {
	due := "-"
	if t.Due != nil {
		due = t.Due.Format(task.DateLayout)
		if t.Overdue(a.now()) {
			due = a.paint(due+" (overdue)", styleRed)
		}
	}

	tags := "-"
	if len(t.Tags) > 0 {
		tags = strings.Join(t.Tags, ", ")
	}

	const timeLayout = "2006-01-02 15:04"
	fields := []struct{ label, value string }{
		{"ID", fmt.Sprint(t.ID)},
		{"Description", t.Description},
		{"Status", a.paint(string(t.Status), statusStyle(t.Status)...)},
		{"Priority", a.paint(string(t.Priority), priorityStyle(t.Priority)...)},
		{"Due", due},
		{"Tags", tags},
		{"Created", t.CreatedAt.Local().Format(timeLayout)},
		{"Updated", t.UpdatedAt.Local().Format(timeLayout)},
	}
	for _, f := range fields {
		fmt.Fprintf(a.out, "%-12s %s\n", f.label+":", f.value)
	}
}

func (a *App) printJSON(v any) error {
	if tasks, ok := v.([]task.TaskEntity); ok && tasks == nil {
		v = []task.TaskEntity{}
	}
	enc := json.NewEncoder(a.out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
