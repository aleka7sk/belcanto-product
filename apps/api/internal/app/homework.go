package app

import (
	"context"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// L.3 homework and practice (domain/homework.md Approved 1.0.0). The
// service validates shapes; lifecycle authority lives in the store.

const (
	maxHomeworkTasks       = 10
	maxHomeworkAttachments = 10
	maxSubmissionMedia     = 10
)

type CreateHomeworkInput struct {
	OccurrenceID       string
	StudentID          string
	Goal               string
	ReadinessCriteria  string
	DueAt              *time.Time
	Tasks              []core.HomeworkTaskInput
	AttachmentMediaIDs []string
	Assign             bool
	IdempotencyKey     string
}

func validateMediaIDs(field string, ids []string, limit int) ([]string, error) {
	if len(ids) > limit {
		return nil, core.E(core.CodeInvalidInput, "too many media references", nil)
	}
	result := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		normalized, err := security.ValidateIdentifier(field, id, 128)
		if err != nil {
			return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
		if seen[normalized] {
			return nil, core.E(core.CodeInvalidInput, "media references must be unique", nil)
		}
		seen[normalized] = true
		result = append(result, normalized)
	}
	return result, nil
}

func (s *Service) CreateHomework(ctx context.Context, principal core.Principal, input CreateHomeworkInput) (core.HomeworkAssignment, error) {
	occurrenceID, err := security.ValidateIdentifier("occurrenceId", input.OccurrenceID, 128)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	studentID, err := security.ValidateIdentifier("studentId", input.StudentID, 128)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	goal, err := security.ValidateText("goal", input.Goal, 1, 2000)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	readiness := ""
	if input.ReadinessCriteria != "" {
		readiness, err = security.ValidateText("readinessCriteria", input.ReadinessCriteria, 1, 1000)
		if err != nil {
			return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	now := s.clock.Now()
	if input.DueAt != nil && !input.DueAt.After(now) {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, "the deadline must be in the future", nil)
	}
	if len(input.Tasks) > maxHomeworkTasks {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, "at most 10 tasks per homework", nil)
	}
	tasks := make([]core.HomeworkTaskInput, 0, len(input.Tasks))
	taskIDs := make([]string, 0, len(input.Tasks))
	for _, task := range input.Tasks {
		title, titleErr := security.ValidateText("task title", task.Title, 1, 200)
		if titleErr != nil {
			return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, titleErr.Error(), nil)
		}
		normalized := core.HomeworkTaskInput{Title: title}
		if task.Description != "" {
			normalized.Description, err = security.ValidateText("task description", task.Description, 1, 1000)
			if err != nil {
				return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
			}
		}
		if task.RecommendedMinutes < 0 || task.RecommendedMinutes > 600 {
			return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, "recommended minutes must be between 0 and 600", nil)
		}
		normalized.RecommendedMinutes = task.RecommendedMinutes
		if task.SkillArea != "" {
			normalized.SkillArea, err = security.ValidateText("task skill area", task.SkillArea, 1, 100)
			if err != nil {
				return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
			}
		}
		if task.SongTitle != "" {
			normalized.SongTitle, err = security.ValidateText("task song", task.SongTitle, 1, 200)
			if err != nil {
				return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
			}
		}
		taskID, idErr := security.NewID("hwt")
		if idErr != nil {
			return core.HomeworkAssignment{}, core.E(core.CodeInternal, "could not create task ids", idErr)
		}
		tasks = append(tasks, normalized)
		taskIDs = append(taskIDs, taskID)
	}
	attachments, err := validateMediaIDs("attachment mediaId", input.AttachmentMediaIDs, maxHomeworkAttachments)
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	homeworkID, err := security.NewID("hw")
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInternal, "could not create the homework id", err)
	}
	fingerprint, err := security.Fingerprint(struct {
		OccurrenceID, StudentID, Goal, Readiness string
		DueAt                                    *time.Time
		Tasks                                    []core.HomeworkTaskInput
		Attachments                              []string
		Assign                                   bool
	}{occurrenceID, studentID, goal, readiness, input.DueAt, tasks, attachments, input.Assign})
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	homework, err := s.store.CreateHomework(ctx, core.CreateHomeworkCommand{
		Principal: principal, HomeworkID: homeworkID, TaskIDs: taskIDs,
		OccurrenceID: occurrenceID, StudentID: studentID,
		Goal: goal, ReadinessCriteria: readiness, DueAt: input.DueAt,
		Tasks: tasks, AttachmentMediaIDs: attachments, Assign: input.Assign,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint, Now: now,
	})
	if err != nil {
		return core.HomeworkAssignment{}, normalizeStoreError("create homework", err)
	}
	return homework, nil
}

