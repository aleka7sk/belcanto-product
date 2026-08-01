package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("configure PostgreSQL pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	return New(pool), nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *Store) Ready(ctx context.Context) error {
	if err := s.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}
	return nil
}

type idempotencyClaim struct {
	replayed bool
	response []byte
}

func claimIdempotency(
	ctx context.Context,
	tx pgx.Tx,
	tenantID string,
	actorAccountID string,
	scope string,
	key string,
	fingerprint []byte,
	now time.Time,
) (idempotencyClaim, error) {
	tag, err := tx.Exec(ctx, `
		INSERT INTO idempotency_records (
			tenant_id, actor_account_id, operation_scope, idempotency_key, payload_fingerprint,
			status, created_at
		) VALUES ($1, $2, $3, $4, $5, 'processing', $6)
		ON CONFLICT (tenant_id, actor_account_id, operation_scope, idempotency_key) DO NOTHING
	`, tenantID, actorAccountID, scope, key, fingerprint, now)
	if err != nil {
		return idempotencyClaim{}, fmt.Errorf("claim idempotency key: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return idempotencyClaim{}, nil
	}

	var storedFingerprint []byte
	var status string
	var responseText *string
	err = tx.QueryRow(ctx, `
		SELECT payload_fingerprint, status, response_json::text
		FROM idempotency_records
		WHERE tenant_id = $1 AND actor_account_id = $2
		  AND operation_scope = $3 AND idempotency_key = $4
		FOR UPDATE
	`, tenantID, actorAccountID, scope, key).Scan(&storedFingerprint, &status, &responseText)
	if err != nil {
		return idempotencyClaim{}, fmt.Errorf("read idempotency record: %w", err)
	}
	if !security.EqualDigest(storedFingerprint, fingerprint) {
		return idempotencyClaim{}, core.E(core.CodeConflict, "Idempotency-Key was reused with a different payload", nil)
	}
	if status != "completed" || responseText == nil {
		return idempotencyClaim{}, core.E(core.CodeConflict, "the idempotent operation is still processing", nil)
	}
	return idempotencyClaim{replayed: true, response: []byte(*responseText)}, nil
}

func completeIdempotency(
	ctx context.Context,
	tx pgx.Tx,
	tenantID string,
	actorAccountID string,
	scope string,
	key string,
	response any,
	now time.Time,
) error {
	encoded, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("encode idempotency response: %w", err)
	}
	tag, err := tx.Exec(ctx, `
		UPDATE idempotency_records
		SET status = 'completed', response_json = $5::jsonb, completed_at = $6
		WHERE tenant_id = $1 AND actor_account_id = $2
		  AND operation_scope = $3 AND idempotency_key = $4
		  AND status = 'processing'
	`, tenantID, actorAccountID, scope, key, encoded, now)
	if err != nil {
		return fmt.Errorf("complete idempotency record: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("complete idempotency record: record is not processing")
	}
	return nil
}

func decodeReplay[T any](claim idempotencyClaim) (T, error) {
	var value T
	if !claim.replayed {
		return value, fmt.Errorf("idempotency claim is not a replay")
	}
	if err := json.Unmarshal(claim.response, &value); err != nil {
		return value, fmt.Errorf("decode idempotency response: %w", err)
	}
	return value, nil
}

func hasActiveRole(ctx context.Context, tx pgx.Tx, tenantID, accountID string, role core.Role) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM accounts a
			JOIN role_grants rg
			  ON rg.tenant_id = a.tenant_id AND rg.account_id = a.id
			WHERE a.tenant_id = $1 AND a.id = $2 AND a.status = 'active'
			  AND rg.role_type = $3 AND rg.status = 'active'
		)
	`, tenantID, accountID, string(role)).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check active role: %w", err)
	}
	return exists, nil
}

func onboardingAuthority(ctx context.Context, tx pgx.Tx, tenantID, accountID string, now time.Time) (string, bool, error) {
	isOwner, err := hasActiveRole(ctx, tx, tenantID, accountID, core.RoleOwner)
	if err != nil {
		return "", false, err
	}
	if isOwner {
		return "", true, nil
	}
	isAdministrator, err := hasActiveRole(ctx, tx, tenantID, accountID, core.RoleAdministrator)
	if err != nil {
		return "", false, err
	}
	if !isAdministrator {
		return "", false, nil
	}
	var delegationID string
	err = tx.QueryRow(ctx, `
		SELECT id
		FROM capability_delegations
		WHERE tenant_id = $1 AND grantee_account_id = $2
		  AND bundle = $3 AND status = 'active'
		  AND (expires_at IS NULL OR expires_at > $4)
		ORDER BY granted_at DESC, id DESC
		LIMIT 1
		FOR SHARE
	`, tenantID, accountID, core.StudentOnboardingManagerV1, now).Scan(&delegationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("check onboarding delegation: %w", err)
	}
	return delegationID, true, nil
}

func lockActivationSubject(ctx context.Context, tx pgx.Tx, tenantID, accountID, studentID string) error {
	subject := "account:" + accountID
	if studentID != "" {
		subject = "student:" + studentID
	}
	key := "belcanto:activation:v1:" + tenantID + "\x00" + subject
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, key); err != nil {
		return fmt.Errorf("lock activation subject: %w", err)
	}
	return nil
}

type auditInput struct {
	tenantID       string
	actorID        string
	operatorID     string
	delegationID   string
	action         string
	targetType     string
	targetID       string
	decision       string
	reason         string
	idempotencyKey string
	metadata       any
	at             time.Time
}

func appendAudit(ctx context.Context, tx pgx.Tx, input auditInput) error {
	metadata := input.metadata
	if metadata == nil {
		metadata = struct{}{}
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode audit metadata: %w", err)
	}
	var idempotencyHash string
	if input.idempotencyKey != "" {
		digest := sha256.Sum256([]byte(input.idempotencyKey))
		idempotencyHash = hex.EncodeToString(digest[:])
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_records (
			tenant_id, actor_account_id, operator_identifier, delegation_id, action, target_type,
			target_id, decision, reason_code, idempotency_key_hash, metadata, recorded_at
		) VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''), $5, $6, $7, $8,
		          NULLIF($9, ''), NULLIF($10, ''), $11::jsonb, $12)
	`, input.tenantID, input.actorID, input.operatorID, input.delegationID, input.action, input.targetType,
		input.targetID, input.decision, input.reason, idempotencyHash, encoded, input.at)
	if err != nil {
		return fmt.Errorf("append audit record: %w", err)
	}
	return nil
}

