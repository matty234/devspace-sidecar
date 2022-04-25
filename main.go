package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/matty234/dev-space-configure/cmd"
)

func handleSignals(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals
		cancel()
	}()
	return ctx

}

func main() {

	ctx := context.Background()
	ctx = handleSignals(ctx)
	cmd.Execute(ctx)
}
