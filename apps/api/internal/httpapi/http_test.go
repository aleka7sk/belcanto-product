package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/httpapi"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/store/memory"
)

const (
	httpOwnerPassword   = "Owner-password-123!"
	httpAdminPassword   = "Admin-password-123!"
	httpTeacherPassword = "Teacher-password-123!"
	httpStudentPassword = "Student-password-123!"
)

type httpFixture struct {
	server        *httptest.Server
	service       *app.Service
	store         *memory.Store
	ownerAccess   string
	adminAccess   string
	teacherAccess string
	owner         core.Principal
	admin         core.Principal
	teacher       core.Principal
}

func newHTTPFixture(t *testing.T) *httpFixture {
	t.Helper()
	ctx := context.Background()
	store := memory.New()
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x71}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})
	ownerLink, _, err := service.BootstrapOwner(ctx, app.BootstrapOwnerInput{
		TenantID: "tenant_http", TenantName: "Belcanto HTTP",
		FullName: "HTTP Owner", Phone: "+77001000001",
		Operator: "http-test-operator", Reason: "HTTP fixture bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap HTTP Owner: %v", err)
	}
	activateDirect(t, service, ownerLink, "+77001000001", httpOwnerPassword, "http-owner")
	ownerTokensOutcome, err := service.SignIn(ctx, "+77001000001", httpOwnerPassword, core.SessionClientInfo{})
	if err != nil {
		t.Fatalf("sign in HTTP Owner: %v", err)
	}
	if ownerTokensOutcome.Tokens == nil {
		t.Fatal("sign-in returned a second-factor challenge; tokens expected")
	}
	ownerTokens := *ownerTokensOutcome.Tokens
	owner, err := service.Authenticate(ctx, ownerTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate HTTP Owner: %v", err)
	}

	adminLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "HTTP Administrator", Phone: "+77001000002", Role: core.RoleAdministrator,
		Operator: "http-test-operator", Reason: "HTTP fixture staff bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap HTTP Administrator: %v", err)
	}
	activateDirect(t, service, adminLink, "+77001000002", httpAdminPassword, "http-admin")
	adminTokensOutcome, err := service.SignIn(ctx, "+77001000002", httpAdminPassword, core.SessionClientInfo{})
	if err != nil {
		t.Fatalf("sign in HTTP Administrator: %v", err)
	}
	if adminTokensOutcome.Tokens == nil {
		t.Fatal("sign-in returned a second-factor challenge; tokens expected")
	}
	adminTokens := *adminTokensOutcome.Tokens
	admin, err := service.Authenticate(ctx, adminTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate HTTP Administrator: %v", err)
	}

	teacherLink, _, err := service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: owner.TenantID, OwnerAccountID: owner.AccountID,
		FullName: "HTTP Teacher", Phone: "+77001000003", Role: core.RoleTeacher,
		Operator: "http-test-operator", Reason: "HTTP fixture staff bootstrap",
	})
	if err != nil {
		t.Fatalf("bootstrap HTTP Teacher: %v", err)
	}
	activateDirect(t, service, teacherLink, "+77001000003", httpTeacherPassword, "http-teacher")
	teacherTokensOutcome, err := service.SignIn(ctx, "+77001000003", httpTeacherPassword, core.SessionClientInfo{})
	if err != nil {
		t.Fatalf("sign in HTTP Teacher: %v", err)
	}
	if teacherTokensOutcome.Tokens == nil {
		t.Fatal("sign-in returned a second-factor challenge; tokens expected")
	}
	teacherTokens := *teacherTokensOutcome.Tokens
	teacher, err := service.Authenticate(ctx, teacherTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate HTTP Teacher: %v", err)
	}
	server := httptest.NewServer(httpapi.New(service))
	t.Cleanup(server.Close)
	return &httpFixture{
		server: server, service: service, store: store, ownerAccess: ownerTokens.AccessToken,
		adminAccess: adminTokens.AccessToken, teacherAccess: teacherTokens.AccessToken,
		owner: owner, admin: admin, teacher: teacher,
	}
}

