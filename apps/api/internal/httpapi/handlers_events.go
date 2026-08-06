package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.2 events and RSVP handlers (Figma Page 24 event catalog, Page 29).

type createEventCategoryRequest struct {
	Name string `json:"name"`
}

func (api *API) createEventCategory(response http.ResponseWriter, request *http.Request) {
	var input createEventCategoryRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	category, err := api.service.CreateEventCategory(request.Context(), authenticated.principal, input.Name)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, category)
}

func (api *API) listEventCategories(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	categories, err := api.service.ListEventCategories(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, categories)
}

type createEventSeriesRequest struct {
	CategoryID      string `json:"categoryId"`
	Title           string `json:"title"`
	Description     string `json:"description,omitempty"`
	HostAccountID   string `json:"hostAccountId"`
	RoomID          string `json:"roomId,omitempty"`
	Capacity        int    `json:"capacity"`
	Weekday         int    `json:"weekday"`
	StartMinutes    int    `json:"startMinutes"`
	DurationMinutes int    `json:"durationMinutes"`
	EffectiveFrom   string `json:"effectiveFrom"`
	EffectiveUntil  string `json:"effectiveUntil,omitempty"`
}

func (api *API) createEventSeries(response http.ResponseWriter, request *http.Request) {
	var input createEventSeriesRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	series, err := api.service.CreateEventSeries(request.Context(), authenticated.principal, app.CreateEventSeriesInput{
		CategoryID: input.CategoryID, Title: input.Title, Description: input.Description,
		HostAccountID: input.HostAccountID, RoomID: input.RoomID, Capacity: input.Capacity,
		Weekday: input.Weekday, StartMinutes: input.StartMinutes,
		DurationMinutes: input.DurationMinutes,
		EffectiveFrom:   input.EffectiveFrom, EffectiveUntil: input.EffectiveUntil,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, series)
}

func (api *API) generateEventSeriesOccurrences(response http.ResponseWriter, request *http.Request) {
	var input generateOccurrencesRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	result, err := api.service.GenerateEventSeriesOccurrences(
		request.Context(), authenticated.principal,
		request.PathValue("seriesId"), input.Weeks, idempotencyKey(request))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, result)
}

type createEventRequest struct {
	CategoryID      string    `json:"categoryId"`
	Title           string    `json:"title"`
	Description     string    `json:"description,omitempty"`
	StartsAt        time.Time `json:"startsAt"`
	DurationMinutes int       `json:"durationMinutes"`
	HostAccountID   string    `json:"hostAccountId"`
	RoomID          string    `json:"roomId,omitempty"`
	Capacity        int       `json:"capacity"`
}

func (api *API) createEvent(response http.ResponseWriter, request *http.Request) {
	var input createEventRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	event, err := api.service.CreateEvent(request.Context(), authenticated.principal, app.CreateEventInput{
		CategoryID: input.CategoryID, Title: input.Title, Description: input.Description,
		StartsAt: input.StartsAt, DurationMinutes: input.DurationMinutes,
		HostAccountID: input.HostAccountID, RoomID: input.RoomID, Capacity: input.Capacity,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, event)
}

func (api *API) listEvents(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	for key, values := range query {
		if key != "from" && key != "to" {
			api.writeError(response, core.E(core.CodeInvalidInput, "unsupported event query parameter", nil))
			return
		}
		if len(values) != 1 || strings.TrimSpace(values[0]) == "" {
			api.writeError(response, core.E(core.CodeInvalidInput, "event query parameters must appear exactly once and be non-empty", nil))
			return
		}
	}
	from, err := time.Parse(time.RFC3339, query.Get("from"))
	if err != nil {
		api.writeError(response, core.E(core.CodeInvalidInput, "from must be an RFC3339 timestamp", nil))
		return
	}
	to, err := time.Parse(time.RFC3339, query.Get("to"))
	if err != nil {
		api.writeError(response, core.E(core.CodeInvalidInput, "to must be an RFC3339 timestamp", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	events, err := api.service.ListEvents(request.Context(), authenticated.principal, from, to)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, events)
}

func (api *API) rsvpToEvent(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	view, err := api.service.RsvpToEvent(request.Context(), authenticated.principal, request.PathValue("occurrenceId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, view)
}

func (api *API) cancelEventRsvp(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	view, err := api.service.CancelEventRsvp(request.Context(), authenticated.principal, request.PathValue("occurrenceId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, view)
}

func (api *API) joinEventWaitlist(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	view, err := api.service.JoinEventWaitlist(request.Context(), authenticated.principal, request.PathValue("occurrenceId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, view)
}

func (api *API) leaveEventWaitlist(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	view, err := api.service.LeaveEventWaitlist(request.Context(), authenticated.principal, request.PathValue("occurrenceId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, view)
}

func (api *API) confirmSpotOffer(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	view, err := api.service.ConfirmSpotOffer(request.Context(), authenticated.principal, request.PathValue("offerId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, view)
}

func (api *API) declineSpotOffer(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	view, err := api.service.DeclineSpotOffer(request.Context(), authenticated.principal, request.PathValue("offerId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, view)
}
