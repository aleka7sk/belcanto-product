package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/security"
)

const maxJSONBodyBytes = 64 << 10

type API struct {
	service *app.Service
	limits  *rateLimiter
}

func New(service *app.Service) http.Handler {
	api := &API{service: service, limits: newRateLimiter()}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", api.health)
	mux.HandleFunc("GET /readyz", api.ready)
	mux.HandleFunc("POST /v1/activations/preview", api.previewActivation)
	mux.HandleFunc("POST /v1/activations/complete", api.completeActivation)
	mux.HandleFunc("POST /v1/sessions", api.signIn)
	mux.HandleFunc("POST /v1/sessions/refresh", api.refresh)
	mux.HandleFunc("POST /v1/sessions/twofa", api.signInWithTwofa)
	mux.HandleFunc("POST /v1/password-resets", api.requestPasswordReset)
	mux.HandleFunc("POST /v1/password-resets/complete", api.completePasswordReset)
	mux.HandleFunc("POST /v1/activations/progress", api.activationProgress)
	mux.HandleFunc("POST /v1/activations/password", api.setActivationPassword)
	mux.HandleFunc("POST /v1/activations/contact", api.startActivationContact)
	mux.HandleFunc("POST /v1/activations/contact/verify", api.verifyActivationContact)
	mux.HandleFunc("POST /v1/activations/twofa", api.startActivationTwofa)
	mux.HandleFunc("POST /v1/activations/twofa/confirm", api.confirmActivationTwofa)
	mux.HandleFunc("POST /v1/activations/finish", api.finishActivation)

	mux.Handle("DELETE /v1/sessions/current", api.authenticated(http.HandlerFunc(api.signOut)))
	mux.Handle("GET /v1/me/bootstrap", api.authenticated(http.HandlerFunc(api.bootstrapView)))
	mux.Handle("GET /v1/me/sessions", api.authenticated(http.HandlerFunc(api.listMySessions)))
	mux.Handle("POST /v1/me/sessions/revoke-others", api.authenticated(http.HandlerFunc(api.revokeOtherSessions)))
	mux.Handle("POST /v1/me/sessions/{sessionId}/revoke", api.authenticated(http.HandlerFunc(api.revokeMySession)))
	mux.Handle("GET /v1/me/security-events", api.authenticated(http.HandlerFunc(api.listSecurityEvents)))
	mux.Handle("GET /v1/me/contacts", api.authenticated(http.HandlerFunc(api.listMyContacts)))
	mux.Handle("POST /v1/me/contacts/change", api.authenticated(http.HandlerFunc(api.startContactChange)))
	mux.Handle("POST /v1/me/contacts/confirm", api.authenticated(http.HandlerFunc(api.confirmContactChange)))
	mux.Handle("GET /v1/me/twofa", api.authenticated(http.HandlerFunc(api.twofaStatus)))
	mux.Handle("POST /v1/me/twofa/enroll", api.authenticated(http.HandlerFunc(api.startTwofaEnrollment)))
	mux.Handle("POST /v1/me/twofa/confirm", api.authenticated(http.HandlerFunc(api.confirmTwofaEnrollment)))
	mux.Handle("POST /v1/me/twofa/disable", api.authenticated(http.HandlerFunc(api.disableTwofa)))
	mux.Handle("GET /v1/me/profile", api.authenticated(http.HandlerFunc(api.myProfile)))
	mux.Handle("PUT /v1/me/profile", api.authenticated(http.HandlerFunc(api.updateMyProfile)))
	mux.Handle("GET /v1/policies", api.authenticated(http.HandlerFunc(api.listPolicies)))
	mux.Handle("POST /v1/me/policy-acceptances", api.authenticated(http.HandlerFunc(api.acceptPolicy)))
	mux.Handle("GET /v1/me/privacy", api.authenticated(http.HandlerFunc(api.privacySettings)))
	mux.Handle("PUT /v1/me/privacy", api.authenticated(http.HandlerFunc(api.updatePrivacySettings)))
	mux.Handle("GET /v1/me/data-exports", api.authenticated(http.HandlerFunc(api.listDataExports)))
	mux.Handle("POST /v1/me/data-exports", api.authenticated(http.HandlerFunc(api.createDataExport)))
	mux.Handle("GET /v1/me/deletion-request", api.authenticated(http.HandlerFunc(api.deletionRequest)))
	mux.Handle("POST /v1/me/deletion-request", api.authenticated(http.HandlerFunc(api.createDeletionRequest)))
	mux.Handle("POST /v1/me/deletion-request/cancel", api.authenticated(http.HandlerFunc(api.cancelDeletionRequest)))
	mux.Handle("GET /v1/staff", api.authenticated(http.HandlerFunc(api.listStaff)))
	mux.Handle("GET /v1/student-onboarding", api.authenticated(http.HandlerFunc(api.listStudentOnboarding)))
	mux.Handle("GET /v1/students", api.authenticated(http.HandlerFunc(api.listStudents)))
	mux.Handle("POST /v1/students/primary-teacher-reassignments", api.authenticated(http.HandlerFunc(api.reassignPrimaryTeachers)))
	mux.Handle("GET /v1/lessons", api.authenticated(http.HandlerFunc(api.listLessons)))
	mux.Handle("POST /v1/lessons", api.authenticated(http.HandlerFunc(api.scheduleLesson)))
	mux.Handle("GET /v1/rooms", api.authenticated(http.HandlerFunc(api.listRooms)))
	mux.Handle("POST /v1/rooms", api.authenticated(http.HandlerFunc(api.createRoom)))
	mux.Handle("GET /v1/lesson-series", api.authenticated(http.HandlerFunc(api.listLessonSeries)))
	mux.Handle("POST /v1/lesson-series", api.authenticated(http.HandlerFunc(api.createLessonSeries)))
	mux.Handle("GET /v1/lesson-series/{seriesId}", api.authenticated(http.HandlerFunc(api.getLessonSeries)))
	mux.Handle("POST /v1/lesson-series/{seriesId}/occurrences", api.authenticated(http.HandlerFunc(api.generateSeriesOccurrences)))
	mux.Handle("POST /v1/lesson-series/{seriesId}/status", api.authenticated(http.HandlerFunc(api.changeLessonSeriesStatus)))
	mux.Handle("GET /v1/event-categories", api.authenticated(http.HandlerFunc(api.listEventCategories)))
	mux.Handle("POST /v1/event-categories", api.authenticated(http.HandlerFunc(api.createEventCategory)))
	mux.Handle("POST /v1/event-series", api.authenticated(http.HandlerFunc(api.createEventSeries)))
	mux.Handle("POST /v1/event-series/{seriesId}/occurrences", api.authenticated(http.HandlerFunc(api.generateEventSeriesOccurrences)))
	mux.Handle("GET /v1/events", api.authenticated(http.HandlerFunc(api.listEvents)))
	mux.Handle("POST /v1/events", api.authenticated(http.HandlerFunc(api.createEvent)))
	mux.Handle("GET /v1/events/{occurrenceId}", api.authenticated(http.HandlerFunc(api.getEvent)))
	mux.Handle("POST /v1/events/{occurrenceId}/rsvp", api.authenticated(http.HandlerFunc(api.rsvpToEvent)))
	mux.Handle("POST /v1/events/{occurrenceId}/rsvp/cancel", api.authenticated(http.HandlerFunc(api.cancelEventRsvp)))
	mux.Handle("POST /v1/events/{occurrenceId}/waitlist", api.authenticated(http.HandlerFunc(api.joinEventWaitlist)))
	mux.Handle("POST /v1/events/{occurrenceId}/waitlist/leave", api.authenticated(http.HandlerFunc(api.leaveEventWaitlist)))
	mux.Handle("POST /v1/event-offers/{offerId}/confirm", api.authenticated(http.HandlerFunc(api.confirmSpotOffer)))
	mux.Handle("POST /v1/event-offers/{offerId}/decline", api.authenticated(http.HandlerFunc(api.declineSpotOffer)))
	mux.Handle("GET /v1/reschedule-requests", api.authenticated(http.HandlerFunc(api.listRescheduleRequests)))
	mux.Handle("POST /v1/reschedule-requests", api.authenticated(http.HandlerFunc(api.createRescheduleRequest)))
	mux.Handle("POST /v1/reschedule-requests/{requestId}/decide", api.authenticated(http.HandlerFunc(api.decideRescheduleRequest)))
	mux.Handle("POST /v1/reschedule-requests/{requestId}/withdraw", api.authenticated(http.HandlerFunc(api.withdrawRescheduleRequest)))
	mux.Handle("PUT /v1/journal-drafts", api.authenticated(http.HandlerFunc(api.saveJournalDraft)))
	mux.Handle("POST /v1/journals/publish", api.authenticated(http.HandlerFunc(api.publishJournal)))
	mux.Handle("GET /v1/journals/{occurrenceId}/{studentId}", api.authenticated(http.HandlerFunc(api.getJournal)))
	mux.Handle("GET /v1/students/{studentId}/journals", api.authenticated(http.HandlerFunc(api.listStudentJournals)))
	mux.Handle("GET /v1/students/{studentId}/progress", api.authenticated(http.HandlerFunc(api.listProgressEvidence)))

	mux.Handle("POST /v1/media", api.authenticated(http.HandlerFunc(api.createMedia)))
	mux.Handle("POST /v1/media/{mediaId}/chunks", api.authenticated(http.HandlerFunc(api.appendMediaChunk)))
	mux.Handle("GET /v1/media/{mediaId}", api.authenticated(http.HandlerFunc(api.getMedia)))
	mux.Handle("POST /v1/media/{mediaId}/access", api.authenticated(http.HandlerFunc(api.signMediaAccess)))
	mux.HandleFunc("GET /v1/media/{mediaId}/content", api.mediaContent)
	mux.Handle("POST /v1/homework", api.authenticated(http.HandlerFunc(api.createHomework)))
	mux.Handle("GET /v1/homework/{homeworkId}", api.authenticated(http.HandlerFunc(api.getHomework)))
	mux.Handle("POST /v1/homework/{homeworkId}/assign", api.authenticated(http.HandlerFunc(api.assignHomework)))
	mux.Handle("POST /v1/homework/{homeworkId}/start", api.authenticated(http.HandlerFunc(api.startHomework)))
	mux.Handle("POST /v1/homework/{homeworkId}/cancel", api.authenticated(http.HandlerFunc(api.cancelHomework)))
	mux.Handle("PUT /v1/homework/{homeworkId}/tasks/{taskId}", api.authenticated(http.HandlerFunc(api.markHomeworkTask)))
	mux.Handle("POST /v1/homework/{homeworkId}/submissions", api.authenticated(http.HandlerFunc(api.submitHomework)))
	mux.Handle("POST /v1/homework/{homeworkId}/review", api.authenticated(http.HandlerFunc(api.reviewHomework)))
	mux.Handle("GET /v1/students/{studentId}/homework", api.authenticated(http.HandlerFunc(api.listStudentHomework)))
	mux.Handle("PUT /v1/lessons/{lessonId}/attendance/{studentId}", api.authenticated(http.HandlerFunc(api.markAttendance)))
	mux.Handle("GET /v1/lessons/{lessonId}/attendance", api.authenticated(http.HandlerFunc(api.listLessonAttendance)))
	mux.Handle("POST /v1/students/{studentId}/songs", api.authenticated(http.HandlerFunc(api.addStudentSong)))
	mux.Handle("GET /v1/students/{studentId}/songs", api.authenticated(http.HandlerFunc(api.listStudentSongs)))
	mux.Handle("POST /v1/songs/{songId}/stage", api.authenticated(http.HandlerFunc(api.changeSongStage)))
	mux.Handle("POST /v1/students/{studentId}/goals", api.authenticated(http.HandlerFunc(api.createGoal)))
	mux.Handle("GET /v1/students/{studentId}/goals", api.authenticated(http.HandlerFunc(api.listStudentGoals)))
	mux.Handle("POST /v1/goals/{goalId}/complete", api.authenticated(http.HandlerFunc(api.completeGoal)))
	mux.Handle("POST /v1/goals/{goalId}/reframe", api.authenticated(http.HandlerFunc(api.reframeGoal)))
	mux.Handle("POST /v1/achievement-definitions", api.authenticated(http.HandlerFunc(api.createAchievementDefinition)))
	mux.Handle("GET /v1/achievement-definitions", api.authenticated(http.HandlerFunc(api.listAchievementDefinitions)))
	mux.Handle("POST /v1/achievement-definitions/{definitionId}/retire", api.authenticated(http.HandlerFunc(api.retireAchievementDefinition)))
	mux.Handle("GET /v1/operations/summary", api.authenticated(http.HandlerFunc(api.operationsSummary)))
	mux.Handle("POST /v1/students/{studentId}/assessments", api.authenticated(http.HandlerFunc(api.createAssessment)))
	mux.Handle("GET /v1/students/{studentId}/assessments", api.authenticated(http.HandlerFunc(api.listStudentAssessments)))
	mux.Handle("GET /v1/assessments/{assessmentId}", api.authenticated(http.HandlerFunc(api.getAssessment)))
	mux.Handle("PUT /v1/assessments/{assessmentId}", api.authenticated(http.HandlerFunc(api.updateAssessmentDraft)))
	mux.Handle("POST /v1/assessments/{assessmentId}/evidence", api.authenticated(http.HandlerFunc(api.addAssessmentEvidence)))
	mux.Handle("DELETE /v1/assessments/{assessmentId}/evidence/{evidenceId}", api.authenticated(http.HandlerFunc(api.removeAssessmentEvidence)))
	mux.Handle("POST /v1/assessments/{assessmentId}/publish", api.authenticated(http.HandlerFunc(api.publishAssessment)))
	mux.Handle("POST /v1/assessments/{assessmentId}/supersede", api.authenticated(http.HandlerFunc(api.supersedeAssessment)))
	mux.Handle("POST /v1/assessments/{assessmentId}/withdraw", api.authenticated(http.HandlerFunc(api.withdrawAssessment)))
	mux.Handle("POST /v1/students/{studentId}/achievements", api.authenticated(http.HandlerFunc(api.awardAchievement)))
	mux.Handle("GET /v1/students/{studentId}/achievements", api.authenticated(http.HandlerFunc(api.listStudentAwards)))
	mux.Handle("POST /v1/achievements/{awardId}/revoke", api.authenticated(http.HandlerFunc(api.revokeAchievement)))
	mux.Handle("GET /v1/me/activity", api.authenticated(http.HandlerFunc(api.activityFeed)))
	mux.Handle("POST /v1/me/activity/read", api.authenticated(http.HandlerFunc(api.markActivityRead)))
	mux.Handle("GET /v1/me/notification-preferences", api.authenticated(http.HandlerFunc(api.notificationPreferences)))
	mux.Handle("PUT /v1/me/notification-preferences", api.authenticated(http.HandlerFunc(api.updateNotificationPreference)))
	mux.Handle("POST /v1/community/posts", api.authenticated(http.HandlerFunc(api.createCommunityPost)))
	mux.Handle("GET /v1/community/posts", api.authenticated(http.HandlerFunc(api.communityFeed)))
	mux.Handle("GET /v1/community/posts/{postId}", api.authenticated(http.HandlerFunc(api.communityPost)))
	mux.Handle("POST /v1/community/posts/{postId}/comments", api.authenticated(http.HandlerFunc(api.addCommunityComment)))
	mux.Handle("POST /v1/community/remove", api.authenticated(http.HandlerFunc(api.removeCommunityContent)))
	mux.Handle("POST /v1/community/reports", api.authenticated(http.HandlerFunc(api.reportCommunityContent)))
	mux.Handle("GET /v1/community/moderation/reports", api.authenticated(http.HandlerFunc(api.moderationQueue)))
	mux.Handle("POST /v1/community/moderation/reports/{reportId}/decide", api.authenticated(http.HandlerFunc(api.decideCommunityReport)))
	mux.Handle("PUT /v1/community/blocks", api.authenticated(http.HandlerFunc(api.blockCommunityMember)))
	mux.Handle("GET /v1/community/blocks", api.authenticated(http.HandlerFunc(api.blockedCommunityMembers)))
	mux.Handle("POST /v1/lessons/teacher-replacements", api.authenticated(http.HandlerFunc(api.replaceLessonTeachers)))
	mux.Handle("GET /v1/lessons/{lessonId}", api.authenticated(http.HandlerFunc(api.getLesson)))
	mux.Handle("POST /v1/access/delegations", api.authenticated(http.HandlerFunc(api.grantDelegation)))
	mux.Handle("POST /v1/access/delegations/{delegationId}/revoke", api.authenticated(http.HandlerFunc(api.revokeDelegation)))
	mux.Handle("POST /v1/students", api.authenticated(http.HandlerFunc(api.createStudent)))
	mux.Handle("PUT /v1/students/{studentId}/first-minute", api.authenticated(http.HandlerFunc(api.publishFirstMinute)))
	mux.Handle("POST /v1/students/{studentId}/activation-invitations", api.authenticated(http.HandlerFunc(api.issueInvitation)))
	mux.Handle("POST /v1/students/{studentId}/activation-invitations/reissue", api.authenticated(http.HandlerFunc(api.reissueInvitation)))
	mux.Handle("POST /v1/activation-invitations/{invitationId}/revoke", api.authenticated(http.HandlerFunc(api.revokeInvitation)))
	mux.HandleFunc("/", api.notFound)

	return api.securityHeaders(api.requestIdentity(api.recoverPanics(mux)))
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code      core.ErrorCode `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"requestId"`
}

