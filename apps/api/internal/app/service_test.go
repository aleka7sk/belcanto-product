package app_test

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/memory"
)

const (
	ownerPassword   = "Owner-password-123!"
	adminPassword   = "Admin-password-123!"
	teacherPassword = "Teacher-password-123!"
	studentPassword = "Student-password-123!"
)

type testClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *testClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *testClock) Advance(duration time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(duration)
	c.mu.Unlock()
}

type fixture struct {
	service *app.Service
	store   *memory.Store
	codec   *security.TokenCodec
	hasher  *security.PasswordHasher
	clock   *testClock
	owner   core.Principal
	admin   core.Principal
	teacher core.Principal
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	clock := &testClock{now: time.Date(2026, time.August, 1, 10, 0, 0, 0, time.UTC)}
	store := memory.New()
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x42}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	hasher := security.NewPasswordHasher()
	service := app.NewService(store, codec, hasher, app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour, Clock: clock,
	})
	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_belcanto", TenantName: "Belcanto",
		FullName: "Belcanto Owner", Phone: "+77000000001",
		Operator: "test-operator", Reason: "test fixture bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, ownerLink), Phone: "+77000000001",
		Password: ownerPassword, IdempotencyKey: "activate-owner",
	}); err != nil {
		t.Fatalf("activate Owner: %v", err)
	}
	ownerTokens := mustSignInTokens(t, service, "+77000000001", ownerPassword)
	owner, err := service.Authenticate(ctx, ownerTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Owner: %v", err)
	}

	adminLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: "tenant_belcanto", OwnerAccountID: owner.AccountID,
		FullName: "Belcanto Administrator", Phone: "+77000000002", Role: core.RoleAdministrator,
		Operator: "test-operator", Reason: "test fixture staff bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap Administrator: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, adminLink), Phone: "+77000000002",
		Password: adminPassword, IdempotencyKey: "activate-admin",
	}); err != nil {
		t.Fatalf("activate Administrator: %v", err)
	}
	teacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: "tenant_belcanto", OwnerAccountID: owner.AccountID,
		FullName: "Belcanto Teacher", Phone: "+77000000003", Role: core.RoleTeacher,
		Operator: "test-operator", Reason: "test fixture staff bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap Teacher: %v", err)
	}
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, teacherLink), Phone: "+77000000003",
		Password: teacherPassword, IdempotencyKey: "activate-teacher",
	}); err != nil {
		t.Fatalf("activate Teacher: %v", err)
	}
	adminTokens := mustSignInTokens(t, service, "+77000000002", adminPassword)
	admin, err := service.Authenticate(ctx, adminTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Administrator: %v", err)
	}
	teacherTokens := mustSignInTokens(t, service, "+77000000003", teacherPassword)
	teacher, err := service.Authenticate(ctx, teacherTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Teacher: %v", err)
	}
	return &fixture{service: service, store: store, codec: codec, hasher: hasher, clock: clock, owner: owner, admin: admin, teacher: teacher}
}

func TestClosedInvitationJourneyAndEffectiveAccess(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	ordinaryView, err := fixture.service.BootstrapView(ctx, fixture.admin)
	if err != nil {
		t.Fatalf("ordinary Administrator bootstrap: %v", err)
	}
	assertStrings(t, ordinaryView.AccessProfiles, []string{})
	assertStrings(t, ordinaryView.Permissions, core.LessonPermissionSetForRoles([]core.Role{core.RoleAdministrator}))
	_, err = fixture.service.CreateStudent(ctx, fixture.admin, studentInput("student-before-grant", "+77000000100", "ENR-100", fixture.teacher.AccountID))
	assertCode(t, err, core.CodeForbidden)

	ownerView, err := fixture.service.BootstrapView(ctx, fixture.owner)
	if err != nil {
		t.Fatalf("Owner bootstrap: %v", err)
	}
	assertStrings(t, ownerView.Permissions, append(core.LessonPermissionSetForRoles([]core.Role{core.RoleOwner}), core.OwnerStudentOnboardingPermissionSet()...))
	assertStrings(t, ownerView.AccessProfiles, []string{})

	delegation, err := fixture.service.GrantDelegation(ctx, fixture.owner, app.GrantDelegationInput{
		AdministratorAccountID: fixture.admin.AccountID, Reason: "Owner-approved student onboarding",
		CurrentPassword: ownerPassword, IdempotencyKey: "grant-admin-1",
	})
	if err != nil {
		t.Fatalf("grant Administrator profile: %v", err)
	}
	if delegation.Bundle != core.StudentOnboardingManagerV1 {
		t.Fatalf("delegation bundle = %q", delegation.Bundle)
	}
	delegationReplay, err := fixture.service.GrantDelegation(ctx, fixture.owner, app.GrantDelegationInput{
		AdministratorAccountID: fixture.admin.AccountID, Reason: "Owner-approved student onboarding",
		CurrentPassword: ownerPassword, IdempotencyKey: "grant-admin-1",
	})
	if err != nil || delegationReplay.ID != delegation.ID {
		t.Fatalf("idempotent delegation replay = %#v, %v", delegationReplay, err)
	}

	delegatedView, err := fixture.service.BootstrapView(ctx, fixture.admin)
	if err != nil {
		t.Fatalf("delegated Administrator bootstrap: %v", err)
	}
	assertStrings(t, delegatedView.AccessProfiles, []string{core.StudentOnboardingManagerV1})
	assertStrings(t, delegatedView.Permissions, append(core.LessonPermissionSetForRoles([]core.Role{core.RoleAdministrator}), core.StudentOnboardingManagerV1PermissionSet()...))

	created, err := fixture.service.CreateStudent(ctx, fixture.admin, studentInput("create-student-1", "+77000000101", "ENR-101", fixture.teacher.AccountID))
	if err != nil {
		t.Fatalf("delegated Administrator creates Student: %v", err)
	}
	replayed, err := fixture.service.CreateStudent(ctx, fixture.admin, studentInput("create-student-1", "+77000000101", "ENR-101", fixture.teacher.AccountID))
	if err != nil || replayed != created {
		t.Fatalf("idempotent Student replay = %#v, %v; want %#v", replayed, err, created)
	}
	different := studentInput("create-student-1", "+77000000101", "ENR-101", fixture.teacher.AccountID)
	different.FullName = "Different payload"
	_, err = fixture.service.CreateStudent(ctx, fixture.admin, different)
	assertCode(t, err, core.CodeConflict)

	firstMinute, err := fixture.service.PublishFirstMinute(ctx, fixture.teacher, app.PublishFirstMinuteInput{
		StudentID: created.StudentID, WhatWorked: "Свободная атака звука",
		CurrentFocus: "Опора в длинной фразе", NextStep: "Три коротких упражнения",
		ExpectedVersion: 0, IdempotencyKey: "first-minute-101",
	})
	if err != nil {
		t.Fatalf("publish First Belcanto Minute: %v", err)
	}
	if firstMinute.Revision != 1 {
		t.Fatalf("first-minute revision = %d, want 1", firstMinute.Revision)
	}

	_, _, err = fixture.service.IssueInvitation(ctx, fixture.admin, created.StudentID, "admin-invite-101", core.InvitationIssue)
	assertCode(t, err, core.CodeForbidden)

	invitation, link, err := fixture.service.IssueInvitation(ctx, fixture.owner, created.StudentID, "invite-101", core.InvitationIssue)
	if err != nil {
		t.Fatalf("issue Student invitation: %v", err)
	}
	replayedInvitation, replayedLink, err := fixture.service.IssueInvitation(ctx, fixture.owner, created.StudentID, "invite-101", core.InvitationIssue)
	if err != nil || replayedInvitation != invitation || replayedLink != link {
		t.Fatalf("idempotent invitation replay mismatch: %#v %q %v", replayedInvitation, replayedLink, err)
	}
	token := tokenFromLink(t, link)
	preview, err := fixture.service.PreviewActivation(ctx, token)
	if err != nil {
		t.Fatalf("preview Student activation: %v", err)
	}
	if preview.DisplayName != "Adult Student" || preview.MaskedPhone == "+77000000101" {
		t.Fatalf("unexpected activation preview: %#v", preview)
	}
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: token, Phone: "+77000000101", Password: studentPassword,
		IdempotencyKey: "activate-student-101",
	}); err != nil {
		t.Fatalf("activate Student: %v", err)
	}
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: token, Phone: "+77000000101", Password: studentPassword,
		IdempotencyKey: "activate-student-101",
	}); err != nil {
		t.Fatalf("idempotent activation replay: %v", err)
	}
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: token, Phone: "+77000000101", Password: "Different-password-123!",
		IdempotencyKey: "activate-student-101",
	}); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("activation key reused with different password = %v", err)
	}
	studentTokens := mustSignInTokens(t, fixture.service, "+77000000101", studentPassword)
	studentPrincipal, err := fixture.service.Authenticate(ctx, studentTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate Student: %v", err)
	}
	studentView, err := fixture.service.BootstrapView(ctx, studentPrincipal)
	if err != nil {
		t.Fatalf("Student bootstrap: %v", err)
	}
	if studentView.StudentID != created.StudentID || studentView.FirstMinute == nil || studentView.FirstMinute.Revision != 1 {
		t.Fatalf("Student bootstrap omitted identity or First Minute: %#v", studentView)
	}

	if err := fixture.service.RevokeDelegation(ctx, fixture.owner, app.RevokeDelegationInput{
		DelegationID: delegation.ID, Reason: "Operational handback",
		CurrentPassword: ownerPassword, IdempotencyKey: "revoke-admin-1",
	}); err != nil {
		t.Fatalf("revoke delegation: %v", err)
	}
	revokedView, err := fixture.service.BootstrapView(ctx, fixture.admin)
	if err != nil {
		t.Fatalf("Administrator bootstrap after revoke: %v", err)
	}
	assertStrings(t, revokedView.AccessProfiles, []string{})
	assertStrings(t, revokedView.Permissions, core.LessonPermissionSetForRoles([]core.Role{core.RoleAdministrator}))
	_, err = fixture.service.CreateStudent(ctx, fixture.admin, studentInput("student-after-revoke", "+77000000102", "ENR-102", fixture.teacher.AccountID))
	assertCode(t, err, core.CodeForbidden)
}

