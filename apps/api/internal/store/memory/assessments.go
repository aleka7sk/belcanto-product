package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.4 assessments — parity with PostgreSQL (domain/assessment.md).

type assessmentRecord struct {
	ID               string
	TenantID         string
	StudentID        string
	AuthorAccountID  string
	AuthorRole       string
	Content          core.AssessmentContent
	Status           string
	SupersededByID   string
	WithdrawalReason string
	PublishedAt      *time.Time
	Evidence         []core.AssessmentEvidence
	Version          int64
	CreatedAt        time.Time
}

func (s *Store) assessmentAssignedTeacher(tenantID, studentID, accountID string) bool {
	for _, assignment := range s.assignments[studentID] {
		if assignment.TenantID == tenantID &&
			assignment.TeacherAccountID == accountID &&
			assignment.Status == "active" {
			return true
		}
	}
	return false
}

func (s *Store) assessmentView(stored *assessmentRecord) core.Assessment {
	view := core.Assessment{
		ID:               stored.ID,
		StudentID:        stored.StudentID,
		Author:           core.TeacherSummary{AccountID: stored.AuthorAccountID},
		AuthorRole:       stored.AuthorRole,
		Type:             stored.Content.Type,
		ContextType:      stored.Content.ContextType,
		ContextID:        stored.Content.ContextID,
		AssessmentDate:   stored.Content.AssessmentDate,
		Summary:          stored.Content.Summary,
		Strengths:        stored.Content.Strengths,
		DevelopmentAreas: stored.Content.DevelopmentAreas,
		Recommendations:  stored.Content.Recommendations,
		Confidence:       stored.Content.Confidence,
		Visibility:       stored.Content.Visibility,
		RelatedSongID:    stored.Content.RelatedSongID,
		RelatedGoalID:    stored.Content.RelatedGoalID,
		Areas:            stored.Content.Areas,
		Status:           stored.Status,
		SupersededByID:   stored.SupersededByID,
		WithdrawalReason: stored.WithdrawalReason,
		Evidence:         append([]core.AssessmentEvidence(nil), stored.Evidence...),
		Version:          stored.Version,
		CreatedAt:        stored.CreatedAt,
	}
	if view.Evidence == nil {
		view.Evidence = []core.AssessmentEvidence{}
	}
	if stored.PublishedAt != nil {
		published := *stored.PublishedAt
		view.PublishedAt = &published
	}
	if account := s.accounts[stored.AuthorAccountID]; account != nil {
		view.Author.FullName = account.FullName
	}
	return view
}

func (s *Store) replayAssessment(operation string, principal core.Principal, key string, fingerprint []byte) (core.Assessment, bool, error) {
	response, ok, err := s.replay(operation, principal.TenantID, principal.AccountID, key, fingerprint)
	if err != nil || !ok {
		return core.Assessment{}, ok, err
	}
	var result core.Assessment
	if err := json.Unmarshal(response, &result); err != nil {
		return core.Assessment{}, true, core.E(core.CodeInternal, "decode idempotent assessment result", err)
	}
	return result, true, nil
}

