package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// L.3 media upload lifecycle. Content is reachable only through
// short-lived sealed access links; storage keys and URLs never enter
// audit, outbox or idempotency records.

var mediaKinds = map[string]bool{
	"audio": true, "video": true, "image": true, "pdf": true,
}

type CreateMediaInput struct {
	Kind           string
	ContentType    string
	ByteSize       int64
	IdempotencyKey string
}

func (s *Service) CreateMedia(ctx context.Context, principal core.Principal, input CreateMediaInput) (core.MediaObject, error) {
	if !mediaKinds[input.Kind] {
		return core.MediaObject{}, core.E(core.CodeInvalidInput, "media kind must be audio, video, image or pdf", nil)
	}
	contentType, err := security.ValidateText("contentType", input.ContentType, 3, 120)
	if err != nil || !strings.Contains(contentType, "/") {
		return core.MediaObject{}, core.E(core.CodeInvalidInput, "contentType must be a media type like audio/m4a", nil)
	}
	if input.ByteSize <= 0 || input.ByteSize > s.maxMediaBytes {
		return core.MediaObject{}, core.E(core.CodeInvalidInput, "byteSize must be positive and within the upload limit", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.MediaObject{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	mediaID, err := security.NewID("med")
	if err != nil {
		return core.MediaObject{}, core.E(core.CodeInternal, "could not create the media id", err)
	}
	fingerprint, err := security.Fingerprint(struct {
		Kind, ContentType string
		ByteSize          int64
	}{input.Kind, contentType, input.ByteSize})
	if err != nil {
		return core.MediaObject{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	object, err := s.store.CreateMediaObject(ctx, core.CreateMediaCommand{
		Principal: principal, MediaID: mediaID, Kind: input.Kind,
		ContentType: contentType, ByteSize: input.ByteSize,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.MediaObject{}, normalizeStoreError("create media", err)
	}
	return object, nil
}

func (s *Service) AppendMediaChunk(ctx context.Context, principal core.Principal, mediaID string, offset int64, data []byte) (core.MediaObject, error) {
	normalizedID, err := security.ValidateIdentifier("mediaId", mediaID, 128)
	if err != nil {
		return core.MediaObject{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if offset < 0 {
		return core.MediaObject{}, core.E(core.CodeInvalidInput, "the upload offset must not be negative", nil)
	}
	if int64(len(data)) > s.maxMediaChunk {
		return core.MediaObject{}, core.E(core.CodeInvalidInput, "the upload chunk exceeds the chunk limit", nil)
	}
	object, err := s.store.AppendMediaChunk(ctx, core.AppendMediaChunkCommand{
		Principal: principal, MediaID: normalizedID, Offset: offset, Data: data,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.MediaObject{}, normalizeStoreError("append media chunk", err)
	}
	return object, nil
}

func (s *Service) GetMedia(ctx context.Context, principal core.Principal, mediaID string) (core.MediaObject, error) {
	normalizedID, err := security.ValidateIdentifier("mediaId", mediaID, 128)
	if err != nil {
		return core.MediaObject{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	object, err := s.store.GetMediaObject(ctx, principal, normalizedID)
	if err != nil {
		return core.MediaObject{}, normalizeStoreError("read media", err)
	}
	return object, nil
}

type MediaAccess struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type mediaAccessClaims struct {
	TenantID  string `json:"t"`
	MediaID   string `json:"m"`
	ExpiresAt int64  `json:"e"`
}

// SignMediaAccess issues a short-lived sealed link for playback. The
// grant re-checks visibility first, so a link only ever exists for a
// viewer who could read the object at signing time.
func (s *Service) SignMediaAccess(ctx context.Context, principal core.Principal, mediaID string) (MediaAccess, error) {
	object, err := s.GetMedia(ctx, principal, mediaID)
	if err != nil {
		return MediaAccess{}, err
	}
	if object.Status != core.MediaStatusReady {
		return MediaAccess{}, core.E(core.CodeInvalidState, "the media upload is not complete", nil)
	}
	if s.mediaBox == nil {
		return MediaAccess{}, core.E(core.CodeInternal, "media access signing is not configured", nil)
	}
	expiresAt := s.clock.Now().Add(s.mediaAccessTTL)
	payload, err := json.Marshal(mediaAccessClaims{
		TenantID: principal.TenantID, MediaID: object.ID, ExpiresAt: expiresAt.Unix(),
	})
	if err != nil {
		return MediaAccess{}, core.E(core.CodeInternal, "could not encode the access claims", err)
	}
	sealed, err := s.mediaBox.Seal(payload)
	if err != nil {
		return MediaAccess{}, core.E(core.CodeInternal, "could not seal the access token", err)
	}
	token := base64.RawURLEncoding.EncodeToString(sealed)
	return MediaAccess{
		URL:       "/v1/media/" + object.ID + "/content?token=" + token,
		ExpiresAt: expiresAt,
	}, nil
}

// MediaContentByToken serves bytes for a sealed access link.
func (s *Service) MediaContentByToken(ctx context.Context, mediaID, token string) ([]byte, string, error) {
	normalizedID, err := security.ValidateIdentifier("mediaId", mediaID, 128)
	if err != nil {
		return nil, "", core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if s.mediaBox == nil {
		return nil, "", core.E(core.CodeInternal, "media access signing is not configured", nil)
	}
	sealed, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, "", core.E(core.CodeUnauthenticated, "the media access link is invalid", nil)
	}
	payload, err := s.mediaBox.Open(sealed)
	if err != nil {
		return nil, "", core.E(core.CodeUnauthenticated, "the media access link is invalid", nil)
	}
	var claims mediaAccessClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, "", core.E(core.CodeUnauthenticated, "the media access link is invalid", nil)
	}
	if claims.MediaID != normalizedID {
		return nil, "", core.E(core.CodeUnauthenticated, "the media access link is invalid", nil)
	}
	if time.Unix(claims.ExpiresAt, 0).Before(s.clock.Now()) {
		return nil, "", core.E(core.CodeUnauthenticated, "the media access link expired", nil)
	}
	content, contentType, err := s.store.MediaContent(ctx, claims.TenantID, normalizedID)
	if err != nil {
		return nil, "", normalizeStoreError("read media content", err)
	}
	return content, contentType, nil
}