func TestInvitationReissueInvalidatesOldLink(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	grantAdministrator(t, fixture, nil)
	student, err := fixture.service.CreateStudent(ctx, fixture.admin, studentInput("create-reissue", "+77000000201", "ENR-201", fixture.teacher.AccountID))
	if err != nil {
		t.Fatalf("create Student: %v", err)
	}
	if _, err := fixture.service.PublishFirstMinute(ctx, fixture.teacher, app.PublishFirstMinuteInput{
		StudentID: student.StudentID, WhatWorked: "A", CurrentFocus: "B", NextStep: "C",
		ExpectedVersion: 0, IdempotencyKey: "first-minute-201",
	}); err != nil {
		t.Fatalf("publish First Minute: %v", err)
	}
	oldInvitation, oldLink, err := fixture.service.IssueInvitation(ctx, fixture.owner, student.StudentID, "issue-201", core.InvitationIssue)
	if err != nil {
		t.Fatalf("issue invitation: %v", err)
	}
	if _, _, err := fixture.service.IssueInvitation(ctx, fixture.admin, student.StudentID, "admin-reissue-201", core.InvitationReissue); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("delegated Administrator reissues invitation = %v", err)
	}
	newInvitation, newLink, err := fixture.service.IssueInvitation(ctx, fixture.owner, student.StudentID, "reissue-201", core.InvitationReissue)
	if err != nil {
		t.Fatalf("reissue invitation: %v", err)
	}
	if oldInvitation.InvitationID == newInvitation.InvitationID || oldLink == newLink {
		t.Fatal("reissue did not create a distinct invitation")
	}
	_, err = fixture.service.PreviewActivation(ctx, tokenFromLink(t, oldLink))
	assertCode(t, err, core.CodeInvalidActivation)
	if _, err := fixture.service.PreviewActivation(ctx, tokenFromLink(t, newLink)); err != nil {
		t.Fatalf("new invitation is invalid: %v", err)
	}
}

func TestInvitationReissueRequiresActiveInvitation(t *testing.T) {
	fixture := newFixture(t)
	student, _ := readyStudentWithoutInvitation(t, fixture, "+77000000211", "ENR-211")
	_, _, err := fixture.service.IssueInvitation(context.Background(), fixture.owner, student.StudentID, "reissue-without-active", core.InvitationReissue)
	assertCode(t, err, core.CodeInvalidState)
	invitation, _, err := fixture.service.IssueInvitation(context.Background(), fixture.owner, student.StudentID, "issue-before-revoke", core.InvitationIssue)
	if err != nil {
		t.Fatalf("issue invitation before revoke: %v", err)
	}
	if err := fixture.service.RevokeInvitation(context.Background(), fixture.admin, invitation.InvitationID, "admin-revoke-before-reissue"); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("delegated Administrator revokes invitation = %v", err)
	}
	if err := fixture.service.RevokeInvitation(context.Background(), fixture.owner, invitation.InvitationID, "revoke-before-reissue"); err != nil {
		t.Fatalf("revoke invitation: %v", err)
	}
	_, _, err = fixture.service.IssueInvitation(context.Background(), fixture.owner, student.StudentID, "reissue-after-revoke", core.InvitationReissue)
	assertCode(t, err, core.CodeInvalidState)
	if _, _, err := fixture.service.IssueInvitation(context.Background(), fixture.owner, student.StudentID, "issue-after-revoke", core.InvitationIssue); err != nil {
		t.Fatalf("ordinary issue after revoke: %v", err)
	}
}

func TestConcurrentActivationConsumesInvitationOnce(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	student, link := readyStudentInvitation(t, fixture, "+77000000301", "ENR-301")
	_ = student
	token := tokenFromLink(t, link)

	const competitors = 4
	var successes atomic.Int32
	var invalid atomic.Int32
	var wait sync.WaitGroup
	wait.Add(competitors)
	for index := 0; index < competitors; index++ {
		go func(index int) {
			defer wait.Done()
			err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
				Token: token, Phone: "+77000000301", Password: studentPassword,
				IdempotencyKey: "activation-competitor-" + string(rune('a'+index)),
			})
			switch {
			case err == nil:
				successes.Add(1)
			case core.IsCode(err, core.CodeInvalidActivation):
				invalid.Add(1)
			default:
				t.Errorf("unexpected activation error: %v", err)
			}
		}(index)
	}
	wait.Wait()
	if successes.Load() != 1 || invalid.Load() != competitors-1 {
		t.Fatalf("activation outcomes: success=%d invalid=%d", successes.Load(), invalid.Load())
	}
}

