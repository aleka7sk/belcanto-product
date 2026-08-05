package httpapi

import (
	"net/http"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// P.1 account profile handlers (Figma Page 32: ACC-01/02).

func (api *API) myProfile(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	view, err := api.service.MyProfile(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, view)
}

type updateProfileRequest struct {
	FullName string `json:"fullName"`
}

func (api *API) updateMyProfile(response http.ResponseWriter, request *http.Request) {
	var input updateProfileRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	view, err := api.service.UpdateMyProfile(request.Context(), authenticated.principal, input.FullName)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, view)
}
