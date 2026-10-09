package cli

import (
	"fmt"

	"github.com/mebn/stremcli/internal/history"
	"github.com/mebn/stremcli/internal/media"
	"github.com/mebn/stremcli/internal/progress"
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
	saved, err := progress.Load()
	if err != nil {
		fmt.Fprintf(a.Err, "warning: %v\n", err)
	}

	// Most recent first.
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		line := e.String()
		var ep *media.Episode
		if e.Kind == media.Series {
			ep = &media.Episode{Season: e.Season, Number: e.Episode}
		}
		if p, ok := saved[progress.Key(e.IMDbID, ep)]; ok {
			switch {
			case p.Finished:
				line += "  [finished]"
			case p.Duration > 0:
				line += fmt.Sprintf("  [%s / %s]", progress.FormatTime(p.Position), progress.FormatTime(p.Duration))
			default:
				line += fmt.Sprintf("  [%s]", progress.FormatTime(p.Position))
			}
		}
		fmt.Fprintln(a.Out, line)
	}
	return nil
}