func TestDelegationExpiryRemovesEffectiveAccess(t *testing.T) {
	fixture := newFixture(t)
	expiresAt := fixture.clock.Now().Add(time.Hour)
	grantAdministrator(t, fixture, &expiresAt)
	view, err := fixture.service.BootstrapView(context.Background(), fixture.admin)
	if err != nil || len(view.Permissions) != 6 {
		t.Fatalf("active delegated view: %#v, %v", view, err)
	}
	fixture.clock.Advance(2 * time.Hour)
	view, err = fixture.service.BootstrapView(context.Background(), fixture.admin)
	if err != nil {
		t.Fatalf("expired delegated view: %v", err)
	}
	assertStrings(t, view.AccessProfiles, []string{})
	assertStrings(t, view.Permissions, core.LessonPermissionSetForRoles([]core.Role{core.RoleAdministrator}))
	_, err = fixture.service.CreateStudent(context.Background(), fixture.admin, studentInput("expired-grant", "+77000000401", "ENR-401", fixture.teacher.AccountID))
	assertCode(t, err, core.CodeForbidden)
}

func TestMutationReplayReauthorizesActorAndScopesIdempotency(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	delegation := grantAdministrator(t, fixture, nil)
	adminInput := studentInput("actor-shared-key", "+77000000411", "ACTOR-411", fixture.teacher.AccountID)
	adminStudent, err := fixture.service.CreateStudent(ctx, fixture.admin, adminInput)
	if err != nil {
		t.Fatalf("delegated Administrator creates Student: %v", err)
	}
	if err := fixture.service.RevokeDelegation(ctx, fixture.owner, app.RevokeDelegationInput{
		DelegationID: delegation.ID, Reason: "Replay authorization test",
		CurrentPassword: ownerPassword, IdempotencyKey: "actor-revoke-delegation",
	}); err != nil {
		t.Fatalf("revoke Administrator delegation: %v", err)
	}
	if _, err := fixture.service.CreateStudent(ctx, fixture.admin, adminInput); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("revoked Administrator exact replay = %v", err)
	}
	ownerInput := studentInput("actor-shared-key", "+77000000412", "ACTOR-412", fixture.teacher.AccountID)
	ownerStudent, err := fixture.service.CreateStudent(ctx, fixture.owner, ownerInput)
	if err != nil || ownerStudent.StudentID == adminStudent.StudentID {
		t.Fatalf("Owner actor-scoped key result = %#v, %v", ownerStudent, err)
	}

	firstInput := app.PublishFirstMinuteInput{
		StudentID: ownerStudent.StudentID, WhatWorked: "Actor worked",
		CurrentFocus: "Actor focus", NextStep: "Actor next",
		ExpectedVersion: 0, IdempotencyKey: "teacher-shared-key",
	}
	if _, err := fixture.service.PublishFirstMinute(ctx, fixture.teacher, firstInput); err != nil {
		t.Fatalf("assigned Teacher publishes first revision: %v", err)
	}
	otherTeacherHash, err := security.NewPasswordHasher().Hash(teacherPassword)
	if err != nil {
		t.Fatalf("hash other Teacher password: %v", err)
	}
	const otherTeacherID = "acct_actor_other_teacher"
	if err := fixture.store.SeedActiveStaff(fixture.owner.TenantID, otherTeacherID, "person_actor_other_teacher", "+77000000413", otherTeacherHash, core.RoleTeacher); err != nil {
		t.Fatalf("seed other Teacher: %v", err)
	}
	if err := fixture.store.SetAssignedTeacherForTest(fixture.owner.TenantID, ownerStudent.StudentID, otherTeacherID); err != nil {
		t.Fatalf("change Teacher assignment: %v", err)
	}
	if _, err := fixture.service.PublishFirstMinute(ctx, fixture.teacher, firstInput); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("removed Teacher exact replay = %v", err)
	}
	second, err := fixture.service.PublishFirstMinute(ctx, core.Principal{
		AccountID: otherTeacherID, TenantID: fixture.owner.TenantID, Roles: []core.Role{core.RoleTeacher},
	}, app.PublishFirstMinuteInput{
		StudentID: ownerStudent.StudentID, WhatWorked: "New Teacher worked",
		CurrentFocus: "New Teacher focus", NextStep: "New Teacher next",
		ExpectedVersion: 1, IdempotencyKey: "teacher-shared-key",
	})
	if err != nil || second.Revision != 2 {
		t.Fatalf("new Teacher actor-scoped key result = %#v, %v", second, err)
	}
}

func TestRefreshRotationDetectsReuseAndRevokesFamily(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	original := mustSignInTokens(t, fixture.service, "+77000000001", ownerPassword)
	rotated, err := fixture.service.Refresh(ctx, original.RefreshToken)
	if err != nil {
		t.Fatalf("rotate refresh token: %v", err)
	}
	if _, err := fixture.service.Authenticate(ctx, rotated.AccessToken); err != nil {
		t.Fatalf("rotated access token is inactive: %v", err)
	}
	if _, err := fixture.service.Refresh(ctx, original.RefreshToken); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("reused refresh token = %v", err)
	}
	if _, err := fixture.service.Authenticate(ctx, rotated.AccessToken); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("refresh family remained active after reuse: %v", err)
	}

	fresh := mustSignInTokens(t, fixture.service, "+77000000001", ownerPassword)
	replacement, err := fixture.service.Refresh(ctx, fresh.RefreshToken)
	if err != nil {
		t.Fatalf("refresh before family logout: %v", err)
	}
	if err := fixture.service.SignOut(ctx, replacement.AccessToken); err != nil {
		t.Fatalf("sign out: %v", err)
	}
	if _, err := fixture.service.Authenticate(ctx, replacement.AccessToken); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("signed-out access token remained active: %v", err)
	}
	if _, err := fixture.service.Refresh(ctx, replacement.RefreshToken); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("signed-out refresh token remained active: %v", err)
	}
}

