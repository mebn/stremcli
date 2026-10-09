package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mebn/stremcli/internal/config"
	"github.com/mebn/stremcli/internal/player"
	"github.com/mebn/stremcli/internal/realdebrid"
	"github.com/mebn/stremcli/internal/torrentio"
)

const apiTokenURL = "https://real-debrid.com/apitoken"

func (a *App) runConfig(ctx context.Context, args []string) error {
	var token, playerName, quality string
	fs := a.newFlagSet("config")
	fs.StringVar(&token, "token", "", "Real-Debrid API token")
	fs.StringVar(&playerName, "player", "", "default player")
	fs.StringVar(&quality, "quality", "", "preferred quality")
	if err := fs.Parse(args); err != nil {
		if isHelp(err) {
			return nil
		}
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if playerName != "" {
		if _, err := player.Lookup(playerName); err != nil {
			return err
		}
		cfg.Player = strings.ToLower(playerName)
	}
	if quality != "" {
		q := torrentio.NormalizeQuality(quality)
		if !validQualities[q] {
			return fmt.Errorf("unknown quality %q (want 4k, 1080p, 720p or 480p)", quality)
		}
		cfg.Quality = q
	}

	switch {
	case token != "":
		if err := a.checkToken(ctx, token); err != nil {
			return err
		}
		cfg.RealDebridToken = token
	case playerName == "" && quality == "":
		// With no flags, prompt for the token interactively until a valid
		// one is given.
		fmt.Fprintf(a.Err, "Get your Real-Debrid API key at %s\n", apiTokenURL)
		for {
			if token, err = a.prompt(ctx, "API key: "); err != nil {
				return err
			}
			if token == "" {
				continue
			}
			err = a.checkToken(ctx, token)
			if errors.Is(err, realdebrid.ErrUnauthorized) {
				fmt.Fprintln(a.Err, "Real-Debrid rejected that key, try again.")
				continue
			}
			if err != nil {
				return err
			}
			break
		}
		cfg.RealDebridToken = token
	}

	path, err := config.Save(cfg)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Err, "Saved config to %s\n", path)
	return nil
}

// checkToken verifies token against Real-Debrid and reports the account.
func (a *App) checkToken(ctx context.Context, token string) error {
	user, err := realdebrid.New(a.HTTP, token).User(ctx)
	if err != nil {
		if errors.Is(err, realdebrid.ErrUnauthorized) {
			return realdebrid.ErrUnauthorized
		}
		return fmt.Errorf("verify API key: %w", err)
	}
	fmt.Fprintf(a.Err, "Logged in as %s (%s)\n", user.Username, user.Type)
	if user.Type != "premium" {
		fmt.Fprintln(a.Err, "warning: streaming needs a premium Real-Debrid account")
	}
	return nil
}

type readResult struct {
	line string
	err  error
}

// prompt prints msg and reads a line from In. It returns ctx.Err() as soon as
// ctx is cancelled (e.g. Ctrl-C) instead of blocking on the read.
func (a *App) prompt(ctx context.Context, msg string) (string, error) {
	if a.lines == nil {
		a.lines = make(chan readResult)
		go func() {
			r := bufio.NewReader(a.In)
			for {
				line, err := r.ReadString('\n')
				a.lines <- readResult{line, err}
				if err != nil {
					return
				}
			}
		}()
	}

	fmt.Fprint(a.Err, msg)
	select {
	case <-ctx.Done():
		fmt.Fprintln(a.Err)
		return "", ctx.Err()
	case res := <-a.lines:
		if res.err != nil && res.line == "" {
			return "", fmt.Errorf("read input: %w", res.err)
		}
		return strings.TrimSpace(res.line), nil
	}
}
