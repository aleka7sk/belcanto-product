package postgres

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.5 community and safety. Audience is role-aware (staff posts stay
// staff-only), nothing is deleted — removal and moderation are statuses
// with tombstones (COM-SAFE-05) — the reporter is never exposed to
// non-moderators, and posting requires accepting the school's community
// guidelines when such a policy exists (COM-SAFE-04, data-driven).

func isStaffPrincipal(ctx context.Context, tx pgx.Tx, tenantID, accountID string) (bool, error) {
	var staff bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM role_grants
			WHERE tenant_id = $1 AND account_id = $2
			  AND role_type IN ('Owner', 'Administrator', 'Teacher') AND status = 'active'
		)
	`, tenantID, accountID).Scan(&staff); err != nil {
		return false, fmt.Errorf("check staff principal: %w", err)
	}
	return staff, nil
}

// communityGuidelinesAccepted: when the school has published a
// community policy, the newest effective version must be accepted
// before writing to the community.
func communityGuidelinesAccepted(ctx context.Context, tx pgx.Tx, tenantID, accountID string) error {
	var accepted *bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM policy_acceptances a
			WHERE a.tenant_id = p.tenant_id AND a.account_id = $2
			  AND a.policy_version_id = p.id
		)
		FROM policy_versions p
		WHERE p.tenant_id = $1 AND p.kind = 'community' AND p.effective_from <= now()
		ORDER BY p.effective_from DESC
		LIMIT 1
	`, tenantID, accountID).Scan(&accepted)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("check community guidelines: %w", err)
	}
	if accepted == nil || !*accepted {
		return core.E(core.CodeForbidden, "community guidelines must be accepted first", nil)
	}
	return nil
}

func communityAuthor(ctx context.Context, reader lessonReader, tenantID, accountID string) (core.CommunityAuthor, error) {
	author := core.CommunityAuthor{AccountID: accountID}
	err := reader.QueryRow(ctx, `
		SELECT person.full_name,
		       COALESCE((
		           SELECT role_type FROM role_grants
		           WHERE tenant_id = account.tenant_id AND account_id = account.id AND status = 'active'
		           ORDER BY CASE role_type
		               WHEN 'Owner' THEN 1 WHEN 'Administrator' THEN 2
		               WHEN 'Teacher' THEN 3 ELSE 4 END
		           LIMIT 1
		       ), '')
		FROM accounts account
		JOIN people person ON person.tenant_id = account.tenant_id AND person.id = account.person_id
		WHERE account.tenant_id = $1 AND account.id = $2
	`, tenantID, accountID).Scan(&author.FullName, &author.Role)
	if err != nil {
		return core.CommunityAuthor{}, fmt.Errorf("read community author: %w", err)
	}
	return author, nil
}

