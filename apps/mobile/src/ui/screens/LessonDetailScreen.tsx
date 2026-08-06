import { LinearGradient } from "expo-linear-gradient";
import { router } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { canReadLessons } from "@/access";
import { useApiClient, type FirstMinute, type Lesson } from "@/api";
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
import { InitialsAvatar } from "../lessonComponents";
import { gradients, semantic, space } from "../tokens";
import { apiErrorMessage, formatLessonDay, formatLessonTime } from "../viewModels";
import { RoleNav } from "./account/shared";

function lessonStarted(lesson: Lesson): boolean {
  return new Date(lesson.startsAt).getTime() <= Date.now();
}

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
  const selfStudentId = bootstrap.studentId;
  if (!bootstrap.roles.includes("Student") || !canReadLessons(bootstrap)) {
    return (
      <PremiumScrollScreen navigation={<RoleNav active="schedule" />}>
        <InlineNotice title="Урок недоступен" body="У этой учётной записи нет ученического доступа к уроку." tone="error" />
        <SecondaryButton label="Вернуться" onPress={() => router.replace("/(protected)")} />
      </PremiumScrollScreen>
    );
  }

  return (
    <PremiumScrollScreen
      gutter={space.s4}
      navigation={<RoleNav active="schedule" />}
      testID="lesson-detail"
    >
      <AmbientGlow />
      <Text style={styles.brand}>BELCANTO</Text>
      {loading ? <PremiumCard><Text style={uiStyles.body}>Открываем урок…</Text></PremiumCard> : null}
      {error ? (
        <ErrorNotice
          actionLabel="Повторить"
          body={error}
          onAction={() => void load()}
          title="Урок не открылся"
        />
      ) : null}
      {lesson ? (
        <>
          <LinearGradient colors={gradients.feature} style={styles.hero}>
            <Text style={styles.eyebrow}>
              {lesson.format === "group"
                ? "ГРУППОВОЕ ЗАНЯТИЕ"
                : lesson.format === "individual"
                  ? "ИНДИВИДУАЛЬНОЕ ЗАНЯТИЕ"
                  : formatLessonDay(lesson.startsAt).toUpperCase()}
            </Text>
            <Text style={styles.time}>{formatLessonTime(lesson.startsAt)}</Text>
            <Text accessibilityRole="header" style={styles.title}>{lesson.title}</Text>
            <Text style={styles.meta}>
              {[
                lesson.format !== undefined
                  ? formatLessonDay(lesson.startsAt)
                  : undefined,
                `${lesson.durationMinutes} мин`,
                lesson.location,
              ]
                .filter(Boolean)
                .join(" · ")}
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
          {selfStudentId !== undefined && lessonStarted(lesson) ? (
            <SecondaryButton
              label="Итог урока"
              onPress={() =>
                router.push({
                  pathname: "/(protected)/journal/[occurrenceId]/[studentId]",
                  params: { occurrenceId: lesson.id, studentId: selfStudentId },
                })
              }
            />
          ) : null}
        </>
      ) : null}
    </PremiumScrollScreen>
  );
}

const EYEBROW = {
  fontFamily: "Onest_600SemiBold",
  fontSize: 10,
  letterSpacing: 1,
  lineHeight: 13,
} as const;

const styles = StyleSheet.create({
  brand: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 11,
    letterSpacing: 2,
    lineHeight: 15,
  },
  stack: { gap: space.s3 },
  hero: {
    borderColor: semantic.borderGlass,
    borderRadius: 24,
    borderWidth: 1,
    marginTop: space.s10,
    padding: space.s6,
  },
  eyebrow: { color: semantic.textGold, ...EYEBROW },
  time: {
    color: semantic.textPrimary,
    fontFamily: "Onest_800ExtraBold",
    fontSize: 42,
    lineHeight: 50,
    marginTop: space.s2,
  },
  title: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 19,
    lineHeight: 23,
    marginTop: space.s2,
  },
  meta: {
    color: semantic.textSecondary,
    fontFamily: "Onest_400Regular",
    fontSize: 14,
    lineHeight: 20,
    marginTop: space.s2,
  },
  teacherRow: { alignItems: "center", flexDirection: "row", gap: space.s3 },
  teacherCopy: { flex: 1 },
  label: { color: semantic.textMuted, ...EYEBROW },
  teacherName: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 16,
    lineHeight: 21,
    marginTop: space.s1,
  },
  goldLabel: { color: semantic.textGold, ...EYEBROW },
  cyanLabel: { color: semantic.accentCyan, ...EYEBROW },
  preparation: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 19,
    lineHeight: 23,
    marginTop: space.s3,
  },
  focus: {
    color: semantic.textPrimary,
    fontFamily: "Onest_500Medium",
    fontSize: 15,
    lineHeight: 23,
    marginTop: space.s3,
  },
});