func TestConcurrentRefreshAndLogoutCannotLeaveFamilyActive(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	for iteration := 0; iteration < 12; iteration++ {
		originalOutcome, err := fixture.service.SignIn(ctx, "+77000000001", ownerPassword, core.SessionClientInfo{})
		if err != nil {
			t.Fatalf("sign in iteration %d: %v", iteration, err)
		}
		if originalOutcome.Tokens == nil {
			t.Fatalf("sign in iteration %d returned a second-factor challenge", iteration)
		}
		original := *originalOutcome.Tokens
		start := make(chan struct{})
		var replacement core.SessionTokens
		var refreshErr error
		var logoutErr error
		var wait sync.WaitGroup
		wait.Add(2)
		go func() {
			defer wait.Done()
			<-start
			replacement, refreshErr = fixture.service.Refresh(ctx, original.RefreshToken)
		}()
		go func() {
			defer wait.Done()
			<-start
			logoutErr = fixture.service.SignOut(ctx, original.AccessToken)
		}()
		close(start)
		wait.Wait()
		if logoutErr != nil {
			t.Fatalf("logout iteration %d: %v", iteration, logoutErr)
		}
		if refreshErr != nil && !core.IsCode(refreshErr, core.CodeUnauthenticated) {
			t.Fatalf("refresh iteration %d: %v", iteration, refreshErr)
		}
		if refreshErr == nil {
			if _, err := fixture.service.Authenticate(ctx, replacement.AccessToken); !core.IsCode(err, core.CodeUnauthenticated) {
				t.Fatalf("race iteration %d left replacement access active: %v", iteration, err)
			}
			if _, err := fixture.service.Refresh(ctx, replacement.RefreshToken); !core.IsCode(err, core.CodeUnauthenticated) {
				t.Fatalf("race iteration %d left replacement refresh active: %v", iteration, err)
			}
		}
	}
}

func TestBootstrapInvitationRecoveryPreservesIdentityAndRoles(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	ownerBootstrap, err := fixture.service.BootstrapOwnerWithAccount(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_recovery", TenantName: "Recovery School",
		FullName: "Recovery Owner", Phone: "+77000000901",
		Operator: "bootstrap-operator", Reason: "initial recovery test bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap pending recovery Owner: %v", err)
	}
	oldOwnerLink := ownerBootstrap.ActivationLink
	ownerAccountID := ownerBootstrap.AccountID
	if ownerAccountID == "" {
		t.Fatal("Owner bootstrap result omitted non-secret AccountID")
	}
	ownerResult, newOwnerLink, err := fixture.service.ReissueBootstrapInvitation(ctx, app.ReissueBootstrapInvitationInput{
		TenantID: "tenant_recovery", AccountID: ownerAccountID,
		Operator: "incident-operator", Reason: "original Owner link was lost",
	})
	if err != nil {
		t.Fatalf("reissue pending Owner invitation: %v", err)
	}
	if ownerResult.AccountID != ownerAccountID || ownerResult.Kind != "owner_bootstrap" || oldOwnerLink == newOwnerLink {
		t.Fatalf("Owner recovery result = %#v, old=%q new=%q", ownerResult, oldOwnerLink, newOwnerLink)
	}
	if _, err := fixture.service.PreviewActivation(ctx, tokenFromLink(t, oldOwnerLink)); !core.IsCode(err, core.CodeInvalidActivation) {
		t.Fatalf("superseded Owner invitation preview = %v", err)
	}
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, newOwnerLink), Phone: "+77000000901",
		Password: ownerPassword, IdempotencyKey: "activate-recovered-owner",
	}); err != nil {
		t.Fatalf("activate recovered Owner: %v", err)
	}
	recoveredOwner := signInPrincipal(t, fixture.service, "+77000000901", ownerPassword)
	assertRoles(t, recoveredOwner.Roles, []core.Role{core.RoleOwner})
	if _, _, err := fixture.service.ReissueBootstrapInvitation(ctx, app.ReissueBootstrapInvitationInput{
		TenantID: "tenant_recovery", AccountID: ownerAccountID,
		Operator: "incident-operator", Reason: "must reject active account",
	}); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("active Owner recovery = %v", err)
	}

	staffBootstrap, err := fixture.service.BootstrapStaffWithAccount(ctx, app.BootstrapStaffInput{
		TenantID: fixture.owner.TenantID, OwnerAccountID: fixture.owner.AccountID,
		FullName: "Recovery Teacher", Phone: "+77000000902", Role: core.RoleTeacher,
		Operator: "bootstrap-operator", Reason: "initial recovery staff bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap pending recovery Teacher: %v", err)
	}
	oldStaffLink := staffBootstrap.ActivationLink
	staffAccountID := staffBootstrap.AccountID
	if staffAccountID == "" {
		t.Fatal("staff bootstrap result omitted non-secret AccountID")
	}
	staffResult, newStaffLink, err := fixture.service.ReissueBootstrapInvitation(ctx, app.ReissueBootstrapInvitationInput{
		TenantID: fixture.owner.TenantID, AccountID: staffAccountID,
		Operator: "incident-operator", Reason: "original Teacher link was lost",
	})
	if err != nil {
		t.Fatalf("reissue pending Teacher invitation: %v", err)
	}
	if staffResult.AccountID != staffAccountID || staffResult.Kind != "staff_activation" || oldStaffLink == newStaffLink {
		t.Fatalf("staff recovery result = %#v", staffResult)
	}
	if _, err := fixture.service.PreviewActivation(ctx, tokenFromLink(t, oldStaffLink)); !core.IsCode(err, core.CodeInvalidActivation) {
		t.Fatalf("superseded staff invitation preview = %v", err)
	}
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, newStaffLink), Phone: "+77000000902",
		Password: teacherPassword, IdempotencyKey: "activate-recovered-teacher",
	}); err != nil {
		t.Fatalf("activate recovered Teacher: %v", err)
	}
	recoveredTeacher := signInPrincipal(t, fixture.service, "+77000000902", teacherPassword)
	assertRoles(t, recoveredTeacher.Roles, []core.Role{core.RoleTeacher})

	audit := fixture.store.AuditRecords()
	foundRecoveryAudit := false
	for _, record := range audit {
		if record.Action == "BootstrapActivationInvitationReissued" && record.TargetID == staffResult.InvitationID {
			foundRecoveryAudit = record.OperatorID == "incident-operator" && record.Reason == "original Teacher link was lost" && record.ActorID == ""
		}
	}
	if !foundRecoveryAudit {
		t.Fatal("bootstrap recovery audit omitted explicit operator/reason attribution")
	}
	if _, _, err := fixture.service.ReissueBootstrapInvitation(ctx, app.ReissueBootstrapInvitationInput{
		TenantID: fixture.owner.TenantID, AccountID: staffAccountID,
		Operator: "", Reason: "missing operator",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("recovery without operator = %v", err)
	}
}

func TestOpaqueTokensRequireExactBase64URLAlphabet(t *testing.T) {
	fixture := newFixture(t)
	malformed := strings.Repeat("!", 43)
	if _, err := fixture.service.PreviewActivation(context.Background(), malformed); !core.IsCode(err, core.CodeInvalidActivation) {
		t.Fatalf("malformed activation token = %v", err)
	}
	if _, err := fixture.service.Refresh(context.Background(), malformed); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("malformed refresh token = %v", err)
	}
	if _, err := fixture.service.Authenticate(context.Background(), malformed); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("malformed access token = %v", err)
	}
}

