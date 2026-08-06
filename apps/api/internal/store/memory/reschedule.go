package memory

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.2 reschedule and cancellation requests — parity with PostgreSQL.

type rescheduleRequest struct {
	ID                   string
	TenantID             string
	OccurrenceID         string
	RequestedByAccountID string
	Kind                 string
	ProposedStartsAt     *time.Time
	Reason               string
	Status               string
	DecidedByAccountID   string
	DecisionNote         string
	DecidedAt            *time.Time
	Version              int64
	CreatedAt            time.Time
}

func (s *Store) rescheduleView(stored *rescheduleRequest) core.RescheduleRequest {
	view := core.RescheduleRequest{
		ID: stored.ID, OccurrenceID: stored.OccurrenceID, Kind: stored.Kind,
		Reason: stored.Reason, Status: stored.Status,
		RequestedBy:  core.TeacherSummary{AccountID: stored.RequestedByAccountID},
		DecisionNote: stored.DecisionNote,
		CreatedAt:    stored.CreatedAt, Version: stored.Version,
	}
	if account := s.accounts[stored.RequestedByAccountID]; account != nil {
		view.RequestedBy.FullName = account.FullName
	}
	if stored.ProposedStartsAt != nil {
		proposed := *stored.ProposedStartsAt
		view.ProposedStartsAt = &proposed
	}
	if stored.DecidedAt != nil {
		decided := *stored.DecidedAt
		view.DecidedAt = &decided
	}
	return view
}

func (s *Store) CreateRescheduleRequest(_ context.Context, command core.CreateRescheduleRequestCommand) (core.RescheduleRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	if response, ok, err := s.replay("create_reschedule_request", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint); ok || err != nil {
		if err != nil {
			return core.RescheduleRequest{}, err
		}
		var result core.RescheduleRequest
		if err := json.Unmarshal(response, &result); err != nil {
			return core.RescheduleRequest{}, core.E(core.CodeInternal, "decode idempotent request result", err)
		}
		return result, nil
	}
	occurrence := s.lessons[command.OccurrenceID]
	if occurrence == nil || occurrence.TenantID != principal.TenantID {
		return core.RescheduleRequest{}, core.E(core.CodeNotFound, "Lesson not found", nil)
	}
	if occurrence.Status != core.LessonScheduled || !occurrence.StartsAt.After(command.Now) {
		return core.RescheduleRequest{}, core.E(core.CodeInvalidState, "only a scheduled future Lesson can be changed", nil)
	}
	if occurrence.TeacherAccountID != principal.AccountID {
		participates := false
		requesterStudent := s.studentIDForAccount(principal.AccountID)
		for _, studentID := range occurrence.StudentIDs {
			if studentID == requesterStudent && requesterStudent != "" {
				participates = true
			}
		}
		if !participates {
			return core.RescheduleRequest{}, core.E(core.CodeForbidden, "only a participant or the Lesson's Teacher can ask for a change", nil)
		}
	}
	for _, existing := range s.rescheduleRequests {
		if existing.TenantID == principal.TenantID && existing.OccurrenceID == command.OccurrenceID &&
			existing.RequestedByAccountID == principal.AccountID && existing.Status == "pending" {
			return core.RescheduleRequest{}, core.E(core.CodeConflict, "an open request already exists for this Lesson", nil)
		}
	}
	stored := &rescheduleRequest{
		ID: command.RequestID, TenantID: principal.TenantID,
		OccurrenceID: command.OccurrenceID, RequestedByAccountID: principal.AccountID,
		Kind: command.Kind, Reason: command.Reason, Status: "pending",
		Version: 0, CreatedAt: command.Now,
	}
	if command.ProposedStartsAt != nil {
		proposed := *command.ProposedStartsAt
		stored.ProposedStartsAt = &proposed
	}
	s.rescheduleRequests[stored.ID] = stored
	result := s.rescheduleView(stored)
	if err := s.completeIdempotency("create_reschedule_request", principal.TenantID, principal.AccountID, command.IdempotencyKey, command.PayloadFingerprint, result); err != nil {
		return core.RescheduleRequest{}, err
	}
	action := "LessonRescheduleRequested"
	if command.Kind == "cancellation" {
		action = "LessonCancellationRequested"
	}
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, action,
		"lesson_occurrence", command.OccurrenceID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, action, stored.ID, command.Now)
	return result, nil
}

