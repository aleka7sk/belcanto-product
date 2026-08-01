package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/postgres"
	"github.com/aleka7sk/belcanto-product/apps/api/migrations"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	if len(arguments) != 1 || arguments[0] != "up" {
		return fmt.Errorf("only the non-destructive command `migrate up` is supported")
	}
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	store, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer store.Close()
	if err := migrations.Up(ctx, store.Pool()); err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, "migrations applied")
	return err
}
