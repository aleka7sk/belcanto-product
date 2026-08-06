package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 media upload lifecycle: pending → uploading → ready, resumable by
// exact offset. Rows carry only adapter storage keys; URLs are never
// stored and never appear in audit or outbox records.

func scanMediaObject(row pgx.Row) (core.MediaObject, error) {
	var object core.MediaObject
	err := row.Scan(&object.ID, &object.Kind, &object.ContentType, &object.ByteSize,
		&object.UploadedBytes, &object.Status, &object.CreatedAt, &object.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.MediaObject{}, core.E(core.CodeNotFound, "media not found", nil)
	}
	if err != nil {
		return core.MediaObject{}, fmt.Errorf("read media object: %w", err)
	}
	object.CreatedAt = object.CreatedAt.UTC()
	object.UpdatedAt = object.UpdatedAt.UTC()
	return object, nil
}

const mediaObjectColumns = `id, kind, content_type, byte_size, uploaded_bytes, status, created_at, updated_at`

func activeAccountExists(ctx context.Context, tx pgx.Tx, tenantID, accountID string) error {
	var active bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM accounts
			WHERE tenant_id = $1 AND id = $2 AND status = 'active'
		)
	`, tenantID, accountID).Scan(&active); err != nil {
		return fmt.Errorf("check active account: %w", err)
	}
	if !active {
		return core.E(core.CodeForbidden, "an active account is required", nil)
	}
	return nil
}

func (s *Store) CreateMediaObject(ctx context.Context, command core.CreateMediaCommand) (core.MediaObject, error) {
	principal := command.Principal
	var object core.MediaObject
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := activeAccountExists(ctx, tx, principal.TenantID, principal.AccountID); err != nil {
			return err
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_media", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			object, err = decodeReplay[core.MediaObject](claim)
			return err
		}
		storageKey := principal.TenantID + "/" + command.MediaID
		if _, err := tx.Exec(ctx, `
			INSERT INTO media_objects (
				id, tenant_id, owner_account_id, kind, content_type,
				byte_size, uploaded_bytes, status, storage_key, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, 0, 'pending', $7, $8, $8)
		`, command.MediaID, principal.TenantID, principal.AccountID, command.Kind,
			command.ContentType, command.ByteSize, storageKey, command.Now); err != nil {
			return mapWriteError(err, "media object conflicts with existing data")
		}
		object, err = scanMediaObject(tx.QueryRow(ctx, `
			SELECT `+mediaObjectColumns+` FROM media_objects
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.MediaID))
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_media", command.IdempotencyKey, object, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "MediaDeclared", targetType: "media_object", targetID: command.MediaID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"kind": command.Kind, "byteSize": command.ByteSize},
			at:       command.Now,
		})
	})
	if err != nil {
		return core.MediaObject{}, err
	}
	return object, nil
}

func (s *Store) AppendMediaChunk(ctx context.Context, command core.AppendMediaChunkCommand) (core.MediaObject, error) {
	principal := command.Principal
	var object core.MediaObject
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var owner, status, storageKey string
		var byteSize, uploaded int64
		err := tx.QueryRow(ctx, `
			SELECT owner_account_id, status, storage_key, byte_size, uploaded_bytes
			FROM media_objects
			WHERE tenant_id = $1 AND id = $2
			FOR UPDATE
		`, principal.TenantID, command.MediaID).Scan(&owner, &status, &storageKey, &byteSize, &uploaded)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "media not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock media object: %w", err)
		}
		if owner != principal.AccountID {
			return core.E(core.CodeForbidden, "only the owner uploads media content", nil)
		}
		if status == core.MediaStatusReady {
			// A fully replayed tail chunk after completion is a no-op.
			if command.Offset+int64(len(command.Data)) <= byteSize {
				object, err = scanMediaObject(tx.QueryRow(ctx, `
					SELECT `+mediaObjectColumns+` FROM media_objects
					WHERE tenant_id = $1 AND id = $2
				`, principal.TenantID, command.MediaID))
				return err
			}
			return core.E(core.CodeInvalidState, "the upload is already complete", nil)
		}
		if status == core.MediaStatusFailed {
			return core.E(core.CodeInvalidState, "the upload is marked failed; declare a new media object", nil)
		}
		if len(command.Data) == 0 {
			return core.E(core.CodeInvalidInput, "an upload chunk must not be empty", nil)
		}
		end := command.Offset + int64(len(command.Data))
		if end > byteSize {
			return core.E(core.CodeInvalidInput, "the chunk exceeds the declared media size", nil)
		}
		if end <= uploaded {
			// Full replay of an already-applied chunk: return current state.
			object, err = scanMediaObject(tx.QueryRow(ctx, `
				SELECT `+mediaObjectColumns+` FROM media_objects
				WHERE tenant_id = $1 AND id = $2
			`, principal.TenantID, command.MediaID))
			return err
		}
		if command.Offset != uploaded {
			return core.E(core.CodeConflict, "the upload resumes at the recorded offset", nil)
		}
		if err := s.media.WriteAt(ctx, storageKey, command.Offset, command.Data); err != nil {
			return fmt.Errorf("write media chunk: %w", err)
		}
		nextStatus := core.MediaStatusUploading
		if end == byteSize {
			nextStatus = core.MediaStatusReady
		}
		if _, err := tx.Exec(ctx, `
			UPDATE media_objects
			SET uploaded_bytes = $3, status = $4, updated_at = $5
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.MediaID, end, nextStatus, command.Now); err != nil {
			return mapWriteError(err, "media upload conflicts with existing data")
		}
		object, err = scanMediaObject(tx.QueryRow(ctx, `
			SELECT `+mediaObjectColumns+` FROM media_objects
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.MediaID))
		if err != nil {
			return err
		}
		if nextStatus == core.MediaStatusReady {
			return appendAudit(ctx, tx, auditInput{
				tenantID: principal.TenantID, actorID: principal.AccountID,
				action: "MediaUploaded", targetType: "media_object", targetID: command.MediaID,
				decision: "allow", metadata: map[string]any{"byteSize": byteSize},
				at: command.Now,
			})
		}
		return nil
	})
	if err != nil {
		return core.MediaObject{}, err
	}
	return object, nil
}

