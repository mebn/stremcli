// Package catalog resolves free-text titles to IMDb IDs using Stremio's
// Cinemeta addon.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

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
