package httpapi

import (
	"net/http"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// P.1 multi-step activation, contact verification and TOTP 2FA
// (Figma Page 32: AUTH-01..05/10, ACC-03/06).

type activationTokenRequest struct {
	Token string `json:"token"`
}

func (api *API) activationProgress(response http.ResponseWriter, request *http.Request) {
	var input activationTokenRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "activation_step", input.Token, 30, 120) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many activation attempts", nil))
		return
	}
	view, err := api.service.ActivationProgress(request.Context(), input.Token)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, view)
}

type activationPasswordRequest struct {
	Token    string `json:"token"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (api *API) setActivationPassword(response http.ResponseWriter, request *http.Request) {
	var input activationPasswordRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "activation_step", input.Token, 30, 120) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many activation attempts", nil))
		return
	}
	if err := api.service.SetActivationPassword(request.Context(), input.Token, input.Phone, input.Password); err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

type activationContactRequest struct {
	Token string `json:"token"`
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

func (api *API) startActivationContact(response http.ResponseWriter, request *http.Request) {
	var input activationContactRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "activation_step", input.Token, 30, 120) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many activation attempts", nil))
		return
	}
	if err := api.service.StartActivationContact(request.Context(), input.Token, core.ContactKind(input.Kind), input.Value); err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

type activationCodeRequest struct {
	Token string `json:"token"`
	Code  string `json:"code"`
}

func (api *API) verifyActivationContact(response http.ResponseWriter, request *http.Request) {
	var input activationCodeRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "activation_step", input.Token, 30, 120) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many activation attempts", nil))
		return
	}
	if err := api.service.VerifyActivationContact(request.Context(), input.Token, input.Code); err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (api *API) startActivationTwofa(response http.ResponseWriter, request *http.Request) {
	var input activationTokenRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "activation_step", input.Token, 30, 120) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many activation attempts", nil))
		return
	}
	start, err := api.service.StartActivationTwofa(request.Context(), input.Token)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, start)
}

type recoveryCodesResponse struct {
	RecoveryCodes []string `json:"recoveryCodes"`
}

func (api *API) confirmActivationTwofa(response http.ResponseWriter, request *http.Request) {
	var input activationCodeRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "activation_step", input.Token, 30, 120) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many activation attempts", nil))
		return
	}
	codes, err := api.service.ConfirmActivationTwofa(request.Context(), input.Token, input.Code)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, recoveryCodesResponse{RecoveryCodes: codes})
}

type activationFinishRequest struct {
	Token string `json:"token"`
	Phone string `json:"phone"`
}

func (api *API) finishActivation(response http.ResponseWriter, request *http.Request) {
	var input activationFinishRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "activation_complete", input.Token, 10, 30) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many activation attempts", nil))
		return
	}
	if err := api.service.FinishActivation(request.Context(), input.Token, input.Phone, idempotencyKey(request)); err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

type twofaSignInRequest struct {
	Challenge string `json:"challenge"`
	Code      string `json:"code"`
}

func (api *API) signInWithTwofa(response http.ResponseWriter, request *http.Request) {
	var input twofaSignInRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "twofa_challenge", input.Challenge, 10, 30) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many second-factor attempts", nil))
		return
	}
	tokens, err := api.service.SignInWithTwofa(request.Context(), input.Challenge, input.Code)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, tokens)
}

func (api *API) listMyContacts(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	contacts, err := api.service.ListVerifiedContacts(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, contacts)
}

type startContactChangeRequest struct {
	Kind            string `json:"kind"`
	Value           string `json:"value"`
	CurrentPassword string `json:"currentPassword"`
}

func (api *API) startContactChange(response http.ResponseWriter, request *http.Request) {
	var input startContactChangeRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	if !api.allowSensitive(request, "session_reauthentication", authenticated.principal.AccountID, 5, 20) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many re-authentication attempts", nil))
		return
	}
	if err := api.service.StartContactChange(
		request.Context(),
		authenticated.principal,
		input.CurrentPassword,
		core.ContactKind(input.Kind),
		input.Value,
	); err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

type confirmContactChangeRequest struct {
	Code string `json:"code"`
}

func (api *API) confirmContactChange(response http.ResponseWriter, request *http.Request) {
	var input confirmContactChangeRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	if !api.allowSensitive(request, "contact_confirmation", authenticated.principal.AccountID, 10, 30) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many confirmation attempts", nil))
		return
	}
	contact, err := api.service.ConfirmContactChange(request.Context(), authenticated.principal, input.Code)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, contact)
}

func (api *API) twofaStatus(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	status, err := api.service.TwofaStatus(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, status)
}

type currentPasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
}

func (api *API) startTwofaEnrollment(response http.ResponseWriter, request *http.Request) {
	var input currentPasswordRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	if !api.allowSensitive(request, "session_reauthentication", authenticated.principal.AccountID, 5, 20) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many re-authentication attempts", nil))
		return
	}
	enrollment, err := api.service.StartTwofaEnrollment(request.Context(), authenticated.principal, input.CurrentPassword)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, enrollment)
}

type twofaCodeRequest struct {
	Code string `json:"code"`
}

func (api *API) confirmTwofaEnrollment(response http.ResponseWriter, request *http.Request) {
	var input twofaCodeRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	if !api.allowSensitive(request, "twofa_enrollment", authenticated.principal.AccountID, 10, 30) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many enrollment attempts", nil))
		return
	}
	codes, err := api.service.ConfirmTwofaEnrollment(request.Context(), authenticated.principal, input.Code)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, recoveryCodesResponse{RecoveryCodes: codes})
}

type disableTwofaRequest struct {
	CurrentPassword string `json:"currentPassword"`
	Code            string `json:"code"`
}

func (api *API) disableTwofa(response http.ResponseWriter, request *http.Request) {
	var input disableTwofaRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	if !api.allowSensitive(request, "session_reauthentication", authenticated.principal.AccountID, 5, 20) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many re-authentication attempts", nil))
		return
	}
	if err := api.service.DisableTwofa(request.Context(), authenticated.principal, input.CurrentPassword, input.Code); err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}
