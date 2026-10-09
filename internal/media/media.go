// Package media holds the domain types shared across the CLI.
package media

import (
	"fmt"
	"strings"
)

// Kind is the type of content being looked up.
type Kind string

const (
	Movie  Kind = "movie"
	Series Kind = "series"
)

// ParseKind accepts user-friendly names for a media kind.
func ParseKind(s string) (Kind, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "movie", "m", "film":
		return Movie, nil
	case "tv", "show", "series", "s":
		return Series, nil
	default:
		return "", fmt.Errorf("unknown type %q (want movie or tv)", s)
	}
}

// Title is a resolved catalog entry.
type Title struct {
	IMDbID string
	Name   string
	Year   string
	Kind   Kind
}

func (t Title) String() string {
	if t.Year == "" {
		return t.Name
	}
	return fmt.Sprintf("%s (%s)", t.Name, t.Year)
}

// Episode identifies a single episode of a series.
type Episode struct {
	Season int
	Number int
}

func (e Episode) String() string {
	return fmt.Sprintf("S%02dE%02d", e.Season, e.Number)
}
