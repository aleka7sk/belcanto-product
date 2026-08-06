package app

import (
	"context"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// L.2 reschedule and cancellation requests (flows J/K/L). DEC-102 stays
// open: requests carry a reason and a decision, never a consequence.

type CreateRescheduleRequestInput struct {
	OccurrenceID     string
	Kind             string
	ProposedStartsAt *time.Time
	Reason           string
	IdempotencyKey   string
}

func (s *Service) CreateRescheduleRequest(ctx context.Context, principal core.Principal, input CreateRescheduleRequestInput) (core.RescheduleRequest, error) {
	occurrenceID, err := security.ValidateIdentifier("occurrenceId", input.OccurrenceID, 128)
	if err != nil {
		return core.RescheduleRequest{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	if input.Kind != "reschedule" && input.Kind != "cancellation" {
		return core.RescheduleRequest{}, core.E(core.CodeInvalidInput, "kind must be reschedule or cancellation", nil)
	}
	now := s.clock.Now()
	var proposed *time.Time
	if input.Kind == "reschedule" {
		if input.ProposedStartsAt == nil {
			return core.RescheduleRequest{}, core.E(core.CodeInvalidInput, "proposedStartsAt is required for a reschedule", nil)
		}
		if !input.ProposedStartsAt.After(now) {
			return core.RescheduleRequest{}, core.E(core.CodeInvalidInput, "proposedStartsAt must be in the future", nil)
		}
		utc := input.ProposedStartsAt.UTC()
		proposed = &utc
	} else if input.ProposedStartsAt != nil {
		return core.RescheduleRequest{}, core.E(core.CodeInvalidInput, "a cancellation carries no proposed time", nil)
	}
	reason, err := security.ValidateText("reason", input.Reason, 1, 500)
	if err != nil {
		return core.RescheduleRequest{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.RescheduleRequest{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	requestID, err := security.NewID("resch")
	if err != nil {
		return core.RescheduleRequest{}, core.E(core.CodeInternal, "could not create the request", err)
	}
	fingerprint, err := security.Fingerprint(struct {
		OccurrenceID, Kind, Reason string
		Proposed                   *time.Time
	}{occurrenceID, input.Kind, reason, proposed})
	if err != nil {
		return core.RescheduleRequest{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	request, err := s.store.CreateRescheduleRequest(ctx, core.CreateRescheduleRequestCommand{
		Principal: principal, RequestID: requestID, OccurrenceID: occurrenceID,
		Kind: input.Kind, ProposedStartsAt: proposed, Reason: reason,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint, Now: now,
	})
	if err != nil {
		return core.RescheduleRequest{}, normalizeStoreError("create reschedule request", err)
	}
	return request, nil
}

func (s *Service) ListRescheduleRequests(ctx context.Context, principal core.Principal) ([]core.RescheduleRequest, error) {
	requests, err := s.store.ListRescheduleRequests(ctx, principal)
	if err != nil {
		return nil, normalizeStoreError("list reschedule requests", err)
	}
	return requests, nil
}

type DecideRescheduleRequestInput struct {
	Approve         bool
	DecisionNote    string
	ExpectedVersion int64
}

func (s *Service) DecideRescheduleRequest(ctx context.Context, principal core.Principal, requestID string, input DecideRescheduleRequestInput) (core.RescheduleRequest, error) {
	normalizedID, err := security.ValidateIdentifier("requestId", requestID, 128)
	if err != nil {
		return core.RescheduleRequest{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	note := ""
	if input.DecisionNote != "" {
		note, err = security.ValidateText("decisionNote", input.DecisionNote, 1, 500)
		if err != nil {
			return core.RescheduleRequest{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	if input.ExpectedVersion < 0 {
		return core.RescheduleRequest{}, core.E(core.CodeInvalidInput, "expectedVersion must be at least 0", nil)
	}
	request, err := s.store.DecideRescheduleRequest(ctx, core.DecideRescheduleRequestCommand{
		Principal: principal, RequestID: normalizedID,
		Approve: input.Approve, DecisionNote: note,
		ExpectedVersion: input.ExpectedVersion, Now: s.clock.Now(),
	})
	if err != nil {
		return core.RescheduleRequest{}, normalizeStoreError("decide reschedule request", err)
	}
	return request, nil
}

func (s *Service) WithdrawRescheduleRequest(ctx context.Context, principal core.Principal, requestID string) (core.RescheduleRequest, error) {
	normalizedID, err := security.ValidateIdentifier("requestId", requestID, 128)
	if err != nil {
		return core.RescheduleRequest{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	request, err := s.store.WithdrawRescheduleRequest(ctx, core.WithdrawRescheduleRequestCommand{
		Principal: principal, RequestID: normalizedID, Now: s.clock.Now(),
	})
	if err != nil {
		return core.RescheduleRequest{}, normalizeStoreError("withdraw reschedule request", err)
	}
	return request, nil
}
