package memory

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 media upload lifecycle — parity with PostgreSQL. Bytes live in
// process; metadata mirrors the resumable pending → uploading → ready
// machine and no deletion path exists (DEC-104 open).

type mediaObject struct {
	ID             string
	TenantID       string
	OwnerAccountID string
	Kind           string
	ContentType    string
	ByteSize       int64
	UploadedBytes  int64
	Status         string
	Bytes          []byte
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func mediaView(stored *mediaObject) core.MediaObject {
	return core.MediaObject{
		ID: stored.ID, Kind: stored.Kind, ContentType: stored.ContentType,
		ByteSize: stored.ByteSize, UploadedBytes: stored.UploadedBytes,
		Status: stored.Status, CreatedAt: stored.CreatedAt, UpdatedAt: stored.UpdatedAt,
	}
}

func (s *Store) CreateMediaObject(_ context.Context, command core.CreateMediaCommand) (core.MediaObject, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if s.activeAccount(principal.AccountID, principal.TenantID) == nil {
		return core.MediaObject{}, core.E(core.CodeForbidden, "an active account is required", nil)
	}
	if response, ok, err := s.replay("create_media", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.MediaObject{}, err
		}
		var result core.MediaObject
		if err := json.Unmarshal(response, &result); err != nil {
			return core.MediaObject{}, core.E(core.CodeInternal, "decode idempotent media result", err)
		}
		return result, nil
	}
	if _, exists := s.mediaObjects[command.MediaID]; exists {
		return core.MediaObject{}, core.E(core.CodeConflict, "media object conflicts with existing data", nil)
	}
	stored := &mediaObject{
		ID: command.MediaID, TenantID: principal.TenantID, OwnerAccountID: principal.AccountID,
		Kind: command.Kind, ContentType: command.ContentType, ByteSize: command.ByteSize,
		Status: core.MediaStatusPending, CreatedAt: command.Now, UpdatedAt: command.Now,
	}
	s.mediaObjects[command.MediaID] = stored
	result := mediaView(stored)
	if err := s.completeIdempotency("create_media", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.MediaObject{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "MediaDeclared",
		"media_object", stored.ID, "allow", "", command.Now, nil)
	return result, nil
}

func (s *Store) AppendMediaChunk(_ context.Context, command core.AppendMediaChunkCommand) (core.MediaObject, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	stored := s.mediaObjects[command.MediaID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.MediaObject{}, core.E(core.CodeNotFound, "media not found", nil)
	}
	if stored.OwnerAccountID != principal.AccountID {
		return core.MediaObject{}, core.E(core.CodeForbidden, "only the owner uploads media content", nil)
	}
	end := command.Offset + int64(len(command.Data))
	if stored.Status == core.MediaStatusReady {
		if end <= stored.ByteSize {
			return mediaView(stored), nil
		}
		return core.MediaObject{}, core.E(core.CodeInvalidState, "the upload is already complete", nil)
	}
	if stored.Status == core.MediaStatusFailed {
		return core.MediaObject{}, core.E(core.CodeInvalidState, "the upload is marked failed; declare a new media object", nil)
	}
	if len(command.Data) == 0 {
		return core.MediaObject{}, core.E(core.CodeInvalidInput, "an upload chunk must not be empty", nil)
	}
	if end > stored.ByteSize {
		return core.MediaObject{}, core.E(core.CodeInvalidInput, "the chunk exceeds the declared media size", nil)
	}
	if end <= stored.UploadedBytes {
		return mediaView(stored), nil
	}
	if command.Offset != stored.UploadedBytes {
		return core.MediaObject{}, core.E(core.CodeConflict, "the upload resumes at the recorded offset", nil)
	}
	if int64(len(stored.Bytes)) < end {
		grown := make([]byte, end)
		copy(grown, stored.Bytes)
		stored.Bytes = grown
	}
	copy(stored.Bytes[command.Offset:], command.Data)
	stored.UploadedBytes = end
	stored.Status = core.MediaStatusUploading
	if end == stored.ByteSize {
		stored.Status = core.MediaStatusReady
		s.appendSecurityAudit(principal.TenantID, principal.AccountID, "MediaUploaded",
			"media_object", stored.ID, "allow", "", command.Now, nil)
	}
	stored.UpdatedAt = command.Now
	return mediaView(stored), nil
}

func (s *Store) mediaVisible(principal core.Principal, stored *mediaObject) bool {
	if stored.OwnerAccountID == principal.AccountID {
		return true
	}
	if actor := s.activeAccount(principal.AccountID, principal.TenantID); actor != nil &&
		(actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "") {
		return true
	}
	studentID := s.studentIDForAccount(principal.AccountID)
	for _, record := range s.homework {
		if record.TenantID != principal.TenantID {
			continue
		}
		participant := record.TeacherAccountID == principal.AccountID ||
			(studentID != "" && record.StudentID == studentID)
		if !participant {
			continue
		}
		for _, mediaID := range record.AttachmentMediaIDs {
			if mediaID == stored.ID {
				return true
			}
		}
		for _, submission := range record.Submissions {
			for _, mediaID := range submission.MediaIDs {
				if mediaID == stored.ID {
					return true
				}
			}
		}
	}
	return false
}

func (s *Store) GetMediaObject(_ context.Context, principal core.Principal, mediaID string) (core.MediaObject, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.mediaObjects[mediaID]
	if stored == nil || stored.TenantID != principal.TenantID || !s.mediaVisible(principal, stored) {
		return core.MediaObject{}, core.E(core.CodeNotFound, "media not found", nil)
	}
	return mediaView(stored), nil
}

func (s *Store) MediaContent(_ context.Context, tenantID, mediaID string) ([]byte, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.mediaObjects[mediaID]
	if stored == nil || stored.TenantID != tenantID {
		return nil, "", core.E(core.CodeNotFound, "media not found", nil)
	}
	if stored.Status != core.MediaStatusReady {
		return nil, "", core.E(core.CodeInvalidState, "the media upload is not complete", nil)
	}
	content := make([]byte, len(stored.Bytes))
	copy(content, stored.Bytes)
	return content, stored.ContentType, nil
}
