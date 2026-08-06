package app_test

import (
	"context"
	"testing"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// TestRepertoireLifecycle pins the StudentSong journey: the assigned
// Teacher adds a song at the first stage, moves it along the design's
// journey with append-only history, CAS protects concurrent edits, and
// the Student reads but never leads the path.
func TestRepertoireLifecycle(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	_, link := readyStudentInvitation(t, fixture, "+77000000901", "ENR-901")
	const studentPassword = "Repertoire-student-pass-1!"
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, link), Phone: "+77000000901", Password: studentPassword,
		IdempotencyKey: "song-activate",
	}); err != nil {
		t.Fatalf("activate student: %v", err)
	}
	student := signInPrincipal(t, fixture.service, "+77000000901", studentPassword)
	directory, err := fixture.service.ListStudents(ctx, fixture.owner, app.ListStudentsInput{})
	if err != nil || len(directory) == 0 {
		t.Fatalf("resolve student id: %v", err)
	}
	studentID := directory[len(directory)-1].StudentID

	if _, err := fixture.service.AddStudentSong(ctx, student, app.AddStudentSongInput{
		StudentID: studentID, Title: "Easy On Me", IdempotencyKey: "song-student-adds",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("student adding a song = %v, want FORBIDDEN", err)
	}
	song, err := fixture.service.AddStudentSong(ctx, fixture.teacher, app.AddStudentSongInput{
		StudentID: studentID, Title: "Easy On Me", Artist: "Adele",
		IdempotencyKey: "song-add",
	})
	if err != nil {
		t.Fatalf("add song: %v", err)
	}
	if song.Stage != core.SongStages[0] || len(song.History) != 1 || song.Version != 1 {
		t.Fatalf("new song = %#v", song)
	}

	if _, err := fixture.service.ChangeSongStage(ctx, fixture.teacher, app.ChangeSongStageInput{
		SongID: song.ID, Stage: "solo_world_tour", ExpectedVersion: song.Version,
		IdempotencyKey: "song-bad-stage",
	}); !core.IsCode(err, core.CodeInvalidInput) {
		t.Fatalf("unknown stage = %v, want INVALID_INPUT", err)
	}
	moved, err := fixture.service.ChangeSongStage(ctx, fixture.teacher, app.ChangeSongStageInput{
		SongID: song.ID, Stage: "technically_stable",
		StageNote:       "Записать припев целиком без потери опоры",
		ExpectedVersion: song.Version, IdempotencyKey: "song-move",
	})
	if err != nil {
		t.Fatalf("change stage: %v", err)
	}
	if moved.Stage != "technically_stable" || len(moved.History) != 2 || moved.Version != 2 {
		t.Fatalf("moved song = %#v", moved)
	}
	if moved.History[0].FromStage != core.SongStages[0] || moved.History[0].ToStage != "technically_stable" {
		t.Fatalf("history head = %#v", moved.History[0])
	}

	if _, err := fixture.service.ChangeSongStage(ctx, fixture.teacher, app.ChangeSongStageInput{
		SongID: song.ID, Stage: "interpretation", ExpectedVersion: song.Version,
		IdempotencyKey: "song-stale",
	}); !core.IsCode(err, core.CodeConflict) {
		t.Fatalf("stale version = %v, want CONFLICT", err)
	}

	list, err := fixture.service.ListStudentSongs(ctx, student, studentID)
	if err != nil || len(list) != 1 {
		t.Fatalf("student repertoire = %#v, %v", list, err)
	}
	if list[0].AssignedBy.AccountID != fixture.teacher.AccountID {
		t.Fatalf("assigned-by = %#v", list[0].AssignedBy)
	}
	teacherList, err := fixture.service.ListStudentSongs(ctx, fixture.teacher, studentID)
	if err != nil || len(teacherList) != 1 {
		t.Fatalf("teacher repertoire = %#v, %v", teacherList, err)
	}
}
