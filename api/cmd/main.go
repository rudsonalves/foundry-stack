package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	err := run(ctx, os.Args[1:])
	stop()

	if err != nil {
		log.Printf("application terminated: %v", err)
		os.Exit(1)
	}
}