type principalContext struct {
	principal core.Principal
	access    string
}

type principalContextKey struct{}

type requestIDContextKey struct{}

func (api *API) authenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		authorization := request.Header.Get("Authorization")
		parts := strings.Split(authorization, " ")
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			api.writeError(response, core.E(core.CodeUnauthenticated, "authentication is required", nil))
			return
		}
		principal, err := api.service.Authenticate(request.Context(), parts[1])
		if err != nil {
			api.writeError(response, err)
			return
		}
		value := principalContext{principal: principal, access: parts[1]}
		ctx := context.WithValue(request.Context(), principalContextKey{}, value)
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

func authenticatedPrincipal(request *http.Request) principalContext {
	value, _ := request.Context().Value(principalContextKey{}).(principalContext)
	return value
}

func (api *API) requestIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestID, err := security.NewID("request")
		if err != nil {
			requestID = "request_unavailable"
		}
		response.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(request.Context(), requestIDContextKey{}, requestID)
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

func (api *API) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Cache-Control", "no-store")
		response.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		response.Header().Set("Referrer-Policy", "no-referrer")
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(response, request)
	})
}

func (api *API) recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if recover() != nil {
				api.writeError(response, core.E(core.CodeInternal, "internal server error", nil))
			}
		}()
		next.ServeHTTP(response, request)
	})
}

