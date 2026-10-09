package player

import (
	"slices"
	"testing"
)

func TestIPCArgs(t *testing.T) {
	p := Player{Binary: "/bin/echo", ipcArgs: players["mpv"].ipcArgs}

	cmd, err := p.command("http://x", Options{Socket: "/tmp/s.sock", Start: 61.4})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/bin/echo", "--input-ipc-server=/tmp/s.sock", "--resume-playback=no", "--start=61", "http://x"}
	if !slices.Equal(cmd.Args, want) {
		t.Errorf("args = %q, want %q", cmd.Args, want)
	}

	cmd, err = p.command("http://x", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"/bin/echo", "http://x"}; !slices.Equal(cmd.Args, want) {
		t.Errorf("without options: args = %q, want %q", cmd.Args, want)
	}
}
