package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/config"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/postgres"
	"github.com/aleka7sk/belcanto-product/apps/api/migrations"
)

func main() {
	var tenantID string
	var accountID string
	var operator string
	var reason string
	flag.StringVar(&tenantID, "tenant-id", "", "existing school tenant identifier")
	flag.StringVar(&accountID, "account-id", "", "existing pending Owner, Administrator, or Teacher account")
	flag.StringVar(&operator, "operator", "", "database operator identity recorded in audit")
	flag.StringVar(&reason, "reason", "", "operational recovery reason recorded in audit")
	flag.Parse()

	if err := run(tenantID, accountID, operator, reason); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "reissue bootstrap invitation:", err)
		os.Exit(1)
	}
}

func run(tenantID, accountID, operator, reason string) error {
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
	if err := migrations.Up(ctx, store.Pool()); err != nil {
		return err
	}
	tokens, err := security.NewTokenCodec(configuration.TokenMasterKey)
	if err != nil {
		return err
	}
	service := app.NewService(store, tokens, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: configuration.PublicActivationBaseURL,
		AccessTTL:         configuration.AccessTTL, RefreshTTL: configuration.RefreshTTL,
		InvitationTTL: configuration.InvitationTTL,
	})
	result, activationLink, err := service.ReissueBootstrapInvitation(ctx, app.ReissueBootstrapInvitationInput{
		TenantID: tenantID, AccountID: accountID, Operator: operator, Reason: reason,
	})
	if err != nil {
		return err
	}
	// The recovered link is a credential. It is printed once and never stored
	// as raw or reversibly encrypted token material.
	_, err = fmt.Fprintf(os.Stdout, "%s activation link for %s (expires %s):\n%s\n",
		result.Kind, result.AccountID, result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"), activationLink)
	return err
}
