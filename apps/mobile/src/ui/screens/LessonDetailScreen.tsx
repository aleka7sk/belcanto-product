import { LinearGradient } from "expo-linear-gradient";
import { router } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { canReadLessons } from "@/access";
import { useApiClient, type FirstMinute, type Lesson } from "@/api";
import { useSession } from "@/session";
import {
  AmbientGlow,
  InlineNotice,
  PremiumCard,
  PremiumScrollScreen,
  SecondaryButton,
  uiStyles,
} from "../components";
import { InitialsAvatar, StudentBottomNavigation } from "../lessonComponents";
import { colors, fonts, gradients, metrics, spacing, typeScale } from "../tokens";
import { apiErrorMessage, formatLessonDay, formatLessonTime } from "../viewModels";

export function LessonDetailScreen({
  lessonId,
  firstMinute,
}: {
  lessonId: string;
  firstMinute: FirstMinute;
}) {
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const bootstrap = state.bootstrap;
  const [lesson, setLesson] = useState<Lesson | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const load = useCallback(async () => {
    if (bootstrap === null || !canReadLessons(bootstrap)) return;
    setLoading(true);
    setError(null);
    try {
      setLesson(
        await runAuthenticated((accessToken) => api.getLesson(accessToken, lessonId)),
      );
    } catch (loadError) {
      setError(apiErrorMessage(loadError));
    } finally {
      setLoading(false);
    }
  }, [api, bootstrap, lessonId, runAuthenticated]);

  useEffect(() => {
    let active = true;
    queueMicrotask(() => { if (active) void load(); });
    return () => { active = false; };
  }, [load]);
  if (bootstrap === null) return null;
  if (!bootstrap.roles.includes("Student") || !canReadLessons(bootstrap)) {
    return (
      <PremiumScrollScreen>
        <InlineNotice title="Урок недоступен" body="У этой учётной записи нет ученического доступа к уроку." tone="error" />
        <SecondaryButton label="Вернуться" onPress={() => router.replace("/(protected)")} />
      </PremiumScrollScreen>
    );
  }

  return (
    <PremiumScrollScreen gutter={metrics.homeGutter}>
      <AmbientGlow />
      <Text style={styles.brand}>BELCANTO</Text>
      {loading ? <PremiumCard><Text style={uiStyles.body}>Открываем урок…</Text></PremiumCard> : null}
      {error ? (
        <View style={styles.stack}>
          <InlineNotice title="Урок не открылся" body={error} tone="error" />
          <SecondaryButton label="Повторить" onPress={() => void load()} />
        </View>
      ) : null}
      {lesson ? (
        <>
          <LinearGradient colors={gradients.feature} style={styles.hero}>
            <Text style={styles.eyebrow}>{formatLessonDay(lesson.startsAt).toUpperCase()}</Text>
            <Text style={styles.time}>{formatLessonTime(lesson.startsAt)}</Text>
            <Text accessibilityRole="header" style={styles.title}>{lesson.title}</Text>
            <Text style={styles.meta}>
              {[`${lesson.durationMinutes} мин`, lesson.location].filter(Boolean).join(" · ")}
            </Text>
          </LinearGradient>

          <PremiumCard>
            <View style={styles.teacherRow}>
              <InitialsAvatar name={lesson.teacher.fullName} />
              <View style={styles.teacherCopy}>
                <Text style={styles.label}>ПЕДАГОГ</Text>
                <Text style={styles.teacherName}>{lesson.teacher.fullName}</Text>
              </View>
            </View>
          </PremiumCard>

          <PremiumCard>
            <Text style={styles.goldLabel}>ПОДГОТОВКА К УРОКУ</Text>
            <Text style={styles.preparation}>{firstMinute.nextStep}</Text>
          </PremiumCard>
          <PremiumCard>
            <Text style={styles.cyanLabel}>ФОКУС УРОКА</Text>
            <Text style={styles.focus}>{firstMinute.currentFocus}</Text>
          </PremiumCard>
        </>
      ) : null}
      <StudentBottomNavigation
        active="schedule"
        onOpenHome={() => router.replace("/(protected)")}
        onOpenSchedule={() => router.replace("/(protected)/schedule")}
      />
    </PremiumScrollScreen>
  );
}

const styles = StyleSheet.create({
  brand: { color: colors.textPrimary, fontFamily: fonts.bold, ...typeScale.homeBrand },
  stack: { gap: spacing.md },
  hero: { borderColor: colors.borderGlass, borderRadius: 24, borderWidth: 1, marginTop: spacing.loose, padding: spacing.xxl },
  eyebrow: { color: colors.textGold, fontFamily: fonts.semibold, ...typeScale.eyebrow },
  time: { color: colors.textPrimary, fontFamily: fonts.extrabold, fontSize: 42, lineHeight: 50, marginTop: spacing.sm },
  title: { color: colors.textPrimary, fontFamily: fonts.bold, marginTop: spacing.sm, ...typeScale.cardTitle },
  meta: { color: colors.textSecondary, fontFamily: fonts.regular, marginTop: spacing.sm, ...typeScale.body },
  teacherRow: { alignItems: "center", flexDirection: "row", gap: spacing.md },
  teacherCopy: { flex: 1 },
  label: { color: colors.textMuted, fontFamily: fonts.semibold, ...typeScale.eyebrow },
  teacherName: { color: colors.textPrimary, fontFamily: fonts.bold, marginTop: spacing.xs, ...typeScale.sectionTitle },
  goldLabel: { color: colors.textGold, fontFamily: fonts.semibold, ...typeScale.eyebrow },
  cyanLabel: { color: colors.cyan, fontFamily: fonts.semibold, ...typeScale.eyebrow },
  preparation: { color: colors.textPrimary, fontFamily: fonts.bold, marginTop: spacing.md, ...typeScale.cardTitle },
  focus: { color: colors.textPrimary, fontFamily: fonts.medium, marginTop: spacing.md, ...typeScale.bodyLarge },
});