func TestDuplicatePhoneConflictsUseOneNeutralMessage(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	grantAdministrator(t, fixture, nil)

	_, _, ownerErr := fixture.service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_duplicate_phone", TenantName: "Duplicate Phone",
		FullName: "Duplicate Owner", Phone: "+77000000001",
		Operator: "test-operator", Reason: "duplicate phone test",
	})
	_, _, staffErr := fixture.service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: fixture.owner.TenantID, OwnerAccountID: fixture.owner.AccountID,
		FullName: "Duplicate Staff", Phone: "+77000000001", Role: core.RoleTeacher,
		Operator: "test-operator", Reason: "duplicate phone test",
	})
	studentInput := studentInput("duplicate-phone-student", "+77000000001", "DUPLICATE-PHONE", fixture.teacher.AccountID)
	_, studentErr := fixture.service.CreateStudent(ctx, fixture.admin, studentInput)

	for name, err := range map[string]error{"Owner": ownerErr, "staff": staffErr, "Student": studentErr} {
		if !core.IsCode(err, core.CodeConflict) {
			t.Fatalf("%s duplicate phone error = %v", name, err)
		}
		var appError *core.AppError
		if !errors.As(err, &appError) || appError.Message != "login identifier is unavailable" {
			t.Fatalf("%s duplicate phone message = %v", name, err)
		}
	}
}

func TestUnexpectedStoreFailuresAreInternal(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	fault := errors.New("database transport failed")
	opaque := strings.Repeat("A", 43)

	tests := []struct {
		name string
		run  func(*app.Service) error
		wrap app.Store
		code core.ErrorCode
	}{
		{"ready", func(service *app.Service) error { return service.Ready(ctx) }, &faultStore{Store: fixture.store, readyErr: fault}, core.CodeUnavailable},
		{"activation preview", func(service *app.Service) error { _, err := service.PreviewActivation(ctx, opaque); return err }, &faultStore{Store: fixture.store, previewErr: fault}, core.CodeInternal},
		{"activation validation", func(service *app.Service) error {
			return service.CompleteActivation(ctx, app.CompleteActivationInput{Token: opaque, Phone: "+77000000001", Password: ownerPassword, IdempotencyKey: "fault-validation"})
		}, &faultStore{Store: fixture.store, validateErr: fault}, core.CodeInternal},
		{"activation completion", func(service *app.Service) error {
			return service.CompleteActivation(ctx, app.CompleteActivationInput{Token: opaque, Phone: "+77000000001", Password: ownerPassword, IdempotencyKey: "fault-completion"})
		}, &faultStore{Store: fixture.store, validatePass: true, completeErr: fault}, core.CodeInternal},
		{"sign in", func(service *app.Service) error {
			_, err := service.SignIn(ctx, "+77000000001", ownerPassword, core.SessionClientInfo{})
			return err
		}, &faultStore{Store: fixture.store, credentialByPhoneErr: fault}, core.CodeInternal},
		{"corrupted sign-in credential", func(service *app.Service) error {
			_, err := service.SignIn(ctx, "+77000000001", ownerPassword, core.SessionClientInfo{})
			return err
		}, &faultStore{Store: fixture.store, credentialByPhoneResult: &core.CredentialRecord{
			AccountID: fixture.owner.AccountID, TenantID: fixture.owner.TenantID,
			Phone: "+77000000001", PasswordHash: "corrupted-phc", Status: "active",
		}}, core.CodeInternal},
		{"refresh", func(service *app.Service) error { _, err := service.Refresh(ctx, opaque); return err }, &faultStore{Store: fixture.store, rotateErr: fault}, core.CodeInternal},
		{"authenticate", func(service *app.Service) error { _, err := service.Authenticate(ctx, opaque); return err }, &faultStore{Store: fixture.store, principalErr: fault}, core.CodeInternal},
		{"Owner reauthentication", func(service *app.Service) error {
			_, err := service.GrantDelegation(ctx, fixture.owner, app.GrantDelegationInput{AdministratorAccountID: fixture.admin.AccountID, Reason: "fault", CurrentPassword: ownerPassword, IdempotencyKey: "fault-reauth"})
			return err
		}, &faultStore{Store: fixture.store, credentialByAccountErr: fault}, core.CodeInternal},
		{"sign out", func(service *app.Service) error { return service.SignOut(ctx, opaque) }, &faultStore{Store: fixture.store, revokeSessionErr: fault}, core.CodeInternal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := fixture.serviceForStore(test.wrap)
			if err := test.run(service); !core.IsCode(err, test.code) {
				t.Fatalf("unexpected store failure = %v, want %s", err, test.code)
			}
		})
	}
}

func TestActivationPrevalidationPrecedesHashingAndHashFailureIsInternal(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	opaque := strings.Repeat("A", 43)

	prevalidationHasher := &instrumentedPasswordHasher{delegate: fixture.hasher}
	prevalidationStore := &faultStore{
		Store:       fixture.store,
		validateErr: core.E(core.CodeInvalidActivation, "activation link is invalid or expired", nil),
	}
	service := fixture.serviceFor(prevalidationStore, prevalidationHasher)
	err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: opaque, Phone: "+77000000001", Password: ownerPassword,
		IdempotencyKey: "prevalidation-before-hash",
	})
	if !core.IsCode(err, core.CodeInvalidActivation) {
		t.Fatalf("prevalidation error = %v", err)
	}
	if prevalidationHasher.hashCalls.Load() != 0 {
		t.Fatalf("invalid token reached Argon2 hash %d time(s)", prevalidationHasher.hashCalls.Load())
	}
	pending, err := fixture.service.BootstrapOwnerWithAccount(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_phone_prevalidation", TenantName: "Phone Prevalidation",
		FullName: "Pending Owner", Phone: "+77000000911",
		Operator: "test-operator", Reason: "phone prevalidation fixture",
	})
	if err != nil {
		t.Fatalf("bootstrap phone prevalidation account: %v", err)
	}
	phoneHasher := &instrumentedPasswordHasher{delegate: fixture.hasher}
	service = fixture.serviceFor(fixture.store, phoneHasher)
	err = service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, pending.ActivationLink), Phone: "+77000000912",
		Password: ownerPassword, IdempotencyKey: "wrong-phone-before-hash",
	})
	if !core.IsCode(err, core.CodeInvalidActivation) {
		t.Fatalf("wrong activation phone = %v", err)
	}
	if phoneHasher.hashCalls.Load() != 0 {
		t.Fatalf("wrong phone reached Argon2 hash %d time(s)", phoneHasher.hashCalls.Load())
	}

	hashFailure := errors.New("entropy source failed")
	failingHasher := &instrumentedPasswordHasher{delegate: fixture.hasher, hashErr: hashFailure}
	service = fixture.serviceFor(&faultStore{Store: fixture.store, validatePass: true}, failingHasher)
	err = service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: opaque, Phone: "+77000000001", Password: ownerPassword,
		IdempotencyKey: "hash-infrastructure-failure",
	})
	if !core.IsCode(err, core.CodeInternal) {
		t.Fatalf("post-validation hash failure = %v, want INTERNAL", err)
	}
	if failingHasher.hashCalls.Load() != 1 {
		t.Fatalf("hash calls = %d, want 1", failingHasher.hashCalls.Load())
	}
}

