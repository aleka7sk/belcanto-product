import type { IsoDateTime, Lesson } from "../api/contracts";
import {
  almatyDayRange,
  latestPastLesson,
  lessonAgendaState,
  lessonAgendaType,
  liveLesson,
} from "./teacherToday";

function lesson(
  id: string,
  startsAtIso: string,
  overrides: Partial<Lesson> = {},
): Lesson {
  return {
    id,
    title: `Урок ${id}`,
    startsAt: startsAtIso as IsoDateTime,
    durationMinutes: 45,
    teacher: { accountId: "account_t", fullName: "Елена Орлова" },
    students: [{ studentId: "student_1", fullName: "Аружан" }],
    status: "scheduled",
    version: 1,
    ...overrides,
  };
}

describe("teacher today derivations (TCH-TODAY-01/02)", () => {
  const now = new Date("2026-08-06T10:00:00Z").getTime(); // 15:00 Almaty

  it("derives the six-state agenda from lifecycle plus wall clock", () => {
    expect(lessonAgendaState(lesson("a", "2026-08-06T11:00:00Z"), now)).toBe("upcoming");
    expect(lessonAgendaState(lesson("b", "2026-08-06T09:30:00Z"), now)).toBe("now");
    expect(lessonAgendaState(lesson("c", "2026-08-06T08:00:00Z"), now)).toBe("completed");
    expect(
      lessonAgendaState(lesson("d", "2026-08-06T11:00:00Z", { status: "cancelled_student" }), now),
    ).toBe("cancelled");
    expect(
      lessonAgendaState(lesson("e", "2026-08-06T11:00:00Z", { status: "rescheduled" }), now),
    ).toBe("changed");
    expect(
      lessonAgendaState(lesson("f", "2026-08-06T11:00:00Z", { status: "no_show" }), now),
    ).toBe("completed");
  });

  it("splits individual and group by participant count (DEC-002)", () => {
    expect(lessonAgendaType(lesson("a", "2026-08-06T11:00:00Z"))).toBe("individual");
    expect(
      lessonAgendaType(
        lesson("b", "2026-08-06T11:00:00Z", {
          students: [
            { studentId: "student_1", fullName: "Аружан" },
            { studentId: "student_2", fullName: "Дана" },
          ],
        }),
      ),
    ).toBe("group");
  });

  it("finds the on-air lesson and the latest finished one", () => {
    const lessons = [
      lesson("past1", "2026-08-06T05:00:00Z"),
      lesson("past2", "2026-08-06T08:00:00Z"),
      lesson("live", "2026-08-06T09:40:00Z"),
      lesson("later", "2026-08-06T12:00:00Z"),
      lesson("cancelled", "2026-08-06T07:00:00Z", { status: "cancelled_school" }),
    ];
    expect(liveLesson(lessons, now)?.id).toBe("live");
    expect(latestPastLesson(lessons, now)?.id).toBe("past2");
  });

  it("frames the Asia/Almaty civil day around the moment", () => {
    const { from, to } = almatyDayRange(new Date("2026-08-06T10:00:00Z"));
    // 2026-08-06 00:00 Almaty = 2026-08-05 19:00 UTC.
    expect(from.toISOString()).toBe("2026-08-05T19:00:00.000Z");
    expect(to.getTime() - from.getTime()).toBe(24 * 60 * 60 * 1000);
  });
});
