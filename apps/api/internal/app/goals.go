package app

import (
	"context"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// L.3 goals and achievements. A goal has an explicit criterion and is
// reframed with a reason, never «failed»; completion carries a decision
// note. Awards are evidence-backed and never numeric (DEC-006).

type CreateGoalInput struct {
	StudentID        string
	Criterion        string
	Description      string
	RelatedSongID    string
	RelatedSkillArea string
	IdempotencyKey   string
}

func (s *Service) CreateGoal(ctx context.Context, principal core.Principal, input CreateGoalInput) (core.StudentGoal, error) {
	studentID, err := security.ValidateIdentifier("studentId", input.StudentID, 128)
	if err != nil {
		return core.StudentGoal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	criterion, err := security.ValidateText("criterion", input.Criterion, 1, 500)
	if err != nil {
		return core.StudentGoal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	description := ""
	if input.Description != "" {
		description, err = security.ValidateText("description", input.Description, 1, 1000)
		if err != nil {
			return core.StudentGoal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	relatedSongID := ""
	if input.RelatedSongID != "" {
		relatedSongID, err = security.ValidateIdentifier("relatedSongId", input.RelatedSongID, 128)
		if err != nil {
			return core.StudentGoal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	relatedSkillArea := ""
	if input.RelatedSkillArea != "" {
		relatedSkillArea, err = security.ValidateText("relatedSkillArea", input.RelatedSkillArea, 1, 100)
		if err != nil {
			return core.StudentGoal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.StudentGoal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	goalID, err := security.NewID("goal")
	if err != nil {
		return core.StudentGoal{}, core.E(core.CodeInternal, "could not create the goal id", err)
	}
	fingerprint, err := security.Fingerprint(struct {
		StudentID, Criterion, Description, SongID, SkillArea string
	}{studentID, criterion, description, relatedSongID, relatedSkillArea})
	if err != nil {
		return core.StudentGoal{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	goal, err := s.store.CreateGoal(ctx, core.CreateGoalCommand{
		Principal: principal, GoalID: goalID, StudentID: studentID,
		Criterion: criterion, Description: description,
		RelatedSongID: relatedSongID, RelatedSkillArea: relatedSkillArea,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.StudentGoal{}, normalizeStoreError("create goal", err)
	}
	return goal, nil
}

type CompleteGoalInput struct {
	GoalID          string
	CompletionNote  string
	ExpectedVersion int
	IdempotencyKey  string
}

func (s *Service) CompleteGoal(ctx context.Context, principal core.Principal, input CompleteGoalInput) (core.StudentGoal, error) {
	goalID, err := security.ValidateIdentifier("goalId", input.GoalID, 128)
	if err != nil {
		return core.StudentGoal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	note, err := security.ValidateText("completionNote", input.CompletionNote, 1, 500)
	if err != nil {
		return core.StudentGoal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if input.ExpectedVersion <= 0 {
		return core.StudentGoal{}, core.E(core.CodeInvalidInput, "expectedVersion must be positive", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.StudentGoal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		GoalID, Note    string
		ExpectedVersion int
	}{goalID, note, input.ExpectedVersion})
	if err != nil {
		return core.StudentGoal{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	goal, err := s.store.CompleteGoal(ctx, core.CompleteGoalCommand{
		Principal: principal, GoalID: goalID, CompletionNote: note,
		ExpectedVersion: input.ExpectedVersion,
		IdempotencyKey:  idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.StudentGoal{}, normalizeStoreError("complete goal", err)
	}
	return goal, nil
}

type ReframeGoalInput struct {
	GoalID          string
	Reason          string
	NewCriterion    string
	NewDescription  string
	ExpectedVersion int
	IdempotencyKey  string
}

func (s *Service) ReframeGoal(ctx context.Context, principal core.Principal, input ReframeGoalInput) ([]core.StudentGoal, error) {
	goalID, err := security.ValidateIdentifier("goalId", input.GoalID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	reason, err := security.ValidateText("reason", input.Reason, 1, 500)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	newCriterion := ""
	if input.NewCriterion != "" {
		newCriterion, err = security.ValidateText("newCriterion", input.NewCriterion, 1, 500)
		if err != nil {
			return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	newDescription := ""
	if input.NewDescription != "" {
		if newCriterion == "" {
			return nil, core.E(core.CodeInvalidInput, "a new description needs a new criterion", nil)
		}
		newDescription, err = security.ValidateText("newDescription", input.NewDescription, 1, 1000)
		if err != nil {
			return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	if input.ExpectedVersion <= 0 {
		return nil, core.E(core.CodeInvalidInput, "expectedVersion must be positive", nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	newGoalID := ""
	if newCriterion != "" {
		newGoalID, err = security.NewID("goal")
		if err != nil {
			return nil, core.E(core.CodeInternal, "could not create the goal id", err)
		}
	}
	fingerprint, err := security.Fingerprint(struct {
		GoalID, Reason, NewCriterion, NewDescription string
		ExpectedVersion                              int
	}{goalID, reason, newCriterion, newDescription, input.ExpectedVersion})
	if err != nil {
		return nil, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	goals, err := s.store.ReframeGoal(ctx, core.ReframeGoalCommand{
		Principal: principal, GoalID: goalID, Reason: reason,
		NewGoalID: newGoalID, NewCriterion: newCriterion, NewDescription: newDescription,
		ExpectedVersion: input.ExpectedVersion,
		IdempotencyKey:  idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return nil, normalizeStoreError("reframe goal", err)
	}
	return goals, nil
}

func (s *Service) ListStudentGoals(ctx context.Context, principal core.Principal, studentID string) ([]core.StudentGoal, error) {
	normalizedID, err := security.ValidateIdentifier("studentId", studentID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	goals, err := s.store.ListStudentGoals(ctx, principal, normalizedID)
	if err != nil {
		return nil, normalizeStoreError("list goals", err)
	}
	return goals, nil
}

type CreateAchievementDefinitionInput struct {
	Name                string
	Description         string
	Category            string
	EvidenceRequirement string
	IdempotencyKey      string
}

func (s *Service) CreateAchievementDefinition(ctx context.Context, principal core.Principal, input CreateAchievementDefinitionInput) (core.AchievementDefinition, error) {
	name, err := security.ValidateText("name", input.Name, 1, 200)
	if err != nil {
		return core.AchievementDefinition{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	description, err := security.ValidateText("description", input.Description, 1, 1000)
	if err != nil {
		return core.AchievementDefinition{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	category, err := security.ValidateText("category", input.Category, 1, 100)
	if err != nil {
		return core.AchievementDefinition{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	evidenceRequirement := ""
	if input.EvidenceRequirement != "" {
		evidenceRequirement, err = security.ValidateText("evidenceRequirement", input.EvidenceRequirement, 1, 500)
		if err != nil {
			return core.AchievementDefinition{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.AchievementDefinition{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	definitionID, err := security.NewID("achdef")
	if err != nil {
		return core.AchievementDefinition{}, core.E(core.CodeInternal, "could not create the definition id", err)
	}
	fingerprint, err := security.Fingerprint(struct {
		Name, Description, Category, Requirement string
	}{name, description, category, evidenceRequirement})
	if err != nil {
		return core.AchievementDefinition{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	definition, err := s.store.CreateAchievementDefinition(ctx, core.CreateAchievementDefinitionCommand{
		Principal: principal, DefinitionID: definitionID,
		Name: name, Description: description, Category: category,
		EvidenceRequirement: evidenceRequirement,
		IdempotencyKey:      idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.AchievementDefinition{}, normalizeStoreError("create achievement definition", err)
	}
	return definition, nil
}

func (s *Service) RetireAchievementDefinition(ctx context.Context, principal core.Principal, definitionID, idempotencyKey string) (core.AchievementDefinition, error) {
	normalizedID, err := security.ValidateIdentifier("definitionId", definitionID, 128)
	if err != nil {
		return core.AchievementDefinition{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	normalizedKey, err := security.ValidateIdempotencyKey(idempotencyKey)
	if err != nil {
		return core.AchievementDefinition{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct{ DefinitionID string }{normalizedID})
	if err != nil {
		return core.AchievementDefinition{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	definition, err := s.store.RetireAchievementDefinition(ctx, core.RetireAchievementDefinitionCommand{
		Principal: principal, DefinitionID: normalizedID,
		IdempotencyKey: normalizedKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.AchievementDefinition{}, normalizeStoreError("retire achievement definition", err)
	}
	return definition, nil
}

func (s *Service) ListAchievementDefinitions(ctx context.Context, principal core.Principal) ([]core.AchievementDefinition, error) {
	definitions, err := s.store.ListAchievementDefinitions(ctx, principal)
	if err != nil {
		return nil, normalizeStoreError("list achievement definitions", err)
	}
	return definitions, nil
}

type AwardAchievementInput struct {
	DefinitionID   string
	StudentID      string
	EvidenceNote   string
	IdempotencyKey string
}

func (s *Service) AwardAchievement(ctx context.Context, principal core.Principal, input AwardAchievementInput) (core.AchievementAward, error) {
	definitionID, err := security.ValidateIdentifier("definitionId", input.DefinitionID, 128)
	if err != nil {
		return core.AchievementAward{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	studentID, err := security.ValidateIdentifier("studentId", input.StudentID, 128)
	if err != nil {
		return core.AchievementAward{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	evidenceNote, err := security.ValidateText("evidenceNote", input.EvidenceNote, 1, 1000)
	if err != nil {
		return core.AchievementAward{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.AchievementAward{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	awardID, err := security.NewID("award")
	if err != nil {
		return core.AchievementAward{}, core.E(core.CodeInternal, "could not create the award id", err)
	}
	fingerprint, err := security.Fingerprint(struct {
		DefinitionID, StudentID, EvidenceNote string
	}{definitionID, studentID, evidenceNote})
	if err != nil {
		return core.AchievementAward{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	award, err := s.store.AwardAchievement(ctx, core.AwardAchievementCommand{
		Principal: principal, AwardID: awardID, DefinitionID: definitionID,
		StudentID: studentID, EvidenceNote: evidenceNote,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.AchievementAward{}, normalizeStoreError("award achievement", err)
	}
	return award, nil
}

func (s *Service) RevokeAchievement(ctx context.Context, principal core.Principal, awardID, reason, idempotencyKey string) (core.AchievementAward, error) {
	normalizedID, err := security.ValidateIdentifier("awardId", awardID, 128)
	if err != nil {
		return core.AchievementAward{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	normalizedReason, err := security.ValidateText("reason", reason, 1, 500)
	if err != nil {
		return core.AchievementAward{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	normalizedKey, err := security.ValidateIdempotencyKey(idempotencyKey)
	if err != nil {
		return core.AchievementAward{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		AwardID, Reason string
	}{normalizedID, normalizedReason})
	if err != nil {
		return core.AchievementAward{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	award, err := s.store.RevokeAchievement(ctx, core.RevokeAchievementCommand{
		Principal: principal, AwardID: normalizedID, Reason: normalizedReason,
		IdempotencyKey: normalizedKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.AchievementAward{}, normalizeStoreError("revoke achievement", err)
	}
	return award, nil
}

func (s *Service) ListStudentAwards(ctx context.Context, principal core.Principal, studentID string) ([]core.AchievementAward, error) {
	normalizedID, err := security.ValidateIdentifier("studentId", studentID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	awards, err := s.store.ListStudentAwards(ctx, principal, normalizedID)
	if err != nil {
		return nil, normalizeStoreError("list achievements", err)
	}
	return awards, nil
}
