package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"river-demo/internal/demo"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if _, err := demo.Run(ctx, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "river demo failed:", err)
		os.Exit(1)
	}
}
