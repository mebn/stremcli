// Package progress remembers how far into each movie or episode the user
// got, in ~/.stremcli/progress.json.
package progress

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/mebn/stremcli/internal/config"
	"github.com/mebn/stremcli/internal/media"
)

const fileName = "progress.json"

// Entry is the saved position in one movie or episode.
type Entry struct {
	Position  float64   `json:"position"`           // seconds
	Duration  float64   `json:"duration,omitempty"` // seconds; 0 if unknown
	Finished  bool      `json:"finished"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Finished reports whether playback stopped close enough to the end (into
// the credits) to count as watched.
func Finished(position, duration float64) bool {
	if duration <= 0 {
		return false
	}
	return duration-position <= math.Max(duration*0.05, 120)
}

// Resumable reports whether e is worth resuming from rather than starting
// over.
func (e Entry) Resumable() bool {
	return !e.Finished && e.Position >= 30
}

// Key identifies a movie (ep == nil) or an episode of a series.
func Key(imdbID string, ep *media.Episode) string {
	if ep == nil {
		return imdbID
	}
	return fmt.Sprintf("%s:%d:%d", imdbID, ep.Season, ep.Number)
}

func path() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// Load returns all saved progress by Key. A missing file yields an empty map.
func Load() (map[string]Entry, error) {
	entries := map[string]Entry{}
	p, err := path()
	if err != nil {
		return entries, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return entries, nil
	}
	if err != nil {
		return entries, fmt.Errorf("read progress: %w", err)
	}
	if err := json.Unmarshal(data, &entries); err != nil {
		return entries, fmt.Errorf("parse %s: %w", p, err)
	}
	return entries, nil
}

// Get returns the saved progress for key, if any.
func Get(key string) (Entry, bool, error) {
	entries, err := Load()
	e, ok := entries[key]
	return e, ok, err
}

// Set saves the progress for key.
func Set(key string, e Entry) error {
	entries, err := Load()
	if err != nil {
		return err
	}
	entries[key] = e

	dir, err := config.EnsureDir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	// Write to a temp file and rename so readers never see a partial file.
	tmp, err := os.CreateTemp(dir, fileName+".*")
	if err != nil {
		return fmt.Errorf("write progress: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return fmt.Errorf("write progress: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write progress: %w", err)
	}
	if err := os.Rename(tmp.Name(), filepath.Join(dir, fileName)); err != nil {
		return fmt.Errorf("write progress: %w", err)
	}
	return nil
}

// FormatTime renders seconds as h:mm:ss or m:ss.
func FormatTime(sec float64) string {
	t := int(sec)
	h, m, s := t/3600, t/60%60, t%60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