func (s *Store) ListRescheduleRequests(_ context.Context, principal core.Principal) ([]core.RescheduleRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	manager := actor != nil && (actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "")
	result := []core.RescheduleRequest{}
	for _, stored := range s.rescheduleRequests {
		if stored.TenantID != principal.TenantID {
			continue
		}
		if !manager && stored.RequestedByAccountID != principal.AccountID {
			continue
		}
		result = append(result, s.rescheduleView(stored))
	}
	sort.Slice(result, func(left, right int) bool {
		leftPending := result[left].Status == "pending"
		rightPending := result[right].Status == "pending"
		if leftPending != rightPending {
			return leftPending
		}
		if !result[left].CreatedAt.Equal(result[right].CreatedAt) {
			return result[left].CreatedAt.After(result[right].CreatedAt)
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}

func (s *Store) DecideRescheduleRequest(_ context.Context, command core.DecideRescheduleRequestCommand) (core.RescheduleRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	actor := s.activeAccount(principal.AccountID, principal.TenantID)
	manager := actor != nil && (actor.Roles[core.RoleOwner] != "" || actor.Roles[core.RoleAdministrator] != "")
	if !manager {
		return core.RescheduleRequest{}, core.E(core.CodeForbidden, "schedule decision permission is required", nil)
	}
	stored := s.rescheduleRequests[command.RequestID]
	if stored == nil || stored.TenantID != principal.TenantID {
		return core.RescheduleRequest{}, core.E(core.CodeNotFound, "request not found", nil)
	}
	if stored.Status != "pending" {
		return core.RescheduleRequest{}, core.E(core.CodeInvalidState, "request is already resolved", nil)
	}
	if stored.Version != command.ExpectedVersion {
		return core.RescheduleRequest{}, core.E(core.CodeConflict, "request version is stale", nil)
	}
	occurrence := s.lessons[stored.OccurrenceID]
	action := "LessonRescheduleDeclined"
	if command.Approve {
		if occurrence == nil || occurrence.Status != core.LessonScheduled {
			return core.RescheduleRequest{}, core.E(core.CodeInvalidState, "Lesson is no longer scheduled", nil)
		}
		if stored.Kind == "reschedule" {
			if stored.ProposedStartsAt == nil || !stored.ProposedStartsAt.After(command.Now) {
				return core.RescheduleRequest{}, core.E(core.CodeInvalidState, "proposed time is already in the past", nil)
			}
			if s.lessonScheduleConflict(principal.TenantID, *stored.ProposedStartsAt, occurrence.DurationMinutes, occurrence.TeacherAccountID, occurrence.StudentIDs, map[string]struct{}{occurrence.ID: {}}) {
				return core.RescheduleRequest{}, core.E(core.CodeConflict, "proposed time overlaps an existing Lesson", nil)
			}
			occurrence.StartsAt = *stored.ProposedStartsAt
			occurrence.Version++
			occurrence.UpdatedAt = command.Now
			action = "LessonRescheduled"
		} else {
			cancelledStatus := core.LessonStatus("cancelled_student")
			if stored.RequestedByAccountID == occurrence.TeacherAccountID {
				cancelledStatus = core.LessonStatus("cancelled_school")
			}
			occurrence.Status = cancelledStatus
			occurrence.Version++
			occurrence.UpdatedAt = command.Now
			action = "LessonCancelled"
		}
		stored.Status = "approved"
	} else {
		stored.Status = "declined"
	}
	decided := command.Now
	stored.DecidedByAccountID = principal.AccountID
	stored.DecisionNote = command.DecisionNote
	stored.DecidedAt = &decided
	stored.Version++
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, action,
		"lesson_occurrence", stored.OccurrenceID, "allow", "", command.Now, nil)
	s.appendOutbox(principal.TenantID, action, stored.OccurrenceID, command.Now)
	return s.rescheduleView(stored), nil
}

func (s *Store) WithdrawRescheduleRequest(_ context.Context, command core.WithdrawRescheduleRequestCommand) (core.RescheduleRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	principal := command.Principal
	stored := s.rescheduleRequests[command.RequestID]
	if stored == nil || stored.TenantID != principal.TenantID ||
		stored.RequestedByAccountID != principal.AccountID || stored.Status != "pending" {
		return core.RescheduleRequest{}, core.E(core.CodeInvalidState, "no pending request of yours to withdraw", nil)
	}
	stored.Status = "withdrawn"
	stored.Version++
	s.appendSecurityAudit(principal.TenantID, principal.AccountID, "LessonRescheduleWithdrawn",
		"reschedule_request", stored.ID, "allow", "", command.Now, nil)
	return s.rescheduleView(stored), nil
}
