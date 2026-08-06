package app

import (
	"context"
	"slices"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// L.5 community and safety (Figma Page 28). Posts are text-only while
// DEC-103 (guardian consent for community media) is open; there is no
// student discovery surface and no Student↔Student DM — chat is a
// separate future slice per the production prompt.

const communityFeedLimit = 100

type CreatePostInput struct {
	Kind            string
	Title           string
	Body            string
	Audience        string
	CommentsEnabled bool
	Pinned          bool
	IdempotencyKey  string
}

func (s *Service) CreateCommunityPost(ctx context.Context, principal core.Principal, input CreatePostInput) (core.CommunityPost, error) {
	kind := input.Kind
	if kind == "" {
		kind = core.PostKindPost
	}
	if kind != core.PostKindPost && kind != core.PostKindAnnouncement {
		return core.CommunityPost{}, core.E(core.CodeInvalidInput, "kind must be post or announcement", nil)
	}
	audience := input.Audience
	if audience == "" {
		audience = core.AudienceSchool
	}
	if audience != core.AudienceSchool && audience != core.AudienceStaff {
		return core.CommunityPost{}, core.E(core.CodeInvalidInput, "audience must be school or staff", nil)
	}
	body, err := security.ValidateText("body", input.Body, 1, 2000)
	if err != nil {
		return core.CommunityPost{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	title := ""
	if input.Title != "" {
		title, err = security.ValidateText("title", input.Title, 1, 200)
		if err != nil {
			return core.CommunityPost{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	if kind == core.PostKindAnnouncement && title == "" {
		return core.CommunityPost{}, core.E(core.CodeInvalidInput, "an announcement carries a title", nil)
	}
	if kind == core.PostKindPost && input.Pinned {
		return core.CommunityPost{}, core.E(core.CodeInvalidInput, "only an announcement is pinned", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.CommunityPost{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		Kind, Title, Body, Audience string
		CommentsEnabled, Pinned     bool
	}{kind, title, body, audience, input.CommentsEnabled, input.Pinned})
	if err != nil {
		return core.CommunityPost{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	ids, err := newIDs("post")
	if err != nil {
		return core.CommunityPost{}, core.E(core.CodeInternal, "could not create identifiers", err)
	}
	post, err := s.store.CreatePost(ctx, core.CreatePostCommand{
		Principal: principal, PostID: ids[0], Kind: kind, Title: title, Body: body,
		Audience: audience, CommentsEnabled: input.CommentsEnabled, Pinned: input.Pinned,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.CommunityPost{}, normalizeStoreError("create post", err)
	}
	return post, nil
}

func (s *Service) CommunityFeed(ctx context.Context, principal core.Principal) ([]core.CommunityPost, error) {
	posts, err := s.store.ListFeed(ctx, principal, communityFeedLimit)
	if err != nil {
		return nil, normalizeStoreError("list feed", err)
	}
	return posts, nil
}

func (s *Service) CommunityPost(ctx context.Context, principal core.Principal, postID string) (core.CommunityPost, error) {
	normalizedID, err := security.ValidateIdentifier("postId", postID, 128)
	if err != nil {
		return core.CommunityPost{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	post, err := s.store.GetPost(ctx, principal, normalizedID)
	if err != nil {
		return core.CommunityPost{}, normalizeStoreError("get post", err)
	}
	return post, nil
}

type AddCommentInput struct {
	PostID         string
	Body           string
	IdempotencyKey string
}

func (s *Service) AddCommunityComment(ctx context.Context, principal core.Principal, input AddCommentInput) (core.CommunityPost, error) {
	postID, err := security.ValidateIdentifier("postId", input.PostID, 128)
	if err != nil {
		return core.CommunityPost{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	body, err := security.ValidateText("body", input.Body, 1, 1000)
	if err != nil {
		return core.CommunityPost{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.CommunityPost{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct{ PostID, Body string }{postID, body})
	if err != nil {
		return core.CommunityPost{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	ids, err := newIDs("comment")
	if err != nil {
		return core.CommunityPost{}, core.E(core.CodeInternal, "could not create identifiers", err)
	}
	post, err := s.store.AddComment(ctx, core.AddCommentCommand{
		Principal: principal, CommentID: ids[0], PostID: postID, Body: body,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.CommunityPost{}, normalizeStoreError("add comment", err)
	}
	return post, nil
}

func validateCommunityTarget(targetType, targetID string) (string, string, error) {
	if targetType != "post" && targetType != "comment" {
		return "", "", core.E(core.CodeInvalidInput, "targetType must be post or comment", nil)
	}
	normalizedID, err := security.ValidateIdentifier("targetId", targetID, 128)
	if err != nil {
		return "", "", core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	return targetType, normalizedID, nil
}

type RemoveContentInput struct {
	TargetType     string
	TargetID       string
	IdempotencyKey string
}

func (s *Service) RemoveCommunityContent(ctx context.Context, principal core.Principal, input RemoveContentInput) (core.CommunityPost, error) {
	targetType, targetID, err := validateCommunityTarget(input.TargetType, input.TargetID)
	if err != nil {
		return core.CommunityPost{}, err
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.CommunityPost{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct{ TargetType, TargetID string }{targetType, targetID})
	if err != nil {
		return core.CommunityPost{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	post, err := s.store.RemoveContent(ctx, core.RemoveContentCommand{
		Principal: principal, TargetType: targetType, TargetID: targetID,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.CommunityPost{}, normalizeStoreError("remove content", err)
	}
	return post, nil
}

type ReportContentInput struct {
	TargetType     string
	TargetID       string
	Reason         string
	Note           string
	IdempotencyKey string
}

func (s *Service) ReportCommunityContent(ctx context.Context, principal core.Principal, input ReportContentInput) (core.CommunityReport, error) {
	targetType, targetID, err := validateCommunityTarget(input.TargetType, input.TargetID)
	if err != nil {
		return core.CommunityReport{}, err
	}
	if !slices.Contains(core.ReportReasons, input.Reason) {
		return core.CommunityReport{}, core.E(core.CodeInvalidInput, "reason must be abuse, personal_data, spam or other", nil)
	}
	note := ""
	if input.Note != "" {
		note, err = security.ValidateText("note", input.Note, 1, 1000)
		if err != nil {
			return core.CommunityReport{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	if input.Reason == "other" && note == "" {
		return core.CommunityReport{}, core.E(core.CodeInvalidInput, "the reason \"other\" carries a note", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.CommunityReport{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		TargetType, TargetID, Reason, Note string
	}{targetType, targetID, input.Reason, note})
	if err != nil {
		return core.CommunityReport{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	ids, err := newIDs("report")
	if err != nil {
		return core.CommunityReport{}, core.E(core.CodeInternal, "could not create identifiers", err)
	}
	report, err := s.store.ReportContent(ctx, core.ReportContentCommand{
		Principal: principal, ReportID: ids[0], TargetType: targetType, TargetID: targetID,
		Reason: input.Reason, Note: note,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.CommunityReport{}, normalizeStoreError("report content", err)
	}
	return report, nil
}

func (s *Service) ModerationQueue(ctx context.Context, principal core.Principal) ([]core.CommunityReport, error) {
	reports, err := s.store.ListModerationQueue(ctx, principal)
	if err != nil {
		return nil, normalizeStoreError("list moderation queue", err)
	}
	return reports, nil
}

type DecideReportInput struct {
	ReportID       string
	Decision       string
	DecisionReason string
	IdempotencyKey string
}

func (s *Service) DecideCommunityReport(ctx context.Context, principal core.Principal, input DecideReportInput) (core.CommunityReport, error) {
	reportID, err := security.ValidateIdentifier("reportId", input.ReportID, 128)
	if err != nil {
		return core.CommunityReport{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if input.Decision != "hidden" && input.Decision != "kept" {
		return core.CommunityReport{}, core.E(core.CodeInvalidInput, "decision must be hidden or kept", nil)
	}
	reason, err := security.ValidateText("reason", input.DecisionReason, 1, 500)
	if err != nil {
		return core.CommunityReport{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.CommunityReport{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		ReportID, Decision, Reason string
	}{reportID, input.Decision, reason})
	if err != nil {
		return core.CommunityReport{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	report, err := s.store.DecideReport(ctx, core.DecideReportCommand{
		Principal: principal, ReportID: reportID, Decision: input.Decision,
		DecisionReason: reason, IdempotencyKey: idempotencyKey,
		PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.CommunityReport{}, normalizeStoreError("decide report", err)
	}
	return report, nil
}

type BlockMemberInput struct {
	BlockedAccountID string
	Blocked          bool
}

func (s *Service) BlockCommunityMember(ctx context.Context, principal core.Principal, input BlockMemberInput) ([]string, error) {
	blockedID, err := security.ValidateIdentifier("accountId", input.BlockedAccountID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if blockedID == principal.AccountID {
		return nil, core.E(core.CodeInvalidInput, "you cannot block yourself", nil)
	}
	blocked, err := s.store.BlockMember(ctx, core.BlockMemberCommand{
		Principal: principal, BlockedAccountID: blockedID,
		Blocked: input.Blocked, Now: s.clock.Now(),
	})
	if err != nil {
		return nil, normalizeStoreError("block member", err)
	}
	return blocked, nil
}

func (s *Service) BlockedCommunityMembers(ctx context.Context, principal core.Principal) ([]string, error) {
	blocked, err := s.store.ListBlockedMembers(ctx, principal)
	if err != nil {
		return nil, normalizeStoreError("list blocked members", err)
	}
	return blocked, nil
}
