package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mebn/stremcli/internal/catalog"
	"github.com/mebn/stremcli/internal/config"
	"github.com/mebn/stremcli/internal/history"
	"github.com/mebn/stremcli/internal/media"
	"github.com/mebn/stremcli/internal/player"
	"github.com/mebn/stremcli/internal/progress"
	"github.com/mebn/stremcli/internal/realdebrid"
	"github.com/mebn/stremcli/internal/torrentio"
)

// maxAttempts caps how many torrents are tried before giving up.
const maxAttempts = 10

var validQualities = map[string]bool{"4k": true, "1080p": true, "720p": true, "480p": true}

type playOptions struct {
	kind       media.Kind
	episode    *media.Episode
	playerName string // empty means just print the link
	player     player.Player
	query      string
	quality    string // empty means any quality
	pick       int    // 1-based result to use; 0 means first cached
	list       bool   // list results instead of playing
}

func (a *App) runPlay(ctx context.Context, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	opts, err := a.parsePlayOptions(args, cfg)
	if err != nil {
		return err
	}

	title, err := catalog.New(a.HTTP).Search(ctx, opts.kind, opts.query)
	if err != nil {
		return err
	}
	return a.play(ctx, cfg, opts, title)
}

// play finds a stream for title and opens it according to opts.
func (a *App) play(ctx context.Context, cfg config.Config, opts playOptions, title media.Title) error {
	label := title.String()
	if opts.episode != nil {
		label += " " + opts.episode.String()
	}
	fmt.Fprintf(a.Err, "Found %s [%s]\n", label, title.IMDbID)

	streams, err := torrentio.New(a.HTTP).Streams(ctx, title, opts.episode)
	if err != nil {
		return err
	}
	if opts.quality != "" {
		streams = torrentio.FilterQuality(streams, opts.quality)
	} else if cfg.Quality != "" {
		streams = torrentio.PreferQuality(streams, cfg.Quality)
	}
	if len(streams) == 0 {
		return fmt.Errorf("no torrents found for %s", label)
	}

	if opts.list {
		a.listStreams(streams)
		return nil
	}

	token := os.Getenv("REAL_DEBRID_TOKEN")
	if token == "" {
		token = cfg.RealDebridToken
	}
	if token == "" {
		return fmt.Errorf("no Real-Debrid token configured; get one at %s and run: stremcli config -token <KEY>", apiTokenURL)
	}

	if opts.pick > 0 {
		if opts.pick > len(streams) {
			return fmt.Errorf("-pick %d is out of range: only %d results (see -list)", opts.pick, len(streams))
		}
		// Only try the chosen result; don't fall back to others.
		streams = streams[opts.pick-1 : opts.pick]
	}

	stream, link, err := a.resolveFirst(ctx, realdebrid.New(a.HTTP, token), streams)
	if err != nil {
		return err
	}

	fmt.Fprintln(a.Out, link)

	if opts.playerName != "" {
		if err := a.launch(opts, progress.Key(title.IMDbID, opts.episode), link); err != nil {
			return err
		}
	}

	entry := history.NewEntry(title, opts.episode, opts.playerName, stream.Label())
	if err := history.Append(entry); err != nil {
		fmt.Fprintf(a.Err, "warning: could not save history: %v\n", err)
	}
	return nil
}

// streamFlags are the flags shared by play and continue.
type streamFlags struct {
	playerName, quality string
	pick                int
	list                bool
}

func (f *streamFlags) register(fs *flag.FlagSet, cfg config.Config) {
	for _, name := range []string{"player", "p"} {
		fs.StringVar(&f.playerName, name, cfg.Player, "player to open the stream in")
	}
	for _, name := range []string{"quality", "q"} {
		fs.StringVar(&f.quality, name, "", "only use results of this quality")
	}
	fs.IntVar(&f.pick, "pick", 0, "use this result number (see -list)")
	for _, name := range []string{"list", "l"} {
		fs.BoolVar(&f.list, name, false, "list results instead of playing")
	}
}

// options validates the flags and turns them into playOptions.
func (f *streamFlags) options() (playOptions, error) {
	opts := playOptions{
		playerName: strings.ToLower(f.playerName),
		quality:    f.quality,
		pick:       f.pick,
		list:       f.list,
	}
	if opts.playerName != "" {
		var err error
		if opts.player, err = player.Lookup(opts.playerName); err != nil {
			return opts, err
		}
	}
	if f.pick < 0 {
		return opts, errors.New("-pick must be >= 1")
	}
	if f.quality != "" && !validQualities[torrentio.NormalizeQuality(f.quality)] {
		return opts, fmt.Errorf("unknown quality %q (want 4k, 1080p, 720p or 480p)", f.quality)
	}
	return opts, nil
}

