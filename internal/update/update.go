// Package update replaces the running stremcli binary with the latest
// GitHub release.
package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/mebn/stremcli/internal/httpx"
)

const repo = "mebn/stremcli"

// Version is set at build time by the release workflow with
// -ldflags "-X github.com/mebn/stremcli/internal/update.Version=vX.Y.Z".
var Version = ""

// Current returns the running version: the one stamped by the release
// build, or the module version from `go install`, or "dev".
func Current() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && strings.HasPrefix(info.Main.Version, "v") {
		return info.Main.Version
	}
	return "dev"
}

// Latest returns the tag of the newest release.
func Latest(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.github.com/repos/"+repo+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := httpx.Do(client, req, &rel); err != nil {
		return "", fmt.Errorf("check latest release: %w", err)
	}
	return rel.TagName, nil
}

// Newer reports whether version a is newer than b. Versions that aren't
// vX.Y.Z (like "dev") are older than any release.
func Newer(a, b string) bool {
	pa, okA := parse(a)
	pb, okB := parse(b)
	switch {
	case !okA:
		return false
	case !okB:
		return true
	}
	for i := range pa {
		if pa[i] != pb[i] {
			return pa[i] > pb[i]
		}
	}
	return false
}

func parse(v string) ([3]int, bool) {
	var out [3]int
	// Drop pre-release/build suffixes like "-rc1" or "+dirty".
	v, _, _ = strings.Cut(strings.TrimPrefix(v, "v"), "-")
	v, _, _ = strings.Cut(v, "+")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

// Install downloads the release binary for tag and atomically replaces the
// running executable with it. It returns the path it replaced.
func Install(ctx context.Context, client *http.Client, tag string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate stremcli: %w", err)
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return "", fmt.Errorf("locate stremcli: %w", err)
	}

	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/stremcli-%s-%s",
		repo, tag, runtime.GOOS, runtime.GOARCH)
	// Download next to the executable so the final rename stays on one
	// filesystem and is atomic.
	tmp, err := os.CreateTemp(filepath.Dir(exe), ".stremcli-update-*")
	if err != nil {
		return "", fmt.Errorf("can't write to %s: %w", filepath.Dir(exe), err)
	}
	defer os.Remove(tmp.Name())

	if err := download(ctx, client, url, tmp); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Chmod(tmp.Name(), 0o755); err != nil {
		return "", err
	}
	if err := os.Rename(tmp.Name(), exe); err != nil {
		return "", fmt.Errorf("replace %s: %w", exe, err)
	}
	return exe, nil
}

// SkillDirs returns copies of the stremcli agent skill installed by copying
// SKILL.md (Codex, or Claude Code before the plugin). Plugin installs are
// updated by Claude Code itself.
func SkillDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	var dirs []string
	for _, d := range []string{".claude/skills/stremcli", ".agents/skills/stremcli"} {
		dir := filepath.Join(home, d)
		if _, err := os.Stat(filepath.Join(dir, "SKILL.md")); err == nil {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

// InstallSkill refreshes SKILL.md in dir from the given tag.
func InstallSkill(ctx context.Context, client *http.Client, tag, dir string) error {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/skills/stremcli/SKILL.md", repo, tag)
	tmp, err := os.CreateTemp(dir, ".update-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	err = download(ctx, client, url, tmp)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), filepath.Join(dir, "SKILL.md")); err != nil {
		return err
	}
	// Older releases installed a helper script the skill no longer uses.
	return os.RemoveAll(filepath.Join(dir, "scripts"))
}

func download(ctx context.Context, client *http.Client, url string, w io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", url, resp.Status)
	}
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	return nil
}
