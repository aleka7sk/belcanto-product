package httpapi

import (
	"net/http"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// L.3 repertoire handlers (STU-GROWTH-07, STU-PRACTICE-10/11).

type addStudentSongRequest struct {
	Title     string `json:"title"`
	Artist    string `json:"artist,omitempty"`
	Stage     string `json:"stage,omitempty"`
	StageNote string `json:"stageNote,omitempty"`
}

func (api *API) addStudentSong(response http.ResponseWriter, request *http.Request) {
	var input addStudentSongRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	song, err := api.service.AddStudentSong(request.Context(), authenticated.principal, app.AddStudentSongInput{
		StudentID: request.PathValue("studentId"),
		Title:     input.Title, Artist: input.Artist,
		Stage: input.Stage, StageNote: input.StageNote,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, song)
}

type changeSongStageRequest struct {
	Stage           string `json:"stage"`
	StageNote       string `json:"stageNote,omitempty"`
	ExpectedVersion int    `json:"expectedVersion"`
}

func (api *API) changeSongStage(response http.ResponseWriter, request *http.Request) {
	var input changeSongStageRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	song, err := api.service.ChangeSongStage(request.Context(), authenticated.principal, app.ChangeSongStageInput{
		SongID: request.PathValue("songId"),
		Stage:  input.Stage, StageNote: input.StageNote,
		ExpectedVersion: input.ExpectedVersion,
		IdempotencyKey:  idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, song)
}

func (api *API) listStudentSongs(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Query()) != 0 {
		api.writeError(response, core.E(core.CodeInvalidInput, "query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	songs, err := api.service.ListStudentSongs(
		request.Context(), authenticated.principal, request.PathValue("studentId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, songs)
}
