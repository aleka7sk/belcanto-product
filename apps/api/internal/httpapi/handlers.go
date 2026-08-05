package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

type activationPreviewRequest struct {
	Token string `json:"token"`
}

func (api *API) previewActivation(response http.ResponseWriter, request *http.Request) {
	var input activationPreviewRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "activation_preview", input.Token, 30, 120) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many activation attempts", nil))
		return
	}
	preview, err := api.service.PreviewActivation(request.Context(), input.Token)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, preview)
}

type activationCompleteRequest struct {
	Token    string `json:"token"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (api *API) completeActivation(response http.ResponseWriter, request *http.Request) {
	var input activationCompleteRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "activation_complete", input.Token, 10, 30) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many activation attempts", nil))
		return
	}
	err := api.service.CompleteActivation(request.Context(), app.CompleteActivationInput{
		Token: input.Token, Phone: input.Phone, Password: input.Password,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

type signInRequest struct {
	Phone       string `json:"phone"`
	Password    string `json:"password"`
	DeviceLabel string `json:"deviceLabel,omitempty"`
	Platform    string `json:"platform,omitempty"`
}

func (api *API) signIn(response http.ResponseWriter, request *http.Request) {
	var input signInRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "sign_in", normalizeRateLimitPhone(input.Phone), 10, 50) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many sign-in attempts", nil))
		return
	}
	outcome, err := api.service.SignIn(request.Context(), input.Phone, input.Password, core.SessionClientInfo{
		DeviceLabel: input.DeviceLabel,
		Platform:    input.Platform,
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, outcome)
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (api *API) refresh(response http.ResponseWriter, request *http.Request) {
	var input refreshRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if !api.allowSensitive(request, "refresh", input.RefreshToken, 30, 100) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many refresh attempts", nil))
		return
	}
	tokens, err := api.service.Refresh(request.Context(), input.RefreshToken)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, tokens)
}

func (api *API) signOut(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	if err := api.service.SignOut(request.Context(), authenticated.access); err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (api *API) bootstrapView(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	view, err := api.service.BootstrapView(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, view)
}

func (api *API) listStaff(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	roles, exists := query["role"]
	if !exists || len(query) != 1 || len(roles) != 1 || roles[0] == "" {
		api.writeError(response, core.E(core.CodeInvalidInput, "exactly one role query is required", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	staff, err := api.service.ListStaff(request.Context(), authenticated.principal, core.Role(roles[0]))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, staff)
}

func (api *API) listStudentOnboarding(response http.ResponseWriter, request *http.Request) {
	if request.URL.RawQuery != "" {
		api.writeError(response, core.E(core.CodeInvalidInput, "student onboarding query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	items, err := api.service.ListStudentOnboarding(request.Context(), authenticated.principal)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, items)
}

func (api *API) listStudents(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	for key, values := range query {
		if key != "asOf" || len(values) != 1 || strings.TrimSpace(values[0]) == "" {
			api.writeError(response, core.E(core.CodeInvalidInput, "only one non-empty asOf query parameter is supported", nil))
			return
		}
	}
	var asOf time.Time
	if raw, exists := query["asOf"]; exists {
		parsed, err := time.Parse(time.RFC3339, raw[0])
		if err != nil {
			api.writeError(response, core.E(core.CodeInvalidInput, "asOf must be an RFC3339 timestamp", nil))
			return
		}
		asOf = parsed
	}
	authenticated := authenticatedPrincipal(request)
	items, err := api.service.ListStudents(request.Context(), authenticated.principal, app.ListStudentsInput{AsOf: asOf})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, items)
}

type scheduleLessonRequest struct {
	Title            string         `json:"title"`
	StartsAt         time.Time      `json:"startsAt"`
	DurationMinutes  int            `json:"durationMinutes"`
	Location         optionalString `json:"location"`
	TeacherAccountID string         `json:"teacherAccountId"`
	StudentIDs       []string       `json:"studentIds"`
}

func (api *API) scheduleLesson(response http.ResponseWriter, request *http.Request) {
	var input scheduleLessonRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	location := ""
	if input.Location.Set {
		location = input.Location.Value
		if strings.TrimSpace(location) == "" {
			api.writeError(response, core.E(core.CodeInvalidInput, "location must not be empty when provided", nil))
			return
		}
	}
	authenticated := authenticatedPrincipal(request)
	result, err := api.service.ScheduleLesson(request.Context(), authenticated.principal, app.ScheduleLessonInput{
		Title: input.Title, StartsAt: input.StartsAt, DurationMinutes: input.DurationMinutes,
		Location: location, TeacherAccountID: input.TeacherAccountID,
		StudentIDs: input.StudentIDs, IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, result)
}

func (api *API) listLessons(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	for key, values := range query {
		if key != "from" && key != "to" && key != "studentId" && key != "teacherAccountId" {
			api.writeError(response, core.E(core.CodeInvalidInput, "unsupported Lesson query parameter", nil))
			return
		}
		if len(values) != 1 || strings.TrimSpace(values[0]) == "" {
			api.writeError(response, core.E(core.CodeInvalidInput, "Lesson query parameters must appear exactly once and be non-empty", nil))
			return
		}
	}
	fromValues, fromExists := query["from"]
	toValues, toExists := query["to"]
	if !fromExists || !toExists || len(fromValues) != 1 || len(toValues) != 1 {
		api.writeError(response, core.E(core.CodeInvalidInput, "exactly one from and to query parameter is required", nil))
		return
	}
	from, err := time.Parse(time.RFC3339, fromValues[0])
	if err != nil {
		api.writeError(response, core.E(core.CodeInvalidInput, "from must be an RFC3339 timestamp", nil))
		return
	}
	to, err := time.Parse(time.RFC3339, toValues[0])
	if err != nil {
		api.writeError(response, core.E(core.CodeInvalidInput, "to must be an RFC3339 timestamp", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	result, err := api.service.ListLessons(request.Context(), authenticated.principal, app.ListLessonsInput{
		From: from, To: to, StudentID: query.Get("studentId"), TeacherAccountID: query.Get("teacherAccountId"),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, result)
}

func (api *API) getLesson(response http.ResponseWriter, request *http.Request) {
	if request.URL.RawQuery != "" {
		api.writeError(response, core.E(core.CodeInvalidInput, "Lesson detail query parameters are not supported", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	result, err := api.service.GetLesson(request.Context(), authenticated.principal, pathID(request, "lessonId"))
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, result)
}

type replaceLessonTeacherRequestTarget struct {
	LessonID                         string  `json:"lessonId"`
	ExpectedVersion                  *int64  `json:"expectedVersion"`
	ExpectedPreviousTeacherAccountID *string `json:"expectedPreviousTeacherAccountId"`
}

type replaceLessonTeachersRequest struct {
	Lessons             []replaceLessonTeacherRequestTarget `json:"lessons"`
	NewTeacherAccountID string                              `json:"newTeacherAccountId"`
}

func (api *API) replaceLessonTeachers(response http.ResponseWriter, request *http.Request) {
	var input replaceLessonTeachersRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	targets := make([]core.ReplaceLessonTeacherTarget, len(input.Lessons))
	for index, item := range input.Lessons {
		if item.ExpectedVersion == nil {
			api.writeError(response, core.E(core.CodeInvalidInput, "expectedVersion is required for every Lesson", nil))
			return
		}
		if item.ExpectedPreviousTeacherAccountID == nil {
			api.writeError(response, core.E(core.CodeInvalidInput, "expectedPreviousTeacherAccountId is required for every Lesson", nil))
			return
		}
		targets[index] = core.ReplaceLessonTeacherTarget{
			LessonID: item.LessonID, ExpectedVersion: *item.ExpectedVersion,
			ExpectedPreviousTeacherAccountID: *item.ExpectedPreviousTeacherAccountID,
		}
	}
	authenticated := authenticatedPrincipal(request)
	result, err := api.service.ReplaceLessonTeachers(request.Context(), authenticated.principal, app.ReplaceLessonTeachersInput{
		Lessons: targets, NewTeacherAccountID: input.NewTeacherAccountID,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, result)
}

type reassignPrimaryTeacherRequestTarget struct {
	StudentID                 string `json:"studentId"`
	ExpectedAssignmentVersion *int64 `json:"expectedAssignmentVersion"`
}

type reassignPrimaryTeachersRequest struct {
	Students            []reassignPrimaryTeacherRequestTarget `json:"students"`
	NewTeacherAccountID string                                `json:"newTeacherAccountId"`
	EffectiveMode       core.PrimaryTeacherEffectiveMode      `json:"effectiveMode"`
	EffectiveFrom       *time.Time                            `json:"effectiveFrom"`
}

func (api *API) reassignPrimaryTeachers(response http.ResponseWriter, request *http.Request) {
	var input reassignPrimaryTeachersRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	targets := make([]core.PrimaryTeacherReassignmentTarget, len(input.Students))
	for index, item := range input.Students {
		if item.ExpectedAssignmentVersion == nil {
			api.writeError(response, core.E(core.CodeInvalidInput, "expectedAssignmentVersion is required for every Student", nil))
			return
		}
		targets[index] = core.PrimaryTeacherReassignmentTarget{
			StudentID: item.StudentID, ExpectedAssignmentVersion: *item.ExpectedAssignmentVersion,
		}
	}
	authenticated := authenticatedPrincipal(request)
	effectiveFrom := time.Time{}
	if input.EffectiveFrom != nil {
		effectiveFrom = *input.EffectiveFrom
	}
	result, err := api.service.ReassignPrimaryTeachers(request.Context(), authenticated.principal, app.ReassignPrimaryTeachersInput{
		Students: targets, NewTeacherAccountID: input.NewTeacherAccountID,
		EffectiveMode: input.EffectiveMode, EffectiveFrom: effectiveFrom,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, result)
}

type grantDelegationRequest struct {
	AdministratorAccountID string     `json:"administratorAccountId"`
	Reason                 string     `json:"reason"`
	ExpiresAt              *time.Time `json:"expiresAt"`
	CurrentPassword        string     `json:"currentPassword"`
}

func (api *API) grantDelegation(response http.ResponseWriter, request *http.Request) {
	var input grantDelegationRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	if !api.allowSensitive(request, "owner_reauthentication", authenticated.principal.AccountID, 5, 20) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many Owner reauthentication attempts", nil))
		return
	}
	result, err := api.service.GrantDelegation(request.Context(), authenticated.principal, app.GrantDelegationInput{
		AdministratorAccountID: input.AdministratorAccountID,
		Reason:                 input.Reason, ExpiresAt: input.ExpiresAt,
		CurrentPassword: input.CurrentPassword, IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, result)
}

type revokeDelegationRequest struct {
	Reason          string `json:"reason"`
	CurrentPassword string `json:"currentPassword"`
}

func (api *API) revokeDelegation(response http.ResponseWriter, request *http.Request) {
	var input revokeDelegationRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	authenticated := authenticatedPrincipal(request)
	if !api.allowSensitive(request, "owner_reauthentication", authenticated.principal.AccountID, 5, 20) {
		api.writeError(response, core.E(core.CodeRateLimited, "too many Owner reauthentication attempts", nil))
		return
	}
	err := api.service.RevokeDelegation(request.Context(), authenticated.principal, app.RevokeDelegationInput{
		DelegationID: pathID(request, "delegationId"), Reason: input.Reason,
		CurrentPassword: input.CurrentPassword, IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

type createStudentRequest struct {
	FullName            string         `json:"fullName"`
	Phone               string         `json:"phone"`
	EnrollmentReference string         `json:"enrollmentReference"`
	TeacherAccountID    string         `json:"teacherAccountId"`
	Locale              optionalString `json:"locale"`
	Timezone            optionalString `json:"timezone"`
	AdultConfirmed      bool           `json:"adultConfirmed"`
}

func (api *API) createStudent(response http.ResponseWriter, request *http.Request) {
	var input createStudentRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	locale := "ru-KZ"
	if input.Locale.Set {
		locale = input.Locale.Value
		if strings.TrimSpace(locale) == "" {
			api.writeError(response, core.E(core.CodeInvalidInput, "locale must not be empty when provided", nil))
			return
		}
	}
	timezone := "Asia/Almaty"
	if input.Timezone.Set {
		timezone = input.Timezone.Value
		if strings.TrimSpace(timezone) == "" {
			api.writeError(response, core.E(core.CodeInvalidInput, "timezone must not be empty when provided", nil))
			return
		}
	}
	authenticated := authenticatedPrincipal(request)
	result, err := api.service.CreateStudent(request.Context(), authenticated.principal, app.CreateStudentInput{
		FullName: input.FullName, Phone: input.Phone,
		EnrollmentReference: input.EnrollmentReference,
		TeacherAccountID:    input.TeacherAccountID, Locale: locale,
		Timezone: timezone, AdultConfirmed: input.AdultConfirmed,
		IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, result)
}

// optionalString preserves the contract-level distinction between an omitted
// optional value and an explicitly empty or null value. JSON null is never a
// valid locale/timezone override.
type optionalString struct {
	Value string
	Set   bool
}

func (value *optionalString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return core.E(core.CodeInvalidInput, "optional string must not be null", nil)
	}
	var decoded string
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	value.Value = decoded
	value.Set = true
	return nil
}

type publishFirstMinuteRequest struct {
	WhatWorked      string `json:"whatWorked"`
	CurrentFocus    string `json:"currentFocus"`
	NextStep        string `json:"nextStep"`
	ExpectedVersion *int64 `json:"expectedVersion"`
}

func (api *API) publishFirstMinute(response http.ResponseWriter, request *http.Request) {
	var input publishFirstMinuteRequest
	if err := api.decodeJSON(response, request, &input); err != nil {
		api.writeError(response, err)
		return
	}
	if input.ExpectedVersion == nil {
		api.writeError(response, core.E(core.CodeInvalidInput, "expectedVersion is required", nil))
		return
	}
	authenticated := authenticatedPrincipal(request)
	result, err := api.service.PublishFirstMinute(request.Context(), authenticated.principal, app.PublishFirstMinuteInput{
		StudentID: pathID(request, "studentId"), WhatWorked: input.WhatWorked,
		CurrentFocus: input.CurrentFocus, NextStep: input.NextStep,
		ExpectedVersion: *input.ExpectedVersion, IdempotencyKey: idempotencyKey(request),
	})
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, result)
}

type invitationResponse struct {
	InvitationID   string    `json:"invitationId"`
	StudentID      string    `json:"studentId"`
	Status         string    `json:"status"`
	ExpiresAt      time.Time `json:"expiresAt"`
	ActivationLink string    `json:"activationLink"`
}

func (api *API) issueInvitation(response http.ResponseWriter, request *http.Request) {
	api.handleInvitation(response, request, core.InvitationIssue)
}

func (api *API) reissueInvitation(response http.ResponseWriter, request *http.Request) {
	api.handleInvitation(response, request, core.InvitationReissue)
}

func (api *API) handleInvitation(response http.ResponseWriter, request *http.Request, mode core.InvitationMode) {
	authenticated := authenticatedPrincipal(request)
	result, link, err := api.service.IssueInvitation(
		request.Context(), authenticated.principal, pathID(request, "studentId"),
		idempotencyKey(request), mode,
	)
	if err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusCreated, invitationResponse{
		InvitationID: result.InvitationID, StudentID: result.StudentID,
		Status: result.Status, ExpiresAt: result.ExpiresAt, ActivationLink: link,
	})
}

func (api *API) revokeInvitation(response http.ResponseWriter, request *http.Request) {
	authenticated := authenticatedPrincipal(request)
	err := api.service.RevokeInvitation(
		request.Context(), authenticated.principal, pathID(request, "invitationId"),
		idempotencyKey(request),
	)
	if err != nil {
		api.writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}
