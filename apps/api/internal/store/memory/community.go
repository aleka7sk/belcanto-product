package memory

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.5 community and safety — parity with PostgreSQL.

type communityPost struct {
	ID              string
	TenantID        string
	AuthorAccountID string
	Kind            string
	Title           string
	Body            string
	Audience        string
	CommentsEnabled bool
	Pinned          bool
	Status          string
	StatusReason    string
	CreatedAt       time.Time
}

type communityComment struct {
	ID              string
	TenantID        string
	PostID          string
	AuthorAccountID string
	Body            string
	Status          string
	StatusReason    string
	CreatedAt       time.Time
}

type communityReport struct {
	ID             string
	TenantID       string
	TargetType     string
	TargetID       string
	Reason         string
	Note           string
	Reporter       string
	Status         string
	Decision       string
	DecisionReason string
	DecidedBy      string
	DecidedAt      *time.Time
	CreatedAt      time.Time
}

func (s *Store) isStaffAccount(tenantID, accountID string) bool {
	account := s.activeAccount(accountID, tenantID)
	if account == nil {
		return false
	}
	return account.Roles[core.RoleOwner] != "" ||
		account.Roles[core.RoleAdministrator] != "" ||
		account.Roles[core.RoleTeacher] != ""
}

func (s *Store) isModeratorAccount(tenantID, accountID string) bool {
	account := s.activeAccount(accountID, tenantID)
	if account == nil {
		return false
	}
	return account.Roles[core.RoleOwner] != "" || account.Roles[core.RoleAdministrator] != ""
}

func (s *Store) communityGuidelinesAccepted(tenantID, accountID string) error {
	var latest *policyVersion
	for _, policy := range s.policies {
		if policy.TenantID != tenantID || policy.Kind != "community" {
			continue
		}
		if latest == nil || policy.EffectiveFrom.After(latest.EffectiveFrom) {
			latest = policy
		}
	}
	if latest == nil {
		return nil
	}
	for _, acceptance := range s.acceptances {
		if acceptance.TenantID == tenantID && acceptance.AccountID == accountID &&
			acceptance.PolicyVersionID == latest.ID {
			return nil
		}
	}
	return core.E(core.CodeForbidden, "community guidelines must be accepted first", nil)
}

func (s *Store) communityAuthorView(tenantID, accountID string) core.CommunityAuthor {
	author := core.CommunityAuthor{AccountID: accountID}
	account := s.accounts[accountID]
	if account == nil || account.TenantID != tenantID {
		return author
	}
	author.FullName = account.FullName
	for _, role := range []core.Role{core.RoleOwner, core.RoleAdministrator, core.RoleTeacher, core.RoleStudent} {
		if account.Roles[role] != "" {
			author.Role = string(role)
			break
		}
	}
	return author
}

func (s *Store) postView(stored *communityPost, moderator bool, blocked map[string]bool) core.CommunityPost {
	post := core.CommunityPost{
		ID: stored.ID, Kind: stored.Kind, Title: stored.Title, Body: stored.Body,
		Audience: stored.Audience, CommentsEnabled: stored.CommentsEnabled,
		Pinned: stored.Pinned, Status: stored.Status,
		Author:    s.communityAuthorView(stored.TenantID, stored.AuthorAccountID),
		CreatedAt: stored.CreatedAt,
		Comments:  []core.CommunityComment{},
	}
	if stored.Status != core.ContentPublished && !moderator {
		post.Title = ""
		post.Body = ""
		post.Author = core.CommunityAuthor{}
	}
	comments := []*communityComment{}
	for _, comment := range s.communityComments {
		if comment.TenantID == stored.TenantID && comment.PostID == stored.ID {
			comments = append(comments, comment)
		}
	}
	sort.Slice(comments, func(left, right int) bool {
		if !comments[left].CreatedAt.Equal(comments[right].CreatedAt) {
			return comments[left].CreatedAt.Before(comments[right].CreatedAt)
		}
		return comments[left].ID < comments[right].ID
	})
	for _, comment := range comments {
		if blocked[comment.AuthorAccountID] {
			continue
		}
		view := core.CommunityComment{
			ID: comment.ID, Status: comment.Status, CreatedAt: comment.CreatedAt,
		}
		if comment.Status == core.ContentPublished {
			post.CommentCount++
		}
		if comment.Status == core.ContentPublished || moderator {
			view.Body = comment.Body
			view.Author = s.communityAuthorView(stored.TenantID, comment.AuthorAccountID)
		}
		post.Comments = append(post.Comments, view)
	}
	return post
}

