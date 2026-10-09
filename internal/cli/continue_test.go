package cli

import (
	"testing"

	"github.com/mebn/stremcli/internal/history"
	"github.com/mebn/stremcli/internal/media"
)

func TestLastEpisode(t *testing.T) {
	entries := []history.Entry{
		{Name: "MobLand", Kind: media.Series, Season: 1, Episode: 2},
		{Name: "Severance", Kind: media.Series, Season: 2, Episode: 1},
		{Name: "Dune", Kind: media.Movie},
	}
	if e, ok := lastEpisode(entries, ""); !ok || e.Name != "Severance" {
		t.Errorf("no query: got %+v, %v; want Severance", e, ok)
	}
	if e, ok := lastEpisode(entries, "mob"); !ok || e.Name != "MobLand" {
		t.Errorf("query mob: got %+v, %v; want MobLand", e, ok)
	}
	if _, ok := lastEpisode(entries, "dune"); ok {
		t.Error("query dune: matched a movie")
	}
}

func TestNextEpisode(t *testing.T) {
	ep := func(s, n int) media.Episode { return media.Episode{Season: s, Number: n} }
	eps := []media.Episode{ep(1, 1), ep(1, 2), ep(2, 1)}
	tests := []struct {
		watched media.Episode
		want    media.Episode
		ok      bool
	}{
		{ep(1, 1), ep(1, 2), true},
		{ep(1, 2), ep(2, 1), true},         // rolls over to next season
		{ep(2, 1), media.Episode{}, false}, // caught up
	}
	for _, tt := range tests {
		got, ok := nextEpisode(eps, tt.watched)
		if got != tt.want || ok != tt.ok {
			t.Errorf("after %s: got %s, %v; want %s, %v", tt.watched, got, ok, tt.want, tt.ok)
		}
	}
}
