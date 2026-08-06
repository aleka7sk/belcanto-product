package httpapi

import (
	"net/http"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.5 community and safety handlers (Figma Page 28).

type createPostRequest struct {
	Kind            string `json:"kind,omitempty"`
	Title           string `json:"title,omitempty"`
	Body            string `json:"body"`
	Audience        string `json:"audience,omitempty"`
	CommentsEnabled bool   `json:"commentsEnabled"`
	Pinned          bool   `json:"pinned,omitempty"`
}

func (api *API) createCommunityPost(response http.ResponseWriter, request *http.Request) {
	var input createPostRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	post, err := api.service.CreateCommunityPost(request.Context(), authenticated.principal, app.CreatePostInput{
		Kind: input.Kind, Title: input.Title, Body: input.Body,
		Audience: input.Audience, CommentsEnabled: input.CommentsEnabled,
		Pinned: input.Pinned, IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, post)
}

func (api *API) communityFeed(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	posts, err := api.service.CommunityFeed(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, posts)
}

func (api *API) communityPost(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	post, err := api.service.CommunityPost(
		request.Context(), authenticated.principal, request.PathValue("postId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, post)
}

type addCommentRequest struct {
	Body string `json:"body"`
}

func (api *API) addCommunityComment(response http.ResponseWriter, request *http.Request) {
	var input addCommentRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	post, err := api.service.AddCommunityComment(request.Context(), authenticated.principal, app.AddCommentInput{
		PostID: request.PathValue("postId"), Body: input.Body,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, post)
}

type removeContentRequest struct {
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
}

func (api *API) removeCommunityContent(response http.ResponseWriter, request *http.Request) {
	var input removeContentRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	post, err := api.service.RemoveCommunityContent(request.Context(), authenticated.principal, app.RemoveContentInput{
		TargetType: input.TargetType, TargetID: input.TargetID,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, post)
}

type reportContentRequest struct {
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Reason     string `json:"reason"`
	Note       string `json:"note,omitempty"`
}

func (api *API) reportCommunityContent(response http.ResponseWriter, request *http.Request) {
	var input reportContentRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	report, err := api.service.ReportCommunityContent(request.Context(), authenticated.principal, app.ReportContentInput{
		TargetType: input.TargetType, TargetID: input.TargetID,
		Reason: input.Reason, Note: input.Note,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, report)
}

func (api *API) moderationQueue(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	reports, err := api.service.ModerationQueue(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, reports)
}

type decideReportRequest struct {
	Decision       string `json:"decision"`
	DecisionReason string `json:"decisionReason"`
}

func (api *API) decideCommunityReport(response http.ResponseWriter, request *http.Request) {
	var input decideReportRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	report, err := api.service.DecideCommunityReport(request.Context(), authenticated.principal, app.DecideReportInput{
		ReportID: request.PathValue("reportId"), Decision: input.Decision,
		DecisionReason: input.DecisionReason, IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, report)
}

type blockMemberRequest struct {
	AccountID string `json:"accountId"`
	Blocked   bool   `json:"blocked"`
}

func (api *API) blockCommunityMember(response http.ResponseWriter, request *http.Request) {
	var input blockMemberRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	blocked, err := api.service.BlockCommunityMember(request.Context(), authenticated.principal, app.BlockMemberInput{
		BlockedAccountID: input.AccountID, Blocked: input.Blocked,
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, map[string][]string{"blocked": blocked})
}

func (api *API) blockedCommunityMembers(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	blocked, err := api.service.BlockedCommunityMembers(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, map[string][]string{"blocked": blocked})
}
