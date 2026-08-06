package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/notify"
)

// TestCommunityLifecycle pins the Page 28 semantics: the guidelines
// gate is data-driven (no accepted newest community policy → no
// posting), announcements and the staff audience are staff-only,
// removal and moderation hide are status tombstones (COM-SAFE-05),
// a report reaches administrators, and a block filters what the
// blocker sees without deleting anything (COM-SAFE-03).
func TestCommunityLifecycle(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	_, link := readyStudentInvitation(t, fixture, "+77000001301", "ENR-1301")
	const studentPassword = "Community-student-1!"
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, link), Phone: "+77000001301", Password: studentPassword,
		IdempotencyKey: "com-activate",
	}); err != nil {
		t.Fatalf("activate student: %v", err)
	}
	student := signInPrincipal(t, fixture.service, "+77000001301", studentPassword)

	// Without a community policy there is no gate.
	early, err := fixture.service.CreateCommunityPost(ctx, student, app.CreatePostInput{
		Body: "Кто идёт на Open Stage?", CommentsEnabled: true, IdempotencyKey: "com-early",
	})
	if err != nil || early.Status != core.ContentPublished {
		t.Fatalf("post before guidelines exist = %#v, %v", early, err)
	}

	// A published community guidelines version gates every new write.
	fixture.store.SeedPolicyVersionForTest("polver_community_1", "tenant_belcanto",
		"community", "1.0", "Правила сообщества", "community-guidelines-1.0",
		fixture.clock.Now().Add(-time.Hour))
	if _, err := fixture.service.CreateCommunityPost(ctx, student, app.CreatePostInput{
		Body: "Ещё пост", IdempotencyKey: "com-gated",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("post without accepting guidelines = %v, want FORBIDDEN", err)
	}
	for _, principal := range []core.Principal{student, fixture.owner, fixture.teacher, fixture.admin} {
		if err := fixture.service.AcceptPolicy(ctx, principal, "polver_community_1"); err != nil {
			t.Fatalf("accept guidelines: %v", err)
		}
	}

	// Announcements, the staff audience and pinning are staff writes.
	if _, err := fixture.service.CreateCommunityPost(ctx, student, app.CreatePostInput{
		Kind: core.PostKindAnnouncement, Title: "Хочу объявление", Body: "х",
		IdempotencyKey: "com-student-announcement",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("student announcement = %v, want FORBIDDEN", err)
	}
	if _, err := fixture.service.CreateCommunityPost(ctx, student, app.CreatePostInput{
		Body: "х", Audience: core.AudienceStaff, IdempotencyKey: "com-student-staff",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("student staff-audience post = %v, want FORBIDDEN", err)
	}
	announcement, err := fixture.service.CreateCommunityPost(ctx, fixture.owner, app.CreatePostInput{
		Kind: core.PostKindAnnouncement, Title: "Отчётный концерт",
		Body: "Репетиции по расписанию в зале.", Pinned: true,
		IdempotencyKey: "com-announcement",
	})
	if err != nil || !announcement.Pinned || announcement.CommentsEnabled {
		t.Fatalf("announcement = %#v, %v", announcement, err)
	}
	if _, err := fixture.service.AddCommunityComment(ctx, student, app.AddCommentInput{
		PostID: announcement.ID, Body: "Ура!", IdempotencyKey: "com-comment-disabled",
	}); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("comment on a closed announcement = %v, want INVALID_STATE", err)
	}
	staffOnly, err := fixture.service.CreateCommunityPost(ctx, fixture.admin, app.CreatePostInput{
		Body: "Планёрка в пятницу.", Audience: core.AudienceStaff,
		IdempotencyKey: "com-staff-post",
	})
	if err != nil {
		t.Fatalf("staff post: %v", err)
	}

	// The student feed hides the staff audience and pins the announcement.
	feed, err := fixture.service.CommunityFeed(ctx, student)
	if err != nil || len(feed) != 2 {
		t.Fatalf("student feed = %d posts, %v", len(feed), err)
	}
	if feed[0].ID != announcement.ID {
		t.Fatalf("pinned announcement is not first: %#v", feed)
	}
	if _, err := fixture.service.CommunityPost(ctx, student, staffOnly.ID); !core.IsCode(err, core.CodeNotFound) {
		t.Fatalf("student opening a staff post = %v, want NOT_FOUND", err)
	}

	// Comments: replay is idempotent, the count reflects published ones.
	withComment, err := fixture.service.AddCommunityComment(ctx, fixture.teacher, app.AddCommentInput{
		PostID: early.ID, Body: "Я иду!", IdempotencyKey: "com-comment",
	})
	if err != nil || withComment.CommentCount != 1 {
		t.Fatalf("comment = %#v, %v", withComment, err)
	}
	replayed, err := fixture.service.AddCommunityComment(ctx, fixture.teacher, app.AddCommentInput{
		PostID: early.ID, Body: "Я иду!", IdempotencyKey: "com-comment",
	})
	if err != nil || replayed.CommentCount != 1 {
		t.Fatalf("comment replay = %#v, %v", replayed, err)
	}
	teacherCommentID := withComment.Comments[0].ID

	// A report with reason «другое» carries a note; filing reaches admins.
	if _, err := fixture.service.ReportCommunityContent(ctx, student, app.ReportContentInput{
		TargetType: "comment", TargetID: teacherCommentID, Reason: "other",
		IdempotencyKey: "com-report-no-note",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("report other without note = %v, want INVALID_INPUT", err)
	}
	report, err := fixture.service.ReportCommunityContent(ctx, student, app.ReportContentInput{
		TargetType: "comment", TargetID: teacherCommentID, Reason: "abuse",
		IdempotencyKey: "com-report",
	})
	if err != nil || report.Status != "new" {
		t.Fatalf("report = %#v, %v", report, err)
	}
	worker := notify.NewWorker(fixture.store, notify.Options{
		Clock: func() time.Time { return fixture.clock.Now() },
	})
	if _, failed, err := worker.DrainOnce(ctx); err != nil || failed != 0 {
		t.Fatalf("drain outbox: %d failed, %v", failed, err)
	}
	adminFeed, err := fixture.service.ActivityFeed(ctx, fixture.admin)
	if err != nil {
		t.Fatalf("admin feed: %v", err)
	}
	sawReport := false
	for _, entry := range adminFeed.Entries {
		if entry.Kind == "CommunityReportFiled" && entry.Category == "important" {
			sawReport = true
		}
	}
	if !sawReport {
		t.Fatalf("report missing from the administrator feed: %#v", adminFeed.Entries)
	}

	// Moderation is Owner/Administrator; the decision hides with a reason.
	if _, err := fixture.service.ModerationQueue(ctx, fixture.teacher); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("teacher moderation queue = %v, want FORBIDDEN", err)
	}
	queue, err := fixture.service.ModerationQueue(ctx, fixture.admin)
	if err != nil || len(queue) != 1 || queue[0].TargetExcerpt == "" {
		t.Fatalf("moderation queue = %#v, %v", queue, err)
	}
	decided, err := fixture.service.DecideCommunityReport(ctx, fixture.admin, app.DecideReportInput{
		ReportID: report.ID, Decision: "hidden",
		DecisionReason: "Нарушение правил сообщества",
		IdempotencyKey: "com-decide",
	})
	if err != nil || decided.Status != "reviewed" || decided.Decision != "hidden" {
		t.Fatalf("decision = %#v, %v", decided, err)
	}
	if _, err := fixture.service.DecideCommunityReport(ctx, fixture.admin, app.DecideReportInput{
		ReportID: report.ID, Decision: "kept", DecisionReason: "повтор",
		IdempotencyKey: "com-decide-again",
	}); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("second decision = %v, want INVALID_STATE", err)
	}

	// The hidden comment is a tombstone for members, visible to moderators.
	studentView, err := fixture.service.CommunityPost(ctx, student, early.ID)
	if err != nil || studentView.CommentCount != 0 || len(studentView.Comments) != 1 {
		t.Fatalf("student view after hide = %#v, %v", studentView, err)
	}
	if studentView.Comments[0].Body != "" || studentView.Comments[0].Status != core.ContentHidden {
		t.Fatalf("hidden comment is not a tombstone: %#v", studentView.Comments[0])
	}
	adminView, err := fixture.service.CommunityPost(ctx, fixture.admin, early.ID)
	if err != nil || adminView.Comments[0].Body == "" {
		t.Fatalf("moderator lost the hidden body: %#v, %v", adminView, err)
	}

	// Removal is the author's status change, never someone else's delete.
	if _, err := fixture.service.RemoveCommunityContent(ctx, fixture.teacher, app.RemoveContentInput{
		TargetType: "post", TargetID: early.ID, IdempotencyKey: "com-remove-foreign",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("foreign removal = %v, want FORBIDDEN", err)
	}
	removed, err := fixture.service.RemoveCommunityContent(ctx, student, app.RemoveContentInput{
		TargetType: "post", TargetID: early.ID, IdempotencyKey: "com-remove",
	})
	if err != nil || removed.Status != core.ContentRemoved {
		t.Fatalf("removal = %#v, %v", removed, err)
	}
	tombstone, err := fixture.service.CommunityPost(ctx, student, early.ID)
	if err != nil || tombstone.Body != "" || tombstone.CreatedAt.IsZero() {
		t.Fatalf("removed post tombstone = %#v, %v", tombstone, err)
	}
	feed, err = fixture.service.CommunityFeed(ctx, student)
	if err != nil || len(feed) != 1 {
		t.Fatalf("feed after removal = %d posts, %v", len(feed), err)
	}

	// A block filters the blocker's view and reverses cleanly.
	visible, err := fixture.service.CreateCommunityPost(ctx, fixture.teacher, app.CreatePostInput{
		Body: "Разбор занятия в четверг.", CommentsEnabled: true, IdempotencyKey: "com-teacher-post",
	})
	if err != nil {
		t.Fatalf("teacher post: %v", err)
	}
	if _, err := fixture.service.AddCommunityComment(ctx, fixture.teacher, app.AddCommentInput{
		PostID: visible.ID, Body: "Жду вопросы.", IdempotencyKey: "com-teacher-comment",
	}); err != nil {
		t.Fatalf("teacher comment: %v", err)
	}
	if _, err := fixture.service.BlockCommunityMember(ctx, student, app.BlockMemberInput{
		BlockedAccountID: student.AccountID, Blocked: true,
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("self-block = %v, want INVALID_INPUT", err)
	}
	blocked, err := fixture.service.BlockCommunityMember(ctx, student, app.BlockMemberInput{
		BlockedAccountID: fixture.teacher.AccountID, Blocked: true,
	})
	if err != nil || len(blocked) != 1 {
		t.Fatalf("block = %#v, %v", blocked, err)
	}
	blockedView, err := fixture.service.CommunityPost(ctx, student, visible.ID)
	if err != nil || len(blockedView.Comments) != 0 {
		t.Fatalf("blocked author's comment still visible: %#v, %v", blockedView, err)
	}
	unblocked, err := fixture.service.BlockCommunityMember(ctx, student, app.BlockMemberInput{
		BlockedAccountID: fixture.teacher.AccountID, Blocked: false,
	})
	if err != nil || len(unblocked) != 0 {
		t.Fatalf("unblock = %#v, %v", unblocked, err)
	}
	restoredView, err := fixture.service.CommunityPost(ctx, student, visible.ID)
	if err != nil || len(restoredView.Comments) != 1 {
		t.Fatalf("comments after unblock = %#v, %v", restoredView, err)
	}
}
