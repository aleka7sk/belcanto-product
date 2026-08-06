package main

import (
	_ "time/tzdata"

	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/config"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/httpapi"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/media"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/postgres"
	"github.com/aleka7sk/belcanto-product/apps/api/migrations"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	configuration, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store, err := postgres.Open(ctx, configuration.DatabaseURL)
	if err != nil {
		return err
	}
	defer store.Close()
	mediaStorage, err := media.NewFSStorage(configuration.MediaDir)
	if err != nil {
		return err
	}
	store.UseMediaStorage(mediaStorage)
	if configuration.AutoMigrate {
		if err := migrations.Up(ctx, store.Pool()); err != nil {
			return err
		}
	}
	tokens, err := security.NewTokenCodec(configuration.TokenMasterKey)
	if err != nil {
		return err
	}
	service := app.NewService(store, tokens, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: configuration.PublicActivationBaseURL,
		AccessTTL:         configuration.AccessTTL,
		RefreshTTL:        configuration.RefreshTTL,
		InvitationTTL:     configuration.InvitationTTL,
	})
	server := &http.Server{
		Addr:              configuration.ListenAddress,
		Handler:           httpapi.New(service),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    64 << 10,
	}

	serveError := make(chan error, 1)
	go func() {
		logger.Info("api listening", "address", configuration.ListenAddress)
		serveError <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			return err
		}
		return nil
	case err := <-serveError:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
