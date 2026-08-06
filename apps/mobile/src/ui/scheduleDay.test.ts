import type { EventOccurrence, IsoDateTime, Lesson } from "@/api/contracts";
import { dateChipAccessibilityLabel } from "./patterns/dateChip";
import { dayAgenda, lessonAgendaEntryState } from "./screens/ScheduleScreen";
import { seriesTimeLabel } from "./screens/SeriesManagementScreen";

describe("schedule day view (STU-SCHEDULE-01)", () => {
  const lesson = (id: string, startsAt: string, status: Lesson["status"] = "scheduled"): Lesson => ({
    id,
    title: "Вокальная техника",
    format: "individual",
    startsAt: startsAt as IsoDateTime,
    durationMinutes: 60,
    teacher: { accountId: "teacher_1", fullName: "Елена" },
    students: [{ studentId: "student_1", fullName: "Аружан" }],
    status,
    version: 0,
  });
  const event = (id: string, startsAt: string, myRsvp?: "confirmed" | "cancelled"): EventOccurrence => ({
    id,
    categoryId: "cat_1",
    categoryName: "Караоке",
    title: "Караоке-пати",
    startsAt: startsAt as IsoDateTime,
    durationMinutes: 90,
    host: { accountId: "teacher_1", fullName: "Елена" },
    capacity: 10,
    confirmedCount: 5,
    status: "scheduled",
    version: 0,
    ...(myRsvp === undefined ? {} : { myRsvp }),
  });

  it("merges lessons with registered events in start order", () => {
    const agenda = dayAgenda(
      [lesson("lesson_1", "2026-08-06T13:30:00Z")],
      [
        event("event_1", "2026-08-06T15:00:00Z", "confirmed"),
        event("event_2", "2026-08-06T10:00:00Z"),
        event("event_3", "2026-08-07T10:00:00Z", "confirmed"),
      ],
      "2026-08-06",
    );
    expect(agenda.map((item) => item.kind)).toEqual(["lesson", "event"]);
    expect(agenda[1]).toMatchObject({ event: { id: "event_1" } });
  });

  it("maps Lesson statuses onto agenda states", () => {
    expect(lessonAgendaEntryState("scheduled")).toBe("upcoming");
    expect(lessonAgendaEntryState("rescheduled")).toBe("changed");
    expect(lessonAgendaEntryState("cancelled_school")).toBe("cancelled");
    expect(lessonAgendaEntryState("completed")).toBe("completed");
  });

  it("announces the full date and activity count for screen readers", () => {
    const date = new Date("2026-08-06T10:00:00Z");
    expect(dateChipAccessibilityLabel(date, 0)).toContain("занятий нет");
    expect(dateChipAccessibilityLabel(date, 2)).toContain("активностей: 2");
  });

  it("labels a weekly series slot in Almaty civil terms", () => {
    expect(seriesTimeLabel({ weekday: 0, startMinutes: 600, durationMinutes: 45 })).toBe(
      "Пн · 10:00 · 45 мин",
    );
    expect(seriesTimeLabel({ weekday: 6, startMinutes: 725, durationMinutes: 90 })).toBe(
      "Вс · 12:05 · 90 мин",
    );
  });
});
