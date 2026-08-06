package app

import (
	"context"
	"slices"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// L.4 assessments (domain/assessment.md; Figma Page 27 TCH-REVIEW-*).
// A professional observation with an explicit context: never a rating,
// never a characterisation of the person. Progress trend automation is
// deliberately absent — domain/progress.md leaves the minimum-evidence
// rule an open question, so trends stay the Teacher's judgement.

// AssessmentEvidenceKinds mirror the sources the domain names: a text
// observation is evidence too.
var assessmentEvidenceKinds = []string{
	"observation", "media", "journal", "homework", "prior_assessment", "self_assessment",
}

type AssessmentContentInput struct {
	Type             string
	ContextType      string
	ContextID        string
	AssessmentDate   string
	Summary          string
	Strengths        string
	DevelopmentAreas string
	Recommendations  string
	Confidence       string
	Visibility       string
	RelatedSongID    string
	RelatedGoalID    string
	Areas            string
}

func validateAssessmentContent(input AssessmentContentInput) (core.AssessmentContent, error) {
	var content core.AssessmentContent
	if !slices.Contains(core.AssessmentTypes, input.Type) {
		return content, core.E(core.CodeInvalidInput, "type must be observation, diagnostic, formative, summative or self", nil)
	}
	if input.Type == "self" {
		// AddStudentSelfAssessment is its own future command; the
		// teacher path never authors on the student's behalf.
		return content, core.E(core.CodeInvalidInput, "self assessment is recorded by the Student, not the Teacher", nil)
	}
	if !slices.Contains(core.AssessmentContexts, input.ContextType) {
		return content, core.E(core.CodeInvalidInput, "an assessment without a context is forbidden", nil)
	}
	if !slices.Contains(core.AssessmentVisibilities, input.Visibility) {
		return content, core.E(core.CodeInvalidInput, "visibility must be teacher_only, student_visible, staff_visible or owner_analytics", nil)
	}
	if _, err := time.Parse("2006-01-02", input.AssessmentDate); err != nil {
		return content, core.E(core.CodeInvalidInput, "assessmentDate must be YYYY-MM-DD", nil)
	}
	switch input.Confidence {
	case "", "low", "medium", "high":
	default:
		return content, core.E(core.CodeInvalidInput, "confidence must be low, medium or high", nil)
	}
	summary := ""
	if input.Summary != "" {
		normalized, err := security.ValidateText("summary", input.Summary, 1, 2000)
		if err != nil {
			return content, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
		summary = normalized
	}
	optional := func(name, value string, limit int) (string, error) {
		if value == "" {
			return "", nil
		}
		normalized, err := security.ValidateText(name, value, 1, limit)
		if err != nil {
			return "", core.E(core.CodeInvalidInput, err.Error(), nil)
		}
		return normalized, nil
	}
	strengths, err := optional("strengths", input.Strengths, 2000)
	if err != nil {
		return content, err
	}
	developmentAreas, err := optional("developmentAreas", input.DevelopmentAreas, 2000)
	if err != nil {
		return content, err
	}
	recommendations, err := optional("recommendations", input.Recommendations, 2000)
	if err != nil {
		return content, err
	}
	areas, err := optional("areas", input.Areas, 500)
	if err != nil {
		return content, err
	}
	contextID := ""
	if input.ContextID != "" {
		contextID, err = security.ValidateIdentifier("contextId", input.ContextID, 128)
		if err != nil {
			return content, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	songID := ""
	if input.RelatedSongID != "" {
		songID, err = security.ValidateIdentifier("relatedSongId", input.RelatedSongID, 128)
		if err != nil {
			return content, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	goalID := ""
	if input.RelatedGoalID != "" {
		goalID, err = security.ValidateIdentifier("relatedGoalId", input.RelatedGoalID, 128)
		if err != nil {
			return content, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	return core.AssessmentContent{
		Type: input.Type, ContextType: input.ContextType, ContextID: contextID,
		AssessmentDate: input.AssessmentDate, Summary: summary,
		Strengths: strengths, DevelopmentAreas: developmentAreas,
		Recommendations: recommendations, Confidence: input.Confidence,
		Visibility: input.Visibility, RelatedSongID: songID,
		RelatedGoalID: goalID, Areas: areas,
	}, nil
}

type CreateAssessmentInput struct {
	StudentID      string
	Content        AssessmentContentInput
	IdempotencyKey string
}

func (s *Service) CreateAssessment(ctx context.Context, principal core.Principal, input CreateAssessmentInput) (core.Assessment, error) {
	studentID, err := security.ValidateIdentifier("studentId", input.StudentID, 128)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	content, err := validateAssessmentContent(input.Content)
	if err != nil {
		return core.Assessment{}, err
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		StudentID string
		Content   core.AssessmentContent
	}{studentID, content})
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	ids, err := newIDs("asmt")
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInternal, "could not create identifiers", err)
	}
	assessment, err := s.store.CreateAssessment(ctx, core.CreateAssessmentCommand{
		Principal: principal, AssessmentID: ids[0], StudentID: studentID, Content: content,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.Assessment{}, normalizeStoreError("create assessment", err)
	}
	return assessment, nil
}

type UpdateAssessmentDraftInput struct {
	AssessmentID    string
	Content         AssessmentContentInput
	ExpectedVersion int64
	IdempotencyKey  string
}

func (s *Service) UpdateAssessmentDraft(ctx context.Context, principal core.Principal, input UpdateAssessmentDraftInput) (core.Assessment, error) {
	assessmentID, err := security.ValidateIdentifier("assessmentId", input.AssessmentID, 128)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	content, err := validateAssessmentContent(input.Content)
	if err != nil {
		return core.Assessment{}, err
	}
	if input.ExpectedVersion < 0 {
		return core.Assessment{}, core.E(core.CodeInvalidInput, "expectedVersion must be at least 0", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		AssessmentID    string
		Content         core.AssessmentContent
		ExpectedVersion int64
	}{assessmentID, content, input.ExpectedVersion})
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	assessment, err := s.store.UpdateAssessmentDraft(ctx, core.UpdateAssessmentDraftCommand{
		Principal: principal, AssessmentID: assessmentID, Content: content,
		ExpectedVersion: input.ExpectedVersion,
		IdempotencyKey:  idempotencyKey, PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.Assessment{}, normalizeStoreError("update assessment", err)
	}
	return assessment, nil
}

type AddAssessmentEvidenceInput struct {
	AssessmentID   string
	Kind           string
	Note           string
	ReferenceID    string
	IdempotencyKey string
}

func (s *Service) AddAssessmentEvidence(ctx context.Context, principal core.Principal, input AddAssessmentEvidenceInput) (core.Assessment, error) {
	assessmentID, err := security.ValidateIdentifier("assessmentId", input.AssessmentID, 128)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if !slices.Contains(assessmentEvidenceKinds, input.Kind) {
		return core.Assessment{}, core.E(core.CodeInvalidInput, "kind must be observation, media, journal, homework, prior_assessment or self_assessment", nil)
	}
	note, err := security.ValidateText("note", input.Note, 1, 1000)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	referenceID := ""
	if input.ReferenceID != "" {
		referenceID, err = security.ValidateIdentifier("referenceId", input.ReferenceID, 128)
		if err != nil {
			return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		AssessmentID, Kind, Note, ReferenceID string
	}{assessmentID, input.Kind, note, referenceID})
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	ids, err := newIDs("evd")
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInternal, "could not create identifiers", err)
	}
	assessment, err := s.store.AddAssessmentEvidence(ctx, core.AddAssessmentEvidenceCommand{
		Principal: principal, AssessmentID: assessmentID, EvidenceID: ids[0],
		Kind: input.Kind, Note: note, ReferenceID: referenceID,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.Assessment{}, normalizeStoreError("add assessment evidence", err)
	}
	return assessment, nil
}

func (s *Service) RemoveAssessmentEvidence(ctx context.Context, principal core.Principal, assessmentID, evidenceID, idempotencyKey string) (core.Assessment, error) {
	normalizedAssessment, err := security.ValidateIdentifier("assessmentId", assessmentID, 128)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	normalizedEvidence, err := security.ValidateIdentifier("evidenceId", evidenceID, 128)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	normalizedKey, err := security.ValidateIdempotencyKey(idempotencyKey)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		AssessmentID, EvidenceID string
	}{normalizedAssessment, normalizedEvidence})
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	assessment, err := s.store.RemoveAssessmentEvidence(ctx, core.RemoveAssessmentEvidenceCommand{
		Principal: principal, AssessmentID: normalizedAssessment, EvidenceID: normalizedEvidence,
		IdempotencyKey: normalizedKey, PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.Assessment{}, normalizeStoreError("remove assessment evidence", err)
	}
	return assessment, nil
}

type PublishAssessmentInput struct {
	AssessmentID    string
	ExpectedVersion int64
	IdempotencyKey  string
}

func (s *Service) PublishAssessment(ctx context.Context, principal core.Principal, input PublishAssessmentInput) (core.Assessment, error) {
	assessmentID, err := security.ValidateIdentifier("assessmentId", input.AssessmentID, 128)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if input.ExpectedVersion < 0 {
		return core.Assessment{}, core.E(core.CodeInvalidInput, "expectedVersion must be at least 0", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		AssessmentID    string
		ExpectedVersion int64
	}{assessmentID, input.ExpectedVersion})
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	assessment, err := s.store.PublishAssessment(ctx, core.PublishAssessmentCommand{
		Principal: principal, AssessmentID: assessmentID, ExpectedVersion: input.ExpectedVersion,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.Assessment{}, normalizeStoreError("publish assessment", err)
	}
	return assessment, nil
}

type SupersedeAssessmentInput struct {
	AssessmentID   string
	Content        AssessmentContentInput
	IdempotencyKey string
}

func (s *Service) SupersedeAssessment(ctx context.Context, principal core.Principal, input SupersedeAssessmentInput) ([]core.Assessment, error) {
	assessmentID, err := security.ValidateIdentifier("assessmentId", input.AssessmentID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	content, err := validateAssessmentContent(input.Content)
	if err != nil {
		return nil, err
	}
	// The replacement publishes immediately, so it must already carry
	// substance (the store re-checks against the copied evidence).
	if content.Summary == "" {
		return nil, core.E(core.CodeInvalidInput, "a correcting version needs a summary", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		AssessmentID string
		Content      core.AssessmentContent
	}{assessmentID, content})
	if err != nil {
		return nil, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	ids, err := newIDs("asmt")
	if err != nil {
		return nil, core.E(core.CodeInternal, "could not create identifiers", err)
	}
	chain, err := s.store.SupersedeAssessment(ctx, core.SupersedeAssessmentCommand{
		Principal: principal, AssessmentID: assessmentID, NewAssessmentID: ids[0],
		Content: content, IdempotencyKey: idempotencyKey,
		PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return nil, normalizeStoreError("supersede assessment", err)
	}
	return chain, nil
}

func (s *Service) WithdrawAssessment(ctx context.Context, principal core.Principal, assessmentID, reason, idempotencyKey string) (core.Assessment, error) {
	normalizedID, err := security.ValidateIdentifier("assessmentId", assessmentID, 128)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	normalizedReason, err := security.ValidateText("reason", reason, 1, 500)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	normalizedKey, err := security.ValidateIdempotencyKey(idempotencyKey)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		AssessmentID, Reason string
	}{normalizedID, normalizedReason})
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	assessment, err := s.store.WithdrawAssessment(ctx, core.WithdrawAssessmentCommand{
		Principal: principal, AssessmentID: normalizedID, Reason: normalizedReason,
		IdempotencyKey: normalizedKey, PayloadFingerprint: fingerprint, Now: s.clock.Now(),
	})
	if err != nil {
		return core.Assessment{}, normalizeStoreError("withdraw assessment", err)
	}
	return assessment, nil
}

func (s *Service) ListStudentAssessments(ctx context.Context, principal core.Principal, studentID string) ([]core.Assessment, error) {
	normalizedID, err := security.ValidateIdentifier("studentId", studentID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	assessments, err := s.store.ListStudentAssessments(ctx, principal, normalizedID)
	if err != nil {
		return nil, normalizeStoreError("list assessments", err)
	}
	return assessments, nil
}

func (s *Service) GetAssessment(ctx context.Context, principal core.Principal, assessmentID string) (core.Assessment, error) {
	normalizedID, err := security.ValidateIdentifier("assessmentId", assessmentID, 128)
	if err != nil {
		return core.Assessment{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	assessment, err := s.store.GetAssessment(ctx, principal, normalizedID)
	if err != nil {
		return core.Assessment{}, normalizeStoreError("get assessment", err)
	}
	return assessment, nil
}
