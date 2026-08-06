import { router } from "expo-router";
import { useCallback, useEffect, useMemo, useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { canReadLessons } from "@/access";
import {
  useApiClient,
  type EventOccurrence,
  type IsoDateTime,
  type Lesson,
} from "@/api";
import { useSession } from "@/session";
import {
  AmbientGlow,
  ErrorNotice,
  InlineNotice,
  PremiumCard,
  PremiumScrollScreen,
  SecondaryButton,
  uiStyles,
} from "../components";
import {
  AgendaEntry,
  type AgendaEntryState,
} from "../patterns/agendaEntry";
import { DateChip } from "../patterns/dateChip";
import { colors, fonts, spacing, metrics, typeScale } from "../tokens";
import { RoleNav } from "./account/shared";
import {
  apiErrorMessage,
  dateKey,
  formatLessonTime,
  nextScheduleDates,
} from "../viewModels";

/**
 * STU-SCHEDULE-01 «Моё расписание»: the personal day view merges core
 * Lessons with the events the Student registered for (DEC-001 keeps
 * them distinct domain types on one agenda). Date chips carry activity
 * indicators; a day with a registered event is marked by shape and
 * border, never color alone.
 */

const MONTH_EYEBROW = new Intl.DateTimeFormat("ru-RU", {
  month: "long",
  year: "numeric",
  timeZone: "Asia/Almaty",
});

const DAY_HEADING = new Intl.DateTimeFormat("ru-RU", {
  weekday: "long",
  day: "numeric",
  month: "long",
  timeZone: "Asia/Almaty",
});

export function lessonAgendaEntryState(status: Lesson["status"]): AgendaEntryState {
  switch (status) {
    case "completed":
      return "completed";
    case "cancelled_school":
    case "cancelled_student":
      return "cancelled";
    case "rescheduled":
      return "changed";
    case "no_show":
      return "completed";
    case "scheduled":
      return "upcoming";
  }
}

function timeRange(startsAt: IsoDateTime, durationMinutes: number): string {
  const end = new Date(new Date(startsAt).getTime() + durationMinutes * 60_000);
  return `${formatLessonTime(startsAt)}–${formatLessonTime(end.toISOString())}`;
}

type AgendaItem =
  | { kind: "lesson"; startsAt: IsoDateTime; lesson: Lesson }
  | { kind: "event"; startsAt: IsoDateTime; event: EventOccurrence };

export function dayAgenda(
  lessons: readonly Lesson[],
  events: readonly EventOccurrence[],
  selectedDate: string,
): AgendaItem[] {
  const items: AgendaItem[] = [];
  for (const lesson of lessons) {
    if (dateKey(new Date(lesson.startsAt)) === selectedDate) {
      items.push({ kind: "lesson", startsAt: lesson.startsAt, lesson });
    }
  }
  for (const event of events) {
    if (event.myRsvp === "confirmed" && dateKey(new Date(event.startsAt)) === selectedDate) {
      items.push({ kind: "event", startsAt: event.startsAt, event });
    }
  }
  return items.sort(
    (left, right) => new Date(left.startsAt).getTime() - new Date(right.startsAt).getTime(),
  );
}

export function ScheduleScreen() {
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const bootstrap = state.bootstrap;
  const dates = useMemo(() => nextScheduleDates(new Date(), 7), []);
  const [selectedDate, setSelectedDate] = useState(() => dateKey(dates[0]!));
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [events, setEvents] = useState<EventOccurrence[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (bootstrap?.studentId === undefined || !canReadLessons(bootstrap)) return;
    setLoading(true);
    setError(null);
    const from = new Date();
    const to = new Date(from.getTime() + 8 * 24 * 60 * 60 * 1000);
    const window = {
      from: from.toISOString() as IsoDateTime,
      to: to.toISOString() as IsoDateTime,
    };
    try {
      const [lessonResult, eventResult] = await Promise.all([
        runAuthenticated((accessToken) =>
          api.listLessons(accessToken, { ...window, studentId: bootstrap.studentId }),
        ),
        runAuthenticated((accessToken) => api.listEvents(accessToken, window)),
      ]);
      setLessons(lessonResult);
      setEvents(eventResult);
    } catch (loadError) {
      setError(apiErrorMessage(loadError));
    } finally {
      setLoading(false);
    }
  }, [api, bootstrap, runAuthenticated]);

  useEffect(() => {
    let active = true;
    queueMicrotask(() => { if (active) void load(); });
    return () => { active = false; };
  }, [load]);

  if (bootstrap === null) return null;
  if (!bootstrap.roles.includes("Student") || !canReadLessons(bootstrap)) {
    return (
      <PremiumScrollScreen navigation={<RoleNav active="schedule" />}>
        <InlineNotice title="Раздел недоступен" body="Расписание доступно ученику и сотрудникам с правом просмотра уроков." tone="error" />
        <SecondaryButton
          label="Каталог событий"
          onPress={() => router.push("/(protected)/events")}
        />
        <SecondaryButton label="Вернуться" onPress={() => router.replace("/(protected)")} />
      </PremiumScrollScreen>
    );
  }

  const registered = events.filter((event) => event.myRsvp === "confirmed");
  const countFor = (date: Date): { total: number; event: boolean } => {
    const key = dateKey(date);
    const dayLessons = lessons.filter(
      (lesson) => dateKey(new Date(lesson.startsAt)) === key,
    ).length;
    const dayEvents = registered.filter(
      (event) => dateKey(new Date(event.startsAt)) === key,
    ).length;
    return { total: dayLessons + dayEvents, event: dayEvents > 0 };
  };
  const todayKey = dateKey(dates[0]!);
  const selected = dates.find((date) => dateKey(date) === selectedDate) ?? dates[0]!;
  const agenda = dayAgenda(lessons, registered, selectedDate);

  return (
    <PremiumScrollScreen
      gutter={metrics.homeGutter}
      navigation={<RoleNav active="schedule" />}
      testID="student-schedule"
    >
      <AmbientGlow />
      <Text style={styles.eyebrow}>{MONTH_EYEBROW.format(selected).toUpperCase()}</Text>
      <Text accessibilityRole="header" style={styles.title}>Моё расписание</Text>
      <Text style={styles.subtitle}>Основные занятия и события, на которые вы записались.</Text>
      <View style={styles.chips}>
        {dates.map((date) => {
          const key = dateKey(date);
          const marks = countFor(date);
          return (
            <DateChip
              date={date}
              eventDay={marks.event}
              itemCount={marks.total}
              key={key}
              onPress={() => setSelectedDate(key)}
              selected={key === selectedDate}
              testID={`schedule-chip-${key}`}
              today={key === todayKey}
            />
          );
        })}
      </View>
      <Text style={styles.dayHeading}>
        {DAY_HEADING.format(selected).replace(",", " ·").toUpperCase()}
      </Text>
      {loading ? <PremiumCard><Text style={uiStyles.body}>Загружаем расписание…</Text></PremiumCard> : null}
      {error ? (
        <ErrorNotice
          actionLabel="Повторить"
          body={error}
          onAction={() => void load()}
          title="Расписание не загрузилось"
        />
      ) : null}
      {!loading && !error && agenda.length === 0 ? (
        <PremiumCard>
          <Text style={uiStyles.sectionTitle}>В этот день занятий нет</Text>
          <Text style={[uiStyles.body, styles.emptyBody]}>Можно выбрать другой день.</Text>
        </PremiumCard>
      ) : null}
      <View style={styles.stack}>
        {agenda.map((item) =>
          item.kind === "lesson" ? (
            <AgendaEntry
              action="Открыть детали"
              eyebrow={
                item.lesson.format === "group"
                  ? "ОСНОВНОЕ · ГРУППОВОЕ"
                  : "ОСНОВНОЕ · ИНДИВИДУАЛЬНОЕ"
              }
              key={item.lesson.id}
              onPress={() =>
                router.push({
                  pathname: "/(protected)/lesson/[lessonId]",
                  params: { lessonId: item.lesson.id },
                })
              }
              state={lessonAgendaEntryState(item.lesson.status)}
              supporting={item.lesson.teacher.fullName}
              testID={`schedule-lesson-${item.lesson.id}`}
              timePlace={[
                timeRange(item.lesson.startsAt, item.lesson.durationMinutes),
                item.lesson.location,
              ]
                .filter(Boolean)
                .join(" · ")}
              title={item.lesson.title}
              type={item.lesson.format === "group" ? "group" : "individual"}
            />
          ) : (
            <AgendaEntry
              action="Открыть событие"
              eyebrow="СОБЫТИЕ · ВЫ УЧАСТВУЕТЕ"
              key={item.event.id}
              onPress={() =>
                router.push({
                  pathname: "/(protected)/events/[occurrenceId]",
                  params: { occurrenceId: item.event.id },
                })
              }
              state={item.event.status === "cancelled" ? "cancelled" : "upcoming"}
              supporting={item.event.categoryName}
              testID={`schedule-event-${item.event.id}`}
              timePlace={timeRange(item.event.startsAt, item.event.durationMinutes)}
              title={item.event.title}
              type="event"
            />
          ),
        )}
      </View>
    </PremiumScrollScreen>
  );
}

const styles = StyleSheet.create({
  eyebrow: { color: colors.textGold, fontFamily: fonts.semibold, marginTop: spacing.loose, ...typeScale.eyebrow },
  title: { color: colors.textPrimary, fontFamily: fonts.extrabold, marginTop: spacing.sm, ...typeScale.screenTitle },
  subtitle: { color: colors.textSecondary, fontFamily: fonts.regular, marginBottom: spacing.xxl, marginTop: spacing.sm, ...typeScale.body },
  chips: { flexDirection: "row", gap: 3, justifyContent: "space-between" },
  dayHeading: { color: colors.textSecondary, fontFamily: fonts.semibold, marginTop: spacing.lg, ...typeScale.eyebrow },
  stack: { gap: spacing.md, marginTop: spacing.md },
  emptyBody: { marginTop: spacing.sm },
});