func (s *Store) blockedSet(tenantID, accountID string) map[string]bool {
	result := map[string]bool{}
	for key := range s.blocks {
		if key.tenantID == tenantID && key.blocker == accountID {
			result[key.blocked] = true
		}
	}
	return result
}

type blockKey struct {
	tenantID, blocker, blocked string
}

func (s *Store) CreatePost(_ context.Context, command core.CreatePostCommand) (core.CommunityPost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if s.activeAccount(principal.AccountID, principal.TenantID) == nil {
		return core.CommunityPost{}, core.E(core.CodeForbidden, "an active account is required", nil)
	}
	if err := s.communityGuidelinesAccepted(principal.TenantID, principal.AccountID); err != nil {
		return core.CommunityPost{}, err
	}
	staff := s.isStaffAccount(principal.TenantID, principal.AccountID)
	if command.Kind == core.PostKindAnnouncement && !staff {
		return core.CommunityPost{}, core.E(core.CodeForbidden, "announcements are published by the school", nil)
	}
	if command.Audience == core.AudienceStaff && !staff {
		return core.CommunityPost{}, core.E(core.CodeForbidden, "the staff audience is for staff posts", nil)
	}
	if response, ok, err := s.replay("create_post", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.CommunityPost{}, err
		}
		var result core.CommunityPost
		if err := json.Unmarshal(response, &result); err != nil {
			return core.CommunityPost{}, core.E(core.CodeInternal, "decode idempotent post result", err)
		}
		return result, nil
	}
	stored := &communityPost{
		ID: command.PostID, TenantID: principal.TenantID, AuthorAccountID: principal.AccountID,
		Kind: command.Kind, Title: command.Title, Body: command.Body,
		Audience: command.Audience, CommentsEnabled: command.CommentsEnabled,
		Pinned: command.Pinned, Status: core.ContentPublished, CreatedAt: command.Now,
	}
	s.communityPosts[command.PostID] = stored
	result := s.postView(stored, true, nil)
	if err := s.completeIdempotency("create_post", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.CommunityPost{}, err
	}
	action := "CommunityPostPublished"
	if command.Kind == core.PostKindAnnouncement {
		action = "CommunityAnnouncementPublished"
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, action,
		"community_post", stored.ID, "allow", "", command.Now, nil)
	s.appendOutboxPayload(principal.TenantID, action, "community_post", stored.ID,
		map[string]any{"postId": stored.ID, "audience": stored.Audience}, command.Now)
	return result, nil
}

func (s *Store) ListFeed(_ context.Context, principal core.Principal, limit int) ([]core.CommunityPost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeAccount(principal.AccountID, principal.TenantID) == nil {
		return nil, core.E(core.CodeForbidden, "an active account is required", nil)
	}
	staff := s.isStaffAccount(principal.TenantID, principal.AccountID)
	moderator := s.isModeratorAccount(principal.TenantID, principal.AccountID)
	blocked := s.blockedSet(principal.TenantID, principal.AccountID)
	posts := []*communityPost{}
	for _, stored := range s.communityPosts {
		if stored.TenantID != principal.TenantID || stored.Status != core.ContentPublished {
			continue
		}
		if stored.Audience == core.AudienceStaff && !staff {
			continue
		}
		posts = append(posts, stored)
	}
	sort.Slice(posts, func(left, right int) bool {
		if posts[left].Pinned != posts[right].Pinned {
			return posts[left].Pinned
		}
		if !posts[left].CreatedAt.Equal(posts[right].CreatedAt) {
			return posts[left].CreatedAt.After(posts[right].CreatedAt)
		}
		return posts[left].ID < posts[right].ID
	})
	result := []core.CommunityPost{}
	for index, stored := range posts {
		if index >= limit {
			break
		}
		view := s.postView(stored, moderator, blocked)
		view.Comments = nil
		result = append(result, view)
	}
	return result, nil
}

