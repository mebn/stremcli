package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/mebn/stremcli/internal/catalog"
	"github.com/mebn/stremcli/internal/config"
	"github.com/mebn/stremcli/internal/history"
	"github.com/mebn/stremcli/internal/media"
	"github.com/mebn/stremcli/internal/progress"
)

// runContinue plays the episode after the last one watched of the most
// recent show in history, or of the show matching the given name.
func (a *App) runContinue(ctx context.Context, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var sf streamFlags
	fs := a.newFlagSet("continue")
	sf.register(fs, cfg)
	positional, err := parseInterspersed(fs, args)
	if err != nil {
		if isHelp(err) {
			return nil
		}
		return err
	}
	opts, err := sf.options()
	if err != nil {
		return err
	}

	entries, err := history.Load()
	if err != nil {
		return err
	}
	last, ok := lastEpisode(entries, strings.Join(positional, " "))
	if !ok {
		if len(positional) > 0 {
			return fmt.Errorf("no show matching %q in history", strings.Join(positional, " "))
		}
		return fmt.Errorf("no shows in history yet; play one with: stremcli -t tv -s 1 -e 1 <title>")
	}

	title := media.Title{IMDbID: last.IMDbID, Name: last.Name, Year: last.Year, Kind: media.Series}
	watched := media.Episode{Season: last.Season, Number: last.Episode}
	opts.kind = media.Series

	// Pick up an episode that was stopped partway through. Without saved
	// progress (e.g. VLC), assume it was watched to the end.
	saved, ok, err := progress.Get(progress.Key(title.IMDbID, &watched))
	if err != nil {
		return err
	}
	if ok && saved.Resumable() {
		fmt.Fprintf(a.Err, "Resuming %s %s, stopped at %s\n", title, watched, progress.FormatTime(saved.Position))
		opts.episode = &watched
		return a.play(ctx, cfg, opts, title)
	}

	eps, err := catalog.New(a.HTTP).Episodes(ctx, title.IMDbID)
	if err != nil {
		return err
	}
	next, ok := nextEpisode(eps, watched)
	if !ok {
		fmt.Fprintf(a.Err, "You're caught up on %s (last watched %s).\n", title, watched)
		return nil
	}
	fmt.Fprintf(a.Err, "Last watched %s %s, continuing with %s\n", title, watched, next)

	opts.episode = &next
	return a.play(ctx, cfg, opts, title)
}

// lastEpisode returns the most recent series entry whose name contains query
// (case-insensitively); an empty query matches any show.
func lastEpisode(entries []history.Entry, query string) (history.Entry, bool) {
	query = strings.ToLower(strings.TrimSpace(query))
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if e.Kind == media.Series && strings.Contains(strings.ToLower(e.Name), query) {
			return e, true
		}
	}
	return history.Entry{}, false
}

// nextEpisode returns the first of eps (sorted) that comes after watched.
func nextEpisode(eps []media.Episode, watched media.Episode) (media.Episode, bool) {
	for _, ep := range eps {
		if ep.Compare(watched) > 0 {
			return ep, true
		}
	}
	return media.Episode{}, false
}
