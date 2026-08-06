package app_test

import (
	"context"
	"testing"

	"github.com/aleka7sk/belcanto-product/apps/api/internal/app"
	"github.com/aleka7sk/belcanto-product/apps/api/internal/core"
)

// TestGoalsAndAchievementsLifecycle pins the growth semantics: a goal
// has an explicit criterion and is reframed with a reason (never
// «failed»), completion carries a decision note, awards are
// evidence-backed against a published definition, and revocation
// preserves the original.
func TestGoalsAndAchievementsLifecycle(t *testing.T) {
	fixture := newFixture(t)
	ctx := context.Background()

	_, link := readyStudentInvitation(t, fixture, "+77000001001", "ENR-1001")
	const studentPassword = "Goal-student-pass-1!"
	if err := fixture.service.CompleteActivation(ctx, app.CompleteActivationInput{
		Token: tokenFromLink(t, link), Phone: "+77000001001", Password: studentPassword,
		IdempotencyKey: "goal-activate",
	}); err != nil {
		t.Fatalf("activate student: %v", err)
	}
	student := signInPrincipal(t, fixture.service, "+77000001001", studentPassword)
	directory, err := fixture.service.ListStudents(ctx, fixture.owner, app.ListStudentsInput{})
	if err != nil || len(directory) == 0 {
		t.Fatalf("resolve student id: %v", err)
	}
	studentID := directory[len(directory)-1].StudentID

	if _, err := fixture.service.CreateGoal(ctx, student, app.CreateGoalInput{
		StudentID: studentID, Criterion: "Цель", IdempotencyKey: "goal-student-creates",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("student creating a goal = %v, want FORBIDDEN", err)
	}
	goal, err := fixture.service.CreateGoal(ctx, fixture.teacher, app.CreateGoalInput{
		StudentID: studentID, Criterion: "Свободный длинный припев",
		Description:      "80% темпа · вдох до фразы · две попытки без повторения ошибки.",
		RelatedSkillArea: "Дыхание", IdempotencyKey: "goal-create",
	})
	if err != nil || goal.Status != core.GoalStatusActive {
		t.Fatalf("create goal = %#v, %v", goal, err)
	}

	reframed, err := fixture.service.ReframeGoal(ctx, fixture.teacher, app.ReframeGoalInput{
		GoalID: goal.ID, Reason: "Фокус сместился на выступление",
		NewCriterion:    "Сохранить свободу в припеве на сцене",
		ExpectedVersion: goal.Version, IdempotencyKey: "goal-reframe",
	})
	if err != nil || len(reframed) != 2 {
		t.Fatalf("reframe goal = %#v, %v", reframed, err)
	}
	if reframed[0].Status != core.GoalStatusCancelled ||
		reframed[0].ReplacedByGoalID != reframed[1].ID ||
		reframed[1].Status != core.GoalStatusActive {
		t.Fatalf("reframe chain = %#v", reframed)
	}
	replacement := reframed[1]

	if _, err := fixture.service.CompleteGoal(ctx, fixture.teacher, app.CompleteGoalInput{
		GoalID: goal.ID, CompletionNote: "поздно", ExpectedVersion: reframed[0].Version,
		IdempotencyKey: "goal-complete-cancelled",
	}); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("completing a cancelled goal = %v, want INVALID_STATE", err)
	}
	completed, err := fixture.service.CompleteGoal(ctx, fixture.teacher, app.CompleteGoalInput{
		GoalID: replacement.ID, CompletionNote: "Подтверждено записью выступления",
		ExpectedVersion: replacement.Version, IdempotencyKey: "goal-complete",
	})
	if err != nil || completed.Status != core.GoalStatusCompleted {
		t.Fatalf("complete goal = %#v, %v", completed, err)
	}
	goals, err := fixture.service.ListStudentGoals(ctx, student, studentID)
	if err != nil || len(goals) != 2 {
		t.Fatalf("student goals = %d, %v", len(goals), err)
	}

	if _, err := fixture.service.CreateAchievementDefinition(ctx, fixture.teacher, app.CreateAchievementDefinitionInput{
		Name: "Первое уверенное выступление", Description: "х", Category: "выступления",
		IdempotencyKey: "achdef-teacher",
	}); !core.IsCode(err, core.CodeForbidden) {
		t.Fatalf("teacher creating a definition = %v, want FORBIDDEN", err)
	}
	definition, err := fixture.service.CreateAchievementDefinition(ctx, fixture.owner, app.CreateAchievementDefinitionInput{
		Name:        "Первое уверенное выступление",
		Description: "Подтверждено видео и отзывом педагога.",
		Category:    "выступления", EvidenceRequirement: "Видео выступления",
		IdempotencyKey: "achdef-create",
	})
	if err != nil || definition.Status != "published" {
		t.Fatalf("create definition = %#v, %v", definition, err)
	}

	award, err := fixture.service.AwardAchievement(ctx, fixture.teacher, app.AwardAchievementInput{
		DefinitionID: definition.ID, StudentID: studentID,
		EvidenceNote:   "Open Stage · подтверждено видео и отзывом педагога",
		IdempotencyKey: "award-create",
	})
	if err != nil || award.Status != "awarded" || award.DefinitionName != definition.Name {
		t.Fatalf("award achievement = %#v, %v", award, err)
	}

	retired, err := fixture.service.RetireAchievementDefinition(ctx, fixture.owner, definition.ID, "achdef-retire")
	if err != nil || retired.Status != "retired" {
		t.Fatalf("retire definition = %#v, %v", retired, err)
	}
	if _, err := fixture.service.AwardAchievement(ctx, fixture.teacher, app.AwardAchievementInput{
		DefinitionID: definition.ID, StudentID: studentID,
		EvidenceNote: "ещё раз", IdempotencyKey: "award-retired",
	}); !core.IsCode(err, core.CodeInvalidState) {
		t.Fatalf("award on retired definition = %v, want INVALID_STATE", err)
	}

	revoked, err := fixture.service.RevokeAchievement(ctx, fixture.teacher, award.ID,
		"Выдано по ошибке другому ученику", "award-revoke")
	if err != nil || revoked.Status != "revoked" || revoked.EvidenceNote != award.EvidenceNote {
		t.Fatalf("revoke award = %#v, %v", revoked, err)
	}
	awards, err := fixture.service.ListStudentAwards(ctx, student, studentID)
	if err != nil || len(awards) != 1 || awards[0].Status != "revoked" {
		t.Fatalf("student awards = %#v, %v", awards, err)
	}
}