func TestHTTPClosedAccessJourney(t *testing.T) {
	fixture := newHTTPFixture(t)

	response := fixture.do(t, http.MethodGet, "/readyz", nil, "", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("readiness status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	var readiness map[string]string
	decodeResponse(t, response, &readiness)
	if readiness["status"] != "ready" {
		t.Fatalf("readiness response = %#v", readiness)
	}

	response = fixture.do(t, http.MethodPost, "/v1/signup", map[string]any{}, "", "")
	assertHTTPError(t, response, http.StatusNotFound, core.CodeNotFound)

	response = fixture.do(t, http.MethodPost, "/v1/activations/preview", map[string]any{
		"token": "invalid", "unexpected": true,
	}, "", "")
	assertHTTPError(t, response, http.StatusUnprocessableEntity, core.CodeInvalidInput)

	response = fixture.do(t, http.MethodGet, "/v1/staff?role=Teacher", nil, fixture.adminAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("ordinary Administrator Teacher discovery status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	var teachers []core.StaffMember
	decodeResponse(t, response, &teachers)
	if len(teachers) != 1 || teachers[0].AccountID != fixture.teacher.AccountID {
		t.Fatalf("ordinary Administrator Teacher discovery = %#v", teachers)
	}
	response = fixture.do(t, http.MethodGet, "/v1/student-onboarding", nil, fixture.adminAccess, "")
	assertHTTPError(t, response, http.StatusForbidden, core.CodeForbidden)
	response = fixture.do(t, http.MethodGet, "/v1/staff?role=Student", nil, fixture.ownerAccess, "")
	assertHTTPError(t, response, http.StatusUnprocessableEntity, core.CodeInvalidInput)
	response = fixture.do(t, http.MethodGet, "/v1/student-onboarding?tenantId=other", nil, fixture.ownerAccess, "")
	assertHTTPError(t, response, http.StatusUnprocessableEntity, core.CodeInvalidInput)

	response = fixture.do(t, http.MethodGet, "/v1/staff?role=Administrator", nil, fixture.ownerAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Owner staff discovery status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	var administrators []core.StaffMember
	decodeResponse(t, response, &administrators)
	if len(administrators) != 1 || administrators[0].AccountID != fixture.admin.AccountID {
		t.Fatalf("Administrator discovery = %#v", administrators)
	}

	response = fixture.do(t, http.MethodPost, "/v1/access/delegations", map[string]any{
		"administratorAccountId": fixture.admin.AccountID,
		"reason":                 "HTTP Owner grant", "expiresAt": nil,
		"currentPassword": httpOwnerPassword,
	}, fixture.ownerAccess, "http-grant-admin")
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("grant status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	var delegation core.DelegationResult
	decodeResponse(t, response, &delegation)
	if delegation.Bundle != core.StudentOnboardingManagerV1 {
		t.Fatalf("delegation = %#v", delegation)
	}
	response = fixture.do(t, http.MethodGet, "/v1/staff?role=Administrator", nil, fixture.ownerAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("delegated Administrator discovery status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	decodeResponse(t, response, &administrators)
	if len(administrators) != 1 || administrators[0].OnboardingDelegationID != delegation.ID {
		t.Fatalf("delegation identity missing from staff discovery: %#v", administrators)
	}

	response = fixture.do(t, http.MethodGet, "/v1/staff?role=Teacher", nil, fixture.adminAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("delegated Teacher discovery status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	teachers = nil
	decodeResponse(t, response, &teachers)
	if len(teachers) != 1 || teachers[0].AccountID != fixture.teacher.AccountID {
		t.Fatalf("Teacher discovery = %#v", teachers)
	}
	response = fixture.do(t, http.MethodGet, "/v1/staff?role=Administrator", nil, fixture.adminAccess, "")
	assertHTTPError(t, response, http.StatusForbidden, core.CodeForbidden)

	invalidOptionalValues := []struct {
		name     string
		locale   any
		timezone any
	}{
		{name: "empty-locale", locale: "", timezone: "Asia/Almaty"},
		{name: "null-locale", locale: nil, timezone: "Asia/Almaty"},
		{name: "empty-timezone", locale: "ru-KZ", timezone: ""},
		{name: "null-timezone", locale: "ru-KZ", timezone: nil},
	}
	for index, test := range invalidOptionalValues {
		response = fixture.do(t, http.MethodPost, "/v1/students", map[string]any{
			"fullName": "Invalid optional Student", "phone": "+7700100010" + string(rune('2'+index)),
			"enrollmentReference": "HTTP-OPTIONAL-" + test.name, "teacherAccountId": fixture.teacher.AccountID,
			"locale": test.locale, "timezone": test.timezone, "adultConfirmed": true,
		}, fixture.adminAccess, "http-invalid-optional-"+test.name)
		assertHTTPError(t, response, http.StatusUnprocessableEntity, core.CodeInvalidInput)
	}

	studentBody := map[string]any{
		"fullName": "HTTP Adult Student", "phone": "+77001000101",
		"enrollmentReference": "HTTP-ENR-101", "teacherAccountId": fixture.teacher.AccountID,
		"adultConfirmed": true,
	}
	response = fixture.do(t, http.MethodPost, "/v1/students", studentBody, fixture.adminAccess, "http-create-student")
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create Student status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	var student core.StudentResult
	decodeResponse(t, response, &student)
	response = fixture.do(t, http.MethodGet, "/v1/student-onboarding", nil, fixture.teacherAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Teacher onboarding queue status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	var queue []core.StudentOnboardingItem
	queue = nil
	decodeResponse(t, response, &queue)
	if len(queue) != 1 || queue[0].StudentID != student.StudentID || queue[0].EnrollmentReference != "HTTP-ENR-101" || queue[0].OnboardingState != core.OnboardingAwaitingFirstMinute || queue[0].StudentVersion != 0 {
		t.Fatalf("initial HTTP onboarding queue = %#v", queue)
	}

	response = fixture.do(t, http.MethodPut, "/v1/students/"+student.StudentID+"/first-minute", map[string]any{
		"whatWorked": "HTTP worked", "currentFocus": "HTTP focus",
		"nextStep": "HTTP next", "expectedVersion": 0,
	}, fixture.teacherAccess, "http-first-minute")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("publish First Minute status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	_ = response.Body.Close()
	response = fixture.do(t, http.MethodGet, "/v1/student-onboarding", nil, fixture.adminAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("ready onboarding queue status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	queue = nil
	decodeResponse(t, response, &queue)
	if len(queue) != 1 || queue[0].OnboardingState != core.OnboardingReadyToInvite || queue[0].StudentVersion != 1 {
		t.Fatalf("ready HTTP onboarding queue = %#v", queue)
	}

	response = fixture.do(t, http.MethodPost, "/v1/students/"+student.StudentID+"/activation-invitations", nil, fixture.adminAccess, "http-admin-invite")
	assertHTTPError(t, response, http.StatusForbidden, core.CodeForbidden)

	response = fixture.do(t, http.MethodPost, "/v1/students/"+student.StudentID+"/activation-invitations", nil, fixture.ownerAccess, "http-invite")
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("issue invitation status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	var invitation struct {
		InvitationID   string    `json:"invitationId"`
		StudentID      string    `json:"studentId"`
		ExpiresAt      time.Time `json:"expiresAt"`
		ActivationLink string    `json:"activationLink"`
	}
	decodeResponse(t, response, &invitation)
	encodedInvitation, err := json.Marshal(invitation)
	if err != nil {
		t.Fatalf("marshal invitation response: %v", err)
	}
	if strings.Contains(strings.ToLower(string(encodedInvitation)), "cipher") {
		t.Fatalf("invitation response leaked recoverable-token field: %s", encodedInvitation)
	}
	response = fixture.do(t, http.MethodPost, "/v1/students/"+student.StudentID+"/activation-invitations/reissue", nil, fixture.adminAccess, "http-admin-reissue")
	assertHTTPError(t, response, http.StatusForbidden, core.CodeForbidden)
	response = fixture.do(t, http.MethodPost, "/v1/activation-invitations/"+invitation.InvitationID+"/revoke", nil, fixture.adminAccess, "http-admin-revoke")
	assertHTTPError(t, response, http.StatusForbidden, core.CodeForbidden)
	token := httpTokenFromLink(t, invitation.ActivationLink)
	response = fixture.do(t, http.MethodGet, "/v1/student-onboarding", nil, fixture.ownerAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("invited onboarding queue status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	queue = nil
	decodeResponse(t, response, &queue)
	if len(queue) != 1 || queue[0].OnboardingState != core.OnboardingInvited || queue[0].InvitationID != invitation.InvitationID || queue[0].InvitationExpiresAt == nil {
		t.Fatalf("invited HTTP onboarding queue = %#v", queue)
	}

	response = fixture.do(t, http.MethodPost, "/v1/activations/preview", map[string]any{"token": token}, "", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("activation preview status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	_ = response.Body.Close()

	response = fixture.do(t, http.MethodPost, "/v1/activations/complete", map[string]any{
		"token": token, "phone": "+77001000101", "password": httpStudentPassword,
	}, "", "http-activate-student")
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("activation status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	if body := readBody(t, response); body != "" {
		t.Fatalf("activation created an implicit response/session: %q", body)
	}
	response = fixture.do(t, http.MethodGet, "/v1/student-onboarding", nil, fixture.adminAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("activated onboarding queue status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	queue = nil
	decodeResponse(t, response, &queue)
	if len(queue) != 1 || queue[0].OnboardingState != core.OnboardingActivated || queue[0].InvitationID != "" || queue[0].InvitationExpiresAt != nil {
		t.Fatalf("activated HTTP onboarding queue = %#v", queue)
	}

	response = fixture.do(t, http.MethodPost, "/v1/sessions", map[string]any{
		"phone": "+77001000101", "password": httpStudentPassword,
	}, "", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Student sign-in status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	var studentOutcome core.SignInOutcome
	decodeResponse(t, response, &studentOutcome)
	if studentOutcome.Tokens == nil || studentOutcome.Tokens.AccessToken == "" || studentOutcome.Tokens.RefreshToken == "" {
		t.Fatalf("Student sign-in outcome = %#v", studentOutcome)
	}
	studentTokens := *studentOutcome.Tokens

	response = fixture.do(t, http.MethodGet, "/v1/me/bootstrap", nil, studentTokens.AccessToken, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Student bootstrap status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	var view core.BootstrapView
	decodeResponse(t, response, &view)
	if view.StudentID != student.StudentID || view.FirstMinute == nil || view.FirstMinute.WhatWorked != "HTTP worked" {
		t.Fatalf("Student bootstrap view = %#v", view)
	}

	response = fixture.do(t, http.MethodPost, "/v1/access/delegations/"+delegation.ID+"/revoke", map[string]any{
		"reason": "HTTP replay authorization test", "currentPassword": httpOwnerPassword,
	}, fixture.ownerAccess, "http-revoke-admin")
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("revoke delegation status = %d, body=%s", response.StatusCode, readBody(t, response))
	}
	_ = response.Body.Close()
	response = fixture.do(t, http.MethodPost, "/v1/students", studentBody, fixture.adminAccess, "http-create-student")
	assertHTTPError(t, response, http.StatusForbidden, core.CodeForbidden)
}

func TestReadinessFailureIsInternal(t *testing.T) {
	store := &readinessFailureStore{Store: memory.New(), err: errors.New("database unavailable")}
	codec, err := security.NewTokenCodec(bytes.Repeat([]byte{0x61}, 32))
	if err != nil {
		t.Fatalf("new token codec: %v", err)
	}
	service := app.NewService(store, codec, security.NewPasswordHasher(), app.Options{
		ActivationBaseURL: "https://app.belcanto.test/activate",
		AccessTTL:         15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
		InvitationTTL: 7 * 24 * time.Hour,
	})
	server := httptest.NewServer(httpapi.New(service))
	defer server.Close()
	response, err := http.Get(server.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET readiness: %v", err)
	}
	assertHTTPError(t, response, http.StatusServiceUnavailable, core.CodeUnavailable)
}

func TestEmptyOwnerCurrentPasswordIsInvalidInput(t *testing.T) {
	fixture := newHTTPFixture(t)
	emptyGrant := fixture.do(t, http.MethodPost, "/v1/access/delegations", map[string]any{
		"administratorAccountId": fixture.admin.AccountID,
		"reason":                 "Empty password contract", "expiresAt": nil,
		"currentPassword": "",
	}, fixture.ownerAccess, "empty-owner-password-grant")
	assertHTTPError(t, emptyGrant, http.StatusUnprocessableEntity, core.CodeInvalidInput)
	emptyRevoke := fixture.do(t, http.MethodPost, "/v1/access/delegations/delegation_missing/revoke", map[string]any{
		"reason": "Empty password contract", "currentPassword": "",
	}, fixture.ownerAccess, "empty-owner-password-revoke")
	assertHTTPError(t, emptyRevoke, http.StatusUnprocessableEntity, core.CodeInvalidInput)
}

func TestOwnerReauthenticationIsRateLimitedAcrossGrantAndRevoke(t *testing.T) {
	fixture := newHTTPFixture(t)
	for attempt := 0; attempt < 5; attempt++ {
		path := "/v1/access/delegations"
		body := map[string]any{
			"administratorAccountId": fixture.admin.AccountID,
			"reason":                 "Wrong password attempt", "expiresAt": nil,
			"currentPassword": "Wrong-owner-password-123!",
		}
		if attempt%2 == 1 {
			path = "/v1/access/delegations/delegation_missing/revoke"
			body = map[string]any{
				"reason":          "Wrong password attempt",
				"currentPassword": "Wrong-owner-password-123!",
			}
		}
		response := fixture.do(t, http.MethodPost, path, body, fixture.ownerAccess, "wrong-owner-password-"+string(rune('a'+attempt)))
		assertHTTPError(t, response, http.StatusUnauthorized, core.CodeUnauthenticated)
	}
	response := fixture.do(t, http.MethodPost, "/v1/access/delegations", map[string]any{
		"administratorAccountId": fixture.admin.AccountID,
		"reason":                 "Sixth password attempt", "expiresAt": nil,
		"currentPassword": "Wrong-owner-password-123!",
	}, fixture.ownerAccess, "wrong-owner-password-f")
	assertHTTPError(t, response, http.StatusTooManyRequests, core.CodeRateLimited)
}

func (fixture *httpFixture) do(t *testing.T, method, path string, body any, accessToken, idempotency string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, fixture.server.URL+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	if idempotency != "" {
		request.Header.Set("Idempotency-Key", idempotency)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	return response
}

func activateDirect(t *testing.T, service *app.Service, link, phone, password, key string) {
	t.Helper()
	if err := service.CompleteActivation(context.Background(), app.CompleteActivationInput{
		Token: httpTokenFromLink(t, link), Phone: phone, Password: password,
		IdempotencyKey: key,
	}); err != nil {
		t.Fatalf("activate directly: %v", err)
	}
}

func httpTokenFromLink(t *testing.T, link string) string {
	t.Helper()
	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse activation link: %v", err)
	}
	values, err := url.ParseQuery(parsed.Fragment)
	if err != nil {
		t.Fatalf("parse activation fragment: %v", err)
	}
	return values.Get("token")
}

func assertHTTPError(t *testing.T, response *http.Response, status int, code core.ErrorCode) {
	t.Helper()
	if response.StatusCode != status {
		t.Fatalf("status = %d, want %d; body=%s", response.StatusCode, status, readBody(t, response))
	}
	var envelope struct {
		Error struct {
			Code      core.ErrorCode `json:"code"`
			Message   string         `json:"message"`
			RequestID string         `json:"requestId"`
		} `json:"error"`
	}
	decodeResponse(t, response, &envelope)
	if envelope.Error.Code != code || envelope.Error.Message == "" || envelope.Error.RequestID == "" {
		t.Fatalf("error envelope = %#v", envelope)
	}
}

func decodeResponse(t *testing.T, response *http.Response, destination any) {
	t.Helper()
	defer response.Body.Close()
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func readBody(t *testing.T, response *http.Response) string {
	t.Helper()
	defer response.Body.Close()
	value, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return string(value)
}

type readinessFailureStore struct {
	app.Store
	err error
}

func (store *readinessFailureStore) Ready(context.Context) error { return store.err }
