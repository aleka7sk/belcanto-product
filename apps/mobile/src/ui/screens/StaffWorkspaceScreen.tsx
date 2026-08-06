import { router } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import {
  Pressable,
  RefreshControl,
  StyleSheet,
  Text,
  View,
} from "react-native";

import {
  canCreateStudents,
  canCreateLessons,
  canDelegateStudentOnboarding,
  canOpenStudentOnboardingQueue,
  canReassignPrimaryTeachers,
  canReplaceLessonTeachers,
} from "@/access";
import { useApiClient, type StudentOnboardingItem } from "@/api";
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
import {
  colors,
  fonts,
  metrics,
  radii,
  spacing,
  typeScale,
} from "../tokens";
import {
  apiErrorMessage,
  formatBelcantoDate,
  onboardingStateCopy,
  roleLabel,
} from "../viewModels";

export function StaffWorkspaceScreen({
  onOpenStudent,
}: {
  onOpenStudent?: (() => void) | undefined;
}) {
  const api = useApiClient();
  const { state, signOut, retryBootstrap, runAuthenticated } = useSession();
  const bootstrap = state.bootstrap;
  const [items, setItems] = useState<StudentOnboardingItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(
    async (refresh = false) => {
      if (bootstrap === null) return;
      if (refresh) {
        setRefreshing(true);
        setError(null);
        try {
          await retryBootstrap();
        } catch (refreshError) {
          setError(apiErrorMessage(refreshError));
        } finally {
          setRefreshing(false);
        }
        return;
      }
      if (!canOpenStudentOnboardingQueue(bootstrap)) {
        setLoading(false);
        setRefreshing(false);
        return;
      }
      setLoading(true);
      setError(null);
      try {
        setItems(
          await runAuthenticated((accessToken) =>
            api.listStudentOnboarding(accessToken),
          ),
        );
      } catch (loadError) {
        setError(apiErrorMessage(loadError));
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    [api, bootstrap, retryBootstrap, runAuthenticated],
  );

  useEffect(() => {
    let active = true;
    queueMicrotask(() => {
      if (active) void load();
    });
    return () => {
      active = false;
    };
  }, [load]);

  if (bootstrap === null) return null;
  const allowed = canOpenStudentOnboardingQueue(bootstrap);
  const createAllowed = canCreateStudents(bootstrap);
  const delegateAllowed = canDelegateStudentOnboarding(bootstrap);
  const createLessonAllowed = canCreateLessons(bootstrap);
  const changeTeacherAllowed =
    canReplaceLessonTeachers(bootstrap) && canReassignPrimaryTeachers(bootstrap);

  const leave = async () => {
    await signOut();
    router.replace("/");
  };

  return (
    <PremiumScrollScreen
      gutter={metrics.homeGutter}
      scrollProps={{
        refreshControl: (
          <RefreshControl
            colors={[colors.violet]}
            onRefresh={() => void load(true)}
            progressBackgroundColor={colors.raised}
            refreshing={refreshing}
            tintColor={colors.violet}
          />
        ),
      }}
    >
      <AmbientGlow />
      <Text style={styles.brand}>BELCANTO</Text>
      <Text style={styles.role}>{roleLabel(bootstrap.roles).toUpperCase()}</Text>
      <Text accessibilityRole="header" style={styles.title}>
        Рабочее пространство
      </Text>
      <Text style={styles.subtitle}>
        Внутреннее расписание, ученики и закрытый доступ — в одном месте.
      </Text>

      {onOpenStudent ? (
        <View style={styles.workspaceSwitch}>
          <SecondaryButton label="Мой профиль ученика" onPress={onOpenStudent} />
        </View>
      ) : null}

      <View style={styles.quickActions}>
        {createLessonAllowed || bootstrap.roles.includes("Teacher") ? (
          <View style={styles.actionGrow}>
            <PrimaryButton
              label="Сегодня"
              onPress={() => router.push("/(protected)/teacher")}
            />
          </View>
        ) : null}
        {createLessonAllowed ? (
          <View style={styles.actionGrow}>
            <PrimaryButton
              label="Занятия"
              onPress={() => router.push("/(protected)/lessons")}
            />
          </View>
        ) : null}
        {changeTeacherAllowed ? (
          <View style={styles.actionGrow}>
            <SecondaryButton
              label="Сменить педагога"
              onPress={() => router.push("/(protected)/teacher-change")}
            />
          </View>
        ) : null}
      </View>

      <View style={styles.quickActions}>
        {createAllowed ? (
          <View style={styles.actionGrow}>
            <SecondaryButton
              label="Добавить ученика"
              onPress={() => router.push("/(protected)/create-student")}
            />
          </View>
        ) : null}
        {delegateAllowed ? (
          <View style={styles.actionGrow}>
            <SecondaryButton
              label="Суперадмин"
              onPress={() => router.push("/(protected)/access")}
            />
          </View>
        ) : null}
      </View>

      {!allowed ? (
        <View style={styles.stackGap}>
          <InlineNotice
            body="Расписание доступно независимо. Для очереди активации владелец может отдельно выдать администратору доступ суперадминистратора."
            title="Очередь доступа скрыта"
          />
          <SecondaryButton
            disabled={refreshing}
            label={refreshing ? "Проверяем доступ…" : "Проверить доступ"}
            onPress={() => void load(true)}
          />
        </View>
      ) : (
        <>
          <View style={styles.sectionHeader}>
            <Text style={uiStyles.sectionTitle}>Очередь доступа</Text>
            <Text style={uiStyles.supporting}>{items.length} учеников</Text>
          </View>

          {error ? (
            <View style={styles.stackGap}>
              <InlineNotice body={error} title="Очередь не загрузилась" tone="error" />
              <SecondaryButton label="Повторить" onPress={() => void load()} />
            </View>
          ) : null}
          {loading ? (
            <PremiumCard>
              <Text style={uiStyles.body}>Загружаем учеников…</Text>
            </PremiumCard>
          ) : null}
          {!loading && !error && items.length === 0 ? (
            <PremiumCard>
              <Text style={uiStyles.sectionTitle}>Очередь пока пустая</Text>
              <Text style={[uiStyles.body, styles.emptyBody]}>
                После создания ученика здесь появится следующий честный шаг.
              </Text>
            </PremiumCard>
          ) : null}
          <View style={styles.cards}>
            {items.map((item) => (
              <OnboardingCard item={item} key={item.studentId} />
            ))}
          </View>
        </>
      )}

      <SecondaryButton label="Выйти" onPress={() => void leave()} />
    </PremiumScrollScreen>
  );
}

function OnboardingCard({ item }: { item: StudentOnboardingItem }) {
  const copy = onboardingStateCopy[item.onboardingState];
  return (
    <Pressable
      accessibilityHint="Открывает действия по ученику"
      accessibilityLabel={`${item.fullName}. ${copy.title}`}
      accessibilityRole="button"
      onPress={() =>
        router.push({
          pathname: "/(protected)/student/[studentId]",
          params: { studentId: item.studentId },
        })
      }
      style={({ pressed }) => [styles.onboardingCard, pressed && styles.cardPressed]}
    >
      <View style={styles.cardHeader}>
        <Text style={styles.stateEyebrow}>{copy.eyebrow}</Text>
        <Text style={styles.chevron}>›</Text>
      </View>
      <Text style={styles.studentName}>{item.fullName}</Text>
      <Text style={styles.stateTitle}>{copy.title}</Text>
      <Text style={styles.stateBody}>{copy.description}</Text>
      {item.invitationExpiresAt ? (
        <Text style={styles.expiry}>
          Приглашение до {formatBelcantoDate(item.invitationExpiresAt)}
        </Text>
      ) : null}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  brand: {
    color: colors.textPrimary,
    fontFamily: fonts.bold,
    marginTop: spacing.sm,
    ...typeScale.brand,
  },
  role: {
    color: colors.textGold,
    fontFamily: fonts.semibold,
    marginTop: metrics.workflowEyebrowTop,
    ...typeScale.eyebrow,
  },
  title: {
    color: colors.textPrimary,
    fontFamily: fonts.extrabold,
    marginTop: spacing.sm,
    ...typeScale.screenTitle,
  },
  subtitle: {
    color: colors.textSecondary,
    fontFamily: fonts.regular,
    marginBottom: spacing.xxl,
    marginTop: spacing.sm,
    ...typeScale.body,
  },
  quickActions: { flexDirection: "row", gap: spacing.md, marginBottom: spacing.section },
  workspaceSwitch: { marginBottom: spacing.xxl },
  actionGrow: { flex: 1 },
  sectionHeader: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
    marginBottom: spacing.md,
  },
  stackGap: { gap: spacing.md },
  emptyBody: { marginTop: spacing.sm },
  cards: { gap: spacing.md },
  onboardingCard: {
    backgroundColor: colors.raised,
    borderColor: colors.borderGlass,
    borderRadius: radii.card,
    borderWidth: metrics.borderWidth,
    minHeight: 164,
    padding: spacing.lg,
  },
  cardPressed: { backgroundColor: colors.surface },
  cardHeader: { alignItems: "center", flexDirection: "row", justifyContent: "space-between" },
  stateEyebrow: {
    color: colors.textGold,
    fontFamily: fonts.semibold,
    ...typeScale.eyebrow,
  },
  chevron: {
    color: colors.textAccent,
    fontFamily: fonts.bold,
    fontSize: 26,
    lineHeight: 30,
  },
  studentName: {
    color: colors.textPrimary,
    fontFamily: fonts.bold,
    marginTop: spacing.sm,
    ...typeScale.cardTitle,
  },
  stateTitle: {
    color: colors.textAccent,
    fontFamily: fonts.semibold,
    marginTop: spacing.sm,
    ...typeScale.supporting,
  },
  stateBody: {
    color: colors.textSecondary,
    fontFamily: fonts.regular,
    marginTop: spacing.xs,
    ...typeScale.supporting,
  },
  expiry: {
    color: colors.textMuted,
    fontFamily: fonts.regular,
    marginTop: spacing.md,
    ...typeScale.label,
  },
});
