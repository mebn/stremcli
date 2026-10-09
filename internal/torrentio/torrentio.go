// Package torrentio finds torrents for a title using the Torrentio Stremio addon.
package torrentio

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mebn/stremcli/internal/httpx"
	"github.com/mebn/stremcli/internal/media"
)

const (
	defaultBaseURL = "https://torrentio.strem.fun"
	// Skip camera rips; results come back sorted by quality.
	defaultOptions = "qualityfilter=scr,cam,unknown"
)

// Stream is a single torrent candidate.
type Stream struct {
	InfoHash string
	// FileIdx is the zero-based index of the target file inside the torrent,
	// or -1 if unknown.
	FileIdx  int
	Filename string
	// Quality is the normalized resolution, e.g. "4k", "1080p", "720p".
	Quality string
	Name    string
	Title   string
}

// Label is a single-line human readable description of the stream.
func (s Stream) Label() string {
	// Name is "<addon>\n<quality tags>"; drop the addon name.
	_, name, _ := strings.Cut(s.Name, "\n")
	title, _, _ := strings.Cut(s.Title, "\n")
	return strings.TrimSpace(name + " " + title)
}

// Details returns the extra metadata lines (seeders, size, source) on one line.
func (s Stream) Details() string {
	_, rest, _ := strings.Cut(s.Title, "\n")
	return strings.Join(strings.Fields(rest), " ")
}

// NormalizeQuality maps quality aliases to the names Torrentio uses.
func NormalizeQuality(q string) string {
	switch q = strings.ToLower(strings.TrimSpace(q)); q {
	case "2160p", "2160", "uhd":
		return "4k"
	case "1080", "720", "480":
		return q + "p"
	default:
		return q
	}
}

// FilterQuality returns the streams matching quality (see NormalizeQuality).
func FilterQuality(streams []Stream, quality string) []Stream {
	quality = NormalizeQuality(quality)
	var out []Stream
	for _, s := range streams {
		if s.Quality == quality {
			out = append(out, s)
		}
	}
	return out
}

// parseQuality extracts the resolution from a stream name such as
// "Torrentio\n4k HDR".
func parseQuality(name string) string {
	_, rest, _ := strings.Cut(name, "\n")
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	return NormalizeQuality(fields[0])
}

// Client queries Torrentio.
type Client struct {
	BaseURL string
	Options string
	HTTP    *http.Client
}

// New returns a Client using the public Torrentio instance.
func New(httpClient *http.Client) *Client {
	return &Client{BaseURL: defaultBaseURL, Options: defaultOptions, HTTP: httpClient}
}

type streamsResponse struct {
	Streams []struct {
		Name          string `json:"name"`
		Title         string `json:"title"`
		InfoHash      string `json:"infoHash"`
		FileIdx       *int   `json:"fileIdx"`
		BehaviorHints struct {
			Filename string `json:"filename"`
		} `json:"behaviorHints"`
	} `json:"streams"`
}

// Streams returns torrent candidates for a movie, or for an episode when ep is non-nil.
func (c *Client) Streams(ctx context.Context, title media.Title, ep *media.Episode) ([]Stream, error) {
	id := title.IMDbID
	if ep != nil {
		id = fmt.Sprintf("%s:%d:%d", id, ep.Season, ep.Number)
	}

	base := c.BaseURL
	if c.Options != "" {
		base += "/" + c.Options
	}
	endpoint := fmt.Sprintf("%s/stream/%s/%s.json", base, title.Kind, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var res streamsResponse
	if err := httpx.Do(c.HTTP, req, &res); err != nil {
		return nil, fmt.Errorf("fetch streams: %w", err)
	}

	streams := make([]Stream, 0, len(res.Streams))
	for _, s := range res.Streams {
		if s.InfoHash == "" {
			continue
		}
		idx := -1
		if s.FileIdx != nil {
			idx = *s.FileIdx
		}
		streams = append(streams, Stream{
			InfoHash: s.InfoHash,
			FileIdx:  idx,
			Filename: s.BehaviorHints.Filename,
			Quality:  parseQuality(s.Name),
			Name:     s.Name,
			Title:    s.Title,
		})
	}
	return streams, nil
}