// mediaViewerScope: the owner, managers, and homework participants
// (the assignment's teacher or student) may read media metadata.
func (s *Store) mediaViewerScope(ctx context.Context, principal core.Principal, mediaID string) (bool, error) {
	var visible bool
	if err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM media_objects m
			WHERE m.tenant_id = $1 AND m.id = $2 AND m.owner_account_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_grants
			WHERE tenant_id = $1 AND account_id = $3
			  AND role_type IN ('Owner', 'Administrator') AND status = 'active'
		) OR EXISTS (
			SELECT 1 FROM homework_attachments ha
			JOIN homework_assignments h
			  ON h.tenant_id = ha.tenant_id AND h.id = ha.homework_id
			LEFT JOIN students st
			  ON st.tenant_id = h.tenant_id AND st.id = h.student_id
			WHERE ha.tenant_id = $1 AND ha.media_id = $2
			  AND (h.teacher_account_id = $3 OR st.account_id = $3)
		) OR EXISTS (
			SELECT 1 FROM practice_submission_media sm
			JOIN practice_submissions ps
			  ON ps.tenant_id = sm.tenant_id AND ps.id = sm.submission_id
			JOIN homework_assignments h
			  ON h.tenant_id = ps.tenant_id AND h.id = ps.homework_id
			LEFT JOIN students st
			  ON st.tenant_id = h.tenant_id AND st.id = h.student_id
			WHERE sm.tenant_id = $1 AND sm.media_id = $2
			  AND (h.teacher_account_id = $3 OR st.account_id = $3)
		)
	`, principal.TenantID, mediaID, principal.AccountID).Scan(&visible); err != nil {
		return false, fmt.Errorf("check media viewer scope: %w", err)
	}
	return visible, nil
}

func (s *Store) GetMediaObject(ctx context.Context, principal core.Principal, mediaID string) (core.MediaObject, error) {
	visible, err := s.mediaViewerScope(ctx, principal, mediaID)
	if err != nil {
		return core.MediaObject{}, err
	}
	if !visible {
		return core.MediaObject{}, core.E(core.CodeNotFound, "media not found", nil)
	}
	return scanMediaObject(s.pool.QueryRow(ctx, `
		SELECT `+mediaObjectColumns+` FROM media_objects
		WHERE tenant_id = $1 AND id = $2
	`, principal.TenantID, mediaID))
}

// MediaContent serves bytes for a signature-verified access link; the
// caller has already proven the grant, so only tenancy and readiness
// are re-checked here.
func (s *Store) MediaContent(ctx context.Context, tenantID, mediaID string) ([]byte, string, error) {
	var status, storageKey, contentType string
	err := s.pool.QueryRow(ctx, `
		SELECT status, storage_key, content_type FROM media_objects
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, mediaID).Scan(&status, &storageKey, &contentType)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", core.E(core.CodeNotFound, "media not found", nil)
	}
	if err != nil {
		return nil, "", fmt.Errorf("read media for content: %w", err)
	}
	if status != core.MediaStatusReady {
		return nil, "", core.E(core.CodeInvalidState, "the media upload is not complete", nil)
	}
	content, err := s.media.Read(ctx, storageKey)
	if err != nil {
		return nil, "", fmt.Errorf("read media content: %w", err)
	}
	return content, contentType, nil
}
