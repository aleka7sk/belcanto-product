package httpapi

import (
	"net/http"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.2 rooms and core lesson series handlers (Figma Pages 24/26/29).

type createRoomRequest struct {
	Name     string `json:"name"`
	Capacity *int   `json:"capacity"`
}

func (api *API) createRoom(response http.ResponseWriter, request *http.Request) {
	var input createRoomRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	room, err := api.service.CreateRoom(request.Context(), authenticated.principal, app.CreateRoomInput{
		Name: input.Name, Capacity: input.Capacity,
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, room)
}

func (api *API) listRooms(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	rooms, err := api.service.ListRooms(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, rooms)
}

type createLessonSeriesRequest struct {
	Format           string   `json:"format"`
	Title            string   `json:"title"`
	TeacherAccountID string   `json:"teacherAccountId"`
	RoomID           string   `json:"roomId,omitempty"`
	Weekday          int      `json:"weekday"`
	StartMinutes     int      `json:"startMinutes"`
	DurationMinutes  int      `json:"durationMinutes"`
	EffectiveFrom    string   `json:"effectiveFrom"`
	EffectiveUntil   string   `json:"effectiveUntil,omitempty"`
	StudentIDs       []string `json:"studentIds"`
}

func (api *API) createLessonSeries(response http.ResponseWriter, request *http.Request) {
	var input createLessonSeriesRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	series, err := api.service.CreateCoreLessonSeries(request.Context(), authenticated.principal, app.CreateCoreLessonSeriesInput{
		Format: input.Format, Title: input.Title,
		TeacherAccountID: input.TeacherAccountID, RoomID: input.RoomID,
		Weekday: input.Weekday, StartMinutes: input.StartMinutes,
		DurationMinutes: input.DurationMinutes,
		EffectiveFrom:   input.EffectiveFrom, EffectiveUntil: input.EffectiveUntil,
		StudentIDs: input.StudentIDs, IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, series)
}

func (api *API) listLessonSeries(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	series, err := api.service.ListCoreLessonSeries(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, series)
}

func (api *API) getLessonSeries(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	series, err := api.service.GetCoreLessonSeries(
		request.Context(), authenticated.principal, request.PathValue("seriesId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, series)
}

type generateOccurrencesRequest struct {
	Weeks int `json:"weeks"`
}

func (api *API) generateSeriesOccurrences(response http.ResponseWriter, request *http.Request) {
	var input generateOccurrencesRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	result, err := api.service.GenerateSeriesOccurrences(
		request.Context(), authenticated.principal,
		request.PathValue("seriesId"), input.Weeks, idempotencyKey(request))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, result)
}
