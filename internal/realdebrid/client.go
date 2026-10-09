// Package realdebrid is a minimal client for the Real-Debrid REST API.
package realdebrid

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mebn/stremcli/internal/httpx"
)

const defaultBaseURL = "https://api.real-debrid.com/rest/1.0"

// Torrent statuses reported by Real-Debrid.
const (
	StatusMagnetConversion = "magnet_conversion"
	StatusWaitingSelection = "waiting_files_selection"
	StatusDownloaded       = "downloaded"
)

// ErrUnauthorized is returned when the API token is invalid or expired.
var ErrUnauthorized = errors.New("real-debrid rejected the API token")

// Client talks to the Real-Debrid API using a private API token.
type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
}

// New returns a Client authenticated with token.
func New(httpClient *http.Client, token string) *Client {
	return &Client{BaseURL: defaultBaseURL, Token: token, HTTP: httpClient}
}

// File is a file inside a torrent.
type File struct {
	ID       int    `json:"id"`
	Path     string `json:"path"`
	Bytes    int64  `json:"bytes"`
	Selected int    `json:"selected"`
}

// TorrentInfo describes a torrent in the user's Real-Debrid account.
type TorrentInfo struct {
	ID       string   `json:"id"`
	Filename string   `json:"filename"`
	Status   string   `json:"status"`
	Progress float64  `json:"progress"`
	Files    []File   `json:"files"`
	Links    []string `json:"links"`
}

// User describes the account that owns the API token.
type User struct {
	Username   string `json:"username"`
	Type       string `json:"type"` // "premium" or "free"
	Expiration string `json:"expiration"`
}

// User returns the account the token belongs to. It doubles as a cheap way
// to check that a token is valid.
func (c *Client) User(ctx context.Context) (User, error) {
	var u User
	err := c.do(ctx, http.MethodGet, "/user", nil, &u)
	return u, err
}

// AddMagnet adds a magnet link and returns the new torrent ID.
func (c *Client) AddMagnet(ctx context.Context, magnet string) (string, error) {
	var res struct {
		ID string `json:"id"`
	}
	err := c.post(ctx, "/torrents/addMagnet", url.Values{"magnet": {magnet}}, &res)
	return res.ID, err
}

// TorrentInfo returns the current state of a torrent.
func (c *Client) TorrentInfo(ctx context.Context, id string) (TorrentInfo, error) {
	var info TorrentInfo
	err := c.do(ctx, http.MethodGet, "/torrents/info/"+url.PathEscape(id), nil, &info)
	return info, err
}

// SelectFiles chooses which files of a torrent to download.
func (c *Client) SelectFiles(ctx context.Context, id string, fileIDs ...string) error {
	form := url.Values{"files": {strings.Join(fileIDs, ",")}}
	return c.post(ctx, "/torrents/selectFiles/"+url.PathEscape(id), form, nil)
}

// DeleteTorrent removes a torrent from the account.
func (c *Client) DeleteTorrent(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/torrents/delete/"+url.PathEscape(id), nil, nil)
}

// Unrestrict converts a hoster link into a direct download URL.
func (c *Client) Unrestrict(ctx context.Context, link string) (string, error) {
	var res struct {
		Download string `json:"download"`
	}
	err := c.post(ctx, "/unrestrict/link", url.Values{"link": {link}}, &res)
	return res.Download, err
}

func (c *Client) post(ctx context.Context, path string, form url.Values, out any) error {
	return c.do(ctx, http.MethodPost, path, strings.NewReader(form.Encode()), out)
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	err = httpx.Do(c.HTTP, req, out)

	var se *httpx.StatusError
	if errors.As(err, &se) && (se.StatusCode == http.StatusUnauthorized || se.StatusCode == http.StatusForbidden) {
		return fmt.Errorf("%w: %s", ErrUnauthorized, se.Body)
	}
	return err
}
