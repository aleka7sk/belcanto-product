import { router } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import { RefreshControl, StyleSheet, Text, View } from "react-native";

import {
  canCreateLessons,
  canReadLessons,
  canReassignPrimaryTeachers,
  canReplaceLessonTeachers,
} from "@/access";
import { useApiClient, type IsoDateTime, type Lesson } from "@/api";
import { useSession } from "@/session";
import {
  ErrorNotice,
  InlineNotice,
  PremiumCard,
  PremiumScrollScreen,
  PrimaryButton,
  SecondaryButton,
  TextAction,
  uiStyles,
} from "../components";
import { ScreenList } from "../screen";
import { semantic, space } from "../tokens";
import { apiErrorMessage, formatLessonDay, formatLessonTime, roleLabel } from "../viewModels";
import { RoleNav } from "./account/shared";

export function splitLessons(lessons: Lesson[], nowMs: number): { upcoming: Lesson[]; past: Lesson[] } {
  return {
    upcoming: lessons.filter(
      (lesson) => new Date(lesson.startsAt).getTime() > nowMs,
    ),
    past: lessons
      .filter((lesson) => new Date(lesson.startsAt).getTime() <= nowMs)
      .sort(
        (left, right) =>
          new Date(right.startsAt).getTime() - new Date(left.startsAt).getTime(),
      ),
  };
}

function splitLessonsNow(lessons: Lesson[]): { upcoming: Lesson[]; past: Lesson[] } {
  return splitLessons(lessons, Date.now());
}

export type StaffLessonRow =
  | { rowKind: "section"; id: string; title: string; count: number }
  | { rowKind: "upcoming"; id: string; lesson: Lesson }
  | { rowKind: "past"; id: string; lesson: Lesson };

/** Two windows, one virtualizable list: future lessons, then journals. */
export function buildStaffLessonRows(upcoming: Lesson[], past: Lesson[]): StaffLessonRow[] {
  const rows: StaffLessonRow[] = [];
  rows.push({ rowKind: "section", id: "section-upcoming", title: "Будущие занятия", count: upcoming.length });
  for (const lesson of upcoming) rows.push({ rowKind: "upcoming", id: lesson.id, lesson });
  if (past.length > 0) {
    rows.push({ rowKind: "section", id: "section-past", title: "Журналы уроков", count: past.length });
    for (const lesson of past) rows.push({ rowKind: "past", id: `past-${lesson.id}`, lesson });
  }
  return rows;
}