// launch opens link in the chosen player, resuming from saved progress and
// recording new progress in the background when the player supports it.
func (a *App) launch(opts playOptions, key, link string) error {
	var po player.Options
	if opts.player.CanTrack() {
		saved, ok, err := progress.Get(key)
		if err != nil {
			fmt.Fprintf(a.Err, "warning: %v\n", err)
		}
		if ok && saved.Resumable() {
			// Back up a little to pick up the thread.
			po.Start = max(saved.Position-5, 0)
			fmt.Fprintf(a.Err, "Resuming at %s\n", progress.FormatTime(po.Start))
		}
		po.Socket = newSocketPath()
	}

	fmt.Fprintf(a.Err, "Opening in %s\n", opts.playerName)
	if err := opts.player.Play(link, po); err != nil {
		return err
	}
	if po.Socket != "" {
		if err := startTracker(po.Socket, key); err != nil {
			fmt.Fprintf(a.Err, "warning: can't track progress: %v\n", err)
		}
	}
	return nil
}

func (a *App) parsePlayOptions(args []string, cfg config.Config) (playOptions, error) {
	var (
		kindStr         string
		season, episode int
		sf              streamFlags
	)
	fs := a.newFlagSet("stremcli")
	for _, name := range []string{"type", "t"} {
		fs.StringVar(&kindStr, name, "movie", "movie or tv")
	}
	for _, name := range []string{"season", "s"} {
		fs.IntVar(&season, name, 0, "season number")
	}
	for _, name := range []string{"episode", "e"} {
		fs.IntVar(&episode, name, 0, "episode number")
	}
	sf.register(fs, cfg)

	positional, err := parseInterspersed(fs, args)
	if err != nil {
		if isHelp(err) {
			return playOptions{}, err
		}
		return playOptions{}, fmt.Errorf("%w\n\n%s", err, usage)
	}

	opts, err := sf.options()
	if err != nil {
		return opts, err
	}
	opts.query = strings.Join(positional, " ")
	if opts.query == "" {
		return opts, fmt.Errorf("missing title\n\n%s", usage)
	}
	if opts.kind, err = media.ParseKind(kindStr); err != nil {
		return opts, err
	}

	switch opts.kind {
	case media.Series:
		if season < 1 || episode < 1 {
			return opts, errors.New("tv shows need -season and -episode (both >= 1)")
		}
		opts.episode = &media.Episode{Season: season, Number: episode}
	case media.Movie:
		if season != 0 || episode != 0 {
			return opts, errors.New("-season and -episode only apply to -type tv")
		}
	}
	return opts, nil
}

// resolveFirst tries streams in order and returns the first one Real-Debrid
// can serve immediately.
func (a *App) resolveFirst(ctx context.Context, rd *realdebrid.Client, streams []torrentio.Stream) (torrentio.Stream, string, error) {
	resolver := realdebrid.NewResolver(rd)

	for i, s := range streams {
		if i == maxAttempts {
			break
		}
		fmt.Fprintf(a.Err, "Trying %s\n", s.Label())

		link, err := resolver.Resolve(ctx, realdebrid.Target{
			InfoHash: s.InfoHash,
			Filename: s.Filename,
			FileIdx:  s.FileIdx,
		})
		switch {
		case err == nil:
			return s, link, nil
		case ctx.Err() != nil:
			return s, "", ctx.Err()
		case errors.Is(err, realdebrid.ErrUnauthorized):
			return s, "", fmt.Errorf("%w (update it with: stremcli config -token <KEY>)", err)
		case errors.Is(err, realdebrid.ErrNotCached):
			fmt.Fprintln(a.Err, "  not cached, skipping")
		default:
			fmt.Fprintf(a.Err, "  failed: %v\n", err)
		}
	}
	return torrentio.Stream{}, "", errors.New("no cached torrent found on Real-Debrid")
}

func (a *App) listStreams(streams []torrentio.Stream) {
	for i, s := range streams {
		fmt.Fprintf(a.Out, "%3d  %s\n", i+1, s.Label())
		if d := s.Details(); d != "" {
			fmt.Fprintf(a.Out, "     %s\n", d)
		}
	}
}