func appendOutbox(ctx context.Context, tx pgx.Tx, tenantID, eventType, aggregateType, aggregateID string, payload any, at time.Time) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode outbox payload: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (
			tenant_id, event_type, aggregate_type, aggregate_id, payload, recorded_at
		) VALUES ($1, $2, $3, $4, $5::jsonb, $6)
	`, tenantID, eventType, aggregateType, aggregateID, encoded, at)
	if err != nil {
		return fmt.Errorf("append outbox event: %w", err)
	}
	return nil
}

func (s *Store) recordDenied(ctx context.Context, input auditInput) {
	input.decision = "deny"
	_ = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		return appendAudit(ctx, tx, input)
	})
}

func mapWriteError(err error, conflictMessage string) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return core.E(core.CodeConflict, conflictMessage, err)
	}
	return err
}

func rolesForAccount(ctx context.Context, runner interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, tenantID, accountID string) ([]core.Role, error) {
	rows, err := runner.Query(ctx, `
		SELECT DISTINCT rg.role_type
		FROM role_grants rg
		JOIN accounts a ON a.tenant_id = rg.tenant_id AND a.id = rg.account_id
		WHERE rg.tenant_id = $1 AND rg.account_id = $2
		  AND rg.status = 'active' AND a.status = 'active'
		ORDER BY rg.role_type
	`, tenantID, accountID)
	if err != nil {
		return nil, fmt.Errorf("read account roles: %w", err)
	}
	defer rows.Close()

	roles := make([]core.Role, 0, 2)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan account role: %w", err)
		}
		roles = append(roles, core.Role(role))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate account roles: %w", err)
	}
	return roles, nil
}
