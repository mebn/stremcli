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

	// With no flags, prompt for the token interactively.
	if token == "" && playerName == "" && quality == "" {
		fmt.Fprintf(a.Err, "Get your Real-Debrid API key at %s\n", apiTokenURL)
		if token, err = a.prompt(ctx, "API key: "); err != nil {
			return err
		}
		if token == "" {
			return errors.New("no API key given")
		}
	}
	if token != "" {
		if err := a.checkToken(ctx, token); err != nil {
			return err
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
	if errors.Is(err, realdebrid.ErrUnauthorized) {
		return realdebrid.ErrUnauthorized
	}
	if err != nil {
		return fmt.Errorf("verify API key: %w", err)
	}
	fmt.Fprintf(a.Err, "Logged in as %s (%s)\n", user.Username, user.Type)
	if user.Type != "premium" {
		fmt.Fprintln(a.Err, "warning: streaming needs a premium Real-Debrid account")
	}
	return nil
}

// prompt prints msg and reads a line from In. It returns as soon as ctx is
// cancelled (Ctrl-C) instead of blocking on the read.
func (a *App) prompt(ctx context.Context, msg string) (string, error) {
	fmt.Fprint(a.Err, msg)
	line := make(chan string, 1)
	go func() {
		s, _ := bufio.NewReader(a.In).ReadString('\n')
		line <- strings.TrimSpace(s)
	}()
	select {
	case <-ctx.Done():
		fmt.Fprintln(a.Err)
		return "", ctx.Err()
	case s := <-line:
		return s, nil
	}
}