func (s *Store) CreateAssessment(_ context.Context, command core.CreateAssessmentCommand) (core.Assessment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	studentRecord := s.students[command.StudentID]
	if studentRecord == nil || studentRecord.TenantID != principal.TenantID {
		return core.Assessment{}, core.E(core.CodeNotFound, "Student not found", nil)
	}
	if !s.assessmentAssignedTeacher(principal.TenantID, command.StudentID, principal.AccountID) {
		return core.Assessment{}, core.E(core.CodeForbidden, "assessments are written by the Student's assigned Teacher", nil)
	}
	if result, ok, err := s.replayAssessment("create_assessment", principal, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		return result, err
	}
	stored := &assessmentRecord{
		ID: command.AssessmentID, TenantID: principal.TenantID, StudentID: command.StudentID,
		AuthorAccountID: principal.AccountID, AuthorRole: "Teacher",
		Content: command.Content, Status: "draft",
		Evidence: []core.AssessmentEvidence{}, Version: 0, CreatedAt: command.Now,
	}
	s.assessments[stored.ID] = stored
	result := s.assessmentView(stored)
	if err := s.completeIdempotency("create_assessment", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.Assessment{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "AssessmentDraftCreated",
		"assessment", stored.ID, "allow", "", command.Now, nil)
	return result, nil
}

func (s *Store) UpdateAssessmentDraft(_ context.Context, command core.UpdateAssessmentDraftCommand) (core.Assessment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if result, ok, err := s.replayAssessment("update_assessment", principal, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		return result, err
	}
	stored := s.assessments[command.AssessmentID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.Assessment{}, core.E(core.CodeNotFound, "assessment not found", nil)
	}
	if stored.AuthorAccountID != principal.AccountID {
		return core.Assessment{}, core.E(core.CodeForbidden, "only the author edits a draft", nil)
	}
	if stored.Status != "draft" {
		return core.Assessment{}, core.E(core.CodeInvalidState, "published assessment content is immutable; supersede it instead", nil)
	}
	if stored.Version != command.ExpectedVersion {
		return core.Assessment{}, core.E(core.CodeConflict, "assessment was changed by someone else", nil)
	}
	stored.Content = command.Content
	stored.Version++
	result := s.assessmentView(stored)
	if err := s.completeIdempotency("update_assessment", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.Assessment{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "AssessmentDraftUpdated",
		"assessment", stored.ID, "allow", "", command.Now, nil)
	return result, nil
}

func (s *Store) assessmentEvidenceGate(principal core.Principal, assessmentID string) (*assessmentRecord, error) {
	stored := s.assessments[assessmentID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return nil, core.E(core.CodeNotFound, "assessment not found", nil)
	}
	if stored.AuthorAccountID != principal.AccountID {
		return nil, core.E(core.CodeForbidden, "only the author manages draft evidence", nil)
	}
	if stored.Status != "draft" {
		return nil, core.E(core.CodeInvalidState, "evidence is edited while the assessment is a draft", nil)
	}
	return stored, nil
}

func (s *Store) AddAssessmentEvidence(_ context.Context, command core.AddAssessmentEvidenceCommand) (core.Assessment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if result, ok, err := s.replayAssessment("add_assessment_evidence", principal, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		return result, err
	}
	stored, err := s.assessmentEvidenceGate(principal, command.AssessmentID)
	if err != nil {
		return core.Assessment{}, err
	}
	stored.Evidence = append(stored.Evidence, core.AssessmentEvidence{
		ID: command.EvidenceID, Kind: command.Kind, Note: command.Note,
		ReferenceID: command.ReferenceID, AddedAt: command.Now,
	})
	result := s.assessmentView(stored)
	if err := s.completeIdempotency("add_assessment_evidence", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.Assessment{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "AssessmentEvidenceAdded",
		"assessment", stored.ID, "allow", "", command.Now, nil)
	return result, nil
}

func (s *Store) RemoveAssessmentEvidence(_ context.Context, command core.RemoveAssessmentEvidenceCommand) (core.Assessment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if result, ok, err := s.replayAssessment("remove_assessment_evidence", principal, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		return result, err
	}
	stored, err := s.assessmentEvidenceGate(principal, command.AssessmentID)
	if err != nil {
		return core.Assessment{}, err
	}
	found := false
	remaining := stored.Evidence[:0]
	for _, entry := range stored.Evidence {
		if entry.ID == command.EvidenceID {
			found = true
			continue
		}
		remaining = append(remaining, entry)
	}
	if !found {
		return core.Assessment{}, core.E(core.CodeNotFound, "evidence not found", nil)
	}
	stored.Evidence = remaining
	result := s.assessmentView(stored)
	if err := s.completeIdempotency("remove_assessment_evidence", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.Assessment{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "AssessmentEvidenceRemoved",
		"assessment", stored.ID, "allow", "", command.Now, nil)
	return result, nil
}

func (s *Store) PublishAssessment(_ context.Context, command core.PublishAssessmentCommand) (core.Assessment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if result, ok, err := s.replayAssessment("publish_assessment", principal, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		return result, err
	}
	stored := s.assessments[command.AssessmentID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.Assessment{}, core.E(core.CodeNotFound, "assessment not found", nil)
	}
	if stored.AuthorAccountID != principal.AccountID {
		return core.Assessment{}, core.E(core.CodeForbidden, "only the author publishes an assessment", nil)
	}
	if stored.Status != "draft" {
		return core.Assessment{}, core.E(core.CodeInvalidState, "only a draft publishes", nil)
	}
	if stored.Version != command.ExpectedVersion {
		return core.Assessment{}, core.E(core.CodeConflict, "assessment was changed by someone else", nil)
	}
	content := stored.Content
	if content.Summary == "" ||
		(content.Strengths == "" && content.DevelopmentAreas == "" &&
			content.Recommendations == "" && len(stored.Evidence) == 0) {
		return core.Assessment{}, core.E(core.CodeInvalidState, "a published assessment needs a summary and at least one observation, strength, development area or recommendation", nil)
	}
	publishedAt := command.Now
	stored.Status = "published"
	stored.PublishedAt = &publishedAt
	stored.Version++
	result := s.assessmentView(stored)
	if err := s.completeIdempotency("publish_assessment", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.Assessment{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "AssessmentPublished",
		"assessment", stored.ID, "allow", "", command.Now, nil)
	if content.Visibility == "student_visible" {
		s.appendOutboxPayload(principal.TenantID, "AssessmentPublished", "assessment", stored.ID,
			map[string]any{"assessmentId": stored.ID, "studentId": stored.StudentID}, command.Now)
	}
	return result, nil
}

func (s *Store) SupersedeAssessment(_ context.Context, command core.SupersedeAssessmentCommand) ([]core.Assessment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if response, ok, err := s.replay("supersede_assessment", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return nil, err
		}
		var chain []core.Assessment
		if err := json.Unmarshal(response, &chain); err != nil {
			return nil, core.E(core.CodeInternal, "decode idempotent supersede result", err)
		}
		return chain, nil
	}
	stored := s.assessments[command.AssessmentID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return nil, core.E(core.CodeNotFound, "assessment not found", nil)
	}
	if stored.AuthorAccountID != principal.AccountID {
		return nil, core.E(core.CodeForbidden, "only the author supersedes an assessment", nil)
	}
	if stored.Status != "published" {
		return nil, core.E(core.CodeInvalidState, "only a published assessment can be superseded", nil)
	}
	publishedAt := command.Now
	replacement := &assessmentRecord{
		ID: command.NewAssessmentID, TenantID: principal.TenantID, StudentID: stored.StudentID,
		AuthorAccountID: principal.AccountID, AuthorRole: "Teacher",
		Content: command.Content, Status: "published", PublishedAt: &publishedAt,
		Evidence: make([]core.AssessmentEvidence, 0, len(stored.Evidence)),
		Version:  0, CreatedAt: command.Now,
	}
	for index, entry := range stored.Evidence {
		carried := entry
		carried.ID = fmt.Sprintf("%s.%d", command.NewAssessmentID, index+1)
		replacement.Evidence = append(replacement.Evidence, carried)
	}
	s.assessments[replacement.ID] = replacement
	stored.Status = "superseded"
	stored.SupersededByID = replacement.ID
	stored.Version++
	chain := []core.Assessment{s.assessmentView(stored), s.assessmentView(replacement)}
	if err := s.completeIdempotency("supersede_assessment", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, chain); err != nil {
		return nil, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "AssessmentSuperseded",
		"assessment", stored.ID, "allow", "", command.Now, nil)
	if command.Content.Visibility == "student_visible" {
		s.appendOutboxPayload(principal.TenantID, "AssessmentPublished", "assessment", replacement.ID,
			map[string]any{"assessmentId": replacement.ID, "studentId": replacement.StudentID}, command.Now)
	}
	return chain, nil
}

func (s *Store) WithdrawAssessment(_ context.Context, command core.WithdrawAssessmentCommand) (core.Assessment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if result, ok, err := s.replayAssessment("withdraw_assessment", principal, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		return result, err
	}
	stored := s.assessments[command.AssessmentID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.Assessment{}, core.E(core.CodeNotFound, "assessment not found", nil)
	}
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	manager := actor != nil && (actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "")
	if stored.AuthorAccountID != principal.AccountID && !manager {
		return core.Assessment{}, core.E(core.CodeForbidden, "withdrawal is the author's or the school's action", nil)
	}
	if stored.Status != "draft" && stored.Status != "published" {
		return core.Assessment{}, core.E(core.CodeInvalidState, "only a draft or published assessment withdraws", nil)
	}
	stored.Status = "withdrawn"
	stored.WithdrawalReason = command.Reason
	stored.Version++
	result := s.assessmentView(stored)
	if err := s.completeIdempotency("withdraw_assessment", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.Assessment{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "AssessmentWithdrawn",
		"assessment", stored.ID, "allow", command.Reason, command.Now, nil)
	return result, nil
}

func (s *Store) assessmentScope(principal core.Principal, studentID string) (assigned bool, manager bool, isSelf bool) {
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	if actor != nil {
		manager = actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != ""
	}
	assigned = s.assessmentAssignedTeacher(principal.TenantID, studentID, principal.AccountID)
	isSelf = s.studentIDForAccount(principal.AccountID) == studentID
	return assigned, manager, isSelf
}

func (s *Store) ListStudentAssessments(_ context.Context, principal core.Principal, studentID string) ([]core.Assessment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	assigned, manager, isSelf := s.assessmentScope(principal, studentID)
	result := []core.Assessment{}
	for _, stored := range s.assessments {
		if stored.TenantID != principal.TenantID || stored.StudentID != studentID {
			continue
		}
		view := s.assessmentView(stored)
		if !core.AssessmentVisible(view, stored.AuthorAccountID == principal.AccountID, assigned, manager, isSelf) {
			continue
		}
		result = append(result, view)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].AssessmentDate != result[right].AssessmentDate {
			return result[left].AssessmentDate > result[right].AssessmentDate
		}
		if !result[left].CreatedAt.Equal(result[right].CreatedAt) {
			return result[left].CreatedAt.After(result[right].CreatedAt)
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}

func (s *Store) GetAssessment(_ context.Context, principal core.Principal, assessmentID string) (core.Assessment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored := s.assessments[assessmentID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.Assessment{}, core.E(core.CodeNotFound, "assessment not found", nil)
	}
	assigned, manager, isSelf := s.assessmentScope(principal, stored.StudentID)
	view := s.assessmentView(stored)
	if !core.AssessmentVisible(view, stored.AuthorAccountID == principal.AccountID, assigned, manager, isSelf) {
		return core.Assessment{}, core.E(core.CodeNotFound, "assessment not found", nil)
	}
	return view, nil
}
