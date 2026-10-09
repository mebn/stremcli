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

	"github.com/mebn/stremcli/internal/history"
	"github.com/mebn/stremcli/internal/player"
)

// trackCommand is the hidden command run in the background to record how
// far playback gets.
const trackCommand = "__track"

// saveEvery limits how often the position is written while playing.
const saveEvery = 15 * time.Second

// newSocketPath returns a fresh path for a player IPC socket. It lives in
// the temp dir because Unix socket paths are limited to ~104 bytes.
func newSocketPath() string {
	b := make([]byte, 4)
	rand.Read(b)
	return filepath.Join(os.TempDir(), "stremcli-"+hex.EncodeToString(b)+".sock")
}

// startTracker launches a detached stremcli that records the playback
// position in the history entry watched at watchedAt, so this process can
// exit while the user watches.
func startTracker(socket string, watchedAt time.Time) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, trackCommand, socket, watchedAt.Format(time.RFC3339Nano))
	// New session: not killed by Ctrl-C in the terminal or when it closes.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

// runTrack implements trackCommand: stremcli __track <socket> <watched_at>.
func (a *App) runTrack(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: stremcli %s <socket> <watched_at>", trackCommand)
	}
	socket := args[0]
	watchedAt, err := time.Parse(time.RFC3339Nano, args[1])
	if err != nil {
		return err
	}
	defer os.Remove(socket)

	var lastSave time.Time
	final, err := player.Track(ctx, socket, func(p player.Progress) {
		if time.Since(lastSave) >= saveEvery {
			lastSave = time.Now()
			history.SetPosition(watchedAt, p.Position, p.Duration)
		}
	})
	if final.Duration > 0 {
		if err := history.SetPosition(watchedAt, final.Position, final.Duration); err != nil {
			return err
		}
	}
	return err
}
