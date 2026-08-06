package httpapi

import (
	"net/http"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.4 attendance handlers (TCH-JOURNAL-01/02).

type markAttendanceRequest struct {
	Status       string `json:"status"`
	LateMinutes  int    `json:"lateMinutes,omitempty"`
	Note         string `json:"note,omitempty"`
	ChangeReason string `json:"changeReason,omitempty"`
}

func (api *API) markAttendance(response http.ResponseWriter, request *http.Request) {
	var input markAttendanceRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	records, err := api.service.MarkAttendance(request.Context(), authenticated.principal, app.MarkAttendanceInput{
		OccurrenceID: request.PathValue("lessonId"),
		StudentID:    request.PathValue("studentId"),
		Status:       input.Status, LateMinutes: input.LateMinutes,
		Note: input.Note, ChangeReason: input.ChangeReason,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, records)
}

func (api *API) listLessonAttendance(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	records, err := api.service.ListLessonAttendance(
		request.Context(), authenticated.principal, request.PathValue("lessonId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, records)
}
