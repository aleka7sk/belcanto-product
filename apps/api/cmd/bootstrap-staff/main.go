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
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/postgres"
	"github.com/aleka7sk/belcanto-product/apps/api/migrations"
)

func main() {
	var tenantID string
	var ownerAccountID string
	var role string
	var fullName string
	var phone string
	var operator string
	var reason string
	flag.StringVar(&tenantID, "tenant-id", "", "existing school tenant identifier")
	flag.StringVar(&ownerAccountID, "owner-account-id", "", "existing active Owner account identifier")
	flag.StringVar(&role, "role", "", "staff role: Administrator or Teacher")
	flag.StringVar(&fullName, "full-name", "", "staff member full name")
	flag.StringVar(&phone, "phone", "", "staff member phone in E.164 format")
	flag.StringVar(&operator, "operator", "", "database operator identity recorded in audit")
	flag.StringVar(&reason, "reason", "", "operational reason recorded in audit")
	flag.Parse()

	if err := run(tenantID, ownerAccountID, core.Role(role), fullName, phone, operator, reason); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "bootstrap staff:", err)
		os.Exit(1)
	}
}

func run(tenantID, ownerAccountID string, role core.Role, fullName, phone, operator, reason string) error {
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
	result, err := service.BootstrapStaffWithAccount(ctx, app.BootstrapStaffInput{
		TenantID: tenantID, OwnerAccountID: ownerAccountID,
		FullName: fullName, Phone: phone, Role: role, Operator: operator, Reason: reason,
	})
	if err != nil {
		return err
	}
	// The link is a credential. It is intentionally printed once and is never
	// persisted as raw or reversibly encrypted token material.
	_, err = fmt.Fprintf(os.Stdout, "%s account ID: %s\n%s activation link (expires %s):\n%s\n",
		role, result.AccountID, role, result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"), result.ActivationLink)
	return err
}
