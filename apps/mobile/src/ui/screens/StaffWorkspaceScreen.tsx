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
import { semantic, space, strokes, typeStyles } from "../tokens";
import {
  apiErrorMessage,
  formatBelcantoDate,
  onboardingStateCopy,
  roleLabel,
} from "../viewModels";
import { RoleNav } from "./account/shared";

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
      gutter={space.s4}
      navigation={<RoleNav active="people" />}
      scrollProps={{
        refreshControl: (
          <RefreshControl
            colors={[semantic.accentViolet]}
            onRefresh={() => void load(true)}
            progressBackgroundColor={semantic.bgRaised}
            refreshing={refreshing}
            tintColor={semantic.accentViolet}
          />
        ),
      }}
      testID="staff-workspace"
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
        {changeTeacherAllowed ? (
          <View style={styles.actionGrow}>
            <SecondaryButton
              label="Серии и кабинеты"
              onPress={() => router.push("/(protected)/series")}
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
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 13,
    letterSpacing: 2.4,
    lineHeight: 17,
    marginTop: space.s2,
  },
  role: {
    color: semantic.textGold,
    fontFamily: "Onest_600SemiBold",
    fontSize: 10,
    letterSpacing: 1,
    lineHeight: 13,
    marginTop: space.s10,
  },
  title: {
    color: semantic.textPrimary,
    fontFamily: "Onest_800ExtraBold",
    fontSize: 28,
    lineHeight: 34,
    marginTop: space.s2,
  },
  subtitle: {
    color: semantic.textSecondary,
    marginBottom: space.s6,
    marginTop: space.s2,
    ...typeStyles.bodyS,
  },
  quickActions: { flexDirection: "row", gap: space.s3, marginBottom: space.s8 },
  workspaceSwitch: { marginBottom: space.s6 },
  actionGrow: { flex: 1 },
  sectionHeader: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
    marginBottom: space.s3,
  },
  stackGap: { gap: space.s3 },
  emptyBody: { marginTop: space.s2 },
  cards: { gap: space.s3 },
  onboardingCard: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderGlass,
    borderRadius: 20,
    borderWidth: strokes.hairline,
    minHeight: 164,
    padding: space.s4,
  },
  cardPressed: { backgroundColor: semantic.bgSurface },
  cardHeader: { alignItems: "center", flexDirection: "row", justifyContent: "space-between" },
  stateEyebrow: {
    color: semantic.textGold,
    fontFamily: "Onest_600SemiBold",
    fontSize: 10,
    letterSpacing: 1,
    lineHeight: 13,
  },
  chevron: {
    color: semantic.textAccent,
    fontFamily: "Onest_700Bold",
    fontSize: 26,
    lineHeight: 30,
  },
  studentName: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 19,
    lineHeight: 23,
    marginTop: space.s2,
  },
  stateTitle: {
    color: semantic.textAccent,
    fontFamily: "Onest_600SemiBold",
    fontSize: 12,
    lineHeight: 17,
    marginTop: space.s2,
  },
  stateBody: {
    color: semantic.textSecondary,
    marginTop: space.s1,
    ...typeStyles.caption,
  },
  expiry: {
    color: semantic.textMuted,
    marginTop: space.s3,
    ...typeStyles.caption,
  },
});