func TestControlledStaffBootstrapAndDiscoveryAuthorization(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	administrators, err := fixture.service.ListStaff(ctx, fixture.owner, core.RoleAdministrator)
	if err != nil || len(administrators) != 1 || administrators[0].AccountID != fixture.admin.AccountID {
		t.Fatalf("Owner Administrator discovery = %#v, %v", administrators, err)
	}
	teachers, err := fixture.service.ListStaff(ctx, fixture.owner, core.RoleTeacher)
	if err != nil || len(teachers) != 1 || teachers[0].AccountID != fixture.teacher.AccountID {
		t.Fatalf("Owner Teacher discovery = %#v, %v", teachers, err)
	}
	teachers, err = fixture.service.ListStaff(ctx, fixture.admin, core.RoleTeacher)
	if err != nil || len(teachers) != 1 || teachers[0].AccountID != fixture.teacher.AccountID {
		t.Fatalf("ordinary Administrator Teacher discovery = %#v, %v", teachers, err)
	}
	_, err = fixture.service.ListStaff(ctx, fixture.admin, core.RoleAdministrator)
	assertCode(t, err, core.CodeForbidden)

	delegation := grantAdministrator(t, fixture, nil)
	teachers, err = fixture.service.ListStaff(ctx, fixture.admin, core.RoleTeacher)
	if err != nil || len(teachers) != 1 || teachers[0].AccountID != fixture.teacher.AccountID {
		t.Fatalf("delegated Administrator Teacher discovery = %#v, %v", teachers, err)
	}
	_, err = fixture.service.ListStaff(ctx, fixture.admin, core.RoleAdministrator)
	assertCode(t, err, core.CodeForbidden)
	administrators, err = fixture.service.ListStaff(ctx, fixture.owner, core.RoleAdministrator)
	if err != nil || len(administrators) != 1 {
		t.Fatalf("Owner delegated Administrator discovery = %#v, %v", administrators, err)
	}
	assertStrings(t, administrators[0].AccessProfiles, []string{core.StudentOnboardingManagerV1})
	if administrators[0].OnboardingDelegationID != delegation.ID {
		t.Fatalf("Administrator discovery delegation id = %q, want %q", administrators[0].OnboardingDelegationID, delegation.ID)
	}

	if _, _, err := fixture.service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: fixture.owner.TenantID, OwnerAccountID: fixture.owner.AccountID,
		FullName: "Invalid Owner", Phone: "+77000000501", Role: core.RoleOwner,
		Operator: "test-operator", Reason: "invalid role test",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("Owner staff bootstrap error = %v", err)
	}
	if _, _, err := fixture.service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: fixture.owner.TenantID, OwnerAccountID: fixture.admin.AccountID,
		FullName: "Unauthorized Teacher", Phone: "+77000000502", Role: core.RoleTeacher,
		Operator: "test-operator", Reason: "authorization test",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("non-Owner staff bootstrap error = %v", err)
	}
	if _, _, err := fixture.service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: fixture.owner.TenantID, OwnerAccountID: fixture.owner.AccountID,
		FullName: "Repeated Administrator", Phone: "+77000000002", Role: core.RoleAdministrator,
		Operator: "test-operator", Reason: "conflict test",
	}); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("repeated staff bootstrap error = %v", err)
	}
	_, err = fixture.service.ListStaff(ctx, fixture.owner, core.RoleStudent)
	assertCode(t, err, core.CodeInvalidInput)
}

func TestStaffDiscoveryIsTenantIsolated(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	otherOwnerLink, _, err := fixture.service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_other", TenantName: "Other School",
		FullName: "Other Owner", Phone: "+77000000601",
		Operator: "test-operator", Reason: "tenant isolation fixture",
	})
	if err != nil {
		t.Fatalf("bootstrap other Owner: %v", err)
	}
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, otherOwnerLink), Phone: "+77000000601",
		Password: ownerPassword, IdempotencyKey: "activate-other-owner",
	}); err != nil {
		t.Fatalf("activate other Owner: %v", err)
	}
	otherOwnerTokens := mustSignInTokens(t, fixture.service, "+77000000601", ownerPassword)
	otherOwner, err := fixture.service.Authenticate(ctx, otherOwnerTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate other Owner: %v", err)
	}
	otherTeacherLink, _, err := fixture.service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: otherOwner.TenantID, OwnerAccountID: otherOwner.AccountID,
		FullName: "Other Teacher", Phone: "+77000000602", Role: core.RoleTeacher,
		Operator: "test-operator", Reason: "tenant isolation fixture",
	})
	if err != nil {
		t.Fatalf("bootstrap other Teacher: %v", err)
	}
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, otherTeacherLink), Phone: "+77000000602",
		Password: teacherPassword, IdempotencyKey: "activate-other-teacher",
	}); err != nil {
		t.Fatalf("activate other Teacher: %v", err)
	}
	otherTeacherTokens := mustSignInTokens(t, fixture.service, "+77000000602", teacherPassword)
	otherTeacher, err := fixture.service.Authenticate(ctx, otherTeacherTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate other Teacher: %v", err)
	}
	otherStudent, err := fixture.service.CreateStudent(ctx, otherOwner, studentInput("other-tenant-student", "+77000000603", "OTHER-603", otherTeacher.AccountID))
	if err != nil {
		t.Fatalf("create other-tenant Student: %v", err)
	}

	firstTenantTeachers, err := fixture.service.ListStaff(ctx, fixture.owner, core.RoleTeacher)
	if err != nil || len(firstTenantTeachers) != 1 || firstTenantTeachers[0].AccountID != fixture.teacher.AccountID {
		t.Fatalf("first tenant leaked staff: %#v, %v", firstTenantTeachers, err)
	}
	otherTenantTeachers, err := fixture.service.ListStaff(ctx, otherOwner, core.RoleTeacher)
	if err != nil || len(otherTenantTeachers) != 1 || otherTenantTeachers[0].FullName != "Other Teacher" {
		t.Fatalf("other tenant staff = %#v, %v", otherTenantTeachers, err)
	}
	firstQueue, err := fixture.service.ListStudentOnboarding(ctx, fixture.owner)
	if err != nil || len(firstQueue) != 0 {
		t.Fatalf("first tenant leaked onboarding queue: %#v, %v", firstQueue, err)
	}
	otherQueue, err := fixture.service.ListStudentOnboarding(ctx, otherOwner)
	if err != nil || len(otherQueue) != 1 || otherQueue[0].StudentID != otherStudent.StudentID {
		t.Fatalf("other tenant onboarding queue = %#v, %v", otherQueue, err)
	}
}

