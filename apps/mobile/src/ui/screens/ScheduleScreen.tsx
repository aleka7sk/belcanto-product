import { router } from "expo-router";
import { useCallback, useEffect, useMemo, useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { canReadLessons } from "@/access";
import { useApiClient, type IsoDateTime, type Lesson } from "@/api";
import { useSession } from "@/session";
import {
  AmbientGlow,
  InlineNotice,
  PremiumCard,
  PremiumScrollScreen,
  SecondaryButton,
  uiStyles,
} from "../components";
import { DateStrip, LessonCard, StudentBottomNavigation } from "../lessonComponents";
import { colors, fonts, metrics, spacing, typeScale } from "../tokens";
import { apiErrorMessage, dateKey, nextScheduleDates } from "../viewModels";

export function ScheduleScreen() {
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const bootstrap = state.bootstrap;
  const dates = useMemo(() => nextScheduleDates(new Date(), 7), []);
  const [selectedDate, setSelectedDate] = useState(() => dateKey(dates[0]!));
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (bootstrap?.studentId === undefined || !canReadLessons(bootstrap)) return;
    setLoading(true);
    setError(null);
    const from = new Date();
    const to = new Date(from.getTime() + 8 * 24 * 60 * 60 * 1000);
    try {
      const result = await runAuthenticated((accessToken) =>
        api.listLessons(accessToken, {
          from: from.toISOString() as IsoDateTime,
          to: to.toISOString() as IsoDateTime,
          studentId: bootstrap.studentId,
        }),
      );
      setLessons(result);
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
      <PremiumScrollScreen>
        <InlineNotice title="Раздел недоступен" body="Расписание доступно ученику и сотрудникам с правом просмотра уроков." tone="error" />
        <SecondaryButton label="Вернуться" onPress={() => router.replace("/(protected)")} />
      </PremiumScrollScreen>
    );
  }
  const visible = lessons
    .filter((lesson) => dateKey(new Date(lesson.startsAt)) === selectedDate)
    .sort(
      (left, right) =>
        new Date(left.startsAt).getTime() - new Date(right.startsAt).getTime(),
    );

  return (
    <PremiumScrollScreen gutter={metrics.homeGutter}>
      <AmbientGlow />
      <Text style={styles.brand}>BELCANTO</Text>
      <Text style={styles.eyebrow}>ВАШЕ ВРЕМЯ</Text>
      <Text accessibilityRole="header" style={styles.title}>Расписание</Text>
      <Text style={styles.subtitle}>Только занятия, созданные школой в Belcanto.</Text>
      <DateStrip dates={dates} onSelect={setSelectedDate} selected={selectedDate} />

      <View style={styles.sectionHeader}>
        <Text style={uiStyles.sectionTitle}>Уроки дня</Text>
        <Text style={uiStyles.supporting}>{visible.length}</Text>
      </View>
      {loading ? <PremiumCard><Text style={uiStyles.body}>Загружаем уроки…</Text></PremiumCard> : null}
      {error ? (
        <View style={styles.stack}>
          <InlineNotice title="Расписание не загрузилось" body={error} tone="error" />
          <SecondaryButton label="Повторить" onPress={() => void load()} />
        </View>
      ) : null}
      {!loading && !error && visible.length === 0 ? (
        <PremiumCard>
          <Text style={uiStyles.sectionTitle}>В этот день занятий нет</Text>
          <Text style={[uiStyles.body, styles.emptyBody]}>Можно выбрать другой день.</Text>
        </PremiumCard>
      ) : null}
      <View style={styles.stack}>
        {visible.map((lesson) => (
          <LessonCard
            key={lesson.id}
            lesson={lesson}
            onPress={() => router.push({ pathname: "/(protected)/lesson/[lessonId]", params: { lessonId: lesson.id } })}
          />
        ))}
      </View>
      <StudentBottomNavigation
        active="schedule"
        onOpenHome={() => router.replace("/(protected)")}
        onOpenSchedule={() => undefined}
      />
    </PremiumScrollScreen>
  );
}

const styles = StyleSheet.create({
  brand: { color: colors.textPrimary, fontFamily: fonts.bold, ...typeScale.homeBrand },
  eyebrow: { color: colors.textGold, fontFamily: fonts.semibold, marginTop: spacing.loose, ...typeScale.eyebrow },
  title: { color: colors.textPrimary, fontFamily: fonts.extrabold, marginTop: spacing.sm, ...typeScale.screenTitle },
  subtitle: { color: colors.textSecondary, fontFamily: fonts.regular, marginBottom: spacing.xxl, marginTop: spacing.sm, ...typeScale.body },
  sectionHeader: { alignItems: "center", flexDirection: "row", justifyContent: "space-between", marginTop: spacing.section },
  stack: { gap: spacing.md },
  emptyBody: { marginTop: spacing.sm },
});
