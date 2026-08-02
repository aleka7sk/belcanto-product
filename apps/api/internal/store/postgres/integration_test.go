package postgres_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/postgres"
	"github.com/aleka7sk/belcanto-product/apps/api/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgreSQLInvitationActivationAndDelegatedOnboarding(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool, schema := isolatedPool(t, ctx, databaseURL)
	if err := migrations.Up(ctx, pool); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
	store := postgres.New(pool)
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x53}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	hasher := security.NewPasswordHasher()
	service := app.NewService(store, codec, hasher, app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})

	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pg", TenantName: "Belcanto PostgreSQL",
		FullName: "PostgreSQL Owner", Phone: "+77002000001",
		Operator: "pg-test-operator", Reason: "PostgreSQL integration bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap Owner: %v", err)
	}
	ownerToken := integrationToken(t, ownerLink)
	assertDigestOnlySchema(t, ctx, pool, schema, ownerToken)
	ownerAccountID := accountIDForToken(t, ctx, pool, codec, ownerToken)
	ownerRecovery, recoveredOwnerLink, err := service.ReissueBootstrapInvitation(ctx, app.ReissueBootstrapInvitationInput{
		TenantID: "tenant_pg", AccountID: ownerAccountID,
		Operator: "pg-recovery-operator", Reason: "lost initial Owner activation link",
	})
	if err != nil {
		t.Fatalf("reissue Owner bootstrap invitation: %v", err)
	}
	if ownerRecovery.Kind != "owner_bootstrap" || ownerRecovery.AccountID != ownerAccountID {
		t.Fatalf("Owner recovery result = %#v", ownerRecovery)
	}
	if _, err := service.PreviewActivation(ctx, ownerToken); !core.IsCode(err, core.CodeInvalidActivation) {
		t.Fatalf("superseded Owner bootstrap link = %v", err)
	}
	activateIntegration(t, ctx, service, recoveredOwnerLink, "+77002000001", "Owner-password-123!", "activate-pg-owner")
	owner := integrationPrincipal(t, ctx, service, "+77002000001", "Owner-password-123!")
	assertRoleGrantCount(t, ctx, pool, owner.TenantID, owner.AccountID, core.RoleOwner, 1)
	assertOperatorAudit(t, ctx, pool, ownerRecovery.InvitationID, "pg-recovery-operator", "lost initial Owner activation link")
	if _, _, err := service.ReissueBootstrapInvitation(ctx, app.ReissueBootstrapInvitationInput{
		TenantID: owner.TenantID, AccountID: owner.AccountID,
		Operator: "pg-recovery-operator", Reason: "active account must reject recovery",
	}); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("active Owner recovery = %v", err)
	}

	if _, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: "not-an-owner",
		FullName: "Forbidden Staff", Phone: "+77002000009", Role: core.RoleTeacher,
		Operator: "pg-test-operator", Reason: "authorization test",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("unauthorized staff bootstrap = %v", err)
	}
	adminLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "PostgreSQL Administrator", Phone: "+77002000002", Role: core.RoleAdministrator,
		Operator: "pg-test-operator", Reason: "PostgreSQL integration staff bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap Administrator: %v", err)
	}
	adminToken := integrationToken(t, adminLink)
	adminAccountID := accountIDForToken(t, ctx, pool, codec, adminToken)
	adminRecovery, recoveredAdminLink, err := service.ReissueBootstrapInvitation(ctx, app.ReissueBootstrapInvitationInput{
		TenantID: owner.TenantID, AccountID: adminAccountID,
		Operator: "pg-recovery-operator", Reason: "lost initial Administrator activation link",
	})
	if err != nil {
		t.Fatalf("reissue Administrator bootstrap invitation: %v", err)
	}
	if _, err := service.PreviewActivation(ctx, adminToken); !core.IsCode(err, core.CodeInvalidActivation) {
		t.Fatalf("superseded Administrator bootstrap link = %v", err)
	}
	activateIntegration(t, ctx, service, recoveredAdminLink, "+77002000002", "Admin-password-123!", "activate-pg-admin")
	administrator := integrationPrincipal(t, ctx, service, "+77002000002", "Admin-password-123!")
	assertRoleGrantCount(t, ctx, pool, owner.TenantID, administrator.AccountID, core.RoleAdministrator, 1)
	assertOperatorAudit(t, ctx, pool, adminRecovery.InvitationID, "pg-recovery-operator", "lost initial Administrator activation link")
	teacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "PostgreSQL Teacher", Phone: "+77002000003", Role: core.RoleTeacher,
		Operator: "pg-test-operator", Reason: "PostgreSQL integration staff bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap Teacher: %v", err)
	}
	activateIntegration(t, ctx, service, teacherLink, "+77002000003", "Teacher-password-123!", "activate-pg-teacher")
	teacher := integrationPrincipal(t, ctx, service, "+77002000003", "Teacher-password-123!")

	staff, err := service.ListStaff(ctx, owner, core.RoleAdministrator)
	if err != nil || len(staff) != 1 || staff[0].AccountID != administrator.AccountID {
		t.Fatalf("Owner staff discovery = %#v, %v", staff, err)
	}
	_, err = service.CreateStudent(ctx, administrator, app.CreateStudentInput{
		FullName: "Before Grant", Phone: "+77002000100", EnrollmentReference: "PG-100",
		TeacherAccountID: teacher.AccountID, AdultConfirmed: true, IdempotencyKey: "pg-before-grant",
	})
	if !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("ordinary Administrator create = %v", err)
	}
	if _, err := service.ListStudentOnboarding(ctx, administrator); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("ordinary Administrator queue read = %v", err)
	}
	delegation, err := service.GrantDelegation(ctx, owner, app.GrantDelegationInput{
		AdministratorAccountID: administrator.AccountID, Reason: "PostgreSQL integration grant",
		CurrentPassword: "Owner-password-123!", IdempotencyKey: "pg-grant-admin",
	})
	if err != nil {
		t.Fatalf("grant Administrator: %v", err)
	}
	view, err := service.BootstrapView(ctx, administrator)
	if err != nil {
		t.Fatalf("Administrator bootstrap view: %v", err)
	}
	expectedAdministratorPermissions := append(core.LessonPermissionSetForRoles([]core.Role{core.RoleAdministrator}), core.StudentOnboardingManagerV1PermissionSet()...)
	if !reflect.DeepEqual(view.AccessProfiles, []string{core.StudentOnboardingManagerV1}) || !reflect.DeepEqual(view.Permissions, expectedAdministratorPermissions) {
		t.Fatalf("effective Administrator access = %#v", view)
	}
	staff, err = service.ListStaff(ctx, owner, core.RoleAdministrator)
	if err != nil || len(staff) != 1 || staff[0].OnboardingDelegationID != delegation.ID {
		t.Fatalf("delegation discovery = %#v, %v", staff, err)
	}
	primaryStudentInput := app.CreateStudentInput{
		FullName: "PostgreSQL Student", Phone: "+77002000101", EnrollmentReference: "PG-101",
		TeacherAccountID: teacher.AccountID, Locale: "ru-KZ", Timezone: "Asia/Almaty",
		AdultConfirmed: true, IdempotencyKey: "pg-create-student",
	}
	student, err := service.CreateStudent(ctx, administrator, primaryStudentInput)
	if err != nil {
		t.Fatalf("create Student: %v", err)
	}
	queue, err := service.ListStudentOnboarding(ctx, teacher)
	if err != nil || len(queue) != 1 || queue[0].StudentID != student.StudentID || queue[0].EnrollmentReference != "PG-101" || queue[0].OnboardingState != core.OnboardingAwaitingFirstMinute || queue[0].StudentVersion != 0 {
		t.Fatalf("initial Teacher queue = %#v, %v", queue, err)
	}
	if _, err := service.PublishFirstMinute(ctx, teacher, app.PublishFirstMinuteInput{
		StudentID: student.StudentID, WhatWorked: "Worked", CurrentFocus: "Focus",
		NextStep: "Next", ExpectedVersion: 0, IdempotencyKey: "pg-first-minute",
	}); err != nil {
		t.Fatalf("publish First Minute: %v", err)
	}
	queue, err = service.ListStudentOnboarding(ctx, administrator)
	if err != nil || len(queue) != 1 || queue[0].OnboardingState != core.OnboardingReadyToInvite || queue[0].StudentVersion != 1 {
		t.Fatalf("ready Administrator queue = %#v, %v", queue, err)
	}
	if _, _, err := service.IssueInvitation(ctx, administrator, student.StudentID, "pg-admin-invite", core.InvitationIssue); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("delegated Administrator issue invitation = %v", err)
	}
	if _, _, err := service.IssueInvitation(ctx, owner, student.StudentID, "pg-reissue-without-active", core.InvitationReissue); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("Owner reissue without active invitation = %v", err)
	}
	oldInvitation, oldLink, err := service.IssueInvitation(ctx, owner, student.StudentID, "pg-issue", core.InvitationIssue)
	if err != nil {
		t.Fatalf("issue invitation: %v", err)
	}
	newInvitation, newLink, err := service.IssueInvitation(ctx, owner, student.StudentID, "pg-reissue", core.InvitationReissue)
	if err != nil {
		t.Fatalf("reissue invitation: %v", err)
	}
	if oldInvitation.InvitationID == newInvitation.InvitationID || oldLink == newLink {
		t.Fatal("PostgreSQL reissue did not replace invitation")
	}
	if err := service.RevokeInvitation(ctx, administrator, newInvitation.InvitationID, "pg-admin-revoke"); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("delegated Administrator revoke invitation = %v", err)
	}
	queue, err = service.ListStudentOnboarding(ctx, owner)
	if err != nil || len(queue) != 1 || queue[0].OnboardingState != core.OnboardingInvited || queue[0].InvitationID != newInvitation.InvitationID {
		t.Fatalf("invited Owner queue = %#v, %v", queue, err)
	}
	if _, err := service.PreviewActivation(ctx, integrationToken(t, oldLink)); !core.IsCode(err, core.CodeInvalidActivation) {
		t.Fatalf("old PostgreSQL invitation preview = %v", err)
	}

	passwordHash, err := hasher.Hash("Student-password-123!")
	if err != nil {
		t.Fatalf("hash concurrent activation password: %v", err)
	}
	newToken := integrationToken(t, newLink)
	assertDigestOnlySchema(t, ctx, pool, schema, newToken)
	fingerprint, err := security.Fingerprint(map[string]string{"phone": "+77002000101", "passwordProof": "integration"})
	if err != nil {
		t.Fatalf("fingerprint concurrent activation: %v", err)
	}
	const competitors = 4
	var successes atomic.Int32
	var invalid atomic.Int32
	var wait sync.WaitGroup
	wait.Add(competitors)
	for index := 0; index < competitors; index++ {
		go func(index int) {
			defer wait.Done()
			err := store.CompleteActivation(ctx, core.ActivationCompleteCommand{
				TokenDigest: codec.Digest(newToken), Phone: "+77002000101",
				PasswordHash: passwordHash, IdempotencyKey: "pg-activation-" + string(rune('a'+index)),
				PayloadFingerprint: fingerprint, Now: time.Now().UTC(),
			})
			switch {
			case err == nil:
				successes.Add(1)
			case core.IsCode(err, core.CodeInvalidActivation):
				invalid.Add(1)
			default:
				t.Errorf("unexpected concurrent activation error: %v", err)
			}
		}(index)
	}
	wait.Wait()
	if successes.Load() != 1 || invalid.Load() != competitors-1 {
		t.Fatalf("PostgreSQL activation outcomes success=%d invalid=%d", successes.Load(), invalid.Load())
	}
	studentPrincipal := integrationPrincipal(t, ctx, service, "+77002000101", "Student-password-123!")
	studentView, err := service.BootstrapView(ctx, studentPrincipal)
	if err != nil || studentView.StudentID != student.StudentID || studentView.FirstMinute == nil {
		t.Fatalf("Student bootstrap = %#v, %v", studentView, err)
	}
	queue, err = service.ListStudentOnboarding(ctx, owner)
	if err != nil || len(queue) != 1 || queue[0].OnboardingState != core.OnboardingActivated || queue[0].InvitationID != "" {
		t.Fatalf("activated Owner queue = %#v, %v", queue, err)
	}

	otherOwnerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_pg_other", TenantName: "Other PostgreSQL School",
		FullName: "Other PostgreSQL Owner", Phone: "+77002000201",
		Operator: "pg-test-operator", Reason: "tenant isolation bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap other PostgreSQL Owner: %v", err)
	}
	otherOwnerPreview, err := service.PreviewActivation(ctx, integrationToken(t, otherOwnerLink))
	if err != nil {
		t.Fatalf("preview other PostgreSQL Owner: %v", err)
	}
	activateIntegration(t, ctx, service, otherOwnerLink, "+77002000201", "Owner-password-123!", "activate-pg-other-owner")
	otherOwner := integrationPrincipal(t, ctx, service, "+77002000201", "Owner-password-123!")
	otherTeacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: otherOwner.TenantID, OwnerAccountID: otherOwner.AccountID,
		FullName: "Other PostgreSQL Teacher", Phone: "+77002000202", Role: core.RoleTeacher,
		Operator: "pg-test-operator", Reason: "tenant isolation staff bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap other PostgreSQL Teacher: %v", err)
	}
	activateIntegration(t, ctx, service, otherTeacherLink, "+77002000202", "Teacher-password-123!", "activate-pg-other-teacher")
	otherTeacher := integrationPrincipal(t, ctx, service, "+77002000202", "Teacher-password-123!")
	otherStudent, err := service.CreateStudent(ctx, otherOwner, app.CreateStudentInput{
		FullName: "Other PostgreSQL Student", Phone: "+77002000203",
		EnrollmentReference: "PG-OTHER-203", TeacherAccountID: otherTeacher.AccountID,
		AdultConfirmed: true, IdempotencyKey: "pg-other-student",
	})
	if err != nil {
		t.Fatalf("create other PostgreSQL Student: %v", err)
	}
	assertConstraintRejected(t, ctx, pool, `
		UPDATE activation_invitations SET student_id = NULL WHERE id = $1
	`, newInvitation.InvitationID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE activation_invitations SET account_id = $2 WHERE id = $1
	`, newInvitation.InvitationID, administrator.AccountID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE activation_invitations SET issued_by_account_id = $2 WHERE id = $1
	`, newInvitation.InvitationID, otherOwner.AccountID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE activation_invitations SET superseded_by_id = $2 WHERE id = $1
	`, oldInvitation.InvitationID, otherOwnerPreview.InvitationID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE sessions SET replaced_by_id = $2 WHERE id = $1
	`, owner.SessionID, otherOwner.SessionID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE role_grants SET granted_by = $3
		WHERE tenant_id = $1 AND account_id = $2 AND role_type = 'Student'
	`, owner.TenantID, student.AccountID, otherOwner.AccountID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE capability_delegations SET revoked_by_account_id = $2 WHERE id = $1
	`, delegation.ID, otherOwner.AccountID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE capability_delegations SET status = 'revoked' WHERE id = $1
	`, delegation.ID)
	secondOwnerGrantID, err := security.NewID("role")
	if err != nil {
		t.Fatalf("generate second Owner role id: %v", err)
	}
	assertConstraintRejected(t, ctx, pool, `
		INSERT INTO role_grants (
			id, tenant_id, account_id, role_type, scope_type, scope_id,
			status, granted_by, granted_at
		) VALUES ($1, $2, $3, 'Owner', 'tenant', $2, 'active', $4, $5)
	`, secondOwnerGrantID, owner.TenantID, administrator.AccountID, owner.AccountID, time.Now().UTC())
	assertConstraintRejected(t, ctx, pool, `
		UPDATE accounts SET status = 'active' WHERE tenant_id = $1 AND id = $2
	`, otherOwner.TenantID, otherStudent.AccountID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE role_grants SET scope_type = 'tenant', scope_id = tenant_id
		WHERE tenant_id = $1 AND account_id = $2 AND role_type = 'Student'
	`, owner.TenantID, student.AccountID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE activation_invitations SET consumed_at = NULL WHERE id = $1
	`, newInvitation.InvitationID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE sessions SET refresh_expires_at = access_expires_at WHERE id = $1
	`, owner.SessionID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE sessions SET access_digest = decode('00', 'hex') WHERE id = $1
	`, owner.SessionID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE activation_invitations SET token_digest = decode('00', 'hex') WHERE id = $1
	`, newInvitation.InvitationID)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE idempotency_records SET status = 'processing'
		WHERE tenant_id = $1 AND actor_account_id = $2
		  AND operation_scope = 'create_student' AND idempotency_key = $3
	`, owner.TenantID, administrator.AccountID, primaryStudentInput.IdempotencyKey)
	assertConstraintRejected(t, ctx, pool, `
		UPDATE idempotency_records SET payload_fingerprint = decode('00', 'hex')
		WHERE tenant_id = $1 AND actor_account_id = $2
		  AND operation_scope = 'create_student' AND idempotency_key = $3
	`, owner.TenantID, administrator.AccountID, primaryStudentInput.IdempotencyKey)
	queue, err = service.ListStudentOnboarding(ctx, owner)
	if err != nil || len(queue) != 1 || queue[0].StudentID != student.StudentID {
		t.Fatalf("primary tenant queue leaked: %#v, %v", queue, err)
	}
	otherQueue, err := service.ListStudentOnboarding(ctx, otherOwner)
	if err != nil || len(otherQueue) != 1 || otherQueue[0].StudentID != otherStudent.StudentID {
		t.Fatalf("other tenant queue = %#v, %v", otherQueue, err)
	}

	// Idempotency is scoped to the authenticated actor, so an Owner may use the
	// same key that an Administrator used without receiving that actor's replay.
	actorScopedStudent, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
		FullName: "Actor Scoped Owner Student", Phone: "+77002000301",
		EnrollmentReference: "PG-ACTOR-301", TeacherAccountID: teacher.AccountID,
		AdultConfirmed: true, IdempotencyKey: primaryStudentInput.IdempotencyKey,
	})
	if err != nil || actorScopedStudent.StudentID == student.StudentID {
		t.Fatalf("Owner actor-scoped idempotency result = %#v, %v", actorScopedStudent, err)
	}

	// Changing the current Teacher assignment must invalidate even an exact
	// replay before the idempotency record is consulted.
	replacementTeacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "Replacement PostgreSQL Teacher", Phone: "+77002000302", Role: core.RoleTeacher,
		Operator: "pg-test-operator", Reason: "assignment replay authorization test",
	})
	if err != nil {
		t.Fatalf("bootstrap replacement Teacher: %v", err)
	}
	activateIntegration(t, ctx, service, replacementTeacherLink, "+77002000302", "Teacher-password-123!", "activate-pg-replacement-teacher")
	replacementTeacher := integrationPrincipal(t, ctx, service, "+77002000302", "Teacher-password-123!")
	replayStudent, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
		FullName: "Teacher Replay Student", Phone: "+77002000303",
		EnrollmentReference: "PG-TEACHER-303", TeacherAccountID: teacher.AccountID,
		AdultConfirmed: true, IdempotencyKey: "pg-teacher-replay-student",
	})
	if err != nil {
		t.Fatalf("create Teacher replay Student: %v", err)
	}
	oldTeacherInput := app.PublishFirstMinuteInput{
		StudentID: replayStudent.StudentID, WhatWorked: "Original Teacher worked",
		CurrentFocus: "Original Teacher focus", NextStep: "Original Teacher next",
		ExpectedVersion: 0, IdempotencyKey: "pg-teacher-actor-key",
	}
	if _, err := service.PublishFirstMinute(ctx, teacher, oldTeacherInput); err != nil {
		t.Fatalf("publish original Teacher revision: %v", err)
	}
	replacementAssignmentID, err := security.NewID("assignment")
	if err != nil {
		t.Fatalf("generate replacement assignment id: %v", err)
	}
	assignmentChangedAt := time.Now().UTC()
	if _, err := pool.Exec(ctx, `
		UPDATE teacher_assignments
		SET status = 'ended', ended_at = $3
		WHERE tenant_id = $1 AND student_id = $2 AND status = 'active'
	`, owner.TenantID, replayStudent.StudentID, assignmentChangedAt); err != nil {
		t.Fatalf("end original Teacher assignment: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO teacher_assignments (
			id, tenant_id, student_id, teacher_account_id, status,
			assigned_by_account_id, assigned_at
		) VALUES ($1, $2, $3, $4, 'active', $5, $6)
	`, replacementAssignmentID, owner.TenantID, replayStudent.StudentID,
		replacementTeacher.AccountID, owner.AccountID, assignmentChangedAt); err != nil {
		t.Fatalf("assign replacement Teacher: %v", err)
	}
	if _, err := service.PublishFirstMinute(ctx, teacher, oldTeacherInput); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("removed Teacher exact replay = %v", err)
	}
	newTeacherRevision, err := service.PublishFirstMinute(ctx, replacementTeacher, app.PublishFirstMinuteInput{
		StudentID: replayStudent.StudentID, WhatWorked: "Replacement Teacher worked",
		CurrentFocus: "Replacement Teacher focus", NextStep: "Replacement Teacher next",
		ExpectedVersion: 1, IdempotencyKey: oldTeacherInput.IdempotencyKey,
	})
	if err != nil || newTeacherRevision.Revision != 2 {
		t.Fatalf("replacement Teacher actor-scoped replay = %#v, %v", newTeacherRevision, err)
	}

	// Hold the same canonical subject lock used by activation and invitation
	// mutation, prove both operations wait on it, then assert exactly one can
	// win after release without a deadlock or split-brain invitation state.
	raceStudent, err := service.CreateStudent(ctx, owner, app.CreateStudentInput{
		FullName: "Activation Reissue Race Student", Phone: "+77002000304",
		EnrollmentReference: "PG-RACE-304", TeacherAccountID: replacementTeacher.AccountID,
		AdultConfirmed: true, IdempotencyKey: "pg-race-student",
	})
	if err != nil {
		t.Fatalf("create activation/reissue race Student: %v", err)
	}
	if _, err := service.PublishFirstMinute(ctx, replacementTeacher, app.PublishFirstMinuteInput{
		StudentID: raceStudent.StudentID, WhatWorked: "Race worked", CurrentFocus: "Race focus",
		NextStep: "Race next", ExpectedVersion: 0, IdempotencyKey: "pg-race-first-minute",
	}); err != nil {
		t.Fatalf("publish race First Minute: %v", err)
	}
	_, raceLink, err := service.IssueInvitation(ctx, owner, raceStudent.StudentID, "pg-race-issue", core.InvitationIssue)
	if err != nil {
		t.Fatalf("issue race invitation: %v", err)
	}
	raceToken := integrationToken(t, raceLink)
	racePasswordHash, err := hasher.Hash("Race-student-password-123!")
	if err != nil {
		t.Fatalf("hash race activation password: %v", err)
	}
	raceFingerprint, err := security.Fingerprint(map[string]string{"race": "activation-reissue"})
	if err != nil {
		t.Fatalf("fingerprint activation/reissue race: %v", err)
	}
	lockTx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin advisory-lock harness: %v", err)
	}
	lockKey := integrationAdvisoryLockKey("activation", owner.TenantID, "student:"+raceStudent.StudentID)
	if _, err := lockTx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, lockKey); err != nil {
		_ = lockTx.Rollback(ctx)
		t.Fatalf("hold activation subject lock: %v", err)
	}
	activationDone := make(chan error, 1)
	reissueDone := make(chan error, 1)
	go func() {
		activationDone <- store.CompleteActivation(ctx, core.ActivationCompleteCommand{
			TokenDigest: codec.Digest(raceToken), Phone: "+77002000304",
			PasswordHash: racePasswordHash, IdempotencyKey: "pg-race-activation",
			PayloadFingerprint: raceFingerprint, Now: time.Now().UTC(),
		})
	}()
	go func() {
		_, _, raceErr := service.IssueInvitation(ctx, owner, raceStudent.StudentID, "pg-race-reissue", core.InvitationReissue)
		reissueDone <- raceErr
	}()
	select {
	case raceErr := <-activationDone:
		_ = lockTx.Rollback(ctx)
		t.Fatalf("activation bypassed held subject lock: %v", raceErr)
	case raceErr := <-reissueDone:
		_ = lockTx.Rollback(ctx)
		t.Fatalf("reissue bypassed held subject lock: %v", raceErr)
	case <-time.After(75 * time.Millisecond):
	}
	if err := lockTx.Commit(ctx); err != nil {
		t.Fatalf("release activation subject lock: %v", err)
	}
	activationErr := <-activationDone
	reissueErr := <-reissueDone
	if (activationErr == nil) == (reissueErr == nil) {
		t.Fatalf("activation/reissue outcomes activation=%v reissue=%v; want exactly one success", activationErr, reissueErr)
	}
	if activationErr != nil && !core.IsCode(activationErr, core.CodeInvalidActivation) {
		t.Fatalf("activation race error = %v", activationErr)
	}
	if reissueErr != nil && !core.IsCode(reissueErr, core.CodeInvalidState) {
		t.Fatalf("reissue race error = %v", reissueErr)
	}
	assertActivationRaceState(t, ctx, pool, owner.TenantID, raceStudent.AccountID, raceStudent.StudentID, activationErr == nil)

	// Refresh rotation and logout serialize on the original session row. No
	// interleaving may leave an active descendant in that session family.
	raceSession, err := service.SignIn(ctx, "+77002000001", "Owner-password-123!")
	if err != nil {
		t.Fatalf("sign in refresh/logout race: %v", err)
	}
	var raceFamilyID string
	if err := pool.QueryRow(ctx, `SELECT family_id FROM sessions WHERE access_digest = $1`, codec.Digest(raceSession.AccessToken)).Scan(&raceFamilyID); err != nil {
		t.Fatalf("read refresh/logout race family: %v", err)
	}
	refreshStart := make(chan struct{})
	refreshDone := make(chan struct {
		tokens core.SessionTokens
		err    error
	}, 1)
	logoutDone := make(chan error, 1)
	go func() {
		<-refreshStart
		tokens, refreshErr := service.Refresh(ctx, raceSession.RefreshToken)
		refreshDone <- struct {
			tokens core.SessionTokens
			err    error
		}{tokens, refreshErr}
	}()
	go func() {
		<-refreshStart
		logoutDone <- service.SignOut(ctx, raceSession.AccessToken)
	}()
	close(refreshStart)
	refreshOutcome := <-refreshDone
	if logoutErr := <-logoutDone; logoutErr != nil {
		t.Fatalf("concurrent family logout: %v", logoutErr)
	}
	if refreshOutcome.err != nil && !core.IsCode(refreshOutcome.err, core.CodeUnauthenticated) {
		t.Fatalf("concurrent refresh: %v", refreshOutcome.err)
	}
	var activeFamilySessions int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM sessions
		WHERE tenant_id = $1 AND account_id = $2 AND family_id = $3 AND status = 'active'
	`, owner.TenantID, owner.AccountID, raceFamilyID).Scan(&activeFamilySessions); err != nil {
		t.Fatalf("count active race family sessions: %v", err)
	}
	if activeFamilySessions != 0 {
		t.Fatalf("refresh/logout race left %d active family session(s)", activeFamilySessions)
	}
	if refreshOutcome.err == nil {
		if _, err := service.Authenticate(ctx, refreshOutcome.tokens.AccessToken); !core.IsCode(err, core.CodeUnauthenticated) {
			t.Fatalf("refresh/logout race left access token active: %v", err)
		}
		if _, err := service.Refresh(ctx, refreshOutcome.tokens.RefreshToken); !core.IsCode(err, core.CodeUnauthenticated) {
			t.Fatalf("refresh/logout race left refresh token active: %v", err)
		}
	}

	if err := service.RevokeDelegation(ctx, owner, app.RevokeDelegationInput{
		DelegationID: delegation.ID, Reason: "Integration test complete",
		CurrentPassword: "Owner-password-123!", IdempotencyKey: "pg-revoke-admin",
	}); err != nil {
		t.Fatalf("revoke Administrator: %v", err)
	}
	view, err = service.BootstrapView(ctx, administrator)
	if err != nil || !reflect.DeepEqual(view.Permissions, core.LessonPermissionSetForRoles([]core.Role{core.RoleAdministrator})) || len(view.AccessProfiles) != 0 {
		t.Fatalf("revoked Administrator access = %#v, %v", view, err)
	}
	if _, err := service.CreateStudent(ctx, administrator, primaryStudentInput); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("revoked Administrator exact replay = %v", err)
	}
}

func integrationAdvisoryLockKey(namespace string, parts ...string) string {
	digest := sha256.New()
	writePart := func(part string) {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(part)))
		_, _ = digest.Write(length[:])
		_, _ = digest.Write([]byte(part))
	}
	writePart(namespace)
	for _, part := range parts {
		writePart(part)
	}
	return "belcanto:lock:v1:" + hex.EncodeToString(digest.Sum(nil))
}

func assertConstraintRejected(t *testing.T, ctx context.Context, pool *pgxpool.Pool, statement string, arguments ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, statement, arguments...); err == nil {
		t.Fatalf("database accepted invariant-breaking statement: %s", strings.Join(strings.Fields(statement), " "))
	} else {
		var postgresError *pgconn.PgError
		if !errors.As(err, &postgresError) || (postgresError.Code != "23503" && postgresError.Code != "23505" && postgresError.Code != "23514" && postgresError.Code != "23P01") {
			t.Fatalf("invariant-breaking statement failed with unexpected error %v", err)
		}
	}
}

func isolatedPool(t *testing.T, ctx context.Context, databaseURL string) (*pgxpool.Pool, string) {
	t.Helper()
	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open integration admin pool: %v", err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Fatalf("ping integration database: %v", err)
	}
	schemaID, err := security.NewID("belcanto_test")
	if err != nil {
		adminPool.Close()
		t.Fatalf("generate test schema: %v", err)
	}
	schema := strings.ReplaceAll(schemaID, "-", "_")
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		adminPool.Close()
		t.Fatalf("create test schema: %v", err)
	}
	configuration, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		_, _ = adminPool.Exec(ctx, "DROP SCHEMA "+identifier+" CASCADE")
		adminPool.Close()
		t.Fatalf("parse integration database URL: %v", err)
	}
	configuration.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, configuration)
	if err != nil {
		_, _ = adminPool.Exec(ctx, "DROP SCHEMA "+identifier+" CASCADE")
		adminPool.Close()
		t.Fatalf("open isolated integration pool: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		cleanupContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = adminPool.Exec(cleanupContext, "DROP SCHEMA "+identifier+" CASCADE")
		adminPool.Close()
	})
	return pool, schema
}

func assertDigestOnlySchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool, schema, rawToken string) {
	t.Helper()
	rows, err := pool.Query(ctx, `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = 'activation_invitations'
		  AND column_name LIKE '%token%'
		ORDER BY column_name
	`, schema)
	if err != nil {
		t.Fatalf("inspect invitation columns: %v", err)
	}
	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			rows.Close()
			t.Fatalf("scan invitation column: %v", err)
		}
		columns = append(columns, column)
	}
	rows.Close()
	if !reflect.DeepEqual(columns, []string{"token_digest"}) {
		t.Fatalf("invitation token columns = %#v, want digest only", columns)
	}
	var storedDigest []byte
	if err := pool.QueryRow(ctx, `SELECT token_digest FROM activation_invitations LIMIT 1`).Scan(&storedDigest); err != nil {
		t.Fatalf("read stored token digest: %v", err)
	}
	if bytes.Equal(storedDigest, []byte(rawToken)) {
		t.Fatal("database stored raw invitation token")
	}
	var leaked bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM audit_records WHERE metadata::text LIKE '%' || $1 || '%')
		    OR EXISTS (SELECT 1 FROM outbox_events WHERE payload::text LIKE '%' || $1 || '%')
		    OR EXISTS (SELECT 1 FROM idempotency_records WHERE response_json::text LIKE '%' || $1 || '%')
	`, rawToken).Scan(&leaked); err != nil {
		t.Fatalf("inspect audit, outbox, and idempotency token leakage: %v", err)
	}
	if leaked {
		t.Fatal("raw invitation token leaked into audit or outbox")
	}
}

