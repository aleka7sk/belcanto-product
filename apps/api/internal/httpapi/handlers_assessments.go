package httpapi

import (
	"net/http"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.4 assessment handlers (Page 27 TCH-REVIEW-*).

type assessmentContentRequest struct {
	Type             string `json:"type"`
	ContextType      string `json:"contextType"`
	ContextID        string `json:"contextId,omitempty"`
	AssessmentDate   string `json:"assessmentDate"`
	Summary          string `json:"summary,omitempty"`
	Strengths        string `json:"strengths,omitempty"`
	DevelopmentAreas string `json:"developmentAreas,omitempty"`
	Recommendations  string `json:"recommendations,omitempty"`
	Confidence       string `json:"confidence,omitempty"`
	Visibility       string `json:"visibility"`
	RelatedSongID    string `json:"relatedSongId,omitempty"`
	RelatedGoalID    string `json:"relatedGoalId,omitempty"`
	Areas            string `json:"areas,omitempty"`
}

func (request assessmentContentRequest) toInput() app.AssessmentContentInput {
	return app.AssessmentContentInput{
		Type: request.Type, ContextType: request.ContextType, ContextID: request.ContextID,
		AssessmentDate: request.AssessmentDate, Summary: request.Summary,
		Strengths: request.Strengths, DevelopmentAreas: request.DevelopmentAreas,
		Recommendations: request.Recommendations, Confidence: request.Confidence,
		Visibility: request.Visibility, RelatedSongID: request.RelatedSongID,
		RelatedGoalID: request.RelatedGoalID, Areas: request.Areas,
	}
}

func (api *API) createAssessment(response http.ResponseWriter, request *http.Request) {
	var input assessmentContentRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	assessment, err := api.service.CreateAssessment(request.Context(), authenticated.principal, app.CreateAssessmentInput{
		StudentID: request.PathValue("studentId"), Content: input.toInput(),
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, assessment)
}

func (api *API) listStudentAssessments(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	assessments, err := api.service.ListStudentAssessments(
		request.Context(), authenticated.principal, request.PathValue("studentId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, assessments)
}

func (api *API) getAssessment(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	assessment, err := api.service.GetAssessment(
		request.Context(), authenticated.principal, request.PathValue("assessmentId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, assessment)
}

type updateAssessmentRequest struct {
	assessmentContentRequest
	ExpectedVersion int64 `json:"expectedVersion"`
}

func (api *API) updateAssessmentDraft(response http.ResponseWriter, request *http.Request) {
	var input updateAssessmentRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	assessment, err := api.service.UpdateAssessmentDraft(request.Context(), authenticated.principal, app.UpdateAssessmentDraftInput{
		AssessmentID: request.PathValue("assessmentId"), Content: input.toInput(),
		ExpectedVersion: input.ExpectedVersion, IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, assessment)
}

type assessmentEvidenceRequest struct {
	Kind        string `json:"kind"`
	Note        string `json:"note"`
	ReferenceID string `json:"referenceId,omitempty"`
}

func (api *API) addAssessmentEvidence(response http.ResponseWriter, request *http.Request) {
	var input assessmentEvidenceRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	assessment, err := api.service.AddAssessmentEvidence(request.Context(), authenticated.principal, app.AddAssessmentEvidenceInput{
		AssessmentID: request.PathValue("assessmentId"),
		Kind:         input.Kind, Note: input.Note, ReferenceID: input.ReferenceID,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, assessment)
}

func (api *API) removeAssessmentEvidence(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	assessment, err := api.service.RemoveAssessmentEvidence(
		request.Context(), authenticated.principal,
		request.PathValue("assessmentId"), request.PathValue("evidenceId"),
		idempotencyKey(request))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, assessment)
}

type publishAssessmentRequest struct {
	ExpectedVersion int64 `json:"expectedVersion"`
}

func (api *API) publishAssessment(response http.ResponseWriter, request *http.Request) {
	var input publishAssessmentRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	assessment, err := api.service.PublishAssessment(request.Context(), authenticated.principal, app.PublishAssessmentInput{
		AssessmentID:    request.PathValue("assessmentId"),
		ExpectedVersion: input.ExpectedVersion, IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, assessment)
}

func (api *API) supersedeAssessment(response http.ResponseWriter, request *http.Request) {
	var input assessmentContentRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	chain, err := api.service.SupersedeAssessment(request.Context(), authenticated.principal, app.SupersedeAssessmentInput{
		AssessmentID: request.PathValue("assessmentId"), Content: input.toInput(),
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, chain)
}

type withdrawAssessmentRequest struct {
	Reason string `json:"reason"`
}

func (api *API) withdrawAssessment(response http.ResponseWriter, request *http.Request) {
	var input withdrawAssessmentRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	assessment, err := api.service.WithdrawAssessment(
		request.Context(), authenticated.principal,
		request.PathValue("assessmentId"), input.Reason, idempotencyKey(request))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, assessment)
}
