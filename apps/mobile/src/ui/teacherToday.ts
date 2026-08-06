import type { Lesson } from "@/api/contracts";
import { parseAlmatyLocalDateTime } from "@/validation/datetime";
import type { AgendaEntryState, AgendaEntryType } from "./patterns/agendaEntry";

/**
 * Teacher today derivations (Figma TCH-TODAY-01/02). The school day is
 * Asia/Almaty civil time; agenda states follow the six-state Lesson
 * lifecycle plus the wall clock.
 */

const ALMATY_DAY_FORMAT = new Intl.DateTimeFormat("ru-KZ", {
  timeZone: "Asia/Almaty",
  day: "2-digit",
  month: "2-digit",
  year: "numeric",
});

/** Civil-day window [00:00, 24:00) in Asia/Almaty around the given moment. */
export function almatyDayRange(now: Date): { from: Date; to: Date } {
  const civilDate = ALMATY_DAY_FORMAT.format(now);
  const start = parseAlmatyLocalDateTime(civilDate, "00:00");
  if (start === null) {
    // Unreachable for a valid clock; fall back to a rolling 24h window.
    return { from: now, to: new Date(now.getTime() + 24 * 60 * 60 * 1000) };
  }
  return { from: new Date(start), to: new Date(start + 24 * 60 * 60 * 1000) };
}

export function lessonAgendaState(lesson: Lesson, nowMs: number): AgendaEntryState {
  if (lesson.status === "cancelled_school" || lesson.status === "cancelled_student") {
    return "cancelled";
  }
  if (lesson.status === "rescheduled") {
    return "changed";
  }
  const startMs = new Date(lesson.startsAt).getTime();
  const endMs = startMs + lesson.durationMinutes * 60_000;
  if (lesson.status === "completed" || lesson.status === "no_show" || endMs <= nowMs) {
    return "completed";
  }
  if (startMs <= nowMs) {
    return "now";
  }
  return "upcoming";
}

export function lessonAgendaType(lesson: Lesson): AgendaEntryType {
  return lesson.students.length > 1 ? "group" : "individual";
}

/** The lesson currently on air, if any. */
export function liveLesson(lessons: readonly Lesson[], nowMs: number): Lesson | null {
  return lessons.find((lesson) => lessonAgendaState(lesson, nowMs) === "now") ?? null;
}

/**
 * The most recent past lesson of the window — the candidate for the
 * «Незавершённый итог» card while its journals are draft or missing.
 */
export function latestPastLesson(
  lessons: readonly Lesson[],
  nowMs: number,
): Lesson | null {
  const past = lessons
    .filter((lesson) => lessonAgendaState(lesson, nowMs) === "completed")
    .filter(
      (lesson) => lesson.status === "scheduled" || lesson.status === "completed",
    )
    .sort(
      (left, right) =>
        new Date(right.startsAt).getTime() - new Date(left.startsAt).getTime(),
    );
  return past[0] ?? null;
}
