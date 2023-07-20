package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/andrebq/inspector/cmd/inspector/inspector"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := inspector.App(os.Stdout).Run(ctx, os.Args); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
