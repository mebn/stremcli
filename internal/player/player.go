// Package player launches external media players.
package player

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

// Player describes how to launch a media player.
type Player struct {
	// Binary is the executable looked up on PATH.
	Binary string
	// MacBinary is the executable inside the macOS app bundle, used when
	// Binary isn't on PATH.
	MacBinary string
	// MacApp is the application name used with `open -a` on macOS when
	// neither Binary nor MacBinary is found.
	MacApp string
	// ipcArgs returns the arguments that make the player serve mpv's JSON IPC
	// on socket and start at start seconds. Nil if the player has no IPC.
	ipcArgs func(socket string, start float64) []string
}

// Options controls how a stream is opened.
type Options struct {
	// Socket, if set, is where the player serves mpv's JSON IPC (see Track).
	// Ignored by players that can't be tracked.
	Socket string
	// Start is the position in seconds to start playing from.
	Start float64
}

var players = map[string]Player{
	"iina": {
		Binary:    "iina",
		MacBinary: "/Applications/IINA.app/Contents/MacOS/iina-cli",
		MacApp:    "IINA",
		// iina-cli passes --mpv-* options through to mpv. IINA's own resume
		// would override -start, so it is turned off.
		ipcArgs: func(socket string, start float64) []string {
			return []string{
				"--no-stdin",
				"--mpv-input-ipc-server=" + socket,
				"--mpv-resume-playback=no",
				fmt.Sprintf("--mpv-start=%.0f", start),
			}
		},
	},
	"vlc": {Binary: "vlc", MacApp: "VLC"},
	"mpv": {
		Binary:    "mpv",
		MacBinary: "/Applications/mpv.app/Contents/MacOS/mpv",
		ipcArgs: func(socket string, start float64) []string {
			return []string{
				"--input-ipc-server=" + socket,
				"--resume-playback=no",
				fmt.Sprintf("--start=%.0f", start),
			}
		},
	},
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

// CanTrack reports whether the player can report its position (see Track).
func (p Player) CanTrack() bool {
	if p.ipcArgs == nil {
		return false
	}
	_, ok := p.binary()
	return ok
}

// Play starts the player with url and returns without waiting for it to exit.
func (p Player) Play(url string, opts Options) error {
	cmd, err := p.command(url, opts)
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", cmd.Path, err)
	}
	return cmd.Process.Release()
}

func (p Player) command(url string, opts Options) (*exec.Cmd, error) {
	if bin, ok := p.binary(); ok {
		var args []string
		if p.ipcArgs != nil && (opts.Socket != "" || opts.Start > 0) {
			args = p.ipcArgs(opts.Socket, opts.Start)
		}
		return exec.Command(bin, append(args, url)...), nil
	}
	if runtime.GOOS == "darwin" && p.MacApp != "" {
		return exec.Command("open", "-a", p.MacApp, url), nil
	}
	return nil, fmt.Errorf("%s not found on PATH", p.Binary)
}

// binary returns the player's executable, if installed.
func (p Player) binary() (string, bool) {
	if bin, err := exec.LookPath(p.Binary); err == nil {
		return bin, true
	}
	if runtime.GOOS == "darwin" && p.MacBinary != "" {
		if _, err := os.Stat(p.MacBinary); err == nil {
			return p.MacBinary, true
		}
	}
	return "", false
}