func (s *Service) homeworkTransitionCommand(homeworkID, reason, idempotencyKey, operation string) (core.HomeworkTransitionCommand, error) {
	normalizedID, err := security.ValidateIdentifier("homeworkId", homeworkID, 128)
	if err != nil {
		return core.HomeworkTransitionCommand{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	normalizedKey, err := security.ValidateIdempotencyKey(idempotencyKey)
	if err != nil {
		return core.HomeworkTransitionCommand{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		HomeworkID, Reason, Operation string
	}{normalizedID, reason, operation})
	if err != nil {
		return core.HomeworkTransitionCommand{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	return core.HomeworkTransitionCommand{
		HomeworkID: normalizedID, Reason: reason,
		IdempotencyKey: normalizedKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	}, nil
}

func (s *Service) AssignHomework(ctx context.Context, principal core.Principal, homeworkID, idempotencyKey string) (core.HomeworkAssignment, error) {
	command, err := s.homeworkTransitionCommand(homeworkID, "", idempotencyKey, "assign")
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	command.Principal = principal
	homework, err := s.store.AssignHomework(ctx, command)
	if err != nil {
		return core.HomeworkAssignment{}, normalizeStoreError("assign homework", err)
	}
	return homework, nil
}

func (s *Service) StartHomework(ctx context.Context, principal core.Principal, homeworkID, idempotencyKey string) (core.HomeworkAssignment, error) {
	command, err := s.homeworkTransitionCommand(homeworkID, "", idempotencyKey, "start")
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	command.Principal = principal
	homework, err := s.store.StartHomework(ctx, command)
	if err != nil {
		return core.HomeworkAssignment{}, normalizeStoreError("start homework", err)
	}
	return homework, nil
}

func (s *Service) CancelHomework(ctx context.Context, principal core.Principal, homeworkID, reason, idempotencyKey string) (core.HomeworkAssignment, error) {
	normalizedReason, err := security.ValidateText("reason", reason, 1, 500)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	command, err := s.homeworkTransitionCommand(homeworkID, normalizedReason, idempotencyKey, "cancel")
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	command.Principal = principal
	homework, err := s.store.CancelHomework(ctx, command)
	if err != nil {
		return core.HomeworkAssignment{}, normalizeStoreError("cancel homework", err)
	}
	return homework, nil
}

func (s *Service) MarkHomeworkTask(ctx context.Context, principal core.Principal, homeworkID, taskID string, done bool) (core.HomeworkAssignment, error) {
	normalizedHomework, err := security.ValidateIdentifier("homeworkId", homeworkID, 128)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	normalizedTask, err := security.ValidateIdentifier("taskId", taskID, 128)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	homework, err := s.store.MarkHomeworkTask(ctx, core.MarkHomeworkTaskCommand{
		Principal: principal, HomeworkID: normalizedHomework, TaskID: normalizedTask,
		Done: done, Now: s.clock.Now(),
	})
	if err != nil {
		return core.HomeworkAssignment{}, normalizeStoreError("mark homework task", err)
	}
	return homework, nil
}

type SubmitHomeworkInput struct {
	HomeworkID      string
	Note            string
	MediaIDs        []string
	ExpectedVersion int
	IdempotencyKey  string
}

func (s *Service) SubmitHomework(ctx context.Context, principal core.Principal, input SubmitHomeworkInput) (core.HomeworkAssignment, error) {
	homeworkID, err := security.ValidateIdentifier("homeworkId", input.HomeworkID, 128)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	note := ""
	if input.Note != "" {
		note, err = security.ValidateText("note", input.Note, 1, 1000)
		if err != nil {
			return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	mediaIDs, err := validateMediaIDs("submission mediaId", input.MediaIDs, maxSubmissionMedia)
	if err != nil {
		return core.HomeworkAssignment{}, err
	}
	if input.ExpectedVersion <= 0 {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, "expectedVersion must be positive", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	submissionID, err := security.NewID("sub")
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInternal, "could not create the submission id", err)
	}
	fingerprint, err := security.Fingerprint(struct {
		HomeworkID, Note string
		MediaIDs         []string
		ExpectedVersion  int
	}{homeworkID, note, mediaIDs, input.ExpectedVersion})
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	homework, err := s.store.SubmitHomework(ctx, core.SubmitHomeworkCommand{
		Principal: principal, SubmissionID: submissionID, HomeworkID: homeworkID,
		Note: note, MediaIDs: mediaIDs, ExpectedVersion: input.ExpectedVersion,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.HomeworkAssignment{}, normalizeStoreError("submit homework", err)
	}
	return homework, nil
}

type ReviewHomeworkInput struct {
	HomeworkID      string
	Decision        string
	Body            string
	NextStep        string
	EvidenceArea    string
	EvidenceNote    string
	ExpectedVersion int
	IdempotencyKey  string
}

func (s *Service) ReviewHomework(ctx context.Context, principal core.Principal, input ReviewHomeworkInput) (core.HomeworkAssignment, error) {
	homeworkID, err := security.ValidateIdentifier("homeworkId", input.HomeworkID, 128)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if input.Decision != core.FeedbackDecisionNeedsRevision && input.Decision != core.FeedbackDecisionAccepted {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, "decision must be needs_revision or accepted", nil)
	}
	body, err := security.ValidateText("body", input.Body, 1, 2000)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	nextStep := ""
	if input.NextStep != "" {
		nextStep, err = security.ValidateText("nextStep", input.NextStep, 1, 1000)
		if err != nil {
			return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	evidenceArea, evidenceNote := "", ""
	if (input.EvidenceArea == "") != (input.EvidenceNote == "") {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, "evidence needs both an area and a note", nil)
	}
	if input.EvidenceArea != "" {
		if input.Decision != core.FeedbackDecisionAccepted {
			return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, "evidence is recorded only with an accepted review (DEC-006)", nil)
		}
		evidenceArea, err = security.ValidateText("evidence area", input.EvidenceArea, 1, 100)
		if err != nil {
			return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
		evidenceNote, err = security.ValidateText("evidence note", input.EvidenceNote, 1, 1000)
		if err != nil {
			return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	if input.ExpectedVersion <= 0 {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, "expectedVersion must be positive", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	feedbackID, err := security.NewID("fb")
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInternal, "could not create the feedback id", err)
	}
	evidenceID := ""
	if evidenceArea != "" {
		evidenceID, err = security.NewID("evd")
		if err != nil {
			return core.HomeworkAssignment{}, core.E(core.CodeInternal, "could not create the evidence id", err)
		}
	}
	fingerprint, err := security.Fingerprint(struct {
		HomeworkID, Decision, Body, NextStep, EvidenceArea, EvidenceNote string
		ExpectedVersion                                                  int
	}{homeworkID, input.Decision, body, nextStep, evidenceArea, evidenceNote, input.ExpectedVersion})
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	homework, err := s.store.ReviewHomework(ctx, core.ReviewHomeworkCommand{
		Principal: principal, FeedbackID: feedbackID, EvidenceID: evidenceID,
		HomeworkID: homeworkID, Decision: input.Decision, Body: body, NextStep: nextStep,
		EvidenceArea: evidenceArea, EvidenceNote: evidenceNote,
		ExpectedVersion: input.ExpectedVersion,
		IdempotencyKey:  idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.HomeworkAssignment{}, normalizeStoreError("review homework", err)
	}
	return homework, nil
}

func (s *Service) GetHomework(ctx context.Context, principal core.Principal, homeworkID string) (core.HomeworkAssignment, error) {
	normalizedID, err := security.ValidateIdentifier("homeworkId", homeworkID, 128)
	if err != nil {
		return core.HomeworkAssignment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	homework, err := s.store.GetHomework(ctx, principal, normalizedID, s.clock.Now())
	if err != nil {
		return core.HomeworkAssignment{}, normalizeStoreError("read homework", err)
	}
	return homework, nil
}

func (s *Service) ListStudentHomework(ctx context.Context, principal core.Principal, studentID string) ([]core.HomeworkAssignment, error) {
	normalizedID, err := security.ValidateIdentifier("studentId", studentID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	homework, err := s.store.ListStudentHomework(ctx, principal, normalizedID, s.clock.Now())
	if err != nil {
		return nil, normalizeStoreError("list homework", err)
	}
	return homework, nil
}
