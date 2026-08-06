package app_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// TestHomeworkLifecycle pins the approved homework model: assigned →
// in_progress → submitted → reviewed → completed with immutable
// history, resumable uploads, revision loop, and named-area evidence
// (never a score) appearing only on acceptance.
func TestHomeworkLifecycle(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	_, link := readyStudentInvitation(t, fixture, "+77000000701", "ENR-701")
	const studentPassword = "Homework-student-pass-1!"
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, link), Phone: "+77000000701", Password: studentPassword,
		IdempotencyKey: "homework-activate",
	}); err != nil {
		t.Fatalf("activate student: %v", err)
	}
	student := signInPrincipal(t, fixture.service, "+77000000701", studentPassword)
	directory, err := fixture.service.ListStudents(ctx, fixture.owner, app.ListStudentsInput{})
	if err != nil || len(directory) == 0 {
		t.Fatalf("resolve student id: %v", err)
	}
	studentID := directory[len(directory)-1].StudentID

	startsAt := fixture.clock.Now().Add(time.Hour)
	lesson, err := fixture.service.ScheduleLesson(ctx, fixture.owner, app.ScheduleLessonInput{
		Title: "Вокал · практика", StartsAt: startsAt, DurationMinutes: 45,
		TeacherAccountID: fixture.teacher.AccountID, StudentIDs: []string{studentID},
		IdempotencyKey: "homework-lesson",
	})
	if err != nil {
		t.Fatalf("schedule lesson: %v", err)
	}

	dueAt := fixture.clock.Now().Add(48 * time.Hour)
	if _, err := fixture.service.CreateHomework(ctx, student, app.CreateHomeworkInput{
		OccurrenceID: lesson.ID, StudentID: studentID, Goal: "Цель",
		Assign: true, IdempotencyKey: "homework-create-student",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("student creating homework = %v, want FORBIDDEN", err)
	}
	homework, err := fixture.service.CreateHomework(ctx, fixture.teacher, app.CreateHomeworkInput{
		OccurrenceID: lesson.ID, StudentID: studentID,
		Goal:              "Три повтора последних 8 секунд в темпе 75%, затем одна контрольная запись.",
		ReadinessCriteria: "Мягкий вход в финальную ноту без подъёма плеч.",
		DueAt:             &dueAt,
		Tasks: []core.HomeworkTaskInput{
			{Title: "Разминка губами", RecommendedMinutes: 3},
			{Title: "Припев в 80% темпа", RecommendedMinutes: 5, SkillArea: "Дыхание"},
		},
		Assign:         true,
		IdempotencyKey: "homework-create",
	})
	if err != nil {
		t.Fatalf("create homework: %v", err)
	}
	if homework.Status != core.HomeworkStatusAssigned || len(homework.Tasks) != 2 || homework.Version != 1 {
		t.Fatalf("created homework = %#v", homework)
	}

	if _, err := fixture.service.SubmitHomework(ctx, student, app.SubmitHomeworkInput{
		HomeworkID: homework.ID, ExpectedVersion: homework.Version,
		IdempotencyKey: "homework-submit-early",
	}); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("submit before start = %v, want INVALID_STATE", err)
	}

	started, err := fixture.service.StartHomework(ctx, student, homework.ID, "homework-start")
	if err != nil || started.Status != core.HomeworkStatusInProgress {
		t.Fatalf("start homework = %#v, %v", started, err)
	}
	marked, err := fixture.service.MarkHomeworkTask(ctx, student, homework.ID, started.Tasks[0].ID, true)
	if err != nil || marked.Tasks[0].Status != "done" {
		t.Fatalf("mark task = %#v, %v", marked, err)
	}

	// Resumable upload: two chunks, replay tolerated, wrong offset conflicts.
	payload := bytes.Repeat([]byte{0xA5}, 96)
	media, err := fixture.service.CreateMedia(ctx, student, app.CreateMediaInput{
		Kind: "audio", ContentType: "audio/m4a", ByteSize: int64(len(payload)),
		IdempotencyKey: "homework-media",
	})
	if err != nil || media.Status != core.MediaStatusPending {
		t.Fatalf("create media = %#v, %v", media, err)
	}
	if _, err := fixture.service.AppendMediaChunk(ctx, student, media.ID, 0, payload[:64]); err != nil {
		t.Fatalf("first chunk: %v", err)
	}
	if _, err := fixture.service.AppendMediaChunk(ctx, student, media.ID, 0, payload[:64]); err != nil {
		t.Fatalf("replayed chunk should be accepted: %v", err)
	}
	if _, err := fixture.service.AppendMediaChunk(ctx, student, media.ID, 32, payload[32:]); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("wrong offset = %v, want CONFLICT", err)
	}
	uploaded, err := fixture.service.AppendMediaChunk(ctx, student, media.ID, 64, payload[64:])
	if err != nil || uploaded.Status != core.MediaStatusReady {
		t.Fatalf("finish upload = %#v, %v", uploaded, err)
	}

	submitted, err := fixture.service.SubmitHomework(ctx, student, app.SubmitHomeworkInput{
		HomeworkID: homework.ID, Note: "Две попытки, вторая лучше",
		MediaIDs: []string{media.ID}, ExpectedVersion: marked.Version,
		IdempotencyKey: "homework-submit",
	})
	if err != nil {
		t.Fatalf("submit homework: %v", err)
	}
	if submitted.Status != core.HomeworkStatusSubmitted || len(submitted.Submissions) != 1 ||
		submitted.Submissions[0].Attempt != 1 || len(submitted.Submissions[0].Media) != 1 {
		t.Fatalf("submitted homework = %#v", submitted)
	}

	if _, err := fixture.service.ReviewHomework(ctx, fixture.teacher, app.ReviewHomeworkInput{
		HomeworkID: homework.ID, Decision: core.FeedbackDecisionNeedsRevision,
		Body: "Финальная нота зажата", EvidenceArea: "Дыхание", EvidenceNote: "не так",
		ExpectedVersion: submitted.Version, IdempotencyKey: "homework-review-bad",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("evidence with needs_revision = %v, want INVALID_INPUT (DEC-006)", err)
	}
	reviewed, err := fixture.service.ReviewHomework(ctx, fixture.teacher, app.ReviewHomeworkInput{
		HomeworkID: homework.ID, Decision: core.FeedbackDecisionNeedsRevision,
		Body:            "Фраза стала устойчивее, финальные 8 секунд требуют внимания.",
		NextStep:        "Три раза в 75% темпа, затем одна финальная запись.",
		ExpectedVersion: submitted.Version, IdempotencyKey: "homework-review-1",
	})
	if err != nil || reviewed.Status != core.HomeworkStatusReviewed || len(reviewed.Feedback) != 1 {
		t.Fatalf("first review = %#v, %v", reviewed, err)
	}

	resubmitted, err := fixture.service.SubmitHomework(ctx, student, app.SubmitHomeworkInput{
		HomeworkID: homework.ID, Note: "Контрольный дубль",
		ExpectedVersion: reviewed.Version, IdempotencyKey: "homework-submit-2",
	})
	if err != nil || resubmitted.Submissions[0].Attempt != 2 {
		t.Fatalf("resubmit = %#v, %v", resubmitted, err)
	}
	accepted, err := fixture.service.ReviewHomework(ctx, fixture.teacher, app.ReviewHomeworkInput{
		HomeworkID: homework.ID, Decision: core.FeedbackDecisionAccepted,
		Body:         "Работа принята.",
		EvidenceArea: "Дыхание", EvidenceNote: "Держит финальную фразу без подъёма плеч",
		ExpectedVersion: resubmitted.Version, IdempotencyKey: "homework-review-2",
	})
	if err != nil || accepted.Status != core.HomeworkStatusCompleted {
		t.Fatalf("accept review = %#v, %v", accepted, err)
	}
	evidence, err := fixture.service.ListProgressEvidence(ctx, student, studentID)
	if err != nil || len(evidence) != 1 {
		t.Fatalf("evidence after acceptance = %#v, %v", evidence, err)
	}
	if evidence[0].SourceKind != "practice" || evidence[0].Area != "Дыхание" {
		t.Fatalf("evidence shape = %#v", evidence[0])
	}

	if _, err := fixture.service.CancelHomework(ctx, fixture.teacher, homework.ID, "поздно", "homework-cancel-late"); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("cancel completed homework = %v, want INVALID_STATE (final)", err)
	}

	// Signed access: the teacher reaches submission media via a sealed
	// short-lived link; the link dies after the TTL.
	access, err := fixture.service.SignMediaAccess(ctx, fixture.teacher, media.ID)
	if err != nil {
		t.Fatalf("sign media access: %v", err)
	}
	mediaID := media.ID
	content, contentType, err := fixture.service.MediaContentByToken(ctx, mediaID, accessToken(t, access.URL))
	if err != nil || contentType != "audio/m4a" || !bytes.Equal(content, payload) {
		t.Fatalf("media content = %d bytes, %q, %v", len(content), contentType, err)
	}
	fixture.clock.Advance(time.Hour)
	if _, _, err := fixture.service.MediaContentByToken(ctx, mediaID, accessToken(t, access.URL)); !core.IsCode(err, core.CodeUnauthenticated) {
		t.Fatalf("expired media link = %v, want UNAUTHENTICATED", err)
	}

	// Lazy expiry: a second assigned homework flips to expired after the
	// deadline and stays in history without any further consequence.
	shortDue := fixture.clock.Now().Add(30 * time.Minute)
	expiring, err := fixture.service.CreateHomework(ctx, fixture.teacher, app.CreateHomeworkInput{
		OccurrenceID: lesson.ID, StudentID: studentID, Goal: "Повторить разминку",
		DueAt: &shortDue, Assign: true, IdempotencyKey: "homework-expiring",
	})
	if err != nil {
		t.Fatalf("create expiring homework: %v", err)
	}
	fixture.clock.Advance(2 * time.Hour)
	expired, err := fixture.service.GetHomework(ctx, student, expiring.ID)
	if err != nil || expired.Status != core.HomeworkStatusExpired {
		t.Fatalf("expired homework = %#v, %v", expired, err)
	}
	if _, err := fixture.service.StartHomework(ctx, student, expiring.ID, "homework-start-expired"); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("start expired homework = %v, want INVALID_STATE", err)
	}

	list, err := fixture.service.ListStudentHomework(ctx, student, studentID)
	if err != nil || len(list) != 2 {
		t.Fatalf("student homework list = %d, %v", len(list), err)
	}
}

func accessToken(t *testing.T, url string) string {
	t.Helper()
	const marker = "token="
	index := len(url)
	for at := 0; at+len(marker) <= len(url); at++ {
		if url[at:at+len(marker)] == marker {
			index = at + len(marker)
			break
		}
	}
	if index >= len(url) {
		t.Fatalf("access url %q has no token", url)
	}
	return url[index:]
}
