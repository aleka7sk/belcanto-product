import type { IsoDateTime, Lesson } from "@/api";
import { buildStaffLessonRows, splitLessons } from "./screens/StaffLessonsScreen";

const lesson = (id: string, startsAt: string): Lesson => ({
  id,
  title: "Вокальная техника",
  format: "individual",
  startsAt: startsAt as IsoDateTime,
  durationMinutes: 60,
  location: "Класс 1",
  status: "scheduled",
  version: 1,
  teacher: { accountId: "t1", fullName: "Педагог" },
  students: [{ studentId: "s1", fullName: "Ученица" }],
});

const NOW = new Date("2026-08-06T12:00:00Z").getTime();

describe("staff lessons virtualized rows", () => {
  it("splits around the given moment and sorts the journal window newest-first", () => {
    const { upcoming, past } = splitLessons(
      [
        lesson("future", "2026-08-07T10:00:00Z"),
        lesson("old", "2026-08-01T10:00:00Z"),
        lesson("recent", "2026-08-05T10:00:00Z"),
      ],
      NOW,
    );
    expect(upcoming.map((entry) => entry.id)).toEqual(["future"]);
    expect(past.map((entry) => entry.id)).toEqual(["recent", "old"]);
  });

  it("always shows the upcoming section and adds journals only when present", () => {
    const emptyRows = buildStaffLessonRows([], []);
    expect(emptyRows.map((row) => row.id)).toEqual(["section-upcoming"]);

    const rows = buildStaffLessonRows(
      [lesson("a", "2026-08-07T10:00:00Z")],
      [lesson("b", "2026-08-05T10:00:00Z")],
    );
    expect(rows.map((row) => row.id)).toEqual([
      "section-upcoming",
      "a",
      "section-past",
      "past-b",
    ]);
  });
});
