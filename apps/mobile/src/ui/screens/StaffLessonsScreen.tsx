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
  uiStyles,
} from "../components";
import { colors, fonts, metrics, spacing, typeScale } from "../tokens";
import { apiErrorMessage, formatLessonDay, formatLessonTime, roleLabel } from "../viewModels";

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
    const from = new Date();
    const to = new Date(from.getTime() + 90 * 24 * 60 * 60 * 1000);
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
      <PremiumScrollScreen>
        <InlineNotice title="Раздел недоступен" body="Нет права просматривать занятия." tone="error" />
        <SecondaryButton label="Назад" onPress={() => router.back()} />
      </PremiumScrollScreen>
    );
  }
  const mayChangeTeacher =
    canReplaceLessonTeachers(bootstrap) && canReassignPrimaryTeachers(bootstrap);

  return (
    <PremiumScrollScreen
      gutter={metrics.homeGutter}
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
        <Text style={uiStyles.supporting}>{lessons.length}</Text>
      </View>
      {loading ? <PremiumCard><Text style={uiStyles.body}>Загружаем расписание…</Text></PremiumCard> : null}
      {error ? (
        <View style={styles.stack}>
          <InlineNotice title="Расписание не загрузилось" body={error} tone="error" />
          <SecondaryButton label="Повторить" onPress={() => void load()} />
        </View>
      ) : null}
      {!loading && !error && lessons.length === 0 ? (
        <PremiumCard>
          <Text style={uiStyles.sectionTitle}>Будущих занятий пока нет</Text>
          <Text style={[uiStyles.body, styles.emptyBody]}>Создайте первое занятие для выбранных учеников.</Text>
        </PremiumCard>
      ) : null}
      <View style={styles.stack}>
        {lessons.map((lesson) => (
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
          </PremiumCard>
        ))}
      </View>
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
