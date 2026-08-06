package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// TestCoreLessonFormatCapacity pins DEC-002 at the service boundary: a
// core lesson holds one student (individual) or at most three (group);
// a fourth student is rejected as invalid input, not as a raw DB error.
func TestCoreLessonFormatCapacity(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	students := make([]string, 0, 4)
	for index := 0; index < 4; index++ {
		phone := []string{"+77000000301", "+77000000302", "+77000000303", "+77000000304"}[index]
		enrollment := []string{"ENR-301", "ENR-302", "ENR-303", "ENR-304"}[index]
		key := []string{"core-cap-1", "core-cap-2", "core-cap-3", "core-cap-4"}[index]
		created, err := fixture.service.CreateStudent(ctx, fixture.owner,
			studentInput(key, phone, enrollment, fixture.teacher.AccountID))
		if err != nil {
			t.Fatalf("create student %d: %v", index, err)
		}
		students = append(students, created.StudentID)
	}

	startsAt := fixture.clock.Now().Add(48 * time.Hour)
	schedule := func(key string, studentIDs []string) (core.Lesson, error) {
		return fixture.service.ScheduleLesson(ctx, fixture.owner, app.ScheduleLessonInput{
			Title: "Вокал · постановка дыхания", StartsAt: startsAt.Add(time.Duration(len(studentIDs)) * 2 * time.Hour),
			DurationMinutes: 45, TeacherAccountID: fixture.teacher.AccountID,
			StudentIDs: studentIDs, IdempotencyKey: key,
		})
	}

	if _, err := schedule("core-cap-four", students); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("four students = %v, want INVALID_INPUT (DEC-002)", err)
	}

	group, err := schedule("core-cap-three", students[:3])
	if err != nil {
		t.Fatalf("three students: %v", err)
	}
	if len(group.Students) != 3 {
		t.Fatalf("group lesson students = %d, want 3", len(group.Students))
	}

	individual, err := schedule("core-cap-one", students[3:4])
	if err != nil {
		t.Fatalf("one student: %v", err)
	}
	if len(individual.Students) != 1 {
		t.Fatalf("individual lesson students = %d, want 1", len(individual.Students))
	}
}

// TestCoreLessonSeriesGeneration pins DEC-004: the weekly series
// materializes occurrences in the school's civil time (Asia/Almaty),
// generation is idempotent per start time, and format constraints hold
// at series creation.
func TestCoreLessonSeriesGeneration(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	created, err := fixture.service.CreateStudent(ctx, fixture.owner,
		studentInput("series-student", "+77000000311", "ENR-311", fixture.teacher.AccountID))
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	room, err := fixture.service.CreateRoom(ctx, fixture.owner, app.CreateRoomInput{Name: "Зал на Абая"})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if _, err := fixture.service.CreateRoom(ctx, fixture.owner, app.CreateRoomInput{Name: "Зал на Абая"}); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("duplicate room name = %v, want CONFLICT", err)
	}

	if _, err := fixture.service.CreateCoreLessonSeries(ctx, fixture.owner, app.CreateCoreLessonSeriesInput{
		Format: "individual", Title: "Вокал", TeacherAccountID: fixture.teacher.AccountID,
		Weekday: 0, StartMinutes: 600, DurationMinutes: 45,
		EffectiveFrom: "2026-08-03", StudentIDs: []string{created.StudentID, created.StudentID},
		IdempotencyKey: "series-dup",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("duplicate students = %v, want INVALID_INPUT", err)
	}

	series, err := fixture.service.CreateCoreLessonSeries(ctx, fixture.owner, app.CreateCoreLessonSeriesInput{
		Format: "individual", Title: "Вокал · индивидуально",
		TeacherAccountID: fixture.teacher.AccountID, RoomID: room.ID,
		Weekday: 0, StartMinutes: 600, DurationMinutes: 45,
		EffectiveFrom: "2026-08-03", StudentIDs: []string{created.StudentID},
		IdempotencyKey: "series-create",
	})
	if err != nil {
		t.Fatalf("create series: %v", err)
	}
	if series.Format != "individual" || len(series.Students) != 1 || series.RoomID != room.ID {
		t.Fatalf("series view = %#v", series)
	}

	generation, err := fixture.service.GenerateSeriesOccurrences(ctx, fixture.owner, series.ID, 4, "series-gen-1")
	if err != nil {
		t.Fatalf("generate occurrences: %v", err)
	}
	if generation.CreatedCount != 4 {
		t.Fatalf("created occurrences = %d, want 4 weekly Mondays", generation.CreatedCount)
	}

	replay, err := fixture.service.GenerateSeriesOccurrences(ctx, fixture.owner, series.ID, 4, "series-gen-1")
	if err != nil {
		t.Fatalf("replay generation: %v", err)
	}
	if replay.CreatedCount != generation.CreatedCount {
		t.Fatalf("replayed generation = %d, want %d", replay.CreatedCount, generation.CreatedCount)
	}

	again, err := fixture.service.GenerateSeriesOccurrences(ctx, fixture.owner, series.ID, 4, "series-gen-2")
	if err != nil {
		t.Fatalf("re-generate occurrences: %v", err)
	}
	if again.CreatedCount != 0 {
		t.Fatalf("second generation created %d, want 0 (idempotent per start time)", again.CreatedCount)
	}

	lessons, err := fixture.service.ListLessons(ctx, fixture.owner, app.ListLessonsInput{
		From: fixture.clock.Now(), To: fixture.clock.Now().Add(40 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("list lessons: %v", err)
	}
	if len(lessons) != 4 {
		t.Fatalf("materialized lessons = %d, want 4", len(lessons))
	}
	almaty, err := time.LoadLocation("Asia/Almaty")
	if err != nil {
		t.Fatalf("load Almaty zone: %v", err)
	}
	for _, lesson := range lessons {
		local := lesson.StartsAt.In(almaty)
		if local.Weekday() != time.Monday || local.Hour() != 10 || local.Minute() != 0 {
			t.Fatalf("occurrence %s at %s, want Monday 10:00 Almaty", lesson.ID, local)
		}
	}
}
