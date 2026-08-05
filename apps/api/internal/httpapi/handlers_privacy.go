package httpapi

import (
	"net/http"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// P.1 policies, privacy and data rights handlers (Figma Page 32:
// ACC-10..12, ACC-14..18).

func (api *API) listPolicies(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	policies, err := api.service.ListPolicies(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, policies)
}

type acceptPolicyRequest struct {
	PolicyVersionID string `json:"policyVersionId"`
}

func (api *API) acceptPolicy(response http.ResponseWriter, request *http.Request) {
	var input acceptPolicyRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	if err := api.service.AcceptPolicy(request.Context(), authenticated.principal, input.PolicyVersionID); err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (api *API) privacySettings(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	settings, err := api.service.PrivacySettings(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, settings)
}

func (api *API) updatePrivacySettings(response http.ResponseWriter, request *http.Request) {
	var input core.PrivacySettings
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	updated, err := api.service.UpdatePrivacySettings(request.Context(), authenticated.principal, input)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, updated)
}

type sensitiveActionRequest struct {
	CurrentPassword string `json:"currentPassword"`
}

func (api *API) createDataExport(response http.ResponseWriter, request *http.Request) {
	var input sensitiveActionRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	if !api.allowSensitive(request, "data_rights", authenticated.principal.AccountID, 5, 20) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many re-authentication attempts", nil))
		return
	}
	export, err := api.service.CreateDataExport(request.Context(), authenticated.principal, input.CurrentPassword)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, export)
}

func (api *API) listDataExports(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	exports, err := api.service.ListDataExports(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, exports)
}

func (api *API) deletionRequest(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	request2, err := api.service.DeletionRequest(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, request2)
}

func (api *API) createDeletionRequest(response http.ResponseWriter, request *http.Request) {
	var input sensitiveActionRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	if !api.allowSensitive(request, "data_rights", authenticated.principal.AccountID, 5, 20) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many re-authentication attempts", nil))
		return
	}
	created, err := api.service.CreateDeletionRequest(request.Context(), authenticated.principal, input.CurrentPassword)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, created)
}

func (api *API) cancelDeletionRequest(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	cancelled, err := api.service.CancelDeletionRequest(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, cancelled)
}
