package memory

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 homework and practice — parity with PostgreSQL
// (domain/homework.md Approved 1.0.0). Completed is final, history is
// append-only, expiry is lazy and consequence-free.

type practiceSubmissionRecord struct {
	ID          string
	Attempt     int
	Note        string
	MediaIDs    []string
	SubmittedAt time.Time
}

type practiceFeedbackRecord struct {
	ID               string
	SubmissionID     string
	TeacherAccountID string
	Decision         string
	Body             string
	NextStep         string
	EvidenceArea     string
	EvidenceNote     string
	CreatedAt        time.Time
}

type homeworkRecord struct {
	ID                 string
	TenantID           string
	OccurrenceID       string
	StudentID          string
	TeacherAccountID   string
	Status             string
	Goal               string
	ReadinessCriteria  string
	CancelReason       string
	DueAt              *time.Time
	Tasks              []core.HomeworkTask
	AttachmentMediaIDs []string
	Submissions        []*practiceSubmissionRecord
	Feedback           []*practiceFeedbackRecord
	Version            int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (s *Store) expireHomeworkRecord(record *homeworkRecord, now time.Time) {
	if record.DueAt == nil || record.DueAt.After(now) {
		return
	}
	if record.Status == core.HomeworkStatusAssigned || record.Status == core.HomeworkStatusInProgress {
		record.Status = core.HomeworkStatusExpired
		record.Version++
		record.UpdatedAt = now
	}
}

func (s *Store) homeworkView(record *homeworkRecord) core.HomeworkAssignment {
	view := core.HomeworkAssignment{
		ID: record.ID, OccurrenceID: record.OccurrenceID, StudentID: record.StudentID,
		Teacher: core.TeacherSummary{AccountID: record.TeacherAccountID},
		Status:  record.Status, Goal: record.Goal,
		ReadinessCriteria: record.ReadinessCriteria, CancelReason: record.CancelReason,
		Version: record.Version, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
		Tasks:       append([]core.HomeworkTask{}, record.Tasks...),
		Attachments: make([]core.MediaObject, 0, len(record.AttachmentMediaIDs)),
		Submissions: make([]core.PracticeSubmission, 0, len(record.Submissions)),
		Feedback:    make([]core.PracticeFeedback, 0, len(record.Feedback)),
	}
	if record.DueAt != nil {
		due := *record.DueAt
		view.DueAt = &due
	}
	if account := s.accounts[record.TeacherAccountID]; account != nil {
		view.Teacher.FullName = account.FullName
	}
	for _, mediaID := range record.AttachmentMediaIDs {
		if stored := s.mediaObjects[mediaID]; stored != nil {
			view.Attachments = append(view.Attachments, mediaView(stored))
		}
	}
	for _, submission := range record.Submissions {
		item := core.PracticeSubmission{
			ID: submission.ID, Attempt: submission.Attempt, Note: submission.Note,
			Media:       make([]core.MediaObject, 0, len(submission.MediaIDs)),
			SubmittedAt: submission.SubmittedAt,
		}
		for _, mediaID := range submission.MediaIDs {
			if stored := s.mediaObjects[mediaID]; stored != nil {
				item.Media = append(item.Media, mediaView(stored))
			}
		}
		view.Submissions = append(view.Submissions, item)
	}
	sort.Slice(view.Submissions, func(left, right int) bool {
		return view.Submissions[left].Attempt > view.Submissions[right].Attempt
	})
	for _, feedback := range record.Feedback {
		item := core.PracticeFeedback{
			ID: feedback.ID, SubmissionID: feedback.SubmissionID,
			Teacher:  core.TeacherSummary{AccountID: feedback.TeacherAccountID},
			Decision: feedback.Decision, Body: feedback.Body, NextStep: feedback.NextStep,
			EvidenceArea: feedback.EvidenceArea, EvidenceNote: feedback.EvidenceNote,
			CreatedAt: feedback.CreatedAt,
		}
		if account := s.accounts[feedback.TeacherAccountID]; account != nil {
			item.Teacher.FullName = account.FullName
		}
		view.Feedback = append(view.Feedback, item)
	}
	sort.Slice(view.Feedback, func(left, right int) bool {
		if !view.Feedback[left].CreatedAt.Equal(view.Feedback[right].CreatedAt) {
			return view.Feedback[left].CreatedAt.After(view.Feedback[right].CreatedAt)
		}
		return view.Feedback[left].ID < view.Feedback[right].ID
	})
	return view
}

func (s *Store) validateReadyMedia(ownerAccountID string, mediaIDs []string) error {
	for _, mediaID := range mediaIDs {
		stored := s.mediaObjects[mediaID]
		if stored == nil {
			return core.E(core.CodeInvalidInput, "attached media was not found", nil)
		}
		if stored.OwnerAccountID != ownerAccountID {
			return core.E(core.CodeForbidden, "attached media must belong to the actor", nil)
		}
		if stored.Status != core.MediaStatusReady {
			return core.E(core.CodeInvalidState, "attached media must finish uploading first", nil)
		}
	}
	return nil
}

func (s *Store) replayHomework(operation string, principal core.Principal, key string, fingerprint []byte) (core.HomeworkAssignment, bool, error) {
	response, ok, err := s.replay(operation, principal.TenantID, principal.AccountID, key, fingerprint)
	if err != nil {
		return core.HomeworkAssignment{}, false, err
	}
	if !ok {
		return core.HomeworkAssignment{}, false, nil
	}
	var result core.HomeworkAssignment
	if err := json.Unmarshal(response, &result); err != nil {
		return core.HomeworkAssignment{}, false, core.E(core.CodeInternal, "decode idempotent homework result", err)
	}
	return result, true, nil
}

func (s *Store) CreateHomework(_ context.Context, command core.CreateHomeworkCommand) (core.HomeworkAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if err := s.journalTeacherAuthority(principal.AccountID, principal.TenantID, command.OccurrenceID, command.StudentID); err != nil {
		return core.HomeworkAssignment{}, err
	}
	if result, ok, err := s.replayHomework("create_homework", principal, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		return result, err
	}
	if err := s.validateReadyMedia(principal.AccountID, command.AttachmentMediaIDs); err != nil {
		return core.HomeworkAssignment{}, err
	}
	status := core.HomeworkStatusDraft
	if command.Assign {
		status = core.HomeworkStatusAssigned
	}
	record := &homeworkRecord{
		ID: command.HomeworkID, TenantID: principal.TenantID,
		OccurrenceID: command.OccurrenceID, StudentID: command.StudentID,
		TeacherAccountID: principal.AccountID, Status: status,
		Goal: command.Goal, ReadinessCriteria: command.ReadinessCriteria,
		AttachmentMediaIDs: append([]string{}, command.AttachmentMediaIDs...),
		Version:            1, CreatedAt: command.Now, UpdatedAt: command.Now,
	}
	if command.DueAt != nil {
		due := *command.DueAt
		record.DueAt = &due
	}
	for index, task := range command.Tasks {
		record.Tasks = append(record.Tasks, core.HomeworkTask{
			ID: command.TaskIDs[index], Position: index + 1, Title: task.Title,
			Description: task.Description, RecommendedMinutes: task.RecommendedMinutes,
			SkillArea: task.SkillArea, SongTitle: task.SongTitle, Status: "pending",
		})
	}
	s.homework[command.HomeworkID] = record
	result := s.homeworkView(record)
	if err := s.completeIdempotency("create_homework", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.HomeworkAssignment{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "HomeworkCreated",
		"homework", record.ID, "allow", "", command.Now, nil)
	if command.Assign {
		s.appendOutbox(principal.TenantID, "HomeworkAssigned", record.ID, command.Now)
	}
	return result, nil
}

func (s *Store) homeworkTransition(operation string, command core.HomeworkTransitionCommand,
	transition func(record *homeworkRecord) (string, error)) (core.HomeworkAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if result, ok, err := s.replayHomework(operation, principal, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		return result, err
	}
	record := s.homework[command.HomeworkID]
	if record == nil || record.TenantID != principal.TenantID {
		return core.HomeworkAssignment{}, core.E(core.CodeNotFound, "homework not found", nil)
	}
	s.expireHomeworkRecord(record, command.Now)
	action, err := transition(record)
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	record.Version++
	record.UpdatedAt = command.Now
	result := s.homeworkView(record)
	if err := s.completeIdempotency(operation, principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.HomeworkAssignment{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, action,
		"homework", record.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, action, record.ID, command.Now)
	return result, nil
}

func (s *Store) AssignHomework(_ context.Context, command core.HomeworkTransitionCommand) (core.HomeworkAssignment, error) {
	principal := command.Principal
	return s.homeworkTransition("assign_homework", command, func(record *homeworkRecord) (string, error) {
		if record.TeacherAccountID != principal.AccountID {
			return "", core.E(core.CodeForbidden, "only the homework's Teacher assigns it", nil)
		}
		if record.Status != core.HomeworkStatusDraft {
			return "", core.E(core.CodeInvalidState, "only a draft homework can be assigned", nil)
		}
		record.Status = core.HomeworkStatusAssigned
		return "HomeworkAssigned", nil
	})
}

func (s *Store) StartHomework(_ context.Context, command core.HomeworkTransitionCommand) (core.HomeworkAssignment, error) {
	principal := command.Principal
	return s.homeworkTransition("start_homework", command, func(record *homeworkRecord) (string, error) {
		if studentID := s.studentIDForAccount(principal.AccountID); studentID == "" || studentID != record.StudentID {
			return "", core.E(core.CodeForbidden, "only the assigned Student performs this action", nil)
		}
		if record.Status != core.HomeworkStatusAssigned {
			return "", core.E(core.CodeInvalidState, "the homework is not awaiting a start", nil)
		}
		record.Status = core.HomeworkStatusInProgress
		return "HomeworkStarted", nil
	})
}

func (s *Store) CancelHomework(_ context.Context, command core.HomeworkTransitionCommand) (core.HomeworkAssignment, error) {
	principal := command.Principal
	return s.homeworkTransition("cancel_homework", command, func(record *homeworkRecord) (string, error) {
		if record.TeacherAccountID != principal.AccountID {
			return "", core.E(core.CodeForbidden, "only the homework's Teacher cancels it", nil)
		}
		switch record.Status {
		case core.HomeworkStatusDraft, core.HomeworkStatusAssigned,
			core.HomeworkStatusInProgress, core.HomeworkStatusSubmitted,
			core.HomeworkStatusReviewed:
		default:
			return "", core.E(core.CodeInvalidState, "the homework is already closed", nil)
		}
		record.Status = core.HomeworkStatusCancelled
		record.CancelReason = command.Reason
		return "HomeworkCancelled", nil
	})
}

func (s *Store) MarkHomeworkTask(_ context.Context, command core.MarkHomeworkTaskCommand) (core.HomeworkAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	record := s.homework[command.HomeworkID]
	if record == nil || record.TenantID != principal.TenantID {
		return core.HomeworkAssignment{}, core.E(core.CodeNotFound, "homework not found", nil)
	}
	s.expireHomeworkRecord(record, command.Now)
	if studentID := s.studentIDForAccount(principal.AccountID); studentID == "" || studentID != record.StudentID {
		return core.HomeworkAssignment{}, core.E(core.CodeForbidden, "only the assigned Student performs this action", nil)
	}
	if record.Status != core.HomeworkStatusInProgress {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidState, "tasks are marked while the homework is in progress", nil)
	}
	found := false
	for index := range record.Tasks {
		if record.Tasks[index].ID == command.TaskID {
			record.Tasks[index].Status = "pending"
			if command.Done {
				record.Tasks[index].Status = "done"
			}
			found = true
		}
	}
	if !found {
		return core.HomeworkAssignment{}, core.E(core.CodeNotFound, "homework task not found", nil)
	}
	record.UpdatedAt = command.Now
	action := "TaskReopened"
	if command.Done {
		action = "TaskCompleted"
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, action,
		"homework_task", command.TaskID, "allow", "", command.Now, nil)
	return s.homeworkView(record), nil
}

func (s *Store) SubmitHomework(_ context.Context, command core.SubmitHomeworkCommand) (core.HomeworkAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if result, ok, err := s.replayHomework("submit_homework", principal, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		return result, err
	}
	record := s.homework[command.HomeworkID]
	if record == nil || record.TenantID != principal.TenantID {
		return core.HomeworkAssignment{}, core.E(core.CodeNotFound, "homework not found", nil)
	}
	s.expireHomeworkRecord(record, command.Now)
	if studentID := s.studentIDForAccount(principal.AccountID); studentID == "" || studentID != record.StudentID {
		return core.HomeworkAssignment{}, core.E(core.CodeForbidden, "only the assigned Student performs this action", nil)
	}
	if record.Status != core.HomeworkStatusInProgress && record.Status != core.HomeworkStatusReviewed {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidState, "the homework is not open for a submission", nil)
	}
	if command.ExpectedVersion != record.Version {
		return core.HomeworkAssignment{}, core.E(core.CodeConflict, "the homework changed; reload and retry", nil)
	}
	if err := s.validateReadyMedia(principal.AccountID, command.MediaIDs); err != nil {
		return core.HomeworkAssignment{}, err
	}
	attempt := 1
	for _, submission := range record.Submissions {
		if submission.Attempt >= attempt {
			attempt = submission.Attempt + 1
		}
	}
	record.Submissions = append(record.Submissions, &practiceSubmissionRecord{
		ID: command.SubmissionID, Attempt: attempt, Note: command.Note,
		MediaIDs: append([]string{}, command.MediaIDs...), SubmittedAt: command.Now,
	})
	record.Status = core.HomeworkStatusSubmitted
	record.Version++
	record.UpdatedAt = command.Now
	result := s.homeworkView(record)
	if err := s.completeIdempotency("submit_homework", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.HomeworkAssignment{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "HomeworkSubmitted",
		"homework", record.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, "HomeworkSubmitted", record.ID, command.Now)
	return result, nil
}

func (s *Store) ReviewHomework(_ context.Context, command core.ReviewHomeworkCommand) (core.HomeworkAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if result, ok, err := s.replayHomework("review_homework", principal, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		return result, err
	}
	record := s.homework[command.HomeworkID]
	if record == nil || record.TenantID != principal.TenantID {
		return core.HomeworkAssignment{}, core.E(core.CodeNotFound, "homework not found", nil)
	}
	if record.TeacherAccountID != principal.AccountID {
		return core.HomeworkAssignment{}, core.E(core.CodeForbidden, "only the homework's Teacher reviews it", nil)
	}
	if record.Status != core.HomeworkStatusSubmitted {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidState, "there is no submission awaiting review", nil)
	}
	if command.ExpectedVersion != record.Version {
		return core.HomeworkAssignment{}, core.E(core.CodeConflict, "the homework changed; reload and retry", nil)
	}
	var latest *practiceSubmissionRecord
	for _, submission := range record.Submissions {
		if latest == nil || submission.Attempt > latest.Attempt {
			latest = submission
		}
	}
	if latest == nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidState, "there is no submission awaiting review", nil)
	}
	for _, feedback := range record.Feedback {
		if feedback.SubmissionID == latest.ID {
			return core.HomeworkAssignment{}, core.E(core.CodeConflict, "practice feedback conflicts with existing data", nil)
		}
	}
	record.Feedback = append(record.Feedback, &practiceFeedbackRecord{
		ID: command.FeedbackID, SubmissionID: latest.ID, TeacherAccountID: principal.AccountID,
		Decision: command.Decision, Body: command.Body, NextStep: command.NextStep,
		EvidenceArea: command.EvidenceArea, EvidenceNote: command.EvidenceNote,
		CreatedAt: command.Now,
	})
	action := "HomeworkReviewed"
	record.Status = core.HomeworkStatusReviewed
	if command.Decision == core.FeedbackDecisionAccepted {
		record.Status = core.HomeworkStatusCompleted
		action = "HomeworkCompleted"
		if command.EvidenceArea != "" {
			s.evidence = append(s.evidence, &progressEvidenceRecord{
				ID: command.EvidenceID, TenantID: principal.TenantID,
				StudentID: record.StudentID, SourceKind: "practice", SourceID: latest.ID,
				Area: command.EvidenceArea, Note: command.EvidenceNote,
				RecordedBy: principal.AccountID, RecordedAt: command.Now,
			})
		}
	}
	record.Version++
	record.UpdatedAt = command.Now
	result := s.homeworkView(record)
	if err := s.completeIdempotency("review_homework", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.HomeworkAssignment{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, action,
		"homework", record.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, action, record.ID, command.Now)
	return result, nil
}

