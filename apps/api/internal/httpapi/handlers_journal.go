package httpapi

import (
	"net/http"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 lesson journal and progress handlers (flows C/D/H/I).

type journalDraftRequest struct {
	OccurrenceID string `json:"occurrenceId"`
	StudentID    string `json:"studentId"`
	WhatWorked   string `json:"whatWorked"`
	CurrentFocus string `json:"currentFocus"`
	NextStep     string `json:"nextStep"`
}

func (api *API) saveJournalDraft(response http.ResponseWriter, request *http.Request) {
	var input journalDraftRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	journal, err := api.service.SaveJournalDraft(request.Context(), authenticated.principal, app.JournalDraftInput{
		OccurrenceID: input.OccurrenceID, StudentID: input.StudentID,
		WhatWorked: input.WhatWorked, CurrentFocus: input.CurrentFocus, NextStep: input.NextStep,
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, journal)
}

type publishJournalRequest struct {
	OccurrenceID   string               `json:"occurrenceId"`
	StudentID      string               `json:"studentId"`
	CorrectionNote string               `json:"correctionNote,omitempty"`
	Evidence       []core.EvidenceInput `json:"evidence,omitempty"`
}

func (api *API) publishJournal(response http.ResponseWriter, request *http.Request) {
	var input publishJournalRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	journal, err := api.service.PublishJournal(request.Context(), authenticated.principal, app.PublishJournalInput{
		OccurrenceID: input.OccurrenceID, StudentID: input.StudentID,
		CorrectionNote: input.CorrectionNote, Evidence: input.Evidence,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, journal)
}

func (api *API) getJournal(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	journal, err := api.service.GetJournal(
		request.Context(), authenticated.principal,
		request.PathValue("occurrenceId"), request.PathValue("studentId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, journal)
}

func (api *API) listStudentJournals(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	journals, err := api.service.ListStudentJournals(
		request.Context(), authenticated.principal, request.PathValue("studentId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, journals)
}

func (api *API) listProgressEvidence(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	evidence, err := api.service.ListProgressEvidence(
		request.Context(), authenticated.principal, request.PathValue("studentId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, evidence)
}
