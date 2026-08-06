package app

import (
	"context"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

// L.3 lesson journals and progress (DEC-006/007). A draft may change
// freely; publishing snapshots it as an immutable version the student
// sees, and a correction always carries an explicit note.

const maxJournalEvidence = 10

type JournalDraftInput struct {
	OccurrenceID string
	StudentID    string
	WhatWorked   string
	CurrentFocus string
	NextStep     string
}

func (s *Service) SaveJournalDraft(ctx context.Context, principal core.Principal, input JournalDraftInput) (core.LessonJournal, error) {
	occurrenceID, err := security.ValidateIdentifier("occurrenceId", input.OccurrenceID, 128)
	if err != nil {
		return core.LessonJournal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	studentID, err := security.ValidateIdentifier("studentId", input.StudentID, 128)
	if err != nil {
		return core.LessonJournal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	whatWorked, err := security.ValidateText("whatWorked", input.WhatWorked, 1, 2000)
	if err != nil {
		return core.LessonJournal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	currentFocus, err := security.ValidateText("currentFocus", input.CurrentFocus, 1, 2000)
	if err != nil {
		return core.LessonJournal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	nextStep, err := security.ValidateText("nextStep", input.NextStep, 1, 2000)
	if err != nil {
		return core.LessonJournal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	journalID, err := security.NewID("jrnl")
	if err != nil {
		return core.LessonJournal{}, core.E(core.CodeInternal, "could not create the journal", err)
	}
	journal, err := s.store.SaveJournalDraft(ctx, core.SaveJournalDraftCommand{
		Principal: principal, JournalID: journalID,
		OccurrenceID: occurrenceID, StudentID: studentID,
		Draft: core.JournalDraft{
			WhatWorked: whatWorked, CurrentFocus: currentFocus, NextStep: nextStep,
		},
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.LessonJournal{}, normalizeStoreError("save journal draft", err)
	}
	return journal, nil
}

type PublishJournalInput struct {
	OccurrenceID   string
	StudentID      string
	CorrectionNote string
	Evidence       []core.EvidenceInput
	IdempotencyKey string
}

func (s *Service) PublishJournal(ctx context.Context, principal core.Principal, input PublishJournalInput) (core.LessonJournal, error) {
	occurrenceID, err := security.ValidateIdentifier("occurrenceId", input.OccurrenceID, 128)
	if err != nil {
		return core.LessonJournal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	studentID, err := security.ValidateIdentifier("studentId", input.StudentID, 128)
	if err != nil {
		return core.LessonJournal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	correctionNote := ""
	if input.CorrectionNote != "" {
		correctionNote, err = security.ValidateText("correctionNote", input.CorrectionNote, 1, 500)
		if err != nil {
			return core.LessonJournal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
		}
	}
	if len(input.Evidence) > maxJournalEvidence {
		return core.LessonJournal{}, core.E(core.CodeInvalidInput, "at most 10 evidence entries per publish", nil)
	}
	evidence := make([]core.EvidenceInput, 0, len(input.Evidence))
	evidenceIDs := make([]string, 0, len(input.Evidence))
	for _, entry := range input.Evidence {
		area, areaErr := security.ValidateText("evidence area", entry.Area, 1, 100)
		if areaErr != nil {
			return core.LessonJournal{}, core.E(core.CodeInvalidInput, areaErr.Error(), nil)
		}
		note, noteErr := security.ValidateText("evidence note", entry.Note, 1, 1000)
		if noteErr != nil {
			return core.LessonJournal{}, core.E(core.CodeInvalidInput, noteErr.Error(), nil)
		}
		evidenceID, idErr := security.NewID("evd")
		if idErr != nil {
			return core.LessonJournal{}, core.E(core.CodeInternal, "could not create evidence ids", idErr)
		}
		evidence = append(evidence, core.EvidenceInput{Area: area, Note: note})
		evidenceIDs = append(evidenceIDs, evidenceID)
	}
	idempotencyKey, err := security.ValidateIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return core.LessonJournal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	fingerprint, err := security.Fingerprint(struct {
		OccurrenceID, StudentID, CorrectionNote string
		Evidence                                []core.EvidenceInput
	}{occurrenceID, studentID, correctionNote, evidence})
	if err != nil {
		return core.LessonJournal{}, core.E(core.CodeInternal, "could not fingerprint the request", err)
	}
	journal, err := s.store.PublishJournal(ctx, core.PublishJournalCommand{
		Principal: principal, OccurrenceID: occurrenceID, StudentID: studentID,
		CorrectionNote: correctionNote, Evidence: evidence, EvidenceIDs: evidenceIDs,
		IdempotencyKey: idempotencyKey, PayloadFingerprint: fingerprint,
		Now: s.clock.Now(),
	})
	if err != nil {
		return core.LessonJournal{}, normalizeStoreError("publish journal", err)
	}
	return journal, nil
}

func (s *Service) GetJournal(ctx context.Context, principal core.Principal, occurrenceID, studentID string) (core.LessonJournal, error) {
	normalizedOccurrence, err := security.ValidateIdentifier("occurrenceId", occurrenceID, 128)
	if err != nil {
		return core.LessonJournal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	normalizedStudent, err := security.ValidateIdentifier("studentId", studentID, 128)
	if err != nil {
		return core.LessonJournal{}, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	journal, err := s.store.GetJournal(ctx, principal, normalizedOccurrence, normalizedStudent)
	if err != nil {
		return core.LessonJournal{}, normalizeStoreError("read journal", err)
	}
	return journal, nil
}

func (s *Service) ListStudentJournals(ctx context.Context, principal core.Principal, studentID string) ([]core.LessonJournal, error) {
	normalizedStudent, err := security.ValidateIdentifier("studentId", studentID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	journals, err := s.store.ListStudentJournals(ctx, principal, normalizedStudent)
	if err != nil {
		return nil, normalizeStoreError("list journals", err)
	}
	return journals, nil
}

func (s *Service) ListProgressEvidence(ctx context.Context, principal core.Principal, studentID string) ([]core.ProgressEvidence, error) {
	normalizedStudent, err := security.ValidateIdentifier("studentId", studentID, 128)
	if err != nil {
		return nil, core.E(core.CodeInvalidInput, err.Error(), nil)
	}
	evidence, err := s.store.ListProgressEvidence(ctx, principal, normalizedStudent)
	if err != nil {
		return nil, normalizeStoreError("list progress evidence", err)
	}
	return evidence, nil
}