func (s *Store) GetHomework(_ context.Context, principal core.Principal, homeworkID string, now time.Time) (core.HomeworkAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := s.homework[homeworkID]
	if record == nil || record.TenantID != principal.TenantID {
		return core.HomeworkAssignment{}, core.E(core.CodeNotFound, "homework not found", nil)
	}
	s.expireHomeworkRecord(record, now)
	manager, isSelf := s.journalViewerScope(principal, record.StudentID)
	if record.TeacherAccountID == principal.AccountID || manager ||
		(isSelf && record.Status != core.HomeworkStatusDraft) {
		return s.homeworkView(record), nil
	}
	return core.HomeworkAssignment{}, core.E(core.CodeNotFound, "homework not found", nil)
}

func (s *Store) ListStudentHomework(_ context.Context, principal core.Principal, studentID string, now time.Time) ([]core.HomeworkAssignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	manager, isSelf := s.journalViewerScope(principal, studentID)
	teaches := false
	for _, record := range s.homework {
		if record.TenantID == principal.TenantID && record.StudentID == studentID &&
			record.TeacherAccountID == principal.AccountID {
			teaches = true
		}
	}
	if assignment := s.assignmentAt(studentID, now); assignment != nil &&
		assignment.TeacherAccountID == principal.AccountID {
		teaches = true
	}
	if !manager && !isSelf && !teaches {
		return nil, core.E(core.CodeForbidden, "homework is visible to the Student and assigned staff", nil)
	}
	result := []core.HomeworkAssignment{}
	for _, record := range s.homework {
		if record.TenantID != principal.TenantID || record.StudentID != studentID {
			continue
		}
		s.expireHomeworkRecord(record, now)
		isTeacher := record.TeacherAccountID == principal.AccountID
		if !manager && !isTeacher && !(isSelf && record.Status != core.HomeworkStatusDraft) {
			continue
		}
		result = append(result, s.homeworkView(record))
	}
	sort.Slice(result, func(left, right int) bool {
		if !result[left].UpdatedAt.Equal(result[right].UpdatedAt) {
			return result[left].UpdatedAt.After(result[right].UpdatedAt)
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}