export function StaffLessonsScreen() {
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const bootstrap = state.bootstrap;
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async (refresh = false) => {
    if (bootstrap === null || !canReadLessons(bootstrap)) return;
    if (refresh) setRefreshing(true);
    else setLoading(true);
    setError(null);
    // Окно включает прошедшие 14 дней: журнал урока публикуется после
    // его начала, и педагогу нужен вход в недавние занятия.
    const now = new Date();
    const from = new Date(now.getTime() - 14 * 24 * 60 * 60 * 1000);
    const to = new Date(now.getTime() + 90 * 24 * 60 * 60 * 1000);
    const teacherOnly =
      bootstrap.roles.includes("Teacher") &&
      !bootstrap.roles.includes("Owner") &&
      !bootstrap.roles.includes("Administrator");
    try {
      setLessons(
        await runAuthenticated((accessToken) =>
          api.listLessons(accessToken, {
            from: from.toISOString() as IsoDateTime,
            to: to.toISOString() as IsoDateTime,
            ...(teacherOnly ? { teacherAccountId: bootstrap.accountId } : {}),
          }),
        ),
      );
    } catch (loadError) {
      setError(apiErrorMessage(loadError));
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [api, bootstrap, runAuthenticated]);

  useEffect(() => {
    let active = true;
    queueMicrotask(() => { if (active) void load(); });
    return () => { active = false; };
  }, [load]);
  if (bootstrap === null) return null;
  if (!canReadLessons(bootstrap)) {
    return (
      <PremiumScrollScreen navigation={<RoleNav active="schedule" />}>
        <InlineNotice title="Раздел недоступен" body="Нет права просматривать занятия." tone="error" />
        <SecondaryButton label="Назад" onPress={() => router.back()} />
      </PremiumScrollScreen>
    );
  }
  const mayChangeTeacher =
    canReplaceLessonTeachers(bootstrap) && canReassignPrimaryTeachers(bootstrap);
  const { upcoming, past } = splitLessonsNow(lessons);
  const rows = buildStaffLessonRows(upcoming, past);

  const header = (
    <View style={styles.header}>
      <Text style={uiStyles.brand}>BELCANTO</Text>
      <Text style={styles.eyebrow}>{roleLabel(bootstrap.roles).toUpperCase()}</Text>
      <Text accessibilityRole="header" style={uiStyles.screenTitle}>Занятия</Text>
      <Text style={styles.subtitle}>
        Расписание создаётся и поддерживается внутри Belcanto.
      </Text>
      <View style={styles.actions}>
        {canCreateLessons(bootstrap) ? (
          <View style={styles.actionGrow}>
            <PrimaryButton label="Создать занятие" onPress={() => router.push("/(protected)/lessons/create")} />
          </View>
        ) : null}
        {mayChangeTeacher ? (
          <View style={styles.actionGrow}>
            <SecondaryButton label="Сменить педагога" onPress={() => router.push("/(protected)/teacher-change")} />
          </View>
        ) : null}
      </View>
      {loading ? <PremiumCard><Text style={uiStyles.body}>Загружаем расписание…</Text></PremiumCard> : null}
      {error ? (
        <ErrorNotice
          actionLabel="Повторить"
          body={error}
          onAction={() => void load()}
          title="Расписание не загрузилось"
        />
      ) : null}
    </View>
  );

  return (
    <ScreenList<StaffLessonRow>
      data={loading || error !== null ? [] : rows}
      keyExtractor={(row) => row.id}
      ListEmptyComponent={
        loading || error !== null ? null : (
          <PremiumCard>
            <Text style={uiStyles.sectionTitle}>Будущих занятий пока нет</Text>
            <Text style={[uiStyles.body, styles.emptyBody]}>Создайте первое занятие для выбранных учеников.</Text>
          </PremiumCard>
        )
      }
      ListFooterComponent={
        <View style={styles.footer}>
          <SecondaryButton label="Рабочее пространство" onPress={() => router.replace({ pathname: "/(protected)", params: { workspace: "staff" } })} />
        </View>
      }
      ListHeaderComponent={header}
      navigation={<RoleNav active="schedule" />}
      refreshControl={
        <RefreshControl
          onRefresh={() => void load(true)}
          refreshing={refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      renderItem={({ item }) =>
        item.rowKind === "section" ? (
          <View style={styles.sectionHeader}>
            <Text style={uiStyles.sectionTitle}>{item.title}</Text>
            <Text style={uiStyles.supporting}>{item.count}</Text>
          </View>
        ) : item.rowKind === "upcoming" ? (
          <PremiumCard>
            <View style={styles.lessonHeader}>
              <View style={styles.lessonCopy}>
                <Text style={styles.lessonTime}>{formatLessonDay(item.lesson.startsAt)} · {formatLessonTime(item.lesson.startsAt)}</Text>
                <Text style={styles.lessonTitle}>{item.lesson.title}</Text>
              </View>
              <Text style={styles.version}>v{item.lesson.version}</Text>
            </View>
            <Text style={styles.lessonMeta}>
              {item.lesson.teacher.fullName} · {item.lesson.students.map((student) => student.fullName).join(", ")}
            </Text>
            {item.lesson.location ? <Text style={styles.lessonMeta}>{item.lesson.location}</Text> : null}
            <TextAction
              align="right"
              label="Контекст урока"
              onPress={() =>
                router.push({
                  pathname: "/(protected)/teacher/lesson/[lessonId]",
                  params: { lessonId: item.lesson.id },
                })
              }
            />
          </PremiumCard>
        ) : (
          <PremiumCard>
            <View style={styles.lessonHeader}>
              <View style={styles.lessonCopy}>
                <Text style={styles.lessonTime}>{formatLessonDay(item.lesson.startsAt)} · {formatLessonTime(item.lesson.startsAt)}</Text>
                <Text style={styles.lessonTitle}>{item.lesson.title}</Text>
              </View>
              <Text style={styles.version}>v{item.lesson.version}</Text>
            </View>
            <Text style={styles.lessonMeta}>{item.lesson.teacher.fullName}</Text>
            {item.lesson.students.map((student) => (
              <TextAction
                align="right"
                key={student.studentId}
                label={`Журнал · ${student.fullName}`}
                onPress={() =>
                  router.push({
                    pathname: "/(protected)/journal/[occurrenceId]/[studentId]",
                    params: { occurrenceId: item.lesson.id, studentId: student.studentId },
                  })
                }
              />
            ))}
          </PremiumCard>
        )
      }
      testID="staff-lessons"
    />
  );
}

const styles = StyleSheet.create({
  header: { gap: space.s2 },
  footer: { marginTop: space.s2 },
  eyebrow: {
    color: semantic.textGold,
    fontFamily: "Onest_600SemiBold",
    fontSize: 10,
    letterSpacing: 1,
    lineHeight: 13,
    marginTop: space.s10,
  },
  subtitle: {
    color: semantic.textSecondary,
    fontFamily: "Onest_400Regular",
    fontSize: 14,
    lineHeight: 20,
    marginBottom: space.s6,
  },
  actions: { flexDirection: "row", gap: space.s3, marginBottom: space.s4 },
  actionGrow: { flex: 1 },
  sectionHeader: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
    marginTop: space.s5,
  },
  emptyBody: { marginTop: space.s2 },
  lessonHeader: { alignItems: "flex-start", flexDirection: "row" },
  lessonCopy: { flex: 1 },
  lessonTime: {
    color: semantic.textGold,
    fontFamily: "Onest_600SemiBold",
    fontSize: 10,
    letterSpacing: 1,
    lineHeight: 13,
  },
  lessonTitle: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 19,
    lineHeight: 23,
    marginTop: space.s2,
  },
  lessonMeta: {
    color: semantic.textSecondary,
    fontFamily: "Onest_400Regular",
    fontSize: 12,
    lineHeight: 17,
    marginTop: space.s2,
  },
  version: {
    color: semantic.textMuted,
    fontFamily: "Onest_500Medium",
    fontSize: 8,
    letterSpacing: 0.8,
    lineHeight: 11,
  },
});
