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

	ListVerifiedContacts(context.Context, core.Principal) ([]core.VerifiedContact, error)
	StartContactChange(context.Context, core.StartContactChangeCommand) error
	ConfirmContactChange(context.Context, core.ConfirmContactChangeCommand) (core.VerifiedContact, error)
	TwofaStatus(context.Context, core.Principal) (core.TwofaStatus, error)
	TwofaSecret(context.Context, string, string) (core.TwofaSecretRecord, error)
	StartTwofaEnrollment(context.Context, core.StartTwofaEnrollmentCommand) error
	ConfirmTwofaEnrollment(context.Context, core.ConfirmTwofaEnrollmentCommand) error
	DisableTwofa(context.Context, core.DisableTwofaCommand) error
	CreateTwofaChallenge(context.Context, core.CreateTwofaChallengeCommand) error
	TwofaChallengeByDigest(context.Context, []byte, time.Time) (core.TwofaChallengeRecord, error)
	ConsumeTwofaChallenge(context.Context, []byte, time.Time) error
	FailTwofaChallenge(context.Context, []byte, time.Time) error
	TryConsumeRecoveryCode(context.Context, string, string, []byte, time.Time) (bool, error)

	ProfileView(context.Context, core.Principal) (core.ProfileView, error)
	UpdateProfile(context.Context, core.UpdateProfileCommand) (core.ProfileView, error)

	ListPolicies(context.Context, core.Principal) ([]core.PolicyVersion, error)
	AcceptPolicy(context.Context, core.AcceptPolicyCommand) error
	PrivacySettings(context.Context, core.Principal) (core.PrivacySettings, error)
	UpdatePrivacySettings(context.Context, core.UpdatePrivacySettingsCommand) (core.PrivacySettings, error)
	CreateDataExport(context.Context, core.CreateDataExportCommand) (core.DataExportRequest, error)
	ListDataExports(context.Context, core.Principal) ([]core.DataExportRequest, error)
	DeletionRequest(context.Context, core.Principal) (core.DeletionRequest, error)
	CreateDeletionRequest(context.Context, core.CreateDeletionRequestCommand) (core.DeletionRequest, error)
	CancelDeletionRequest(context.Context, core.CancelDeletionRequestCommand) (core.DeletionRequest, error)

	ActivationProgress(context.Context, []byte, time.Time) (core.ActivationProgressView, error)
	SetActivationPassword(context.Context, core.SetActivationPasswordCommand) error
	StartActivationContact(context.Context, core.StartActivationContactCommand) error
	VerifyActivationContact(context.Context, core.VerifyActivationContactCommand) error
	SetActivationTwofa(context.Context, core.SetActivationTwofaCommand) error
	ActivationTwofaSecret(context.Context, []byte, time.Time) ([]byte, error)
	ConfirmActivationTwofa(context.Context, core.ConfirmActivationTwofaCommand) error
	FinishActivation(context.Context, core.FinishActivationCommand) error

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
	CreateRoom(context.Context, core.CreateRoomCommand) (core.Room, error)
	ListRooms(context.Context, core.Principal) ([]core.Room, error)
	CreateCoreLessonSeries(context.Context, core.CreateCoreLessonSeriesCommand) (core.CoreLessonSeries, error)
	ListCoreLessonSeries(context.Context, core.Principal) ([]core.CoreLessonSeries, error)
	GetCoreLessonSeries(context.Context, core.Principal, string) (core.CoreLessonSeries, error)
	GenerateSeriesOccurrences(context.Context, core.GenerateSeriesOccurrencesCommand) (core.SeriesOccurrenceGenerationResult, error)

	SaveJournalDraft(context.Context, core.SaveJournalDraftCommand) (core.LessonJournal, error)
	PublishJournal(context.Context, core.PublishJournalCommand) (core.LessonJournal, error)
	GetJournal(context.Context, core.Principal, string, string) (core.LessonJournal, error)
	ListStudentJournals(context.Context, core.Principal, string) ([]core.LessonJournal, error)
	ListProgressEvidence(context.Context, core.Principal, string) ([]core.ProgressEvidence, error)

	CreateMediaObject(context.Context, core.CreateMediaCommand) (core.MediaObject, error)
	AppendMediaChunk(context.Context, core.AppendMediaChunkCommand) (core.MediaObject, error)
	GetMediaObject(context.Context, core.Principal, string) (core.MediaObject, error)
	MediaContent(context.Context, string, string) ([]byte, string, error)

	CreateHomework(context.Context, core.CreateHomeworkCommand) (core.HomeworkAssignment, error)
	AssignHomework(context.Context, core.HomeworkTransitionCommand) (core.HomeworkAssignment, error)
	StartHomework(context.Context, core.HomeworkTransitionCommand) (core.HomeworkAssignment, error)
	CancelHomework(context.Context, core.HomeworkTransitionCommand) (core.HomeworkAssignment, error)
	MarkHomeworkTask(context.Context, core.MarkHomeworkTaskCommand) (core.HomeworkAssignment, error)
	SubmitHomework(context.Context, core.SubmitHomeworkCommand) (core.HomeworkAssignment, error)
	ReviewHomework(context.Context, core.ReviewHomeworkCommand) (core.HomeworkAssignment, error)
	GetHomework(context.Context, core.Principal, string, time.Time) (core.HomeworkAssignment, error)
	ListStudentHomework(context.Context, core.Principal, string, time.Time) ([]core.HomeworkAssignment, error)

	CreateRescheduleRequest(context.Context, core.CreateRescheduleRequestCommand) (core.RescheduleRequest, error)
	ListRescheduleRequests(context.Context, core.Principal) ([]core.RescheduleRequest, error)
	DecideRescheduleRequest(context.Context, core.DecideRescheduleRequestCommand) (core.RescheduleRequest, error)
	WithdrawRescheduleRequest(context.Context, core.WithdrawRescheduleRequestCommand) (core.RescheduleRequest, error)

	CreateEventCategory(context.Context, core.CreateEventCategoryCommand) (core.EventCategory, error)
	ListEventCategories(context.Context, core.Principal) ([]core.EventCategory, error)
	CreateEventSeries(context.Context, core.CreateEventSeriesCommand) (core.EventSeries, error)
	GetEventSeries(context.Context, core.Principal, string) (core.EventSeries, error)
	GenerateEventOccurrences(context.Context, core.GenerateEventOccurrencesCommand) (core.SeriesOccurrenceGenerationResult, error)
	CreateEventOccurrence(context.Context, core.CreateEventOccurrenceCommand) (core.EventOccurrence, error)
	ListEventOccurrences(context.Context, core.Principal, core.EventListQuery) ([]core.EventOccurrence, error)
	GetEventOccurrence(context.Context, core.Principal, string) (core.EventOccurrence, error)
	RsvpToEvent(context.Context, core.EventSeatCommand) (core.EventOccurrence, error)
	CancelEventRsvp(context.Context, core.EventSeatCommand) (core.EventOccurrence, error)
	JoinEventWaitlist(context.Context, core.EventSeatCommand) (core.EventOccurrence, error)
	LeaveEventWaitlist(context.Context, core.EventSeatCommand) (core.EventOccurrence, error)
	ConfirmSpotOffer(context.Context, core.SpotOfferDecisionCommand) (core.EventOccurrence, error)
	DeclineSpotOffer(context.Context, core.SpotOfferDecisionCommand) (core.EventOccurrence, error)

	ScheduleLesson(context.Context, core.ScheduleLessonCommand) (core.Lesson, error)
	ListLessons(context.Context, core.Principal, core.LessonListQuery, time.Time) ([]core.Lesson, error)
	GetLesson(context.Context, core.Principal, string, time.Time) (core.Lesson, error)
	ReplaceLessonTeachers(context.Context, core.ReplaceLessonTeachersCommand) (core.LessonTeacherReplacementResult, error)
	ReassignPrimaryTeachers(context.Context, core.ReassignPrimaryTeachersCommand) (core.PrimaryTeacherReassignmentResult, error)
}
