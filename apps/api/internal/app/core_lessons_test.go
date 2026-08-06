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
