// Package history records what the user has watched in ~/.stremcli/history.jsonl.
package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/mebn/stremcli/internal/config"
	"github.com/mebn/stremcli/internal/media"
)

const fileName = "history.jsonl"

// Entry is one watched item.
type Entry struct {
	WatchedAt time.Time  `json:"watched_at"`
	IMDbID    string     `json:"imdb_id"`
	Name      string     `json:"name"`
	Year      string     `json:"year,omitempty"`
	Kind      media.Kind `json:"kind"`
	Season    int        `json:"season,omitempty"`
	Episode   int        `json:"episode,omitempty"`
	Player    string     `json:"player,omitempty"`
	Source    string     `json:"source,omitempty"`
}

// NewEntry builds an entry for title (and ep, if it's a series) watched now.
func NewEntry(title media.Title, ep *media.Episode, player, source string) Entry {
	e := Entry{
		WatchedAt: time.Now(),
		IMDbID:    title.IMDbID,
		Name:      title.Name,
		Year:      title.Year,
		Kind:      title.Kind,
		Player:    player,
		Source:    source,
	}
	if ep != nil {
		e.Season, e.Episode = ep.Season, ep.Number
	}
	return e
}

func (e Entry) String() string {
	s := fmt.Sprintf("%s  %s", e.WatchedAt.Local().Format("2006-01-02 15:04"), e.Name)
	if e.Year != "" {
		s += fmt.Sprintf(" (%s)", e.Year)
	}
	if e.Kind == media.Series {
		s += " " + media.Episode{Season: e.Season, Number: e.Episode}.String()
	}
	return s
}

func path() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// Append adds an entry to the history file.
func Append(e Entry) error {
	dir, err := config.EnsureDir()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, fileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open history: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(e); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	return nil
}

// Load returns all history entries, oldest first. Malformed lines are skipped.
func Load() ([]Entry, error) {
	p, err := path()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open history: %w", err)
	}
	defer f.Close()

	var entries []Entry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var e Entry
		if json.Unmarshal(sc.Bytes(), &e) == nil {
			entries = append(entries, e)
		}
	}
	return entries, sc.Err()
}
