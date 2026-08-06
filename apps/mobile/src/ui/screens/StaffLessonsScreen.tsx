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
  AmbientGlow,
  InlineNotice,
  PremiumCard,
  PremiumScrollScreen,
  PrimaryButton,
  SecondaryButton,
  TextAction,
  uiStyles,
} from "../components";
import { colors, fonts, metrics, spacing, typeScale } from "../tokens";
import { apiErrorMessage, formatLessonDay, formatLessonTime, roleLabel } from "../viewModels";
import { RoleNav } from "./account/shared";

function splitLessons(lessons: Lesson[]): { upcoming: Lesson[]; past: Lesson[] } {
  const nowMs = Date.now();
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
  const { upcoming, past } = splitLessons(lessons);

  return (
    <PremiumScrollScreen
      gutter={metrics.homeGutter}
      navigation={<RoleNav active="schedule" />}
      testID="staff-lessons"
      scrollProps={{
        refreshControl: (
          <RefreshControl
            onRefresh={() => void load(true)}
            refreshing={refreshing}
            tintColor={colors.violet}
          />
        ),
      }}
    >
      <AmbientGlow />
      <Text style={styles.brand}>BELCANTO</Text>
      <Text style={styles.eyebrow}>{roleLabel(bootstrap.roles).toUpperCase()}</Text>
      <Text accessibilityRole="header" style={styles.title}>Занятия</Text>
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
      <View style={styles.sectionHeader}>
        <Text style={uiStyles.sectionTitle}>Будущие занятия</Text>
        <Text style={uiStyles.supporting}>{upcoming.length}</Text>
      </View>
      {loading ? <PremiumCard><Text style={uiStyles.body}>Загружаем расписание…</Text></PremiumCard> : null}
      {error ? (
        <View style={styles.stack}>
          <InlineNotice title="Расписание не загрузилось" body={error} tone="error" />
          <SecondaryButton label="Повторить" onPress={() => void load()} />
        </View>
      ) : null}
      {!loading && !error && upcoming.length === 0 ? (
        <PremiumCard>
          <Text style={uiStyles.sectionTitle}>Будущих занятий пока нет</Text>
          <Text style={[uiStyles.body, styles.emptyBody]}>Создайте первое занятие для выбранных учеников.</Text>
        </PremiumCard>
      ) : null}
      <View style={styles.stack}>
        {upcoming.map((lesson) => (
          <PremiumCard key={lesson.id}>
            <View style={styles.lessonHeader}>
              <View style={styles.lessonCopy}>
                <Text style={styles.lessonTime}>{formatLessonDay(lesson.startsAt)} · {formatLessonTime(lesson.startsAt)}</Text>
                <Text style={styles.lessonTitle}>{lesson.title}</Text>
              </View>
              <Text style={styles.version}>v{lesson.version}</Text>
            </View>
            <Text style={styles.lessonMeta}>
              {lesson.teacher.fullName} · {lesson.students.map((student) => student.fullName).join(", ")}
            </Text>
            {lesson.location ? <Text style={styles.lessonMeta}>{lesson.location}</Text> : null}
            <TextAction
              align="right"
              label="Контекст урока"
              onPress={() =>
                router.push({
                  pathname: "/(protected)/teacher/lesson/[lessonId]",
                  params: { lessonId: lesson.id },
                })
              }
            />
          </PremiumCard>
        ))}
      </View>
      {past.length > 0 ? (
        <>
          <View style={styles.sectionHeader}>
            <Text style={uiStyles.sectionTitle}>Журналы уроков</Text>
            <Text style={uiStyles.supporting}>{past.length}</Text>
          </View>
          <View style={styles.stack}>
            {past.map((lesson) => (
              <PremiumCard key={lesson.id}>
                <View style={styles.lessonHeader}>
                  <View style={styles.lessonCopy}>
                    <Text style={styles.lessonTime}>{formatLessonDay(lesson.startsAt)} · {formatLessonTime(lesson.startsAt)}</Text>
                    <Text style={styles.lessonTitle}>{lesson.title}</Text>
                  </View>
                  <Text style={styles.version}>v{lesson.version}</Text>
                </View>
                <Text style={styles.lessonMeta}>{lesson.teacher.fullName}</Text>
                {lesson.students.map((student) => (
                  <TextAction
                    align="right"
                    key={student.studentId}
                    label={`Журнал · ${student.fullName}`}
                    onPress={() =>
                      router.push({
                        pathname: "/(protected)/journal/[occurrenceId]/[studentId]",
                        params: { occurrenceId: lesson.id, studentId: student.studentId },
                      })
                    }
                  />
                ))}
              </PremiumCard>
            ))}
          </View>
        </>
      ) : null}
      <SecondaryButton label="Рабочее пространство" onPress={() => router.replace({ pathname: "/(protected)", params: { workspace: "staff" } })} />
    </PremiumScrollScreen>
  );
}

const styles = StyleSheet.create({
  brand: { color: colors.textPrimary, fontFamily: fonts.bold, ...typeScale.brand },
  eyebrow: { color: colors.textGold, fontFamily: fonts.semibold, marginTop: metrics.workflowEyebrowTop, ...typeScale.eyebrow },
  title: { color: colors.textPrimary, fontFamily: fonts.extrabold, marginTop: spacing.sm, ...typeScale.screenTitle },
  subtitle: { color: colors.textSecondary, fontFamily: fonts.regular, marginBottom: spacing.section, marginTop: spacing.sm, ...typeScale.body },
  actions: { flexDirection: "row", gap: spacing.md },
  actionGrow: { flex: 1 },
  sectionHeader: { alignItems: "center", flexDirection: "row", justifyContent: "space-between", marginTop: spacing.section },
  stack: { gap: spacing.md },
  emptyBody: { marginTop: spacing.sm },
  lessonHeader: { alignItems: "flex-start", flexDirection: "row" },
  lessonCopy: { flex: 1 },
  lessonTime: { color: colors.textGold, fontFamily: fonts.semibold, ...typeScale.eyebrow },
  lessonTitle: { color: colors.textPrimary, fontFamily: fonts.bold, marginTop: spacing.sm, ...typeScale.cardTitle },
  lessonMeta: { color: colors.textSecondary, fontFamily: fonts.regular, marginTop: spacing.sm, ...typeScale.supporting },
  version: { color: colors.textMuted, fontFamily: fonts.medium, ...typeScale.micro },
});
