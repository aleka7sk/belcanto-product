// dev-contact-code plays the role of the absent delivery worker in
// local development: contact confirmation codes are derived from the
// verification identifier with the master key and never stored in
// plaintext (see security.TokenCodec.ContactVerificationCode), so this
// tool re-derives the code for open verifications and prints it to the
// operator instead of sending an email or SMS. Development only — it
// refuses to run with APP_ENV=production.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/config"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

func main() {
	var verificationID string
	flag.StringVar(&verificationID, "verification-id", "",
		"print the code for one verification; default lists every open verification")
	flag.Parse()

	if err := run(verificationID); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "dev-contact-code:", err)
		os.Exit(1)
	}
}

func run(verificationID string) error {
	configuration, err := config.Load()
	if err != nil {
		return err
	}
	if configuration.Environment == "production" {
		return fmt.Errorf("this tool is for development; codes in production are delivered, not printed")
	}
	codec, err := security.NewTokenCodec(configuration.TokenMasterKey)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := pgxpool.New(ctx, configuration.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	if verificationID != "" {
		fmt.Println(codec.ContactVerificationCode(verificationID))
		return nil
	}

	rows, err := pool.Query(ctx, `
		SELECT id, contact_kind, normalized_value, purpose, expires_at, attempts_remaining
		FROM contact_verifications
		WHERE consumed_at IS NULL AND superseded_at IS NULL AND expires_at > now()
		ORDER BY created_at DESC
	`)
	if err != nil {
		return fmt.Errorf("list open verifications: %w", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		var id, kind, value, purpose string
		var expiresAt time.Time
		var attempts int
		if err := rows.Scan(&id, &kind, &value, &purpose, &expiresAt, &attempts); err != nil {
			return fmt.Errorf("scan verification: %w", err)
		}
		found = true
		fmt.Printf("%s %s · %s · attempts %d · expires %s\n  code: %s\n",
			kind, value, purpose, attempts,
			expiresAt.Local().Format("15:04"), codec.ContactVerificationCode(id))
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate verifications: %w", err)
	}
	if !found {
		fmt.Println("no open contact verifications — start the confirmation step in the app first")
	}
	return nil
}
