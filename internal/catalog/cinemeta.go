// Package catalog resolves free-text titles to IMDb IDs using Stremio's
// Cinemeta addon.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/mebn/stremcli/internal/httpx"
	"github.com/mebn/stremcli/internal/media"
)

const defaultBaseURL = "https://v3-cinemeta.strem.io"

// ErrNotFound is returned when a search yields no results.
var ErrNotFound = errors.New("no matching title found")

// Client searches the Cinemeta catalog.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// New returns a Client using the public Cinemeta instance.
func New(httpClient *http.Client) *Client {
	return &Client{BaseURL: defaultBaseURL, HTTP: httpClient}
}

type searchResponse struct {
	Metas []struct {
		ID          string `json:"id"`
		IMDbID      string `json:"imdb_id"`
		Name        string `json:"name"`
		ReleaseInfo string `json:"releaseInfo"`
	} `json:"metas"`
}

// Search returns the best match for query.
func (c *Client) Search(ctx context.Context, kind media.Kind, query string) (media.Title, error) {
	endpoint := fmt.Sprintf("%s/catalog/%s/top/search=%s.json",
		c.BaseURL, kind, url.PathEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return media.Title{}, err
	}

	var res searchResponse
	if err := httpx.Do(c.HTTP, req, &res); err != nil {
		return media.Title{}, fmt.Errorf("search catalog: %w", err)
	}

	for _, m := range res.Metas {
		id := m.IMDbID
		if id == "" {
			id = m.ID
		}
		if id == "" {
			continue
		}
		return media.Title{IMDbID: id, Name: m.Name, Year: m.ReleaseInfo, Kind: kind}, nil
	}
	return media.Title{}, fmt.Errorf("%w: %q", ErrNotFound, query)
}

type metaResponse struct {
	Meta struct {
		Videos []struct {
			Season   int       `json:"season"`
			Episode  int       `json:"episode"`
			Released time.Time `json:"released"`
		} `json:"videos"`
	} `json:"meta"`
}

// Episodes returns the aired episodes of a series in order, leaving out
// specials (season 0).
func (c *Client) Episodes(ctx context.Context, imdbID string) ([]media.Episode, error) {
	endpoint := fmt.Sprintf("%s/meta/series/%s.json", c.BaseURL, url.PathEscape(imdbID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var res metaResponse
	if err := httpx.Do(c.HTTP, req, &res); err != nil {
		return nil, fmt.Errorf("fetch episodes: %w", err)
	}

	now := time.Now()
	var eps []media.Episode
	for _, v := range res.Meta.Videos {
		if v.Season < 1 || v.Episode < 1 || v.Released.IsZero() || v.Released.After(now) {
			continue
		}
		eps = append(eps, media.Episode{Season: v.Season, Number: v.Episode})
	}
	slices.SortFunc(eps, media.Episode.Compare)
	return slices.Compact(eps), nil
}