func readCommunityPost(ctx context.Context, reader lessonReader, tenantID, postID string, moderator bool, blocked map[string]bool) (core.CommunityPost, error) {
	var post core.CommunityPost
	var title, statusReason *string
	var authorID string
	err := reader.QueryRow(ctx, `
		SELECT id, kind, title, body, audience, comments_enabled, pinned,
		       status, status_reason, author_account_id, created_at
		FROM community_posts
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, postID).Scan(&post.ID, &post.Kind, &title, &post.Body, &post.Audience,
		&post.CommentsEnabled, &post.Pinned, &post.Status, &statusReason, &authorID, &post.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.CommunityPost{}, core.E(core.CodeNotFound, "post not found", nil)
	}
	if err != nil {
		return core.CommunityPost{}, fmt.Errorf("read community post: %w", err)
	}
	if title != nil {
		post.Title = *title
	}
	post.CreatedAt = post.CreatedAt.UTC()
	author, err := communityAuthor(ctx, reader, tenantID, authorID)
	if err != nil {
		return core.CommunityPost{}, err
	}
	post.Author = author
	if post.Status != core.ContentPublished && !moderator {
		// Tombstone (COM-SAFE-05): the mark and moment stay, the words
		// and the author leave; the reason is never revealed here.
		post.Title = ""
		post.Body = ""
		post.Author = core.CommunityAuthor{}
	}
	post.Comments = make([]core.CommunityComment, 0)
	rows, err := reader.Query(ctx, `
		SELECT c.id, c.author_account_id, c.body, c.status, c.created_at
		FROM community_comments c
		WHERE c.tenant_id = $1 AND c.post_id = $2
		ORDER BY c.created_at, c.id
	`, tenantID, postID)
	if err != nil {
		return core.CommunityPost{}, fmt.Errorf("read post comments: %w", err)
	}
	defer rows.Close()
	pending := []core.CommunityComment{}
	authorIDs := []string{}
	for rows.Next() {
		var comment core.CommunityComment
		var commentAuthor string
		if err := rows.Scan(&comment.ID, &commentAuthor, &comment.Body, &comment.Status, &comment.CreatedAt); err != nil {
			return core.CommunityPost{}, fmt.Errorf("scan post comment: %w", err)
		}
		comment.CreatedAt = comment.CreatedAt.UTC()
		comment.Author = core.CommunityAuthor{AccountID: commentAuthor}
		pending = append(pending, comment)
		authorIDs = append(authorIDs, commentAuthor)
	}
	if err := rows.Err(); err != nil {
		return core.CommunityPost{}, fmt.Errorf("iterate post comments: %w", err)
	}
	rows.Close()
	for index := range pending {
		comment := &pending[index]
		if blocked[comment.Author.AccountID] {
			// The blocker stops seeing direct interactions.
			continue
		}
		if comment.Status == core.ContentPublished {
			post.CommentCount++
		}
		if comment.Status == core.ContentPublished || moderator {
			commentAuthorView, err := communityAuthor(ctx, reader, tenantID, authorIDs[index])
			if err != nil {
				return core.CommunityPost{}, err
			}
			comment.Author = commentAuthorView
		} else {
			// Comment tombstone keeps thread context.
			comment.Body = ""
			comment.Author = core.CommunityAuthor{}
		}
		post.Comments = append(post.Comments, *comment)
	}
	return post, nil
}

func (s *Store) blockedSetFor(ctx context.Context, tenantID, accountID string) (map[string]bool, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT blocked_account_id FROM community_blocks
		WHERE tenant_id = $1 AND blocker_account_id = $2
	`, tenantID, accountID)
	if err != nil {
		return nil, fmt.Errorf("list blocks: %w", err)
	}
	defer rows.Close()
	result := map[string]bool{}
	for rows.Next() {
		var blocked string
		if err := rows.Scan(&blocked); err != nil {
			return nil, fmt.Errorf("scan block: %w", err)
		}
		result[blocked] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate blocks: %w", err)
	}
	return result, nil
}

