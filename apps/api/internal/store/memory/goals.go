package memory

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 goals and achievements — parity with PostgreSQL.

type studentGoal struct {
	ID               string
	TenantID         string
	StudentID        string
	Criterion        string
	Description      string
	RelatedSongID    string
	RelatedSkillArea string
	Status           string
	CompletionNote   string
	CancelReason     string
	ReplacedByGoalID string
	CreatedBy        string
	Version          int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type achievementDefinition struct {
	ID                  string
	TenantID            string
	Name                string
	Description         string
	Category            string
	EvidenceRequirement string
	Status              string
	DefinitionVersion   int
	CreatedBy           string
	CreatedAt           time.Time
	RetiredAt           *time.Time
}

type achievementAward struct {
	ID                string
	TenantID          string
	DefinitionID      string
	DefinitionVersion int
	StudentID         string
	EvidenceNote      string
	Status            string
	RevokeReason      string
	RevokedAt         *time.Time
	AwardedBy         string
	AwardedAt         time.Time
}

func (s *Store) goalView(stored *studentGoal) core.StudentGoal {
	view := core.StudentGoal{
		ID: stored.ID, StudentID: stored.StudentID, Criterion: stored.Criterion,
		Description: stored.Description, RelatedSongID: stored.RelatedSongID,
		RelatedSkillArea: stored.RelatedSkillArea, Status: stored.Status,
		CompletionNote: stored.CompletionNote, CancelReason: stored.CancelReason,
		ReplacedByGoalID: stored.ReplacedByGoalID,
		CreatedBy:        core.TeacherSummary{AccountID: stored.CreatedBy},
		Version:          stored.Version, CreatedAt: stored.CreatedAt, UpdatedAt: stored.UpdatedAt,
	}
	if account := s.accounts[stored.CreatedBy]; account != nil {
		view.CreatedBy.FullName = account.FullName
	}
	return view
}

func (s *Store) CreateGoal(_ context.Context, command core.CreateGoalCommand) (core.StudentGoal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if err := s.repertoireMarkerAuthority(principal.AccountID, principal.TenantID, command.StudentID, command.Now); err != nil {
		return core.StudentGoal{}, err
	}
	if response, ok, err := s.replay("create_goal", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.StudentGoal{}, err
		}
		var result core.StudentGoal
		if err := json.Unmarshal(response, &result); err != nil {
			return core.StudentGoal{}, core.E(core.CodeInternal, "decode idempotent goal result", err)
		}
		return result, nil
	}
	stored := &studentGoal{
		ID: command.GoalID, TenantID: principal.TenantID, StudentID: command.StudentID,
		Criterion: command.Criterion, Description: command.Description,
		RelatedSongID: command.RelatedSongID, RelatedSkillArea: command.RelatedSkillArea,
		Status: core.GoalStatusActive, CreatedBy: principal.AccountID, Version: 1,
		CreatedAt: command.Now, UpdatedAt: command.Now,
	}
	s.goals[command.GoalID] = stored
	result := s.goalView(stored)
	if err := s.completeIdempotency("create_goal", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.StudentGoal{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "GoalCreated",
		"student_goal", stored.ID, "allow", "", command.Now, nil)
	return result, nil
}

func (s *Store) CompleteGoal(_ context.Context, command core.CompleteGoalCommand) (core.StudentGoal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if response, ok, err := s.replay("complete_goal", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.StudentGoal{}, err
		}
		var result core.StudentGoal
		if err := json.Unmarshal(response, &result); err != nil {
			return core.StudentGoal{}, core.E(core.CodeInternal, "decode idempotent goal result", err)
		}
		return result, nil
	}
	stored := s.goals[command.GoalID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.StudentGoal{}, core.E(core.CodeNotFound, "goal not found", nil)
	}
	if err := s.repertoireMarkerAuthority(principal.AccountID, principal.TenantID, stored.StudentID, command.Now); err != nil {
		return core.StudentGoal{}, err
	}
	if stored.Status != core.GoalStatusActive {
		return core.StudentGoal{}, core.E(core.CodeInvalidState, "only an active goal completes", nil)
	}
	if command.ExpectedVersion != stored.Version {
		return core.StudentGoal{}, core.E(core.CodeConflict, "the goal changed; reload and retry", nil)
	}
	stored.Status = core.GoalStatusCompleted
	stored.CompletionNote = command.CompletionNote
	stored.Version++
	stored.UpdatedAt = command.Now
	result := s.goalView(stored)
	if err := s.completeIdempotency("complete_goal", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.StudentGoal{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "GoalCompleted",
		"student_goal", stored.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, "GoalCompleted", stored.ID, command.Now)
	return result, nil
}

func (s *Store) ReframeGoal(_ context.Context, command core.ReframeGoalCommand) ([]core.StudentGoal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if response, ok, err := s.replay("reframe_goal", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return nil, err
		}
		var result []core.StudentGoal
		if err := json.Unmarshal(response, &result); err != nil {
			return nil, core.E(core.CodeInternal, "decode idempotent goal result", err)
		}
		return result, nil
	}
	stored := s.goals[command.GoalID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return nil, core.E(core.CodeNotFound, "goal not found", nil)
	}
	if err := s.repertoireMarkerAuthority(principal.AccountID, principal.TenantID, stored.StudentID, command.Now); err != nil {
		return nil, err
	}
	if stored.Status != core.GoalStatusActive {
		return nil, core.E(core.CodeInvalidState, "only an active goal is reframed", nil)
	}
	if command.ExpectedVersion != stored.Version {
		return nil, core.E(core.CodeConflict, "the goal changed; reload and retry", nil)
	}
	stored.Status = core.GoalStatusCancelled
	stored.CancelReason = command.Reason
	stored.Version++
	stored.UpdatedAt = command.Now
	goals := []core.StudentGoal{}
	if command.NewCriterion != "" {
		replacement := &studentGoal{
			ID: command.NewGoalID, TenantID: principal.TenantID, StudentID: stored.StudentID,
			Criterion: command.NewCriterion, Description: command.NewDescription,
			Status: core.GoalStatusActive, CreatedBy: principal.AccountID, Version: 1,
			CreatedAt: command.Now, UpdatedAt: command.Now,
		}
		s.goals[command.NewGoalID] = replacement
		stored.ReplacedByGoalID = command.NewGoalID
		goals = append([]core.StudentGoal{s.goalView(stored)}, s.goalView(replacement))
	} else {
		goals = append(goals, s.goalView(stored))
	}
	if err := s.completeIdempotency("reframe_goal", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, goals); err != nil {
		return nil, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "GoalReframed",
		"student_goal", stored.ID, "allow", command.Reason, command.Now, nil)
	return goals, nil
}

