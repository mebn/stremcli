package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/mebn/stremcli/internal/player"
	"github.com/mebn/stremcli/internal/progress"
)

// trackCommand is the hidden command run in the background to record how
// far playback gets.
const trackCommand = "__track"

// saveEvery limits how often progress is written while playing.
const saveEvery = 15 * time.Second

// newSocketPath returns a fresh path for a player IPC socket. It lives in
// the temp dir because Unix socket paths are limited to ~104 bytes.
func newSocketPath() string {
	b := make([]byte, 4)
	rand.Read(b)
	return filepath.Join(os.TempDir(), "stremcli-"+hex.EncodeToString(b)+".sock")
}

// startTracker launches a detached stremcli that records progress for key
// from the player's socket, so this process can exit while the user watches.
func startTracker(socket, key string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, trackCommand, socket, key)
	// New session: not killed by Ctrl-C in the terminal or when it closes.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

// runTrack implements trackCommand: stremcli __track <socket> <key>.
func (a *App) runTrack(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: stremcli %s <socket> <key>", trackCommand)
	}
	socket, key := args[0], args[1]
	defer os.Remove(socket)

	save := func(p player.Progress) error {
		return progress.Set(key, progress.Entry{
			Position:  p.Position,
			Duration:  p.Duration,
			Finished:  p.Ended || progress.Finished(p.Position, p.Duration),
			UpdatedAt: time.Now(),
		})
	}

	var lastSave time.Time
	final, err := player.Track(ctx, socket, func(p player.Progress) {
		if time.Since(lastSave) >= saveEvery && p.Position > 0 {
			lastSave = time.Now()
			save(p)
		}
	})
	if final.Position > 0 || final.Ended {
		if err := save(final); err != nil {
			return err
		}
	}
	return err
}