func TestStudentOnboardingQueueStatesAndAssignmentScope(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	_, err := fixture.service.ListStudentOnboarding(ctx, fixture.admin)
	assertCode(t, err, core.CodeForbidden)
	grantAdministrator(t, fixture, nil)

	otherTeacherLink, _, err := fixture.service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: fixture.owner.TenantID, OwnerAccountID: fixture.owner.AccountID,
		FullName: "Other Assigned Teacher", Phone: "+77000000701", Role: core.RoleTeacher,
		Operator: "test-operator", Reason: "queue fixture",
	})
	if err != nil {
		t.Fatalf("bootstrap other assigned Teacher: %v", err)
	}
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, otherTeacherLink), Phone: "+77000000701",
		Password: teacherPassword, IdempotencyKey: "activate-other-assigned-teacher",
	}); err != nil {
		t.Fatalf("activate other assigned Teacher: %v", err)
	}
	otherTeacherTokens := mustSignInTokens(t, fixture.service, "+77000000701", teacherPassword)
	otherTeacher, err := fixture.service.Authenticate(ctx, otherTeacherTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate other assigned Teacher: %v", err)
	}

	first, err := fixture.service.CreateStudent(ctx, fixture.admin, studentInput("queue-first", "+77000000711", "QUEUE-711", fixture.teacher.AccountID))
	if err != nil {
		t.Fatalf("create first queued Student: %v", err)
	}
	second, err := fixture.service.CreateStudent(ctx, fixture.admin, studentInput("queue-second", "+77000000712", "QUEUE-712", otherTeacher.AccountID))
	if err != nil {
		t.Fatalf("create second queued Student: %v", err)
	}
	items, err := fixture.service.ListStudentOnboarding(ctx, fixture.admin)
	if err != nil || len(items) != 2 {
		t.Fatalf("delegated queue = %#v, %v", items, err)
	}
	firstItem := onboardingItem(t, items, first.StudentID)
	if firstItem.OnboardingState != core.OnboardingAwaitingFirstMinute || firstItem.StudentVersion != 0 || firstItem.InvitationID != "" || firstItem.EnrollmentReference != "QUEUE-711" {
		t.Fatalf("initial onboarding state = %#v", firstItem)
	}
	assignedItems, err := fixture.service.ListStudentOnboarding(ctx, fixture.teacher)
	if err != nil || len(assignedItems) != 1 || assignedItems[0].StudentID != first.StudentID {
		t.Fatalf("first Teacher queue = %#v, %v", assignedItems, err)
	}
	otherAssignedItems, err := fixture.service.ListStudentOnboarding(ctx, otherTeacher)
	if err != nil || len(otherAssignedItems) != 1 || otherAssignedItems[0].StudentID != second.StudentID {
		t.Fatalf("other Teacher queue = %#v, %v", otherAssignedItems, err)
	}

	if _, err := fixture.service.PublishFirstMinute(ctx, fixture.teacher, app.PublishFirstMinuteInput{
		StudentID: first.StudentID, WhatWorked: "Queue worked", CurrentFocus: "Queue focus",
		NextStep: "Queue next", ExpectedVersion: firstItem.StudentVersion,
		IdempotencyKey: "queue-first-minute",
	}); err != nil {
		t.Fatalf("publish queued First Minute: %v", err)
	}
	items, err = fixture.service.ListStudentOnboarding(ctx, fixture.owner)
	if err != nil {
		t.Fatalf("Owner queue after First Minute: %v", err)
	}
	firstItem = onboardingItem(t, items, first.StudentID)
	if firstItem.OnboardingState != core.OnboardingReadyToInvite || firstItem.StudentVersion != 1 {
		t.Fatalf("ready onboarding state = %#v", firstItem)
	}

	invitation, link, err := fixture.service.IssueInvitation(ctx, fixture.owner, first.StudentID, "queue-invitation", core.InvitationIssue)
	if err != nil {
		t.Fatalf("issue queued invitation: %v", err)
	}
	items, err = fixture.service.ListStudentOnboarding(ctx, fixture.admin)
	if err != nil {
		t.Fatalf("queue after invitation: %v", err)
	}
	firstItem = onboardingItem(t, items, first.StudentID)
	if firstItem.OnboardingState != core.OnboardingInvited || firstItem.InvitationID != invitation.InvitationID || firstItem.InvitationExpiresAt == nil {
		t.Fatalf("invited onboarding state = %#v", firstItem)
	}
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, link), Phone: "+77000000711",
		Password: studentPassword, IdempotencyKey: "queue-activation",
	}); err != nil {
		t.Fatalf("activate queued Student: %v", err)
	}
	items, err = fixture.service.ListStudentOnboarding(ctx, fixture.owner)
	if err != nil {
		t.Fatalf("queue after activation: %v", err)
	}
	firstItem = onboardingItem(t, items, first.StudentID)
	if firstItem.OnboardingState != core.OnboardingActivated || firstItem.InvitationID != "" || firstItem.InvitationExpiresAt != nil {
		t.Fatalf("activated onboarding state = %#v", firstItem)
	}
}

func TestPersistedInputBoundsAndSemanticValidation(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()
	tooLongKey := strings.Repeat("k", 129)
	input := studentInput(tooLongKey, "+77000000801", "BOUND-801", fixture.teacher.AccountID)
	_, err := fixture.service.CreateStudent(ctx, fixture.owner, input)
	assertCode(t, err, core.CodeInvalidInput)

	input = studentInput("bounds-full-name", "+77000000802", "BOUND-802", fixture.teacher.AccountID)
	input.FullName = strings.Repeat("я", 201)
	_, err = fixture.service.CreateStudent(ctx, fixture.owner, input)
	assertCode(t, err, core.CodeInvalidInput)

	input = studentInput("bounds-enrollment", "+77000000803", strings.Repeat("E", 101), fixture.teacher.AccountID)
	_, err = fixture.service.CreateStudent(ctx, fixture.owner, input)
	assertCode(t, err, core.CodeInvalidInput)

	input = studentInput("bounds-locale", "+77000000804", "BOUND-804", fixture.teacher.AccountID)
	input.Locale = "not_a_locale"
	_, err = fixture.service.CreateStudent(ctx, fixture.owner, input)
	assertCode(t, err, core.CodeInvalidInput)

	input = studentInput("bounds-timezone", "+77000000805", "BOUND-805", fixture.teacher.AccountID)
	input.Timezone = "Mars/Olympus_Mons"
	_, err = fixture.service.CreateStudent(ctx, fixture.owner, input)
	assertCode(t, err, core.CodeInvalidInput)

	_, err = fixture.service.GrantDelegation(ctx, fixture.owner, app.GrantDelegationInput{
		AdministratorAccountID: fixture.admin.AccountID,
		Reason:                 strings.Repeat("r", 501), CurrentPassword: ownerPassword,
		IdempotencyKey: "bounds-reason",
	})
	assertCode(t, err, core.CodeInvalidInput)

	_, err = fixture.service.PublishFirstMinute(ctx, fixture.teacher, app.PublishFirstMinuteInput{
		StudentID: "student_valid", WhatWorked: strings.Repeat("w", 501),
		CurrentFocus: "Focus", NextStep: "Next", ExpectedVersion: 0,
		IdempotencyKey: "bounds-first-minute",
	})
	assertCode(t, err, core.CodeInvalidInput)
}

func studentInput(key, phone, enrollment, teacherAccountID string) app.CreateStudentInput {
	return app.CreateStudentInput{
		FullName: "Adult Student", Phone: phone, EnrollmentReference: enrollment,
		TeacherAccountID: teacherAccountID, Locale: "ru-KZ", Timezone: "Asia/Almaty",
		AdultConfirmed: true, IdempotencyKey: key,
	}
}

func grantAdministrator(t *testing.T, fixture *fixture, expiresAt *time.Time) core.DelegationResult {
	t.Helper()
	result, err := fixture.service.GrantDelegation(context.Background(), fixture.owner, app.GrantDelegationInput{
		AdministratorAccountID: fixture.admin.AccountID, Reason: "Test delegation",
		ExpiresAt: expiresAt, CurrentPassword: ownerPassword,
		IdempotencyKey: "grant-helper",
	})
	if err != nil {
		t.Fatalf("grant Administrator: %v", err)
	}
	return result
}