func (s *Store) CreatePost(ctx context.Context, command core.CreatePostCommand) (core.CommunityPost, error) {
	principal := command.Principal
	var post core.CommunityPost
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := activeAccountExists(ctx, tx, principal.TenantID, principal.AccountID); err != nil {
			return err
		}
		if err := communityGuidelinesAccepted(ctx, tx, principal.TenantID, principal.AccountID); err != nil {
			return err
		}
		staff, err := isStaffPrincipal(ctx, tx, principal.TenantID, principal.AccountID)
		if err != nil {
			return err
		}
		if command.Kind == core.PostKindAnnouncement && !staff {
			return core.E(core.CodeForbidden, "announcements are published by the school", nil)
		}
		if command.Audience == core.AudienceStaff && !staff {
			return core.E(core.CodeForbidden, "the staff audience is for staff posts", nil)
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_post", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			post, err = decodeReplay[core.CommunityPost](claim)
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO community_posts (
				id, tenant_id, author_account_id, kind, title, body, audience,
				comments_enabled, pinned, status, created_at, updated_at
			) VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9, 'published', $10, $10)
		`, command.PostID, principal.TenantID, principal.AccountID, command.Kind,
			command.Title, command.Body, command.Audience,
			command.CommentsEnabled, command.Pinned, command.Now); err != nil {
			return mapWriteError(err, "post conflicts with existing data")
		}
		post, err = readCommunityPost(ctx, tx, principal.TenantID, command.PostID, true, nil)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "create_post", command.IdempotencyKey, post, command.Now); err != nil {
			return err
		}
		action := "CommunityPostPublished"
		if command.Kind == core.PostKindAnnouncement {
			action = "CommunityAnnouncementPublished"
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: action, targetType: "community_post", targetID: command.PostID,
			decision: "allow", idempotencyKey: command.IdempotencyKey,
			metadata: map[string]any{"audience": command.Audience},
			at:       command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, action, "community_post", command.PostID,
			map[string]any{"postId": command.PostID, "audience": command.Audience}, command.Now)
	})
	if err != nil {
		return core.CommunityPost{}, err
	}
	return post, nil
}

func (s *Store) communityViewerScope(ctx context.Context, principal core.Principal) (staff, moderator bool, err error) {
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := activeAccountExists(ctx, tx, principal.TenantID, principal.AccountID); err != nil {
			return err
		}
		isStaff, staffErr := isStaffPrincipal(ctx, tx, principal.TenantID, principal.AccountID)
		if staffErr != nil {
			return staffErr
		}
		staff = isStaff
		authority, authorityErr := lessonManagementAuthority(ctx, tx, principal.TenantID, principal.AccountID)
		if authorityErr != nil {
			return authorityErr
		}
		moderator = authority
		return nil
	})
	return staff, moderator, err
}

func (s *Store) ListFeed(ctx context.Context, principal core.Principal, limit int) ([]core.CommunityPost, error) {
	staff, moderator, err := s.communityViewerScope(ctx, principal)
	if err != nil {
		return nil, err
	}
	blocked, err := s.blockedSetFor(ctx, principal.TenantID, principal.AccountID)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id FROM community_posts
		WHERE tenant_id = $1 AND status = 'published'
		  AND (audience = 'school' OR $2::boolean)
		ORDER BY pinned DESC, created_at DESC
		LIMIT $3
	`, principal.TenantID, staff, limit)
	if err != nil {
		return nil, fmt.Errorf("list community feed: %w", err)
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan feed id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate feed ids: %w", err)
	}
	rows.Close()
	result := make([]core.CommunityPost, 0, len(ids))
	for _, id := range ids {
		post, err := readCommunityPost(ctx, s.pool, principal.TenantID, id, moderator, blocked)
		if err != nil {
			return nil, err
		}
		post.Comments = nil
		result = append(result, post)
	}
	return result, nil
}

func (s *Store) GetPost(ctx context.Context, principal core.Principal, postID string) (core.CommunityPost, error) {
	staff, moderator, err := s.communityViewerScope(ctx, principal)
	if err != nil {
		return core.CommunityPost{}, err
	}
	blocked, err := s.blockedSetFor(ctx, principal.TenantID, principal.AccountID)
	if err != nil {
		return core.CommunityPost{}, err
	}
	post, err := readCommunityPost(ctx, s.pool, principal.TenantID, postID, moderator, blocked)
	if err != nil {
		return core.CommunityPost{}, err
	}
	if post.Audience == core.AudienceStaff && !staff {
		return core.CommunityPost{}, core.E(core.CodeNotFound, "post not found", nil)
	}
	return post, nil
}

