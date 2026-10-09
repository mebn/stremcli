package torrentio

import (
	"slices"
	"testing"
)

func TestPreferQuality(t *testing.T) {
	streams := []Stream{
		{Title: "a", Quality: "4k"},
		{Title: "b", Quality: "1080p"},
		{Title: "c", Quality: "720p"},
		{Title: "d", Quality: "1080p"},
	}
	var got []string
	for _, s := range PreferQuality(streams, "1080P") {
		got = append(got, s.Title)
	}
	if want := []string{"b", "d", "a", "c"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