func (s *Store) GetPost(_ context.Context, principal core.Principal, postID string) (core.CommunityPost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeAccount(principal.AccountID, principal.TenantID) == nil {
		return core.CommunityPost{}, core.E(core.CodeForbidden, "an active account is required", nil)
	}
	stored := s.communityPosts[postID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.CommunityPost{}, core.E(core.CodeNotFound, "post not found", nil)
	}
	staff := s.isStaffAccount(principal.TenantID, principal.AccountID)
	if stored.Audience == core.AudienceStaff && !staff {
		return core.CommunityPost{}, core.E(core.CodeNotFound, "post not found", nil)
	}
	moderator := s.isModeratorAccount(principal.TenantID, principal.AccountID)
	blocked := s.blockedSet(principal.TenantID, principal.AccountID)
	return s.postView(stored, moderator, blocked), nil
}

func (s *Store) AddComment(_ context.Context, command core.AddCommentCommand) (core.CommunityPost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if s.activeAccount(principal.AccountID, principal.TenantID) == nil {
		return core.CommunityPost{}, core.E(core.CodeForbidden, "an active account is required", nil)
	}
	if err := s.communityGuidelinesAccepted(principal.TenantID, principal.AccountID); err != nil {
		return core.CommunityPost{}, err
	}
	stored := s.communityPosts[command.PostID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.CommunityPost{}, core.E(core.CodeNotFound, "post not found", nil)
	}
	if stored.Audience == core.AudienceStaff && !s.isStaffAccount(principal.TenantID, principal.AccountID) {
		return core.CommunityPost{}, core.E(core.CodeNotFound, "post not found", nil)
	}
	if stored.Status != core.ContentPublished {
		return core.CommunityPost{}, core.E(core.CodeInvalidState, "the post is not open for replies", nil)
	}
	if !stored.CommentsEnabled {
		return core.CommunityPost{}, core.E(core.CodeInvalidState, "replies are disabled for this post", nil)
	}
	if response, ok, err := s.replay("add_comment", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.CommunityPost{}, err
		}
		var result core.CommunityPost
		if err := json.Unmarshal(response, &result); err != nil {
			return core.CommunityPost{}, core.E(core.CodeInternal, "decode idempotent comment result", err)
		}
		return result, nil
	}
	s.communityComments[command.CommentID] = &communityComment{
		ID: command.CommentID, TenantID: principal.TenantID, PostID: command.PostID,
		AuthorAccountID: principal.AccountID, Body: command.Body,
		Status: core.ContentPublished, CreatedAt: command.Now,
	}
	result := s.postView(stored, false, nil)
	if err := s.completeIdempotency("add_comment", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.CommunityPost{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "CommunityCommentAdded",
		"community_comment", command.CommentID, "allow", "", command.Now, nil)
	return result, nil
}

