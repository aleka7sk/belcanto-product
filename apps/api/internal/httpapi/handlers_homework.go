package httpapi

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 homework, practice and media handlers (flow E). Media content is
// served only through sealed short-lived links; the chunk endpoint is
// resumable by exact offset.

type createMediaRequest struct {
	Kind        string `json:"kind"`
	ContentType string `json:"contentType"`
	ByteSize    int64  `json:"byteSize"`
}

func (api *API) createMedia(response http.ResponseWriter, request *http.Request) {
	var input createMediaRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	object, err := api.service.CreateMedia(request.Context(), authenticated.principal, app.CreateMediaInput{
		Kind: input.Kind, ContentType: input.ContentType, ByteSize: input.ByteSize,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, object)
}

const maxChunkBody = 8 << 20

func (api *API) appendMediaChunk(response http.ResponseWriter, request *http.Request) {
	offsetHeader := request.Header.Get("Upload-Offset")
	offset, err := strconv.ParseInt(offsetHeader, 10, 64)
	if err != nil || offset < 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "the Upload-Offset header must be a non-negative integer", nil))
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(response, request.Body, maxChunkBody))
	if err != nil {
		api.writeError(response, core.E(core.CodeInvalidInput, "the upload chunk could not be read", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	object, err := api.service.AppendMediaChunk(
		request.Context(), authenticated.principal,
		request.PathValue("mediaId"), offset, body)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, object)
}

func (api *API) getMedia(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	object, err := api.service.GetMedia(
		request.Context(), authenticated.principal, request.PathValue("mediaId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, object)
}

func (api *API) signMediaAccess(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	access, err := api.service.SignMediaAccess(
		request.Context(), authenticated.principal, request.PathValue("mediaId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, access)
}

func (api *API) mediaContent(response http.ResponseWriter, request *http.Request) {
	token := request.URL.Query().Get("token")
	content, contentType, err := api.service.MediaContentByToken(
		request.Context(), request.PathValue("mediaId"), token)
	if err != nil {
		api.writeError(response, err)
		return
	}
	response.Header().Set("Content-Type", contentType)
	response.Header().Set("Cache-Control", "private, no-store")
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write(content)
}

type homeworkTaskRequest struct {
	Title              string `json:"title"`
	Description        string `json:"description,omitempty"`
	RecommendedMinutes int    `json:"recommendedMinutes,omitempty"`
	SkillArea          string `json:"skillArea,omitempty"`
	SongTitle          string `json:"songTitle,omitempty"`
}

type createHomeworkRequest struct {
	OccurrenceID       string                `json:"occurrenceId"`
	StudentID          string                `json:"studentId"`
	Goal               string                `json:"goal"`
	ReadinessCriteria  string                `json:"readinessCriteria,omitempty"`
	DueAt              *time.Time            `json:"dueAt,omitempty"`
	Tasks              []homeworkTaskRequest `json:"tasks,omitempty"`
	AttachmentMediaIDs []string              `json:"attachmentMediaIds,omitempty"`
	Assign             bool                  `json:"assign,omitempty"`
}

func (api *API) createHomework(response http.ResponseWriter, request *http.Request) {
	var input createHomeworkRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	tasks := make([]core.HomeworkTaskInput, 0, len(input.Tasks))
	for _, task := range input.Tasks {
		tasks = append(tasks, core.HomeworkTaskInput{
			Title: task.Title, Description: task.Description,
			RecommendedMinutes: task.RecommendedMinutes,
			SkillArea:          task.SkillArea, SongTitle: task.SongTitle,
		})
	}
	authenticated := authenticatedPrincipal(request)
	homework, err := api.service.CreateHomework(request.Context(), authenticated.principal, app.CreateHomeworkInput{
		OccurrenceID: input.OccurrenceID, StudentID: input.StudentID,
		Goal: input.Goal, ReadinessCriteria: input.ReadinessCriteria,
		DueAt: input.DueAt, Tasks: tasks,
		AttachmentMediaIDs: input.AttachmentMediaIDs, Assign: input.Assign,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, homework)
}

func (api *API) getHomework(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	homework, err := api.service.GetHomework(
		request.Context(), authenticated.principal, request.PathValue("homeworkId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, homework)
}

func (api *API) assignHomework(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	homework, err := api.service.AssignHomework(
		request.Context(), authenticated.principal,
		request.PathValue("homeworkId"), idempotencyKey(request))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, homework)
}

func (api *API) startHomework(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	homework, err := api.service.StartHomework(
		request.Context(), authenticated.principal,
		request.PathValue("homeworkId"), idempotencyKey(request))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, homework)
}

type cancelHomeworkRequest struct {
	Reason string `json:"reason"`
}

func (api *API) cancelHomework(response http.ResponseWriter, request *http.Request) {
	var input cancelHomeworkRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	homework, err := api.service.CancelHomework(
		request.Context(), authenticated.principal,
		request.PathValue("homeworkId"), input.Reason, idempotencyKey(request))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, homework)
}

type markHomeworkTaskRequest struct {
	Done bool `json:"done"`
}

func (api *API) markHomeworkTask(response http.ResponseWriter, request *http.Request) {
	var input markHomeworkTaskRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	homework, err := api.service.MarkHomeworkTask(
		request.Context(), authenticated.principal,
		request.PathValue("homeworkId"), request.PathValue("taskId"), input.Done)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, homework)
}

type submitHomeworkRequest struct {
	Note            string   `json:"note,omitempty"`
	MediaIDs        []string `json:"mediaIds,omitempty"`
	ExpectedVersion int      `json:"expectedVersion"`
}

func (api *API) submitHomework(response http.ResponseWriter, request *http.Request) {
	var input submitHomeworkRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	homework, err := api.service.SubmitHomework(request.Context(), authenticated.principal, app.SubmitHomeworkInput{
		HomeworkID: request.PathValue("homeworkId"), Note: input.Note,
		MediaIDs: input.MediaIDs, ExpectedVersion: input.ExpectedVersion,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, homework)
}

type reviewHomeworkRequest struct {
	Decision        string `json:"decision"`
	Body            string `json:"body"`
	NextStep        string `json:"nextStep,omitempty"`
	EvidenceArea    string `json:"evidenceArea,omitempty"`
	EvidenceNote    string `json:"evidenceNote,omitempty"`
	ExpectedVersion int    `json:"expectedVersion"`
}

func (api *API) reviewHomework(response http.ResponseWriter, request *http.Request) {
	var input reviewHomeworkRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	homework, err := api.service.ReviewHomework(request.Context(), authenticated.principal, app.ReviewHomeworkInput{
		HomeworkID: request.PathValue("homeworkId"), Decision: input.Decision,
		Body: input.Body, NextStep: input.NextStep,
		EvidenceArea: input.EvidenceArea, EvidenceNote: input.EvidenceNote,
		ExpectedVersion: input.ExpectedVersion, IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, homework)
}

func (api *API) listStudentHomework(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	homework, err := api.service.ListStudentHomework(
		request.Context(), authenticated.principal, request.PathValue("studentId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, homework)
}
