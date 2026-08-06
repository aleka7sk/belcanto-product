package postgres_test

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/postgres"
	"github.com/aleka7sk/belcanto-product/apps/api/migrations"
)

// TestPostgreSQLCommunity proves the Page 28 safety semantics on real
// SQL: the data-driven guidelines gate, the staff audience filter, the
// author-removal and moderation-hide tombstones (COM-SAFE-05), and the
// triggers that reject deleting community history outright.
func TestPostgreSQLCommunity(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool, _ := isolatedPool(t, ctx, databaseURL)
	if err := migrations.Up(ctx, pool); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
	store := postgres.New(pool)
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x73}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})

	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pgcom", TenantName: "Belcanto PG Community",
		FullName: "PG Community Owner", Phone: "+77001200001",
		Operator: "pg-community-operator", Reason: "community integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, ownerLink), Phone: "+77001200001",
		Password: "Pg-community-owner-1!", IdempotencyKey: "pgcom-activate-owner",
	}); err != nil {
		t.Fatalf("activate Owner: %v", err)
	}
	ownerOutcome, err := service.SignIn(ctx, "+77001200001", "Pg-community-owner-1!", core.SessionClientInfo{})
	if err != nil || ownerOutcome.Tokens == nil {
		t.Fatalf("sign in Owner: %v", err)
	}
	owner, err := service.Authenticate(ctx, ownerOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Owner: %v", err)
	}

	teacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "PG Community Teacher", Phone: "+77001200002", Role: core.RoleTeacher,
		Operator: "pg-community-operator", Reason: "community integration",
	})
	if err != nil {
		t.Fatalf("bootstrap Teacher: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, teacherLink), Phone: "+77001200002",
		Password: "Pg-community-teacher-1!", IdempotencyKey: "pgcom-activate-teacher",
	}); err != nil {
		t.Fatalf("activate Teacher: %v", err)
	}
	teacherOutcome, err := service.SignIn(ctx, "+77001200002", "Pg-community-teacher-1!", core.SessionClientInfo{})
	if err != nil || teacherOutcome.Tokens == nil {
		t.Fatalf("sign in Teacher: %v", err)
	}
	teacher, err := service.Authenticate(ctx, teacherOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Teacher: %v", err)
	}

	created, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
		FullName: "PG Community Student", Phone: "+77001200101", EnrollmentReference: "PGCOM-101",
		TeacherAccountID: teacher.AccountID, Locale: "ru-KZ", Timezone: "Asia/Almaty",
		AdultConfirmed: true, IdempotencyKey: "pgcom-create-student",
	})
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	if _, err := service.PublishFirstMinute(ctx, teacher, app.PublishFirstMinuteInput{
		StudentID: created.StudentID, WhatWorked: "Опора",
		CurrentFocus: "Верх", NextStep: "Легато",
		IdempotencyKey: "pgcom-first-minute",
	}); err != nil {
		t.Fatalf("publish first minute: %v", err)
	}
	_, invitationLink, err := service.IssueInvitation(ctx, owner, created.StudentID, "pgcom-invite", core.InvitationIssue)
	if err != nil {
		t.Fatalf("issue invitation: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, invitationLink), Phone: "+77001200101",
		Password: "Pg-community-student-1!", IdempotencyKey: "pgcom-activate-student",
	}); err != nil {
		t.Fatalf("activate student: %v", err)
	}
	studentOutcome, err := service.SignIn(ctx, "+77001200101", "Pg-community-student-1!", core.SessionClientInfo{})
	if err != nil || studentOutcome.Tokens == nil {
		t.Fatalf("sign in student: %v", err)
	}
	student, err := service.Authenticate(ctx, studentOutcome.Tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate student: %v", err)
	}

	// The guidelines gate is data-driven: a published community policy
	// version blocks writes until the member accepts it.
	if _, err := pool.Exec(ctx, `
		INSERT INTO policy_versions (id, tenant_id, kind, version, title, body_ref, effective_from, created_at)
		VALUES ('polver_pgcom_1', $1, 'community', '1.0', 'Правила сообщества', 'community-1.0', now() - interval '1 hour', now())
	`, owner.TenantID); err != nil {
		t.Fatalf("insert community policy: %v", err)
	}
	if _, err := service.CreateCommunityPost(ctx, student, app.CreatePostInput{
		Body: "До принятия правил", IdempotencyKey: "pgcom-gated",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("post without accepting guidelines = %v, want FORBIDDEN", err)
	}
	for _, principal := range []core.Principal{student, teacher, owner} {
		if err := service.AcceptPolicy(ctx, principal, "polver_pgcom_1"); err != nil {
			t.Fatalf("accept guidelines: %v", err)
		}
	}

	post, err := service.CreateCommunityPost(ctx, student, app.CreatePostInput{
		Body: "Кто идёт на Open Stage?", CommentsEnabled: true,
		IdempotencyKey: "pgcom-post",
	})
	if err != nil || post.Status != core.ContentPublished {
		t.Fatalf("create post = %#v, %v", post, err)
	}
	staffOnly, err := service.CreateCommunityPost(ctx, owner, app.CreatePostInput{
		Body: "Планёрка в пятницу.", Audience: core.AudienceStaff,
		IdempotencyKey: "pgcom-staff-post",
	})
	if err != nil {
		t.Fatalf("staff post: %v", err)
	}
	if _, err := service.CommunityPost(ctx, student, staffOnly.ID); !core.IsCode(err, core.CodeNotFound) {
		t.Fatalf("student opening a staff post = %v, want NOT_FOUND", err)
	}
	feed, err := service.CommunityFeed(ctx, student)
	if err != nil || len(feed) != 1 {
		t.Fatalf("student feed = %d posts, %v", len(feed), err)
	}

	withComment, err := service.AddCommunityComment(ctx, teacher, app.AddCommentInput{
		PostID: post.ID, Body: "Я иду!", IdempotencyKey: "pgcom-comment",
	})
	if err != nil || withComment.CommentCount != 1 {
		t.Fatalf("comment = %#v, %v", withComment, err)
	}
	commentID := withComment.Comments[0].ID

	// Moderation hide → the member sees a tombstone, the moderator the body.
	report, err := service.ReportCommunityContent(ctx, student, app.ReportContentInput{
		TargetType: "comment", TargetID: commentID, Reason: "abuse",
		IdempotencyKey: "pgcom-report",
	})
	if err != nil || report.Status != "new" {
		t.Fatalf("report = %#v, %v", report, err)
	}
	queue, err := service.ModerationQueue(ctx, owner)
	if err != nil || len(queue) != 1 || queue[0].TargetExcerpt != "Я иду!" {
		t.Fatalf("moderation queue = %#v, %v", queue, err)
	}
	if _, err := service.DecideCommunityReport(ctx, owner, app.DecideReportInput{
		ReportID: report.ID, Decision: "hidden",
		DecisionReason: "Нарушение правил сообщества",
		IdempotencyKey: "pgcom-decide",
	}); err != nil {
		t.Fatalf("decide report: %v", err)
	}
	studentView, err := service.CommunityPost(ctx, student, post.ID)
	if err != nil || studentView.CommentCount != 0 || len(studentView.Comments) != 1 ||
		studentView.Comments[0].Body != "" || studentView.Comments[0].Status != core.ContentHidden {
		t.Fatalf("member view after hide = %#v, %v", studentView, err)
	}
	moderatorView, err := service.CommunityPost(ctx, owner, post.ID)
	if err != nil || moderatorView.Comments[0].Body != "Я иду!" {
		t.Fatalf("moderator view after hide = %#v, %v", moderatorView, err)
	}

	// Author removal is a status change with the moment preserved.
	removed, err := service.RemoveCommunityContent(ctx, student, app.RemoveContentInput{
		TargetType: "post", TargetID: post.ID, IdempotencyKey: "pgcom-remove",
	})
	if err != nil || removed.Status != core.ContentRemoved || removed.Body != "" {
		t.Fatalf("removal = %#v, %v", removed, err)
	}
	var statusReason string
	if err := pool.QueryRow(ctx, `
		SELECT status_reason FROM community_posts WHERE id = $1
	`, post.ID).Scan(&statusReason); err != nil || statusReason == "" {
		t.Fatalf("status reason after removal = %q, %v", statusReason, err)
	}

	// Community history refuses deletion outright.
	if _, err := pool.Exec(ctx, `DELETE FROM community_posts WHERE id = $1`, post.ID); err == nil {
		t.Fatal("deleting a community post must be rejected")
	}
	if _, err := pool.Exec(ctx, `DELETE FROM community_comments WHERE id = $1`, commentID); err == nil {
		t.Fatal("deleting a community comment must be rejected")
	}
	if _, err := pool.Exec(ctx, `DELETE FROM community_reports WHERE id = $1`, report.ID); err == nil {
		t.Fatal("deleting a community report must be rejected")
	}

	// Blocks are the one reversible safety row.
	blocked, err := service.BlockCommunityMember(ctx, student, app.BlockMemberInput{
		BlockedAccountID: teacher.AccountID, Blocked: true,
	})
	if err != nil || len(blocked) != 1 {
		t.Fatalf("block = %#v, %v", blocked, err)
	}
	unblocked, err := service.BlockCommunityMember(ctx, student, app.BlockMemberInput{
		BlockedAccountID: teacher.AccountID, Blocked: false,
	})
	if err != nil || len(unblocked) != 0 {
		t.Fatalf("unblock = %#v, %v", unblocked, err)
	}
}