func (api *API) health(response http.ResponseWriter, _ *http.Request) {
	api.writeJSON(response, http.StatusOK, map[string]string{"status": "ok"})
}

func (api *API) ready(response http.ResponseWriter, request *http.Request) {
	if err := api.service.Ready(request.Context()); err != nil {
		api.writeError(response, err)
		return
	}
	api.writeJSON(response, http.StatusOK, map[string]string{"status": "ready"})
}

func (api *API) notFound(response http.ResponseWriter, _ *http.Request) {
	api.writeError(response, core.E(core.CodeNotFound, "route not found", nil))
}

func (api *API) decodeJSON(response http.ResponseWriter, request *http.Request, destination any) error {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return core.E(core.CodeInvalidInput, "Content-Type must be application/json", nil)
	}
	request.Body = http.MaxBytesReader(response, request.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return core.E(core.CodeInvalidInput, "request body is invalid", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return core.E(core.CodeInvalidInput, "request body must contain one JSON object", err)
	}
	return nil
}

func (api *API) writeJSON(response http.ResponseWriter, status int, body any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(body)
}

func (api *API) writeError(response http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := core.CodeInternal
	message := "internal server error"
	var appError *core.AppError
	if errors.As(err, &appError) {
		code = appError.Code
		message = appError.Message
		switch appError.Code {
		case core.CodeInvalidInput:
			status = http.StatusUnprocessableEntity
		case core.CodeUnauthenticated:
			status = http.StatusUnauthorized
		case core.CodeForbidden:
			status = http.StatusForbidden
		case core.CodeNotFound:
			status = http.StatusNotFound
		case core.CodeConflict, core.CodeInvalidState:
			status = http.StatusConflict
		case core.CodeInvalidActivation:
			status = http.StatusBadRequest
		case core.CodeRateLimited:
			status = http.StatusTooManyRequests
		case core.CodeUnavailable:
			status = http.StatusServiceUnavailable
			message = "service unavailable"
		case core.CodeInternal:
			status = http.StatusInternalServerError
			message = "internal server error"
		}
	}
	if status == http.StatusUnauthorized {
		response.Header().Set("WWW-Authenticate", "Bearer")
	}
	if status == http.StatusTooManyRequests {
		response.Header().Set("Retry-After", "60")
	}
	api.writeJSON(response, status, errorEnvelope{Error: errorBody{
		Code: code, Message: message, RequestID: response.Header().Get("X-Request-ID"),
	}})
}

func (api *API) allowSensitive(request *http.Request, operation, subject string, subjectCapacity, ipCapacity int) bool {
	ip := remoteIP(request)
	if !api.limits.allow(operation+":ip:"+ip, ipCapacity, time.Minute/time.Duration(ipCapacity)) {
		return false
	}
	key := operation + ":subject:" + privacyKey(subject)
	return api.limits.allow(key, subjectCapacity, time.Minute/time.Duration(subjectCapacity))
}

func remoteIP(request *http.Request) string {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return host
	}
	return request.RemoteAddr
}

func privacyKey(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:8])
}

func idempotencyKey(request *http.Request) string {
	return strings.TrimSpace(request.Header.Get("Idempotency-Key"))
}

func pathID(request *http.Request, name string) string {
	return strings.TrimSpace(request.PathValue(name))
}

func normalizeRateLimitPhone(value string) string {
	phone, err := security.NormalizePhone(value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return phone
}
