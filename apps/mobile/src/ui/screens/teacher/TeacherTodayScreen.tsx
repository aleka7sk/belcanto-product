import { router } from "expo-router";

import { ApiError, useApiClient } from "@/api";
import type { IsoDateTime, Lesson, LessonJournal } from "@/api/contracts";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
} from "../../patterns/accountPatterns";
import { AgendaEntry } from "../../patterns/agendaEntry";
import {
  almatyDayRange,
  latestPastLesson,
  lessonAgendaState,
  lessonAgendaType,
  liveLesson,
} from "../../teacherToday";
import { apiErrorMessage, formatLessonTime } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * Teacher today cockpit (Figma TCH-TODAY-01/02). The Asia/Almaty civil
 * day drives the agenda; the on-air lesson takes the «Сейчас» heading;
 * the most recent finished lesson surfaces its unfinished journal (a
 * draft or a not-yet-started recap) as the continue card. The design's
 * lesson-goal and student-prep cards wait for their own models.
 */

interface UnfinishedJournal {
  lesson: Lesson;
  studentId: string;
  studentName: string;
  state: "draft" | "missing";
}

interface TodayView {
  lessons: Lesson[];
  unfinished: UnfinishedJournal | null;
}

function todayAgenda(lessons: readonly Lesson[]): {
  live: Lesson | null;
  rest: { lesson: Lesson; state: ReturnType<typeof lessonAgendaState> }[];
} {
  const nowMs = Date.now();
  const live = liveLesson(lessons, nowMs);
  return {
    live,
    rest: lessons
      .filter((lesson) => lesson.id !== live?.id)
      .map((lesson) => ({ lesson, state: lessonAgendaState(lesson, nowMs) })),
  };
}

export function TeacherTodayScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { state } = useSession();
  const bootstrap = state.bootstrap;

  const today = useAccountResource<TodayView>(async (accessToken) => {
    if (bootstrap === null) return { lessons: [], unfinished: null };
    const { from, to } = almatyDayRange(new Date());
    const teacherOnly =
      bootstrap.roles.includes("Teacher") &&
      !bootstrap.roles.includes("Owner") &&
      !bootstrap.roles.includes("Administrator");
    const lessons = await api.listLessons(accessToken, {
      from: from.toISOString() as IsoDateTime,
      to: to.toISOString() as IsoDateTime,
      ...(teacherOnly ? { teacherAccountId: bootstrap.accountId } : {}),
    });
    const past = latestPastLesson(lessons, Date.now());
    let unfinished: UnfinishedJournal | null = null;
    if (past !== null && past.teacher.accountId === bootstrap.accountId) {
      const journals = await Promise.all(
        past.students.map(async (student) => {
          try {
            const journal: LessonJournal = await api.getJournal(
              accessToken,
              past.id,
              student.studentId,
            );
            return { student, journal };
          } catch (cause) {
            if (cause instanceof ApiError && cause.code === "NOT_FOUND") {
              return { student, journal: null };
            }
            throw cause;
          }
        }),
      );
      for (const entry of journals) {
        if (entry.journal === null) {
          unfinished = {
            lesson: past,
            studentId: entry.student.studentId,
            studentName: entry.student.fullName,
            state: "missing",
          };
          break;
        }
        if (entry.journal.status === "draft") {
          unfinished = {
            lesson: past,
            studentId: entry.student.studentId,
            studentName: entry.student.fullName,
            state: "draft",
          };
          break;
        }
      }
    }
    return { lessons, unfinished };
  });

  const view = today.value;
  if (view === null) {
    return (
      <AccountScreenShell navigation={<AccountNav active="today" />} testID="teacher-today-loading">
        {today.error !== null ? (
          <InlineNotice
            body={apiErrorMessage(today.error)}
            title={message("common.retry")}
            tone="error"
          />
        ) : null}
      </AccountScreenShell>
    );
  }

  const { live, rest } = todayAgenda(view.lessons);
  const empty = view.lessons.length === 0;

  const formatLine = (lesson: Lesson) =>
    [
      lesson.students.length > 1
        ? message("tch.format.group", { count: lesson.students.length })
        : message("tch.format.individual"),
      lesson.location,
    ]
      .filter((part): part is string => part !== undefined && part !== "")
      .join(" · ");

  const openLesson = (lesson: Lesson) =>
    router.push({
      pathname: "/(protected)/teacher/lesson/[lessonId]",
      params: { lessonId: lesson.id },
    });

  return (
    <AccountScreenShell navigation={<AccountNav active="today" />} testID="teacher-today">
      <ScreenHeading
        eyebrow={message("tch.today.eyebrow")}
        subtitle={
          empty
            ? message("tch.today.empty.body")
            : message("tch.today.subtitle", { count: view.lessons.length })
        }
        title={empty ? message("tch.today.empty.title") : message("tch.today.title")}
      />
      {live !== null ? (
        <>
          <ScreenHeading
            eyebrow={message("tch.today.now", { time: formatLessonTime(live.startsAt) })}
            subtitle={formatLine(live)}
            title={live.title}
          />
          <BlockAction
            label={message("tch.today.openLesson")}
            onPress={() => openLesson(live)}
            testID="teacher-today-open-live"
          />
        </>
      ) : null}
      {view.unfinished !== null ? (
        <StatusCard
          body={message(
            view.unfinished.state === "draft"
              ? "tch.unfinished.draft"
              : "tch.unfinished.missing",
            {
              name: view.unfinished.studentName,
              date: formatLessonTime(view.unfinished.lesson.startsAt),
            },
          )}
          status={message("tch.unfinished.action")}
          title={message("tch.unfinished.title")}
          tone="warning"
        />
      ) : null}
      {view.unfinished !== null ? (
        <BlockAction
          kind="secondary"
          label={message("tch.unfinished.action")}
          onPress={() =>
            router.push({
              pathname: "/(protected)/journal/[occurrenceId]/[studentId]",
              params: {
                occurrenceId: view.unfinished!.lesson.id,
                studentId: view.unfinished!.studentId,
              },
            })
          }
          testID="teacher-today-continue-journal"
        />
      ) : null}
      {rest.map(({ lesson, state: agendaState }) => (
        <AgendaEntry
          action={message("tch.today.entry.open")}
          eyebrow={formatLessonTime(lesson.startsAt)}
          key={lesson.id}
          onPress={() => openLesson(lesson)}
          state={agendaState}
          supporting={lesson.students.map((student) => student.fullName).join(", ")}
          testID={`teacher-today-${lesson.id}`}
          timePlace={formatLine(lesson)}
          title={lesson.title}
          type={lessonAgendaType(lesson)}
        />
      ))}
      <BlockAction
        kind="secondary"
        label={message("tch.today.allLessons")}
        onPress={() => router.push("/(protected)/lessons")}
      />
    </AccountScreenShell>
  );
}
