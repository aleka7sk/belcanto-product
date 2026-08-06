package httpapi

import (
	"net/http"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.2 reschedule and cancellation request handlers (flows J/K/L).

type createRescheduleRequestRequest struct {
	OccurrenceID     string     `json:"occurrenceId"`
	Kind             string     `json:"kind"`
	ProposedStartsAt *time.Time `json:"proposedStartsAt,omitempty"`
	Reason           string     `json:"reason"`
}

func (api *API) createRescheduleRequest(response http.ResponseWriter, request *http.Request) {
	var input createRescheduleRequestRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	created, err := api.service.CreateRescheduleRequest(request.Context(), authenticated.principal, app.CreateRescheduleRequestInput{
		OccurrenceID: input.OccurrenceID, Kind: input.Kind,
		ProposedStartsAt: input.ProposedStartsAt, Reason: input.Reason,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, created)
}

func (api *API) listRescheduleRequests(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	requests, err := api.service.ListRescheduleRequests(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, requests)
}

type decideRescheduleRequestRequest struct {
	Approve         bool   `json:"approve"`
	DecisionNote    string `json:"decisionNote,omitempty"`
	ExpectedVersion int64  `json:"expectedVersion"`
}

func (api *API) decideRescheduleRequest(response http.ResponseWriter, request *http.Request) {
	var input decideRescheduleRequestRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	decided, err := api.service.DecideRescheduleRequest(
		request.Context(), authenticated.principal, request.PathValue("requestId"),
		app.DecideRescheduleRequestInput{
			Approve: input.Approve, DecisionNote: input.DecisionNote,
			ExpectedVersion: input.ExpectedVersion,
		})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, decided)
}

func (api *API) withdrawRescheduleRequest(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	withdrawn, err := api.service.WithdrawRescheduleRequest(
		request.Context(), authenticated.principal, request.PathValue("requestId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, withdrawn)
}