func (s *Store) RemoveContent(_ context.Context, command core.RemoveContentCommand) (core.CommunityPost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if response, ok, err := s.replay("remove_content", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.CommunityPost{}, err
		}
		var result core.CommunityPost
		if err := json.Unmarshal(response, &result); err != nil {
			return core.CommunityPost{}, core.E(core.CodeInternal, "decode idempotent removal result", err)
		}
		return result, nil
	}
	var postID, authorID, status string
	if command.TargetType == "comment" {
		comment := s.communityComments[command.TargetID]
		if comment == nil || comment.TenantID != principal.TenantID {
			return core.CommunityPost{}, core.E(core.CodeNotFound, "content not found", nil)
		}
		postID, authorID, status = comment.PostID, comment.AuthorAccountID, comment.Status
	} else {
		post := s.communityPosts[command.TargetID]
		if post == nil || post.TenantID != principal.TenantID {
			return core.CommunityPost{}, core.E(core.CodeNotFound, "content not found", nil)
		}
		postID, authorID, status = post.ID, post.AuthorAccountID, post.Status
	}
	if authorID != principal.AccountID {
		return core.CommunityPost{}, core.E(core.CodeForbidden, "only the author removes their content", nil)
	}
	if status != core.ContentPublished {
		return core.CommunityPost{}, core.E(core.CodeInvalidState, "the content is already unavailable", nil)
	}
	if command.TargetType == "comment" {
		s.communityComments[command.TargetID].Status = core.ContentRemoved
		s.communityComments[command.TargetID].StatusReason = "removed by the author"
	} else {
		s.communityPosts[command.TargetID].Status = core.ContentRemoved
		s.communityPosts[command.TargetID].StatusReason = "removed by the author"
	}
	result := s.postView(s.communityPosts[postID], false, nil)
	if err := s.completeIdempotency("remove_content", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.CommunityPost{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "CommunityContentRemoved",
		"community_"+command.TargetType, command.TargetID, "allow", "", command.Now, nil)
	return result, nil
}

func (s *Store) ReportContent(_ context.Context, command core.ReportContentCommand) (core.CommunityReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if s.activeAccount(principal.AccountID, principal.TenantID) == nil {
		return core.CommunityReport{}, core.E(core.CodeForbidden, "an active account is required", nil)
	}
	exists := false
	if command.TargetType == "comment" {
		comment := s.communityComments[command.TargetID]
		exists = comment != nil && comment.TenantID == principal.TenantID
	} else {
		post := s.communityPosts[command.TargetID]
		exists = post != nil && post.TenantID == principal.TenantID
	}
	if !exists {
		return core.CommunityReport{}, core.E(core.CodeNotFound, "content not found", nil)
	}
	if response, ok, err := s.replay("report_content", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.CommunityReport{}, err
		}
		var result core.CommunityReport
		if err := json.Unmarshal(response, &result); err != nil {
			return core.CommunityReport{}, core.E(core.CodeInternal, "decode idempotent report result", err)
		}
		return result, nil
	}
	stored := &communityReport{
		ID: command.ReportID, TenantID: principal.TenantID,
		TargetType: command.TargetType, TargetID: command.TargetID,
		Reason: command.Reason, Note: command.Note,
		Reporter: principal.AccountID, Status: "new", CreatedAt: command.Now,
	}
	s.communityReports[command.ReportID] = stored
	result := core.CommunityReport{
		ID: stored.ID, TargetType: stored.TargetType, TargetID: stored.TargetID,
		Reason: stored.Reason, Note: stored.Note, Status: stored.Status,
		CreatedAt: stored.CreatedAt,
	}
	if err := s.completeIdempotency("report_content", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.CommunityReport{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "CommunityReportFiled",
		"community_report", stored.ID, "allow", "", command.Now, nil)
	s.appendOutboxPayload(principal.TenantID, "CommunityReportFiled", "community_report", stored.ID,
		map[string]any{"reportId": stored.ID, "reason": stored.Reason}, command.Now)
	return result, nil
}

func (s *Store) reportView(stored *communityReport) core.CommunityReport {
	view := core.CommunityReport{
		ID: stored.ID, TargetType: stored.TargetType, TargetID: stored.TargetID,
		Reason: stored.Reason, Note: stored.Note, Status: stored.Status,
		Decision: stored.Decision, DecisionReason: stored.DecisionReason,
		CreatedAt: stored.CreatedAt,
	}
	if stored.DecidedAt != nil {
		decided := *stored.DecidedAt
		view.DecidedAt = &decided
	}
	if stored.TargetType == "comment" {
		if comment := s.communityComments[stored.TargetID]; comment != nil {
			view.TargetExcerpt = comment.Body
		}
	} else if post := s.communityPosts[stored.TargetID]; post != nil {
		view.TargetExcerpt = post.Body
	}
	if len(view.TargetExcerpt) > 200 {
		view.TargetExcerpt = view.TargetExcerpt[:200]
	}
	return view
}

