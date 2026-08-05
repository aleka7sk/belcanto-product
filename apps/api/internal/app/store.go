package app

import (
	"context"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

type Store interface {
	Ready(context.Context) error
	BootstrapOwner(context.Context, core.BootstrapOwnerCommand) error
	BootstrapStaff(context.Context, core.BootstrapStaffCommand) error
	ReissueBootstrapInvitation(context.Context, core.ReissueBootstrapInvitationCommand) (core.BootstrapInvitationResult, error)
	PreviewActivation(context.Context, []byte, time.Time) (core.ActivationPreview, error)
	ValidateActivation(context.Context, core.ActivationValidationCommand) (bool, error)
	CompleteActivation(context.Context, core.ActivationCompleteCommand) error

	CredentialByPhone(context.Context, string) (core.CredentialRecord, error)
	CredentialByAccount(context.Context, string) (core.CredentialRecord, error)
	CreateSession(context.Context, string, string, core.SessionMaterial) error
	PrincipalByAccessDigest(context.Context, []byte, time.Time) (core.Principal, error)
	RotateSession(context.Context, []byte, core.SessionMaterial, time.Time) (string, string, error)
	RevokeSession(context.Context, []byte, time.Time) error

	ListSessions(context.Context, core.Principal, time.Time) ([]core.SessionDevice, error)
	RevokeSessionByID(context.Context, core.RevokeSessionByIDCommand) error
	RevokeOtherSessions(context.Context, core.RevokeOtherSessionsCommand) (int, error)
	ListSecurityEvents(context.Context, core.Principal, core.SecurityEventsQuery) ([]core.SecurityEvent, error)
	CreatePasswordReset(context.Context, core.CreatePasswordResetCommand) error
	CompletePasswordReset(context.Context, core.CompletePasswordResetCommand) error

	GrantDelegation(context.Context, core.GrantDelegationCommand) (core.DelegationResult, error)
	RevokeDelegation(context.Context, core.RevokeDelegationCommand) error
	CreateStudent(context.Context, core.CreateStudentCommand) (core.StudentResult, error)
	PublishFirstMinute(context.Context, core.PublishFirstMinuteCommand) (core.FirstMinute, error)
	IssueInvitation(context.Context, core.IssueInvitationCommand) (core.InvitationResult, error)
	RevokeInvitation(context.Context, core.RevokeInvitationCommand) error
	BootstrapView(context.Context, core.Principal, time.Time) (core.BootstrapView, error)
	ListStaff(context.Context, core.Principal, core.Role, time.Time) ([]core.StaffMember, error)
	ListStudentOnboarding(context.Context, core.Principal, time.Time) ([]core.StudentOnboardingItem, error)
	ListStudents(context.Context, core.Principal, time.Time, time.Time) ([]core.StudentDirectoryItem, error)
	ScheduleLesson(context.Context, core.ScheduleLessonCommand) (core.Lesson, error)
	ListLessons(context.Context, core.Principal, core.LessonListQuery, time.Time) ([]core.Lesson, error)
	GetLesson(context.Context, core.Principal, string, time.Time) (core.Lesson, error)
	ReplaceLessonTeachers(context.Context, core.ReplaceLessonTeachersCommand) (core.LessonTeacherReplacementResult, error)
	ReassignPrimaryTeachers(context.Context, core.ReassignPrimaryTeachersCommand) (core.PrimaryTeacherReassignmentResult, error)
}