func (s *Store) AddComment(ctx context.Context, command core.AddCommentCommand) (core.CommunityPost, error) {
	principal := command.Principal
	var post core.CommunityPost
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := activeAccountExists(ctx, tx, principal.TenantID, principal.AccountID); err != nil {
			return err
		}
		if err := communityGuidelinesAccepted(ctx, tx, principal.TenantID, principal.AccountID); err != nil {
			return err
		}
		var status, audience string
		var commentsEnabled bool
		err := tx.QueryRow(ctx, `
			SELECT status, audience, comments_enabled FROM community_posts
			WHERE tenant_id = $1 AND id = $2
			FOR UPDATE
		`, principal.TenantID, command.PostID).Scan(&status, &audience, &commentsEnabled)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "post not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock post for comment: %w", err)
		}
		if audience == core.AudienceStaff {
			staff, staffErr := isStaffPrincipal(ctx, tx, principal.TenantID, principal.AccountID)
			if staffErr != nil {
				return staffErr
			}
			if !staff {
				return core.E(core.CodeNotFound, "post not found", nil)
			}
		}
		if status != core.ContentPublished {
			return core.E(core.CodeInvalidState, "the post is not open for replies", nil)
		}
		if !commentsEnabled {
			return core.E(core.CodeInvalidState, "replies are disabled for this post", nil)
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "add_comment", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			post, err = decodeReplay[core.CommunityPost](claim)
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO community_comments (
				id, tenant_id, post_id, author_account_id, body, status, created_at
			) VALUES ($1, $2, $3, $4, $5, 'published', $6)
		`, command.CommentID, principal.TenantID, command.PostID, principal.AccountID,
			command.Body, command.Now); err != nil {
			return mapWriteError(err, "comment conflicts with existing data")
		}
		post, err = readCommunityPost(ctx, tx, principal.TenantID, command.PostID, false, nil)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "add_comment", command.IdempotencyKey, post, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "CommunityCommentAdded", targetType: "community_comment",
			targetID: command.CommentID, decision: "allow",
			idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	})
	if err != nil {
		return core.CommunityPost{}, err
	}
	return post, nil
}

func (s *Store) RemoveContent(ctx context.Context, command core.RemoveContentCommand) (core.CommunityPost, error) {
	principal := command.Principal
	var post core.CommunityPost
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "remove_content", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			post, err = decodeReplay[core.CommunityPost](claim)
			return err
		}
		table := "community_posts"
		if command.TargetType == "comment" {
			table = "community_comments"
		}
		var authorID, status, postID string
		if command.TargetType == "comment" {
			err = tx.QueryRow(ctx, `
				SELECT author_account_id, status, post_id FROM community_comments
				WHERE tenant_id = $1 AND id = $2 FOR UPDATE
			`, principal.TenantID, command.TargetID).Scan(&authorID, &status, &postID)
		} else {
			postID = command.TargetID
			err = tx.QueryRow(ctx, `
				SELECT author_account_id, status FROM community_posts
				WHERE tenant_id = $1 AND id = $2 FOR UPDATE
			`, principal.TenantID, command.TargetID).Scan(&authorID, &status)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "content not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock content for removal: %w", err)
		}
		if authorID != principal.AccountID {
			return core.E(core.CodeForbidden, "only the author removes their content", nil)
		}
		if status != core.ContentPublished {
			return core.E(core.CodeInvalidState, "the content is already unavailable", nil)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE `+table+`
			SET status = 'removed', status_reason = 'removed by the author'
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.TargetID); err != nil {
			return mapWriteError(err, "content conflicts with existing data")
		}
		post, err = readCommunityPost(ctx, tx, principal.TenantID, postID, false, nil)
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "remove_content", command.IdempotencyKey, post, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "CommunityContentRemoved", targetType: "community_" + command.TargetType,
			targetID: command.TargetID, decision: "allow",
			idempotencyKey: command.IdempotencyKey, at: command.Now,
		})
	})
	if err != nil {
		return core.CommunityPost{}, err
	}
	return post, nil
}

func (s *Store) ReportContent(ctx context.Context, command core.ReportContentCommand) (core.CommunityReport, error) {
	principal := command.Principal
	var report core.CommunityReport
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := activeAccountExists(ctx, tx, principal.TenantID, principal.AccountID); err != nil {
			return err
		}
		table := "community_posts"
		if command.TargetType == "comment" {
			table = "community_comments"
		}
		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (SELECT 1 FROM `+table+` WHERE tenant_id = $1 AND id = $2)
		`, principal.TenantID, command.TargetID).Scan(&exists); err != nil {
			return fmt.Errorf("check report target: %w", err)
		}
		if !exists {
			return core.E(core.CodeNotFound, "content not found", nil)
		}
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "report_content", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			report, err = decodeReplay[core.CommunityReport](claim)
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO community_reports (
				id, tenant_id, target_type, target_id, reason, note,
				reporter_account_id, status, created_at
			) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, 'new', $8)
		`, command.ReportID, principal.TenantID, command.TargetType, command.TargetID,
			command.Reason, command.Note, principal.AccountID, command.Now); err != nil {
			return mapWriteError(err, "report conflicts with existing data")
		}
		report = core.CommunityReport{
			ID: command.ReportID, TargetType: command.TargetType, TargetID: command.TargetID,
			Reason: command.Reason, Note: command.Note, Status: "new", CreatedAt: command.Now,
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "report_content", command.IdempotencyKey, report, command.Now); err != nil {
			return err
		}
		// The reporter appears in the audit trail (staff-only) but never
		// in any member-facing view.
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "CommunityReportFiled", targetType: "community_report",
			targetID: command.ReportID, decision: "allow",
			idempotencyKey: command.IdempotencyKey,
			metadata:       map[string]any{"reason": command.Reason, "targetType": command.TargetType},
			at:             command.Now,
		}); err != nil {
			return err
		}
		return appendOutbox(ctx, tx, principal.TenantID, "CommunityReportFiled", "community_report", command.ReportID,
			map[string]any{"reportId": command.ReportID, "reason": command.Reason}, command.Now)
	})
	if err != nil {
		return core.CommunityReport{}, err
	}
	return report, nil
}

func (s *Store) moderatorAuthority(ctx context.Context, principal core.Principal) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		authority, err := lessonManagementAuthority(ctx, tx, principal.TenantID, principal.AccountID)
		if err != nil {
			return err
		}
		if !authority {
			return core.E(core.CodeForbidden, "moderation is a school capability", nil)
		}
		return nil
	})
}

func scanCommunityReport(row pgx.Row) (core.CommunityReport, error) {
	var report core.CommunityReport
	var note, decision, decisionReason *string
	err := row.Scan(&report.ID, &report.TargetType, &report.TargetID, &report.Reason,
		&note, &report.Status, &decision, &decisionReason, &report.DecidedAt, &report.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return core.CommunityReport{}, core.E(core.CodeNotFound, "report not found", nil)
	}
	if err != nil {
		return core.CommunityReport{}, fmt.Errorf("read report: %w", err)
	}
	if note != nil {
		report.Note = *note
	}
	if decision != nil {
		report.Decision = *decision
	}
	if decisionReason != nil {
		report.DecisionReason = *decisionReason
	}
	report.CreatedAt = report.CreatedAt.UTC()
	if report.DecidedAt != nil {
		utc := report.DecidedAt.UTC()
		report.DecidedAt = &utc
	}
	return report, nil
}

const communityReportColumns = `
	id, target_type, target_id, reason, note, status, decision,
	decision_reason, decided_at, created_at`

func (s *Store) ListModerationQueue(ctx context.Context, principal core.Principal) ([]core.CommunityReport, error) {
	if err := s.moderatorAuthority(ctx, principal); err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT `+communityReportColumns+` FROM community_reports
		WHERE tenant_id = $1
		ORDER BY CASE status WHEN 'new' THEN 0 ELSE 1 END, created_at DESC
		LIMIT 100
	`, principal.TenantID)
	if err != nil {
		return nil, fmt.Errorf("list moderation queue: %w", err)
	}
	defer rows.Close()
	result := []core.CommunityReport{}
	for rows.Next() {
		report, err := scanCommunityReport(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, report)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate moderation queue: %w", err)
	}
	for index := range result {
		table := "community_posts"
		column := "body"
		if result[index].TargetType == "comment" {
			table = "community_comments"
		}
		var excerpt string
		if err := s.pool.QueryRow(ctx, `
			SELECT left(`+column+`, 200) FROM `+table+`
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, result[index].TargetID).Scan(&excerpt); err == nil {
			result[index].TargetExcerpt = excerpt
		}
	}
	return result, nil
}

func (s *Store) DecideReport(ctx context.Context, command core.DecideReportCommand) (core.CommunityReport, error) {
	principal := command.Principal
	if err := s.moderatorAuthority(ctx, principal); err != nil {
		return core.CommunityReport{}, err
	}
	var report core.CommunityReport
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		claim, err := claimIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "decide_report", command.IdempotencyKey, command.PayloadFingerprint, command.Now)
		if err != nil {
			return err
		}
		if claim.replayed {
			report, err = decodeReplay[core.CommunityReport](claim)
			return err
		}
		var targetType, targetID, status string
		err = tx.QueryRow(ctx, `
			SELECT target_type, target_id, status FROM community_reports
			WHERE tenant_id = $1 AND id = $2 FOR UPDATE
		`, principal.TenantID, command.ReportID).Scan(&targetType, &targetID, &status)
		if errors.Is(err, pgx.ErrNoRows) {
			return core.E(core.CodeNotFound, "report not found", nil)
		}
		if err != nil {
			return fmt.Errorf("lock report: %w", err)
		}
		if status != "new" {
			return core.E(core.CodeInvalidState, "the report is already reviewed", nil)
		}
		if command.Decision == "hidden" {
			table := "community_posts"
			if targetType == "comment" {
				table = "community_comments"
			}
			if _, err := tx.Exec(ctx, `
				UPDATE `+table+`
				SET status = 'hidden', status_reason = $3
				WHERE tenant_id = $1 AND id = $2 AND status = 'published'
			`, principal.TenantID, targetID, command.DecisionReason); err != nil {
				return mapWriteError(err, "content conflicts with existing data")
			}
		}
		if _, err := tx.Exec(ctx, `
			UPDATE community_reports
			SET status = 'reviewed', decision = $3, decision_reason = $4,
			    decided_by_account_id = $5, decided_at = $6
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.ReportID, command.Decision,
			command.DecisionReason, principal.AccountID, command.Now); err != nil {
			return mapWriteError(err, "report conflicts with existing data")
		}
		report, err = scanCommunityReport(tx.QueryRow(ctx, `
			SELECT `+communityReportColumns+` FROM community_reports
			WHERE tenant_id = $1 AND id = $2
		`, principal.TenantID, command.ReportID))
		if err != nil {
			return err
		}
		if err := completeIdempotency(ctx, tx, principal.TenantID, principal.AccountID, "decide_report", command.IdempotencyKey, report, command.Now); err != nil {
			return err
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "CommunityReportDecided", targetType: "community_report",
			targetID: command.ReportID, decision: "allow",
			reason:         command.DecisionReason,
			idempotencyKey: command.IdempotencyKey,
			metadata:       map[string]any{"decision": command.Decision, "targetType": targetType, "targetId": targetID},
			at:             command.Now,
		})
	})
	if err != nil {
		return core.CommunityReport{}, err
	}
	return report, nil
}

