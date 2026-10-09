package player

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"
)

// Progress is how far playback got.
type Progress struct {
	Position float64 // seconds
	Duration float64 // seconds; 0 if unknown
	// Ended is true if playback reached the end of the file.
	Ended bool
}

// socketWait is how long Track waits for the player to open its socket.
const socketWait = time.Minute

// Track follows playback through the mpv JSON IPC socket the player was
// started with (see Options.Socket). It calls onUpdate as the position
// changes and returns the final progress once the file stops playing or the
// player quits.
func Track(ctx context.Context, socket string, onUpdate func(Progress)) (Progress, error) {
	var p Progress
	conn, err := dialWait(ctx, socket)
	if err != nil {
		return p, err
	}
	defer conn.Close()
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	for _, prop := range []string{"time-pos", "duration"} {
		cmd := fmt.Sprintf(`{"command":["observe_property",0,%q]}`+"\n", prop)
		if _, err := conn.Write([]byte(cmd)); err != nil {
			return p, fmt.Errorf("player ipc: %w", err)
		}
	}

	var loaded bool
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		var msg struct {
			Event  string   `json:"event"`
			Name   string   `json:"name"`
			Data   *float64 `json:"data"`
			Reason string   `json:"reason"`
		}
		if json.Unmarshal(sc.Bytes(), &msg) != nil {
			continue
		}
		switch msg.Event {
		case "file-loaded":
			loaded = true
		case "property-change":
			if msg.Data == nil {
				continue // unset while loading or after stopping
			}
			if msg.Name == "time-pos" {
				p.Position = *msg.Data
			} else {
				p.Duration = *msg.Data
			}
			onUpdate(p)
		case "end-file":
			// The player may load something else in the same window
			// afterwards; that's not ours to track.
			if loaded {
				p.Ended = msg.Reason == "eof"
				return p, nil
			}
		}
	}
	// The player quit and closed the socket.
	if err := ctx.Err(); err != nil {
		return p, err
	}
	return p, nil
}

func dialWait(ctx context.Context, socket string) (net.Conn, error) {
	deadline := time.Now().Add(socketWait)
	var d net.Dialer
	for {
		conn, err := d.DialContext(ctx, "unix", socket)
		if err == nil {
			return conn, nil
		}
		if ctx.Err() != nil || time.Now().After(deadline) {
			return nil, errors.Join(errors.New("player ipc socket never came up"), err)
		}
		select {
		case <-ctx.Done():
		case <-time.After(500 * time.Millisecond):
		}
	}
}
