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
	var tenantName string
	var fullName string
	var phone string
	var operator string
	var reason string
	flag.StringVar(&tenantID, "tenant-id", "", "stable school tenant identifier")
	flag.StringVar(&tenantName, "tenant-name", "", "school display name")
	flag.StringVar(&fullName, "full-name", "", "Owner full name")
	flag.StringVar(&phone, "phone", "", "Owner phone in E.164 format")
	flag.StringVar(&operator, "operator", "", "database operator identity recorded in audit")
	flag.StringVar(&reason, "reason", "", "operational reason recorded in audit")
	flag.Parse()

	if err := run(tenantID, tenantName, fullName, phone, operator, reason); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "bootstrap Owner:", err)
		os.Exit(1)
	}
}

func run(tenantID, tenantName, fullName, phone, operator, reason string) error {
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
		AccessTTL:         configuration.AccessTTL,
		RefreshTTL:        configuration.RefreshTTL,
		InvitationTTL:     configuration.InvitationTTL,
	})
	result, err := service.BootstrapOwnerWithAccount(ctx, app.BootstrapOwnerInput{
		TenantID: tenantID, TenantName: tenantName, FullName: fullName, Phone: phone,
		Operator: operator, Reason: reason,
	})
	if err != nil {
		return err
	}

	// The link is a credential. It is intentionally printed once to stdout and
	// never logged or stored in recoverable form by the application database.
	_, err = fmt.Fprintf(os.Stdout, "Owner account ID: %s\nOwner activation link (expires %s):\n%s\n",
		result.AccountID, result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"), result.ActivationLink)
	return err
}