func (s *Store) ListModerationQueue(_ context.Context, principal core.Principal) ([]core.CommunityReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.isModeratorAccount(principal.TenantID, principal.AccountID) {
		return nil, core.E(core.CodeForbidden, "moderation is a school capability", nil)
	}
	result := []core.CommunityReport{}
	for _, stored := range s.communityReports {
		if stored.TenantID != principal.TenantID {
			continue
		}
		result = append(result, s.reportView(stored))
	}
	sort.Slice(result, func(left, right int) bool {
		leftNew := result[left].Status == "new"
		rightNew := result[right].Status == "new"
		if leftNew != rightNew {
			return leftNew
		}
		if !result[left].CreatedAt.Equal(result[right].CreatedAt) {
			return result[left].CreatedAt.After(result[right].CreatedAt)
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}

func (s *Store) DecideReport(_ context.Context, command core.DecideReportCommand) (core.CommunityReport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if !s.isModeratorAccount(principal.TenantID, principal.AccountID) {
		return core.CommunityReport{}, core.E(core.CodeForbidden, "moderation is a school capability", nil)
	}
	if response, ok, err := s.replay("decide_report", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.CommunityReport{}, err
		}
		var result core.CommunityReport
		if err := json.Unmarshal(response, &result); err != nil {
			return core.CommunityReport{}, core.E(core.CodeInternal, "decode idempotent decision result", err)
		}
		return result, nil
	}
	stored := s.communityReports[command.ReportID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.CommunityReport{}, core.E(core.CodeNotFound, "report not found", nil)
	}
	if stored.Status != "new" {
		return core.CommunityReport{}, core.E(core.CodeInvalidState, "the report is already reviewed", nil)
	}
	if command.Decision == "hidden" {
		if stored.TargetType == "comment" {
			if comment := s.communityComments[stored.TargetID]; comment != nil &&
				comment.Status == core.ContentPublished {
				comment.Status = core.ContentHidden
				comment.StatusReason = command.DecisionReason
			}
		} else if post := s.communityPosts[stored.TargetID]; post != nil &&
			post.Status == core.ContentPublished {
			post.Status = core.ContentHidden
			post.StatusReason = command.DecisionReason
		}
	}
	decidedAt := command.Now
	stored.Status = "reviewed"
	stored.Decision = command.Decision
	stored.DecisionReason = command.DecisionReason
	stored.DecidedBy = principal.AccountID
	stored.DecidedAt = &decidedAt
	result := s.reportView(stored)
	if err := s.completeIdempotency("decide_report", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.CommunityReport{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "CommunityReportDecided",
		"community_report", stored.ID, "allow", command.DecisionReason, command.Now, nil)
	return result, nil
}

func (s *Store) BlockMember(ctx context.Context, command core.BlockMemberCommand) ([]string, error) {
	s.mu.Lock()
	principal := command.Principal
	if s.activeAccount(principal.AccountID, principal.TenantID) == nil {
		s.mu.Unlock()
		return nil, core.E(core.CodeForbidden, "an active account is required", nil)
	}
	key := blockKey{tenantID: principal.TenantID, blocker: principal.AccountID, blocked: command.BlockedAccountID}
	if command.Blocked {
		s.blocks[key] = command.Now
	} else {
		delete(s.blocks, key)
	}
	action := "CommunityMemberBlocked"
	if !command.Blocked {
		action = "CommunityMemberUnblocked"
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, action,
		"community_block", command.BlockedAccountID, "allow", "", command.Now, nil)
	s.mu.Unlock()
	return s.ListBlockedMembers(ctx, principal)
}

func (s *Store) ListBlockedMembers(_ context.Context, principal core.Principal) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []string{}
	for key := range s.blocks {
		if key.tenantID == principal.TenantID && key.blocker == principal.AccountID {
			result = append(result, key.blocked)
		}
	}
	sort.Strings(result)
	return result, nil
}