func (s *Store) ListStudentGoals(_ context.Context, principal core.Principal, studentID string) ([]core.StudentGoal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.growthViewerScope(principal, studentID); err != nil {
		return nil, err
	}
	result := []core.StudentGoal{}
	for _, stored := range s.goals {
		if stored.TenantID != principal.TenantID || stored.StudentID != studentID {
			continue
		}
		result = append(result, s.goalView(stored))
	}
	sort.Slice(result, func(left, right int) bool {
		if !result[left].UpdatedAt.Equal(result[right].UpdatedAt) {
			return result[left].UpdatedAt.After(result[right].UpdatedAt)
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}

// growthViewerScope: student self, teachers, managers.
func (s *Store) growthViewerScope(principal core.Principal, studentID string) error {
	manager, isSelf := s.journalViewerScope(principal, studentID)
	if manager || isSelf {
		return nil
	}
	if assignment := s.assignmentAt(studentID, timeNowFallback()); assignment != nil &&
		assignment.TeacherAccountID == principal.AccountID {
		return nil
	}
	for _, stored := range s.journals {
		if stored.TenantID == principal.TenantID && stored.StudentID == studentID &&
			stored.TeacherAccountID == principal.AccountID {
			return nil
		}
	}
	return core.E(core.CodeForbidden, "goals are visible to the Student and assigned staff", nil)
}

func (s *Store) definitionView(stored *achievementDefinition) core.AchievementDefinition {
	view := core.AchievementDefinition{
		ID: stored.ID, Name: stored.Name, Description: stored.Description,
		Category: stored.Category, EvidenceRequirement: stored.EvidenceRequirement,
		Status: stored.Status, DefinitionVersion: stored.DefinitionVersion,
		CreatedAt: stored.CreatedAt,
	}
	if stored.RetiredAt != nil {
		retired := *stored.RetiredAt
		view.RetiredAt = &retired
	}
	return view
}

func (s *Store) CreateAchievementDefinition(_ context.Context, command core.CreateAchievementDefinitionCommand) (core.AchievementDefinition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	if actor == nil || (actor.Roles[core.RoleOwner] == "" && actor.Roles[core.RoleAdministrator] == "") {
		return core.AchievementDefinition{}, core.E(core.CodeForbidden, "the achievement catalog is managed by the school", nil)
	}
	if response, ok, err := s.replay("create_achievement_definition", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.AchievementDefinition{}, err
		}
		var result core.AchievementDefinition
		if err := json.Unmarshal(response, &result); err != nil {
			return core.AchievementDefinition{}, core.E(core.CodeInternal, "decode idempotent definition result", err)
		}
		return result, nil
	}
	for _, existing := range s.achievementDefs {
		if existing.TenantID == principal.TenantID && existing.Name == command.Name {
			return core.AchievementDefinition{}, core.E(core.CodeConflict, "achievement definition conflicts with existing data", nil)
		}
	}
	stored := &achievementDefinition{
		ID: command.DefinitionID, TenantID: principal.TenantID,
		Name: command.Name, Description: command.Description, Category: command.Category,
		EvidenceRequirement: command.EvidenceRequirement,
		Status:              "published", DefinitionVersion: 1,
		CreatedBy: principal.AccountID, CreatedAt: command.Now,
	}
	s.achievementDefs[command.DefinitionID] = stored
	result := s.definitionView(stored)
	if err := s.completeIdempotency("create_achievement_definition", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.AchievementDefinition{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "AchievementDefinitionPublished",
		"achievement_definition", stored.ID, "allow", "", command.Now, nil)
	return result, nil
}

func (s *Store) RetireAchievementDefinition(_ context.Context, command core.RetireAchievementDefinitionCommand) (core.AchievementDefinition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	if actor == nil || (actor.Roles[core.RoleOwner] == "" && actor.Roles[core.RoleAdministrator] == "") {
		return core.AchievementDefinition{}, core.E(core.CodeForbidden, "the achievement catalog is managed by the school", nil)
	}
	if response, ok, err := s.replay("retire_achievement_definition", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.AchievementDefinition{}, err
		}
		var result core.AchievementDefinition
		if err := json.Unmarshal(response, &result); err != nil {
			return core.AchievementDefinition{}, core.E(core.CodeInternal, "decode idempotent definition result", err)
		}
		return result, nil
	}
	stored := s.achievementDefs[command.DefinitionID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.AchievementDefinition{}, core.E(core.CodeNotFound, "achievement definition not found", nil)
	}
	if stored.Status != "published" {
		return core.AchievementDefinition{}, core.E(core.CodeInvalidState, "only a published definition retires", nil)
	}
	retiredAt := command.Now
	stored.Status = "retired"
	stored.RetiredAt = &retiredAt
	result := s.definitionView(stored)
	if err := s.completeIdempotency("retire_achievement_definition", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.AchievementDefinition{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "AchievementDefinitionRetired",
		"achievement_definition", stored.ID, "allow", "", command.Now, nil)
	return result, nil
}

func (s *Store) ListAchievementDefinitions(_ context.Context, principal core.Principal) ([]core.AchievementDefinition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeAccount(principal.AccountID, principal.TenantID) == nil {
		return nil, core.E(core.CodeForbidden, "an active account is required", nil)
	}
	result := []core.AchievementDefinition{}
	for _, stored := range s.achievementDefs {
		if stored.TenantID != principal.TenantID {
			continue
		}
		result = append(result, s.definitionView(stored))
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].Category != result[right].Category {
			return result[left].Category < result[right].Category
		}
		return result[left].Name < result[right].Name
	})
	return result, nil
}

func (s *Store) awardView(stored *achievementAward) core.AchievementAward {
	view := core.AchievementAward{
		ID: stored.ID, DefinitionID: stored.DefinitionID,
		StudentID: stored.StudentID, EvidenceNote: stored.EvidenceNote,
		Status: stored.Status, RevokeReason: stored.RevokeReason,
		AwardedBy: core.TeacherSummary{AccountID: stored.AwardedBy},
		AwardedAt: stored.AwardedAt, DefinitionVersion: stored.DefinitionVersion,
	}
	if definition := s.achievementDefs[stored.DefinitionID]; definition != nil {
		view.DefinitionName = definition.Name
		view.Category = definition.Category
	}
	if account := s.accounts[stored.AwardedBy]; account != nil {
		view.AwardedBy.FullName = account.FullName
	}
	if stored.RevokedAt != nil {
		revoked := *stored.RevokedAt
		view.RevokedAt = &revoked
	}
	return view
}

func (s *Store) AwardAchievement(_ context.Context, command core.AwardAchievementCommand) (core.AchievementAward, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if err := s.repertoireMarkerAuthority(principal.AccountID, principal.TenantID, command.StudentID, command.Now); err != nil {
		return core.AchievementAward{}, err
	}
	if response, ok, err := s.replay("award_achievement", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.AchievementAward{}, err
		}
		var result core.AchievementAward
		if err := json.Unmarshal(response, &result); err != nil {
			return core.AchievementAward{}, core.E(core.CodeInternal, "decode idempotent award result", err)
		}
		return result, nil
	}
	definition := s.achievementDefs[command.DefinitionID]
	if definition == nil || definition.TenantID != principal.TenantID {
		return core.AchievementAward{}, core.E(core.CodeNotFound, "achievement definition not found", nil)
	}
	if definition.Status != "published" {
		return core.AchievementAward{}, core.E(core.CodeInvalidState, "a retired definition does not create new awards", nil)
	}
	stored := &achievementAward{
		ID: command.AwardID, TenantID: principal.TenantID,
		DefinitionID: command.DefinitionID, DefinitionVersion: definition.DefinitionVersion,
		StudentID: command.StudentID, EvidenceNote: command.EvidenceNote,
		Status: "awarded", AwardedBy: principal.AccountID, AwardedAt: command.Now,
	}
	s.awards[command.AwardID] = stored
	result := s.awardView(stored)
	if err := s.completeIdempotency("award_achievement", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.AchievementAward{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "AchievementAwarded",
		"achievement_award", stored.ID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, "AchievementAwarded", stored.ID, command.Now)
	return result, nil
}

func (s *Store) RevokeAchievement(_ context.Context, command core.RevokeAchievementCommand) (core.AchievementAward, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if response, ok, err := s.replay("revoke_achievement", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.AchievementAward{}, err
		}
		var result core.AchievementAward
		if err := json.Unmarshal(response, &result); err != nil {
			return core.AchievementAward{}, core.E(core.CodeInternal, "decode idempotent award result", err)
		}
		return result, nil
	}
	stored := s.awards[command.AwardID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.AchievementAward{}, core.E(core.CodeNotFound, "award not found", nil)
	}
	if err := s.repertoireMarkerAuthority(principal.AccountID, principal.TenantID, stored.StudentID, command.Now); err != nil {
		return core.AchievementAward{}, err
	}
	if stored.Status != "awarded" {
		return core.AchievementAward{}, core.E(core.CodeInvalidState, "the award is already revoked", nil)
	}
	revokedAt := command.Now
	stored.Status = "revoked"
	stored.RevokeReason = command.Reason
	stored.RevokedAt = &revokedAt
	result := s.awardView(stored)
	if err := s.completeIdempotency("revoke_achievement", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.AchievementAward{}, err
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "AchievementRevoked",
		"achievement_award", stored.ID, "allow", command.Reason, command.Now, nil)
	return result, nil
}

func (s *Store) ListStudentAwards(_ context.Context, principal core.Principal, studentID string) ([]core.AchievementAward, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.growthViewerScope(principal, studentID); err != nil {
		return nil, err
	}
	result := []core.AchievementAward{}
	for _, stored := range s.awards {
		if stored.TenantID != principal.TenantID || stored.StudentID != studentID {
			continue
		}
		result = append(result, s.awardView(stored))
	}
	sort.Slice(result, func(left, right int) bool {
		if !result[left].AwardedAt.Equal(result[right].AwardedAt) {
			return result[left].AwardedAt.After(result[right].AwardedAt)
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}
