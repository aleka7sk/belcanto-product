import {
  decodeAchievementAward,
  decodeStudentGoal,
  decodeStudentGoals,
} from "./contracts";

describe("goal and achievement contracts (STU-GROWTH-04/08)", () => {
  const goal = {
    id: "goal_1",
    studentId: "student_1",
    criterion: "Свободный длинный припев",
    description: "80% темпа · вдох до фразы",
    status: "cancelled",
    cancelReason: "Фокус сместился на выступление",
    replacedByGoalId: "goal_2",
    createdBy: { accountId: "account_t", fullName: "Елена Орлова" },
    version: 2,
    createdAt: "2026-07-16T10:00:00Z",
    updatedAt: "2026-08-06T15:00:00Z",
  };

  const award = {
    id: "award_1",
    definitionId: "achdef_1",
    definitionName: "Первое уверенное выступление",
    category: "выступления",
    studentId: "student_1",
    evidenceNote: "Open Stage · подтверждено видео и отзывом педагога",
    status: "awarded",
    awardedBy: { accountId: "account_t", fullName: "Елена Орлова" },
    awardedAt: "2026-07-28T19:00:00Z",
    definitionVersion: 1,
  };

  it("binds closed goals to their reason or decision note", () => {
    expect(decodeStudentGoal(goal).replacedByGoalId).toBe("goal_2");
    expect(() =>
      decodeStudentGoal({ ...goal, cancelReason: undefined }),
    ).toThrow("StudentGoal");
    expect(() =>
      decodeStudentGoal({
        ...goal,
        status: "completed",
        cancelReason: undefined,
        replacedByGoalId: undefined,
      }),
    ).toThrow("StudentGoal");
    expect(() =>
      decodeStudentGoal({
        ...goal,
        status: "active",
        cancelReason: undefined,
      }),
    ).toThrow("StudentGoal");
  });

  it("keeps revocation paired and evidence mandatory on awards", () => {
    expect(decodeAchievementAward(award).status).toBe("awarded");
    expect(() =>
      decodeAchievementAward({ ...award, status: "revoked" }),
    ).toThrow("AchievementAward");
    const revoked = decodeAchievementAward({
      ...award,
      status: "revoked",
      revokeReason: "Выдано по ошибке",
      revokedAt: "2026-08-01T10:00:00Z",
    });
    expect(revoked.evidenceNote).toBe(award.evidenceNote);
  });

  it("rejects smuggled points or ratings (DEC-006)", () => {
    expect(() => decodeStudentGoal({ ...goal, score: 5 })).toThrow("StudentGoal");
    expect(() => decodeAchievementAward({ ...award, xp: 100 })).toThrow(
      "AchievementAward",
    );
    expect(() => decodeStudentGoals([goal, goal])).toThrow("StudentGoalList");
  });
});
