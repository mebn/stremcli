package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/mebn/stremcli/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	app := cli.New(os.Stdin, os.Stdout, os.Stderr)
	if err := app.Run(ctx, os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		if errors.Is(err, context.Canceled) {
			os.Exit(130) // interrupted with Ctrl-C
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
