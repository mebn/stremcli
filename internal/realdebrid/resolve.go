package realdebrid

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"
)

// ErrNotCached is returned when a torrent isn't instantly available on Real-Debrid.
var ErrNotCached = errors.New("torrent is not cached on Real-Debrid")

var videoExts = map[string]bool{
	".mkv": true, ".mp4": true, ".avi": true, ".m4v": true,
	".mov": true, ".wmv": true, ".ts": true, ".webm": true,
}

// Target describes which file of a torrent should be streamed.
type Target struct {
	InfoHash string
	// Filename is the expected name of the file, if known.
	Filename string
	// FileIdx is the zero-based index of the file in the torrent, or -1.
	FileIdx int
}

// Resolver turns torrents into direct streaming links.
type Resolver struct {
	Client *Client
	// CacheWait is how long to wait for a torrent to become available
	// before treating it as uncached.
	CacheWait    time.Duration
	PollInterval time.Duration
}

// NewResolver returns a Resolver with sensible defaults.
func NewResolver(c *Client) *Resolver {
	return &Resolver{Client: c, CacheWait: 8 * time.Second, PollInterval: time.Second}
}

// Resolve adds the torrent to Real-Debrid, selects the target file and returns
// a direct download URL. Uncached torrents are removed and ErrNotCached is returned.
func (r *Resolver) Resolve(ctx context.Context, t Target) (string, error) {
	id, err := r.Client.AddMagnet(ctx, "magnet:?xt=urn:btih:"+t.InfoHash)
	if err != nil {
		return "", fmt.Errorf("add magnet: %w", err)
	}

	link, err := r.resolveTorrent(ctx, id, t)
	if err != nil {
		// Don't leave stale torrents in the user's account. Use a fresh
		// context so cleanup still runs if ctx was cancelled.
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = r.Client.DeleteTorrent(cleanupCtx, id)
		return "", err
	}
	return link, nil
}

func (r *Resolver) resolveTorrent(ctx context.Context, id string, t Target) (string, error) {
	info, err := r.waitFor(ctx, id, func(i TorrentInfo) bool {
		return i.Status != StatusMagnetConversion
	})
	if err != nil {
		return "", err
	}

	if info.Status == StatusWaitingSelection {
		file, ok := pickFile(info.Files, t)
		if !ok {
			return "", errors.New("no playable file in torrent")
		}
		if err := r.Client.SelectFiles(ctx, id, strconv.Itoa(file.ID)); err != nil {
			return "", fmt.Errorf("select files: %w", err)
		}
	}

	info, err = r.waitFor(ctx, id, func(i TorrentInfo) bool {
		return i.Status == StatusDownloaded
	})
	if err != nil {
		return "", err
	}
	if len(info.Links) == 0 {
		return "", errors.New("torrent has no links")
	}

	link, err := r.Client.Unrestrict(ctx, info.Links[0])
	if err != nil {
		return "", fmt.Errorf("unrestrict link: %w", err)
	}
	return link, nil
}

// waitFor polls a torrent until done reports true or CacheWait elapses.
func (r *Resolver) waitFor(ctx context.Context, id string, done func(TorrentInfo) bool) (TorrentInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, r.CacheWait)
	defer cancel()

	ticker := time.NewTicker(r.PollInterval)
	defer ticker.Stop()

	for {
		info, err := r.Client.TorrentInfo(ctx, id)
		if err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return info, ErrNotCached
			}
			return info, fmt.Errorf("torrent info: %w", err)
		}
		if done(info) {
			return info, nil
		}

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return info, ErrNotCached
			}
			return info, ctx.Err()
		case <-ticker.C:
		}
	}
}

// pickFile chooses the file to stream: by expected filename, then by index,
// then the largest video file.
func pickFile(files []File, t Target) (File, bool) {
	if t.Filename != "" {
		for _, f := range files {
			if strings.EqualFold(path.Base(f.Path), t.Filename) {
				return f, true
			}
		}
	}

	// Real-Debrid file IDs are 1-based and follow torrent order.
	if t.FileIdx >= 0 {
		for _, f := range files {
			if f.ID == t.FileIdx+1 && isVideo(f.Path) {
				return f, true
			}
		}
	}

	var best File
	found := false
	for _, f := range files {
		if isVideo(f.Path) && (!found || f.Bytes > best.Bytes) {
			best, found = f, true
		}
	}
	return best, found
}

func isVideo(p string) bool {
	return videoExts[strings.ToLower(path.Ext(p))]
}