func readyStudentInvitation(t *testing.T, fixture *fixture, phone, enrollment string) (core.StudentResult, string) {
	t.Helper()
	student, _ := readyStudentWithoutInvitation(t, fixture, phone, enrollment)
	_, link, err := fixture.service.IssueInvitation(context.Background(), fixture.owner, student.StudentID, "invite-"+enrollment, core.InvitationIssue)
	if err != nil {
		t.Fatalf("issue ready invitation: %v", err)
	}
	return student, link
}

func readyStudentWithoutInvitation(t *testing.T, fixture *fixture, phone, enrollment string) (core.StudentResult, core.FirstMinute) {
	t.Helper()
	grantAdministrator(t, fixture, nil)
	student, err := fixture.service.CreateStudent(context.Background(), fixture.admin, studentInput("create-"+enrollment, phone, enrollment, fixture.teacher.AccountID))
	if err != nil {
		t.Fatalf("create ready Student: %v", err)
	}
	firstMinute, err := fixture.service.PublishFirstMinute(context.Background(), fixture.teacher, app.PublishFirstMinuteInput{
		StudentID: student.StudentID, WhatWorked: "Worked", CurrentFocus: "Focus",
		NextStep: "Next", ExpectedVersion: 0, IdempotencyKey: "first-" + enrollment,
	})
	if err != nil {
		t.Fatalf("publish ready First Minute: %v", err)
	}
	return student, firstMinute
}

func tokenFromLink(t *testing.T, link string) string {
	t.Helper()
	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse activation link: %v", err)
	}
	fragment, err := url.ParseQuery(parsed.Fragment)
	if err != nil {
		t.Fatalf("parse activation fragment: %v", err)
	}
	token := fragment.Get("token")
	if token == "" {
		t.Fatalf("activation link has no token: %q", link)
	}
	return token
}

func assertCode(t *testing.T, err error, code core.ErrorCode) {
	t.Helper()
	if !core.IsCode(err, code) {
		t.Fatalf("error = %v, want code %s", err, code)
	}
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("strings = %#v, want %#v", got, want)
	}
}

func onboardingItem(t *testing.T, items []core.StudentOnboardingItem, studentID string) core.StudentOnboardingItem {
	t.Helper()
	for _, item := range items {
		if item.StudentID == studentID {
			return item
		}
	}
	t.Fatalf("student %s not found in onboarding queue %#v", studentID, items)
	return core.StudentOnboardingItem{}
}

type faultStore struct {
	app.Store
	readyErr                error
	previewErr              error
	validateErr             error
	validatePass            bool
	completeErr             error
	credentialByPhoneErr    error
	credentialByPhoneResult *core.CredentialRecord
	credentialByAccountErr  error
	rotateErr               error
	principalErr            error
	revokeSessionErr        error
}

func (s *faultStore) Ready(ctx context.Context) error {
	if s.readyErr != nil {
		return s.readyErr
	}
	return s.Store.Ready(ctx)
}

func (s *faultStore) PreviewActivation(ctx context.Context, digest []byte, now time.Time) (core.ActivationPreview, error) {
	if s.previewErr != nil {
		return core.ActivationPreview{}, s.previewErr
	}
	return s.Store.PreviewActivation(ctx, digest, now)
}

func (s *faultStore) ValidateActivation(ctx context.Context, command core.ActivationValidationCommand) (bool, error) {
	if s.validateErr != nil {
		return false, s.validateErr
	}
	if s.validatePass {
		return false, nil
	}
	return s.Store.ValidateActivation(ctx, command)
}

func (s *faultStore) CompleteActivation(ctx context.Context, command core.ActivationCompleteCommand) error {
	if s.completeErr != nil {
		return s.completeErr
	}
	return s.Store.CompleteActivation(ctx, command)
}

func (s *faultStore) CredentialByPhone(ctx context.Context, phone string) (core.CredentialRecord, error) {
	if s.credentialByPhoneErr != nil {
		return core.CredentialRecord{}, s.credentialByPhoneErr
	}
	if s.credentialByPhoneResult != nil {
		return *s.credentialByPhoneResult, nil
	}
	return s.Store.CredentialByPhone(ctx, phone)
}

func (s *faultStore) CredentialByAccount(ctx context.Context, accountID string) (core.CredentialRecord, error) {
	if s.credentialByAccountErr != nil {
		return core.CredentialRecord{}, s.credentialByAccountErr
	}
	return s.Store.CredentialByAccount(ctx, accountID)
}

func (s *faultStore) RotateSession(ctx context.Context, digest []byte, material core.SessionMaterial, now time.Time) (string, string, error) {
	if s.rotateErr != nil {
		return "", "", s.rotateErr
	}
	return s.Store.RotateSession(ctx, digest, material, now)
}

func (s *faultStore) PrincipalByAccessDigest(ctx context.Context, digest []byte, now time.Time) (core.Principal, error) {
	if s.principalErr != nil {
		return core.Principal{}, s.principalErr
	}
	return s.Store.PrincipalByAccessDigest(ctx, digest, now)
}

func (s *faultStore) RevokeSession(ctx context.Context, digest []byte, now time.Time) error {
	if s.revokeSessionErr != nil {
		return s.revokeSessionErr
	}
	return s.Store.RevokeSession(ctx, digest, now)
}

func (fixture *fixture) serviceForStore(store app.Store) *app.Service {
	return fixture.serviceFor(store, fixture.hasher)
}

func (fixture *fixture) serviceFor(store app.Store, passwords app.PasswordService) *app.Service {
	return app.NewService(store, fixture.codec, passwords, app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour, Clock: fixture.clock,
	})
}

type instrumentedPasswordHasher struct {
	delegate  *security.PasswordHasher
	hashCalls atomic.Int32
	hashErr   error
}

func (hasher *instrumentedPasswordHasher) NormalizeAndValidate(password string) (string, error) {
	return hasher.delegate.NormalizeAndValidate(password)
}

func (hasher *instrumentedPasswordHasher) Hash(password string) (string, error) {
	hasher.hashCalls.Add(1)
	if hasher.hashErr != nil {
		return "", hasher.hashErr
	}
	return hasher.delegate.Hash(password)
}

func (hasher *instrumentedPasswordHasher) VerifyCredential(password, encoded string) (bool, error) {
	return hasher.delegate.VerifyCredential(password, encoded)
}

func (hasher *instrumentedPasswordHasher) DummyHash() string {
	return hasher.delegate.DummyHash()
}

func signInPrincipal(t *testing.T, service *app.Service, phone, password string) core.Principal {
	t.Helper()
	tokens := mustSignInTokens(t, service, phone, password)
	principal, err := service.Authenticate(context.Background(), tokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate principal: %v", err)
	}
	return principal
}

func mustSignInTokens(t *testing.T, service *app.Service, phone, password string) core.SessionTokens {
	t.Helper()
	outcome, err := service.SignIn(context.Background(), phone, password, core.SessionClientInfo{})
	if err != nil {
		t.Fatalf("sign in: %v", err)
	}
	if outcome.Tokens == nil {
		t.Fatal("sign-in returned a second-factor challenge; tokens expected")
	}
	return *outcome.Tokens
}

func assertRoles(t *testing.T, got, want []core.Role) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("roles = %#v, want %#v", got, want)
	}
}
