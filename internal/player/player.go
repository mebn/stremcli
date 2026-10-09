// Package player launches external media players.
package player

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

// Player describes how to launch a media player.
type Player struct {
	// Binary is the executable looked up on PATH.
	Binary string
	// MacApp is the application name used with `open -a` on macOS when
	// Binary isn't on PATH.
	MacApp string
}

var players = map[string]Player{
	"iina": {Binary: "iina", MacApp: "IINA"},
	"vlc":  {Binary: "vlc", MacApp: "VLC"},
	"mpv":  {Binary: "mpv"},
}

// Names returns the supported player names.
func Names() []string {
	names := make([]string, 0, len(players))
	for n := range players {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Lookup returns the player registered under name.
func Lookup(name string) (Player, error) {
	p, ok := players[strings.ToLower(name)]
	if !ok {
		return Player{}, fmt.Errorf("unknown player %q (supported: %s)", name, strings.Join(Names(), ", "))
	}
	return p, nil
}

// Play starts the player with url and returns without waiting for it to exit.
func (p Player) Play(url string) error {
	cmd, err := p.command(url)
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", cmd.Path, err)
	}
	return cmd.Process.Release()
}

func (p Player) command(url string) (*exec.Cmd, error) {
	if bin, err := exec.LookPath(p.Binary); err == nil {
		return exec.Command(bin, url), nil
	}
	if runtime.GOOS == "darwin" && p.MacApp != "" {
		return exec.Command("open", "-a", p.MacApp, url), nil
	}
	return nil, fmt.Errorf("%s not found on PATH", p.Binary)
}
