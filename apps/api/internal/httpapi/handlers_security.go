package httpapi

import (
	"net/http"
	"strconv"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// P.1 session security handlers (Figma Page 32: ACC-05/08/09, AUTH-07/08).

func (api *API) listMySessions(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	devices, err := api.service.ListSessions(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, devices)
}

type revokeSessionRequest struct {
	CurrentPassword string `json:"currentPassword"`
}

func (api *API) revokeMySession(response http.ResponseWriter, request *http.Request) {
	var input revokeSessionRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	if !api.allowSensitive(request, "session_reauthentication", authenticated.principal.AccountID, 5, 20) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many re-authentication attempts", nil))
		return
	}
	if err := api.service.RevokeSessionByID(
		request.Context(),
		authenticated.principal,
		request.PathValue("sessionId"),
		input.CurrentPassword,
	); err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

type revokeOtherSessionsResponse struct {
	RevokedCount int `json:"revokedCount"`
}

func (api *API) revokeOtherSessions(response http.ResponseWriter, request *http.Request) {
	var input revokeSessionRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	if !api.allowSensitive(request, "session_reauthentication", authenticated.principal.AccountID, 5, 20) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many re-authentication attempts", nil))
		return
	}
	revoked, err := api.service.RevokeOtherSessions(request.Context(), authenticated.principal, input.CurrentPassword)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, revokeOtherSessionsResponse{RevokedCount: revoked})
}

func (api *API) listSecurityEvents(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	for key := range query {
		if key != "cursor" && key != "limit" {
			api.writeError(response, core.E(core.CodeInvalidInput, "unsupported query parameter", nil))
			return
		}
	}
	if len(query["cursor"]) > 1 || len(query["limit"]) > 1 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters must not repeat", nil))
		return
	}
	limit := 0
	if raw := query.Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			api.writeError(response, core.E(core.CodeInvalidInput, "limit must be a positive integer", nil))
			return
		}
		limit = parsed
	}
	authenticated := authenticatedPrincipal(request)
	page, err := api.service.ListSecurityEvents(request.Context(), authenticated.principal, query.Get("cursor"), limit)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, page)
}

type requestPasswordResetRequest struct {
	Phone string `json:"phone"`
}

func (api *API) requestPasswordReset(response http.ResponseWriter, request *http.Request) {
	var input requestPasswordResetRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "password_reset", normalizeRateLimitPhone(input.Phone), 5, 20) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many recovery attempts", nil))
		return
	}
	if err := api.service.RequestPasswordReset(request.Context(), input.Phone); err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

type completePasswordResetRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

func (api *API) completePasswordReset(response http.ResponseWriter, request *http.Request) {
	var input completePasswordResetRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "password_reset_complete", input.Token, 10, 30) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many recovery attempts", nil))
		return
	}
	if err := api.service.CompletePasswordReset(request.Context(), input.Token, input.NewPassword); err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}
