package cli

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mebn/stremcli/internal/update"
)

func (a *App) runUpdate(ctx context.Context, args []string) error {
	var check bool
	fs := a.newFlagSet("update")
	fs.BoolVar(&check, "check", false, "only report whether an update is available")
	if err := fs.Parse(args); err != nil {
		if isHelp(err) {
			return nil
		}
		return err
	}

	current := update.Current()
	latest, err := update.Latest(ctx, a.HTTP)
	if err != nil {
		return err
	}
	if !update.Newer(latest, current) {
		fmt.Fprintf(a.Err, "stremcli %s is up to date\n", current)
		return nil
	}
	if check {
		fmt.Fprintf(a.Err, "Update available: %s -> %s (run: stremcli update)\n", current, latest)
		return nil
	}

	fmt.Fprintf(a.Err, "Updating stremcli %s -> %s\n", current, latest)
	// The binary is several MB; allow more time than API calls get.
	dl := &http.Client{Timeout: 5 * time.Minute}
	path, err := update.Install(ctx, dl, latest)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Err, "Installed %s\n", path)

	for _, dir := range update.SkillDirs() {
		if err := update.InstallSkill(ctx, dl, latest, dir); err != nil {
			fmt.Fprintf(a.Err, "warning: could not update skill in %s: %v\n", dir, err)
			continue
		}
		fmt.Fprintf(a.Err, "Updated skill in %s\n", dir)
	}
	return nil
}