func activateIntegration(t *testing.T, ctx context.Context, service *app.Service, link, phone, password, key string) {
	t.Helper()
	if err := service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: integrationToken(t, link), Phone: phone, Password: password,
		IdempotencyKey: key,
	}); err != nil {
		t.Fatalf("complete integration activation: %v", err)
	}
}

func integrationPrincipal(t *testing.T, ctx context.Context, service *app.Service, phone, password string) core.Principal {
	t.Helper()
	tokens, err := service.SignIn(ctx, phone, password)
	if err != nil {
		t.Fatalf("integration sign in: %v", err)
	}
	principal, err := service.Authenticate(ctx, tokens.AccessToken)
	if err != nil {
		t.Fatalf("integration authenticate: %v", err)
	}
	return principal
}

func integrationToken(t *testing.T, link string) string {
	t.Helper()
	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse integration activation link: %v", err)
	}
	values, err := url.ParseQuery(parsed.Fragment)
	if err != nil {
		t.Fatalf("parse integration activation fragment: %v", err)
	}
	return values.Get("token")
}

func accountIDForToken(t *testing.T, ctx context.Context, pool *pgxpool.Pool, codec *security.TokenCodec, rawToken string) string {
	t.Helper()
	var accountID string
	if err := pool.QueryRow(ctx, `
		SELECT account_id FROM activation_invitations WHERE token_digest = $1
	`, codec.Digest(rawToken)).Scan(&accountID); err != nil {
		t.Fatalf("read pending account for activation token: %v", err)
	}
	return accountID
}

func assertRoleGrantCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID, accountID string, role core.Role, want int) {
	t.Helper()
	var count int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM role_grants
		WHERE tenant_id = $1 AND account_id = $2 AND role_type = $3
	`, tenantID, accountID, string(role)).Scan(&count); err != nil {
		t.Fatalf("count %s role grants: %v", role, err)
	}
	if count != want {
		t.Fatalf("%s role grants = %d, want %d", role, count, want)
	}
}

func assertOperatorAudit(t *testing.T, ctx context.Context, pool *pgxpool.Pool, targetID, operator, reason string) {
	t.Helper()
	var count int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM audit_records
		WHERE action = 'BootstrapActivationInvitationReissued'
		  AND target_id = $1 AND operator_identifier = $2 AND reason_code = $3
		  AND actor_account_id IS NULL
	`, targetID, operator, reason).Scan(&count); err != nil {
		t.Fatalf("inspect bootstrap recovery audit: %v", err)
	}
	if count != 1 {
		t.Fatalf("bootstrap recovery audit rows = %d, want 1", count)
	}
}

func assertActivationRaceState(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID, accountID, studentID string, activationWon bool) {
	t.Helper()
	var accountStatus string
	var issued int
	var consumed int
	if err := pool.QueryRow(ctx, `
		SELECT status FROM accounts WHERE tenant_id = $1 AND id = $2
	`, tenantID, accountID).Scan(&accountStatus); err != nil {
		t.Fatalf("read activation race account: %v", err)
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE status = 'issued'),
		       count(*) FILTER (WHERE status = 'consumed')
		FROM activation_invitations
		WHERE tenant_id = $1 AND student_id = $2
	`, tenantID, studentID).Scan(&issued, &consumed); err != nil {
		t.Fatalf("read activation race invitations: %v", err)
	}
	if activationWon {
		if accountStatus != "active" || issued != 0 || consumed != 1 {
			t.Fatalf("activation-winning race state account=%s issued=%d consumed=%d", accountStatus, issued, consumed)
		}
		return
	}
	if accountStatus != "pending_activation" || issued != 1 || consumed != 0 {
		t.Fatalf("reissue-winning race state account=%s issued=%d consumed=%d", accountStatus, issued, consumed)
	}
}
