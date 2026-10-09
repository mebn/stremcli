// Package cli wires the stremcli commands together.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"
)

const usage = `stremcli - stream movies and TV shows through Real-Debrid

Usage:
  stremcli [flags] <title>          find a stream and play it
  stremcli config [-token KEY] [-player NAME]
                                    save settings to ~/.stremcli/config.json
  stremcli history [-n N]           show watch history

Play flags:
  -type, -t     movie or tv (default movie)
  -season, -s   season number (tv only)
  -episode, -e  episode number (tv only)
  -player, -p   iina, vlc or mpv (default from config; prints the link if unset)
  -quality, -q  only use results of this quality: 4k, 1080p, 720p or 480p
  -list, -l     list numbered results instead of playing
  -pick N       use result number N from -list (no fallback to other results)

The Real-Debrid token is read from $REAL_DEBRID_TOKEN, falling back to the
saved config. Get yours at https://real-debrid.com/apitoken.

Examples:
  stremcli config -token ABC123
  stremcli -p iina "The Matrix"
  stremcli -t tv -s 1 -e 3 -p vlc "Breaking Bad"
  stremcli -q 1080p -l "Dune"
  stremcli -q 1080p -pick 2 -p iina "Dune"
`

// App holds the I/O streams and shared dependencies for commands.
type App struct {
	In   io.Reader
	Out  io.Writer // results (e.g. the stream link)
	Err  io.Writer // progress and diagnostics
	HTTP *http.Client
}

// New returns an App using the given streams.
func New(in io.Reader, out, errOut io.Writer) *App {
	return &App{
		In:   in,
		Out:  out,
		Err:  errOut,
		HTTP: &http.Client{Timeout: 30 * time.Second},
	}
}

// Run dispatches args (without the program name) to a command.
func (a *App) Run(ctx context.Context, args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "config":
			return a.runConfig(args[1:])
		case "history":
			return a.runHistory(args[1:])
		case "help", "-h", "-help", "--help":
			fmt.Fprint(a.Out, usage)
			return nil
		}
	}
	return a.runPlay(ctx, args)
}

func (a *App) newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(a.Err)
	fs.Usage = func() { fmt.Fprint(a.Err, usage) }
	return fs
}

// parseInterspersed parses flags that may appear before, after or between
// positional arguments, returning the positionals.
func parseInterspersed(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		if fs.NArg() == 0 {
			return positional, nil
		}
		positional = append(positional, fs.Arg(0))
		args = fs.Args()[1:]
	}
}

// isHelp reports whether err came from -h/-help, which the flag set already
// handled by printing usage.
func isHelp(err error) bool {
	return errors.Is(err, flag.ErrHelp)
}
