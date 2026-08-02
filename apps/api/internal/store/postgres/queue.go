package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/jackc/pgx/v5"
)

func (s *Store) ListStudentOnboarding(ctx context.Context, principal core.Principal, now time.Time) ([]core.StudentOnboardingItem, error) {
	result := make([]core.StudentOnboardingItem, 0)
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		delegationID, onboardingAllowed, err := onboardingAuthority(ctx, tx, principal.TenantID, principal.AccountID, now)
		if err != nil {
			return err
		}
		teacherAllowed, err := hasActiveRole(ctx, tx, principal.TenantID, principal.AccountID, core.RoleTeacher)
		if err != nil {
			return err
		}
		if !onboardingAllowed && !teacherAllowed {
			return core.E(core.CodeForbidden, "student onboarding read permission is required", nil)
		}
		teacherOnly := !onboardingAllowed && teacherAllowed
		projectionAt, err := currentAssignmentProjectionTime(ctx, tx, principal.TenantID, now)
		if err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `
			SELECT s.id, p.full_name, s.enrollment_reference,
			       ta.teacher_account_id, s.version, a.status,
			       EXISTS (
			           SELECT 1 FROM first_minute_revisions fm
			           WHERE fm.tenant_id = s.tenant_id AND fm.student_id = s.id
			       ) AS has_first_minute,
			       current_invitation.id, current_invitation.expires_at
			FROM students s
			JOIN people p ON p.tenant_id = s.tenant_id AND p.id = s.person_id
			JOIN accounts a ON a.tenant_id = s.tenant_id AND a.id = s.account_id
			JOIN teacher_assignments ta
			  ON ta.tenant_id = s.tenant_id AND ta.student_id = s.id AND ta.status = 'active'
			 AND ta.effective_from <= $3
			 AND (ta.effective_until IS NULL OR $3 < ta.effective_until)
			LEFT JOIN LATERAL (
			    SELECT i.id, i.expires_at
			    FROM activation_invitations i
			    WHERE i.tenant_id = s.tenant_id AND i.student_id = s.id
			      AND i.status = 'issued' AND i.expires_at > $5
			    ORDER BY i.issued_at DESC, i.id DESC
			    LIMIT 1
			) current_invitation ON true
			WHERE s.tenant_id = $1 AND s.status = 'active'
			  AND (NOT $4::boolean OR ta.teacher_account_id = $2)
			ORDER BY p.full_name, s.id
		`, principal.TenantID, principal.AccountID, projectionAt, teacherOnly, now)
		if err != nil {
			return fmt.Errorf("list student onboarding queue: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var item core.StudentOnboardingItem
			var accountStatus string
			var hasFirstMinute bool
			var invitationID *string
			var invitationExpiresAt *time.Time
			if err := rows.Scan(
				&item.StudentID,
				&item.FullName,
				&item.EnrollmentReference,
				&item.TeacherAccountID,
				&item.StudentVersion,
				&accountStatus,
				&hasFirstMinute,
				&invitationID,
				&invitationExpiresAt,
			); err != nil {
				return fmt.Errorf("scan student onboarding item: %w", err)
			}
			if invitationID != nil {
				item.InvitationID = *invitationID
				item.InvitationExpiresAt = invitationExpiresAt
			}
			item.OnboardingState = onboardingState(accountStatus, hasFirstMinute, invitationID != nil)
			if item.OnboardingState == core.OnboardingActivated {
				item.InvitationID = ""
				item.InvitationExpiresAt = nil
			}
			result = append(result, item)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate student onboarding queue: %w", err)
		}
		if err := appendAudit(ctx, tx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			delegationID: delegationID, action: "StudentOnboardingListed",
			targetType: "student_onboarding_queue", targetID: "queue",
			decision: "allow", at: now,
		}); err != nil {
			return err
		}
		return nil
	})
	if core.IsCode(err, core.CodeForbidden) {
		s.recordDenied(ctx, auditInput{
			tenantID: principal.TenantID, actorID: principal.AccountID,
			action: "StudentOnboardingListed", targetType: "student_onboarding_queue",
			targetID: "queue", reason: "student_onboarding_read_not_allowed", at: now,
		})
	}
	return result, err
}

func onboardingState(accountStatus string, hasFirstMinute, hasInvitation bool) core.OnboardingState {
	if accountStatus != "pending_activation" {
		return core.OnboardingActivated
	}
	if hasInvitation {
		return core.OnboardingInvited
	}
	if hasFirstMinute {
		return core.OnboardingReadyToInvite
	}
	return core.OnboardingAwaitingFirstMinute
}
