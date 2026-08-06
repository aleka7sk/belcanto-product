package httpapi

import (
	"net/http"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.5 activity and notification-preference handlers (ACT-01..03).

func (api *API) activityFeed(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	feed, err := api.service.ActivityFeed(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, feed)
}

type markActivityReadRequest struct {
	UpTo time.Time `json:"upTo"`
}

type markActivityReadResponse struct {
	Marked int `json:"marked"`
}

func (api *API) markActivityRead(response http.ResponseWriter, request *http.Request) {
	var input markActivityReadRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	marked, err := api.service.MarkActivityRead(request.Context(), authenticated.principal, input.UpTo)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, markActivityReadResponse{Marked: marked})
}

func (api *API) notificationPreferences(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	preferences, err := api.service.NotificationPreferences(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, preferences)
}

type updateNotificationPreferenceRequest struct {
	Category    string `json:"category"`
	PushEnabled bool   `json:"pushEnabled"`
}

func (api *API) updateNotificationPreference(response http.ResponseWriter, request *http.Request) {
	var input updateNotificationPreferenceRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	preferences, err := api.service.UpdateNotificationPreference(
		request.Context(), authenticated.principal, input.Category, input.PushEnabled)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, preferences)
}
