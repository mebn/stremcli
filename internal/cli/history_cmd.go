package cli

import (
	"fmt"

	"github.com/mebn/stremcli/internal/history"
)

func (a *App) runHistory(args []string) error {
	var limit int
	fs := a.newFlagSet("history")
	fs.IntVar(&limit, "n", 20, "number of entries to show (0 for all)")
	if err := fs.Parse(args); err != nil {
		if isHelp(err) {
			return nil
		}
		return err
	}

	entries, err := history.Load()
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Fprintln(a.Err, "No history yet.")
		return nil
	}

	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}
	// Most recent first.
	for i := len(entries) - 1; i >= 0; i-- {
		fmt.Fprintln(a.Out, entries[i])
	}
	return nil
}
