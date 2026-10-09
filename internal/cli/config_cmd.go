package cli

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/mebn/stremcli/internal/config"
	"github.com/mebn/stremcli/internal/player"
)

const apiTokenURL = "https://real-debrid.com/apitoken"

func (a *App) runConfig(args []string) error {
	var token, playerName string
	fs := a.newFlagSet("config")
	fs.StringVar(&token, "token", "", "Real-Debrid API token")
	fs.StringVar(&playerName, "player", "", "default player")
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

	// With no flags, prompt for the token interactively.
	if token == "" && playerName == "" {
		fmt.Fprintf(a.Err, "Get your Real-Debrid API key at %s\n", apiTokenURL)
		if token, err = a.prompt("API key: "); err != nil {
			return err
		}
		if token == "" {
			return errors.New("no token given")
		}
	}

	if token != "" {
		cfg.RealDebridToken = token
	}
	if playerName != "" {
		if _, err := player.Lookup(playerName); err != nil {
			return err
		}
		cfg.Player = strings.ToLower(playerName)
	}

	path, err := config.Save(cfg)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Err, "Saved config to %s\n", path)
	return nil
}

func (a *App) prompt(msg string) (string, error) {
	fmt.Fprint(a.Err, msg)
	line, err := bufio.NewReader(a.In).ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("read input: %w", err)
	}
	return strings.TrimSpace(line), nil
}
