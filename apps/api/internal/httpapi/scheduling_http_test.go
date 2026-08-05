package httpapi_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

func TestHTTPSchedulingAndContinuityContract(t *testing.T) {
	fixture := newHTTPFixture(t)
	ctx := context.Background()
	secondTeacherLink, _, err := fixture.service.BootstrapStaff(ctx, app.BootstrapStaffInput{
		TenantID: fixture.owner.TenantID, OwnerAccountID: fixture.owner.AccountID,
		FullName: "HTTP Replacement Teacher", Phone: "+77001001001", Role: core.RoleTeacher,
		Operator: "http-test-operator", Reason: "HTTP scheduling fixture",
	})
	if err != nil {
		t.Fatalf("bootstrap replacement Teacher: %v", err)
	}
	activateDirect(t, fixture.service, secondTeacherLink, "+77001001001", httpTeacherPassword, "http-replacement-teacher")
	secondTeacherTokens, err := fixture.service.SignIn(ctx, "+77001001001", httpTeacherPassword, core.SessionClientInfo{})
	if err != nil {
		t.Fatalf("sign in replacement Teacher: %v", err)
	}
	secondTeacher, err := fixture.service.Authenticate(ctx, secondTeacherTokens.AccessToken)
	if err != nil {
		t.Fatalf("authenticate replacement Teacher: %v", err)
	}

	response := fixture.do(t, http.MethodPost, "/v1/students", map[string]any{
		"fullName": "HTTP Pending Scheduling Student", "phone": "+77001001002",
		"enrollmentReference": "HTTP-SCHEDULE-1002", "teacherAccountId": fixture.teacher.AccountID,
		"adultConfirmed": true,
	}, fixture.ownerAccess, "http-schedule-create-student")
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create pending scheduling Student = %d, %s", response.StatusCode, readBody(t, response))
	}
	var student core.StudentResult
	decodeResponse(t, response, &student)

	response = fixture.do(t, http.MethodGet, "/v1/students", nil, fixture.adminAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("ordinary Administrator Student directory = %d, %s", response.StatusCode, readBody(t, response))
	}
	var directory []core.StudentDirectoryItem
	decodeResponse(t, response, &directory)
	if len(directory) != 1 || directory[0].StudentID != student.StudentID || directory[0].PrimaryTeacherAssignmentVersion != 0 || directory[0].PrimaryTeacher.Status != core.AssignedTeacherActive {
		t.Fatalf("Student directory = %#v", directory)
	}
	response = fixture.do(t, http.MethodGet, "/v1/students?asOf="+url.QueryEscape(startsAtForDirectory().Format(time.RFC3339)), nil, fixture.adminAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("asOf Student directory = %d, %s", response.StatusCode, readBody(t, response))
	}
	response = fixture.do(t, http.MethodGet, "/v1/students?unexpected=true", nil, fixture.adminAccess, "")
	assertHTTPError(t, response, http.StatusUnprocessableEntity, core.CodeInvalidInput)

	startsAt := time.Now().UTC().Add(4 * time.Hour).Truncate(time.Second)
	lessonBody := map[string]any{
		"title": "HTTP Premium Lesson", "startsAt": startsAt,
		"durationMinutes": 60, "location": "Studio A",
		"teacherAccountId": fixture.teacher.AccountID, "studentIds": []string{student.StudentID},
	}
	response = fixture.do(t, http.MethodPost, "/v1/lessons", lessonBody, fixture.adminAccess, "http-schedule-lesson")
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("schedule Lesson = %d, %s", response.StatusCode, readBody(t, response))
	}
	var lesson core.Lesson
	decodeResponse(t, response, &lesson)
	if lesson.Version != 0 || lesson.Teacher.AccountID != fixture.teacher.AccountID || len(lesson.Students) != 1 {
		t.Fatalf("scheduled Lesson = %#v", lesson)
	}

	listPath := "/v1/lessons?from=" + url.QueryEscape(time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)) +
		"&to=" + url.QueryEscape(time.Now().UTC().Add(8*time.Hour).Format(time.RFC3339))
	response = fixture.do(t, http.MethodGet, listPath, nil, fixture.teacherAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Teacher Lesson list = %d, %s", response.StatusCode, readBody(t, response))
	}
	var lessons []core.Lesson
	decodeResponse(t, response, &lessons)
	if len(lessons) != 1 || lessons[0].ID != lesson.ID {
		t.Fatalf("Teacher Lesson list = %#v", lessons)
	}
	response = fixture.do(t, http.MethodGet, "/v1/lessons?from=bad&to=bad", nil, fixture.adminAccess, "")
	assertHTTPError(t, response, http.StatusUnprocessableEntity, core.CodeInvalidInput)
	response = fixture.do(t, http.MethodGet, "/v1/lessons/"+lesson.ID, nil, fixture.adminAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Lesson detail = %d, %s", response.StatusCode, readBody(t, response))
	}

	replacementBody := map[string]any{
		"lessons": []map[string]any{{"lessonId": lesson.ID, "expectedVersion": 0,
			"expectedPreviousTeacherAccountId": fixture.teacher.AccountID}},
		"newTeacherAccountId": secondTeacher.AccountID,
	}
	replacementWithReason := map[string]any{
		"lessons": []map[string]any{{"lessonId": lesson.ID, "expectedVersion": 0,
			"expectedPreviousTeacherAccountId": fixture.teacher.AccountID}},
		"newTeacherAccountId": secondTeacher.AccountID, "reason": "sensitive free prose",
	}
	response = fixture.do(t, http.MethodPost, "/v1/lessons/teacher-replacements", replacementWithReason, fixture.adminAccess, "http-reason-rejected")
	assertHTTPError(t, response, http.StatusUnprocessableEntity, core.CodeInvalidInput)
	response = fixture.do(t, http.MethodPost, "/v1/lessons/teacher-replacements", replacementBody, fixture.teacherAccess, "http-teacher-replace-denied")
	assertHTTPError(t, response, http.StatusForbidden, core.CodeForbidden)
	response = fixture.do(t, http.MethodPost, "/v1/lessons/teacher-replacements", replacementBody, fixture.adminAccess, "http-admin-replace")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Administrator replacement = %d, %s", response.StatusCode, readBody(t, response))
	}
	var replacement core.LessonTeacherReplacementResult
	decodeResponse(t, response, &replacement)
	if replacement.UpdatedCount != 1 || replacement.Lessons[0].Version != 1 || replacement.Lessons[0].Teacher.AccountID != secondTeacher.AccountID {
		t.Fatalf("replacement result = %#v", replacement)
	}

	missingVersionBody := map[string]any{
		"students":            []map[string]any{{"studentId": student.StudentID}},
		"newTeacherAccountId": secondTeacher.AccountID,
		"effectiveMode":       "scheduled", "effectiveFrom": time.Now().UTC().Add(time.Minute),
	}
	response = fixture.do(t, http.MethodPost, "/v1/students/primary-teacher-reassignments", missingVersionBody, fixture.adminAccess, "http-missing-assignment-version")
	assertHTTPError(t, response, http.StatusUnprocessableEntity, core.CodeInvalidInput)
	reassignmentBody := map[string]any{
		"students":            []map[string]any{{"studentId": student.StudentID, "expectedAssignmentVersion": 0}},
		"newTeacherAccountId": secondTeacher.AccountID,
		"effectiveMode":       "immediate",
	}
	response = fixture.do(t, http.MethodPost, "/v1/students/primary-teacher-reassignments", reassignmentBody, fixture.teacherAccess, "http-teacher-reassign-denied")
	assertHTTPError(t, response, http.StatusForbidden, core.CodeForbidden)
	if err := fixture.store.SetAccountStatusForTest(fixture.owner.TenantID, fixture.teacher.AccountID, "suspended"); err != nil {
		t.Fatalf("disable departed HTTP Teacher: %v", err)
	}
	response = fixture.do(t, http.MethodGet, "/v1/students", nil, fixture.adminAccess, "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("directory with departed HTTP Teacher = %d, %s", response.StatusCode, readBody(t, response))
	}
	decodeResponse(t, response, &directory)
	if len(directory) != 1 || directory[0].PrimaryTeacher.Status != core.AssignedTeacherInactive {
		t.Fatalf("departed HTTP Teacher hidden from directory = %#v", directory)
	}
	response = fixture.do(t, http.MethodPost, "/v1/students/primary-teacher-reassignments", reassignmentBody, fixture.adminAccess, "http-admin-reassign")
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("Administrator reassignment = %d, %s", response.StatusCode, readBody(t, response))
	}
	var reassignment core.PrimaryTeacherReassignmentResult
	decodeResponse(t, response, &reassignment)
	if reassignment.ReassignedCount != 1 || reassignment.Assignments[0].Version != 1 {
		t.Fatalf("reassignment result = %#v", reassignment)
	}
}

func startsAtForDirectory() time.Time {
	return time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
}
