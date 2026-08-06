package httpapi

import (
	"net/http"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 goal and achievement handlers (STU-GROWTH-04/08).

type createGoalRequest struct {
	Criterion        string `json:"criterion"`
	Description      string `json:"description,omitempty"`
	RelatedSongID    string `json:"relatedSongId,omitempty"`
	RelatedSkillArea string `json:"relatedSkillArea,omitempty"`
}

func (api *API) createGoal(response http.ResponseWriter, request *http.Request) {
	var input createGoalRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	goal, err := api.service.CreateGoal(request.Context(), authenticated.principal, app.CreateGoalInput{
		StudentID: request.PathValue("studentId"),
		Criterion: input.Criterion, Description: input.Description,
		RelatedSongID: input.RelatedSongID, RelatedSkillArea: input.RelatedSkillArea,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, goal)
}

func (api *API) listStudentGoals(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	goals, err := api.service.ListStudentGoals(
		request.Context(), authenticated.principal, request.PathValue("studentId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, goals)
}

type completeGoalRequest struct {
	CompletionNote  string `json:"completionNote"`
	ExpectedVersion int    `json:"expectedVersion"`
}

func (api *API) completeGoal(response http.ResponseWriter, request *http.Request) {
	var input completeGoalRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	goal, err := api.service.CompleteGoal(request.Context(), authenticated.principal, app.CompleteGoalInput{
		GoalID:         request.PathValue("goalId"),
		CompletionNote: input.CompletionNote, ExpectedVersion: input.ExpectedVersion,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, goal)
}

type reframeGoalRequest struct {
	Reason          string `json:"reason"`
	NewCriterion    string `json:"newCriterion,omitempty"`
	NewDescription  string `json:"newDescription,omitempty"`
	ExpectedVersion int    `json:"expectedVersion"`
}

func (api *API) reframeGoal(response http.ResponseWriter, request *http.Request) {
	var input reframeGoalRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	goals, err := api.service.ReframeGoal(request.Context(), authenticated.principal, app.ReframeGoalInput{
		GoalID: request.PathValue("goalId"), Reason: input.Reason,
		NewCriterion: input.NewCriterion, NewDescription: input.NewDescription,
		ExpectedVersion: input.ExpectedVersion, IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, goals)
}

type createAchievementDefinitionRequest struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	Category            string `json:"category"`
	EvidenceRequirement string `json:"evidenceRequirement,omitempty"`
}

func (api *API) createAchievementDefinition(response http.ResponseWriter, request *http.Request) {
	var input createAchievementDefinitionRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	definition, err := api.service.CreateAchievementDefinition(request.Context(), authenticated.principal, app.CreateAchievementDefinitionInput{
		Name: input.Name, Description: input.Description, Category: input.Category,
		EvidenceRequirement: input.EvidenceRequirement,
		IdempotencyKey:      idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, definition)
}

func (api *API) listAchievementDefinitions(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	definitions, err := api.service.ListAchievementDefinitions(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, definitions)
}

func (api *API) retireAchievementDefinition(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	definition, err := api.service.RetireAchievementDefinition(
		request.Context(), authenticated.principal,
		request.PathValue("definitionId"), idempotencyKey(request))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, definition)
}

type awardAchievementRequest struct {
	DefinitionID string `json:"definitionId"`
	EvidenceNote string `json:"evidenceNote"`
}

func (api *API) awardAchievement(response http.ResponseWriter, request *http.Request) {
	var input awardAchievementRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	award, err := api.service.AwardAchievement(request.Context(), authenticated.principal, app.AwardAchievementInput{
		DefinitionID:   input.DefinitionID,
		StudentID:      request.PathValue("studentId"),
		EvidenceNote:   input.EvidenceNote,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, award)
}

type revokeAchievementRequest struct {
	Reason string `json:"reason"`
}

func (api *API) revokeAchievement(response http.ResponseWriter, request *http.Request) {
	var input revokeAchievementRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	award, err := api.service.RevokeAchievement(
		request.Context(), authenticated.principal,
		request.PathValue("awardId"), input.Reason, idempotencyKey(request))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, award)
}

func (api *API) listStudentAwards(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	awards, err := api.service.ListStudentAwards(
		request.Context(), authenticated.principal, request.PathValue("studentId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, awards)
}