func (s *Store) BlockMember(ctx context.Context, command core.BlockMemberCommand) ([]string, error) {
	principal := command.Principal
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := activeAccountExists(ctx, tx, principal.TenantID, principal.AccountID); err != nil {
			return err
		}
		if command.Blocked {
			if _, err := tx.Exec(ctx, `
				INSERT INTO community_blocks (tenant_id, blocker_account_id, blocked_account_id, created_at)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT DO NOTHING
			`, principal.TenantID, principal.AccountID, command.BlockedAccountID, command.Now); err != nil {
				return mapWriteError(err, "block conflicts with existing data")
			}
		} else {
			if _, err := tx.Exec(ctx, `
				DELETE FROM community_blocks
				WHERE tenant_id = $1 AND blocker_account_id = $2 AND blocked_account_id = $3
			`, principal.TenantID, principal.AccountID, command.BlockedAccountID); err != nil {
				return fmt.Errorf("remove block: %w", err)
			}
		}
		action := "CommunityMemberBlocked"
		if !command.Blocked {
			action = "CommunityMemberUnblocked"
		}
		return appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: action, targetType: "community_block", targetID: command.BlockedAccountID,
			decision: "allow", at: command.Now,
		})
	})
	if err != nil {
		return nil, err
	}
	return s.ListBlockedMembers(ctx, principal)
}

func (s *Store) ListBlockedMembers(ctx context.Context, principal core.Principal) ([]string, error) {
	blocked, err := s.blockedSetFor(ctx, principal.TenantID, principal.AccountID)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(blocked))
	for accountID := range blocked {
		result = append(result, accountID)
	}
	sort.Strings(result)
	return result, nil
}
