// Command roost turns a machine you own into an always-on dev box you can
// reach from your phone and laptop. See https://github.com/CemOlmez/roost.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cemolmez/roost/internal/cli"
	"github.com/cemolmez/roost/internal/host"
)

// Set by the release build (-ldflags "-X main.version=...").
var (
	version = "dev"
	commit  = "none"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	app := &cli.App{
		Host:    host.Local{},
		In:      os.Stdin,
		Out:     os.Stdout,
		Err:     os.Stderr,
		Version: version,
		Commit:  commit,
	}
	code := cli.Execute(ctx, app, os.Args[1:])
	stop()
	os.Exit(code)
}
