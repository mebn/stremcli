// Package httpx contains small helpers for JSON-over-HTTP APIs.
package httpx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const userAgent = "stremcli/1.0"

// StatusError is returned for non-2xx responses.
type StatusError struct {
	Method     string
	URL        string
	StatusCode int
	Body       string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("%s %s: %d %s: %s",
		e.Method, e.URL, e.StatusCode, http.StatusText(e.StatusCode), e.Body)
}

// Do sends req and decodes a JSON response body into out (if out is non-nil).
// Non-2xx responses are returned as errors that include the response body.
func Do(client *http.Client, req *http.Request, out any) error {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return &StatusError{
			Method:     req.Method,
			URL:        req.URL.Redacted(),
			StatusCode: resp.StatusCode,
			Body:       strings.TrimSpace(string(body)),
		}
	}

	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode %s: %w", req.URL.Redacted(), err)
	}
	return nil
}
