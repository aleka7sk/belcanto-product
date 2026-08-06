import { router, useLocalSearchParams } from "expo-router";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  AccessibilityInfo,
  Share,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";

import {
  canIssueStudentInvitations,
  canOpenStudentOnboardingQueue,
  canReissueStudentInvitations,
  canRevokeStudentInvitations,
} from "@/access";
import {
  useApiClient,
  type InvitationResult,
  type StudentOnboardingItem,
} from "@/api";
import {
  createIntentIdempotency,
  prepareFirstMinute,
  prepareInvitation,
  prepareRevokeInvitation,
  type FirstMinuteDraft,
} from "@/controllers";
import { useSession } from "@/session";
import {
  AmbientGlow,
  FeatureCard,
  InlineNotice,
  LoadingScreen,
  PremiumCard,
  PremiumScrollScreen,
  PremiumTextField,
  PrimaryButton,
  SecondaryButton,
  uiStyles,
} from "../components";
import { semantic, space, typeStyles } from "../tokens";
import {
  apiErrorMessage,
  formIssueMap,
  formatBelcantoDate,
  onboardingStateCopy,
} from "../viewModels";
import { RoleNav } from "./account/shared";

type FirstMinuteErrors = Partial<
  Record<keyof FirstMinuteDraft, string | undefined>
>;

export function OnboardingDetailScreen() {
  const params = useLocalSearchParams<{ studentId?: string | string[] }>();
  const studentId = Array.isArray(params.studentId)
    ? params.studentId[0]
    : params.studentId;
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const bootstrap = state.bootstrap;
  const [item, setItem] = useState<StudentOnboardingItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [whatWorked, setWhatWorked] = useState("");
  const [currentFocus, setCurrentFocus] = useState("");
  const [nextStep, setNextStep] = useState("");
  const [firstMinuteErrors, setFirstMinuteErrors] = useState<FirstMinuteErrors>({});
  const [actionError, setActionError] = useState<string | null>(null);
  const [busyAction, setBusyAction] = useState<
    "first-minute" | "invitation" | "revoke" | null
  >(null);
  const [issuedInvitation, setIssuedInvitation] = useState<InvitationResult | null>(null);
  const whatWorkedRef = useRef<TextInput>(null);
  const currentFocusRef = useRef<TextInput>(null);
  const nextStepRef = useRef<TextInput>(null);
  const firstMinuteIdempotency = useMemo(() => createIntentIdempotency(), []);
  const invitationIdempotency = useMemo(() => createIntentIdempotency(), []);
  const revokeIdempotency = useMemo(() => createIntentIdempotency(), []);

  const load = useCallback(async () => {
    if (
      bootstrap === null ||
      studentId === undefined ||
      !canOpenStudentOnboardingQueue(bootstrap)
    ) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const items = await runAuthenticated((accessToken) =>
        api.listStudentOnboarding(accessToken),
      );
      const found = items.find((candidate) => candidate.studentId === studentId);
      if (found === undefined) {
        setError("Ученик не найден в доступной вам очереди.");
        setItem(null);
      } else {
        setItem(found);
      }
    } catch (loadError) {
      setError(apiErrorMessage(loadError));
    } finally {
      setLoading(false);
    }
  }, [api, bootstrap, runAuthenticated, studentId]);

  useEffect(() => {
    let active = true;
    queueMicrotask(() => {
      if (active) void load();
    });
    return () => {
      active = false;
    };
  }, [load]);

  if (loading) return <LoadingScreen label="Открываем путь ученика" />;
  if (bootstrap === null) {
    return <LoadingScreen label="Проверяем доступ" />;
  }
  if (studentId === undefined) {
    return (
      <PremiumScrollScreen contentStyle={styles.centered} navigation={<RoleNav active="people" />}>
        <InlineNotice
          body="В адресе экрана отсутствует идентификатор ученика. Вернитесь в очередь и откройте запись заново."
          title="Некорректная ссылка"
          tone="error"
        />
        <SecondaryButton label="К очереди" onPress={() => router.replace("/(protected)")} />
      </PremiumScrollScreen>
    );
  }
  if (!canOpenStudentOnboardingQueue(bootstrap)) {
    return (
      <PremiumScrollScreen contentStyle={styles.centered} navigation={<RoleNav active="people" />}>
        <InlineNotice
          body="У этой учётной записи нет доступа к очереди учеников."
          title="Нет разрешения"
          tone="error"
        />
        <SecondaryButton label="Назад" onPress={() => router.back()} />
      </PremiumScrollScreen>
    );
  }
  if (error || item === null) {
    return (
      <PremiumScrollScreen contentStyle={styles.centered} navigation={<RoleNav active="people" />}>
        <InlineNotice
          body={error ?? "Ученик не найден."}
          title="Путь не открылся"
          tone="error"
        />
        <View style={styles.stackGap}>
          <PrimaryButton label="Повторить" onPress={() => void load()} />
          <SecondaryButton label="Назад" onPress={() => router.back()} />
        </View>
      </PremiumScrollScreen>
    );
  }

  const copy = onboardingStateCopy[item.onboardingState];
  const assignedTeacher =
    bootstrap.roles.includes("Teacher") &&
    item.teacherAccountId === bootstrap.accountId;
  const canPublishFirstMinute =
    assignedTeacher && item.onboardingState === "awaiting_first_minute";
  const canIssue =
    item.onboardingState === "ready_to_invite" &&
    canIssueStudentInvitations(bootstrap);
  const canReissue =
    item.onboardingState === "invited" &&
    canReissueStudentInvitations(bootstrap);
  const canRevoke =
    item.onboardingState === "invited" &&
    item.invitationId !== undefined &&
    canRevokeStudentInvitations(bootstrap);

  const publishFirstMinute = async () => {
    const result = prepareFirstMinute(
      {
        studentId,
        whatWorked,
        currentFocus,
        nextStep,
        expectedVersion: item.studentVersion,
      },
      firstMinuteIdempotency.key(),
    );
    if (!result.ok) {
      const nextErrors = formIssueMap(result.issues);
      setFirstMinuteErrors(nextErrors);
      const first = nextErrors.whatWorked
        ? whatWorkedRef
        : nextErrors.currentFocus
          ? currentFocusRef
          : nextStepRef;
      first.current?.focus();
      AccessibilityInfo.announceForAccessibility("Проверьте первый ориентир");
      return;
    }
    setFirstMinuteErrors({});
    setActionError(null);
    setBusyAction("first-minute");
    try {
      await runAuthenticated((accessToken) =>
        api.publishFirstMinute(
          accessToken,
          result.value.studentId,
          result.value.body,
          result.value.idempotencyKey,
        ),
      );
      firstMinuteIdempotency.complete();
      AccessibilityInfo.announceForAccessibility("Первый ориентир опубликован");
      await load();
    } catch (publishError) {
      const message = apiErrorMessage(publishError);
      setActionError(message);
      AccessibilityInfo.announceForAccessibility(message);
    } finally {
      setBusyAction(null);
    }
  };

  const issueInvitation = async (mode: "issue" | "reissue") => {
    const result = prepareInvitation(
      { studentId },
      invitationIdempotency.key(),
      mode,
    );
    if (!result.ok) return;
    setActionError(null);
    setBusyAction("invitation");
    try {
      const invitation =
        mode === "issue"
          ? await runAuthenticated((accessToken) =>
              api.issueInvitation(
                accessToken,
                result.value.studentId,
                result.value.idempotencyKey,
              ),
            )
          : await runAuthenticated((accessToken) =>
              api.reissueInvitation(
                accessToken,
                result.value.studentId,
                result.value.idempotencyKey,
              ),
            );
      invitationIdempotency.complete();
      setIssuedInvitation(invitation);
      AccessibilityInfo.announceForAccessibility("Приглашение готово к отправке");
      await load();
    } catch (invitationError) {
      const message = apiErrorMessage(invitationError);
      setActionError(message);
      AccessibilityInfo.announceForAccessibility(message);
    } finally {
      setBusyAction(null);
    }
  };

  const revokeInvitation = async () => {
    if (item.invitationId === undefined) return;
    const result = prepareRevokeInvitation(
      { invitationId: item.invitationId },
      revokeIdempotency.key(),
    );
    if (!result.ok) return;
    setActionError(null);
    setBusyAction("revoke");
    try {
      await runAuthenticated((accessToken) =>
        api.revokeInvitation(
          accessToken,
          result.value.invitationId,
          result.value.idempotencyKey,
        ),
      );
      revokeIdempotency.complete();
      setIssuedInvitation(null);
      AccessibilityInfo.announceForAccessibility("Приглашение отозвано");
      await load();
    } catch (revokeError) {
      const message = apiErrorMessage(revokeError);
      setActionError(message);
      AccessibilityInfo.announceForAccessibility(message);
    } finally {
      setBusyAction(null);
    }
  };

  const shareInvitation = async () => {
    if (issuedInvitation === null) return;
    try {
      await Share.share({
        message: `Belcanto уже подготовил ваш профиль. Активируйте доступ по персональной ссылке: ${issuedInvitation.activationLink}`,
        title: `Приглашение для ${item.fullName}`,
      });
    } catch {
      setActionError(
        "Системное меню отправки не открылось. Выделите ссылку выше и скопируйте её вручную.",
      );
    }
  };

  return (
    <PremiumScrollScreen keyboardAware contentStyle={styles.content} navigation={<RoleNav active="people" />}>
      <AmbientGlow />
      <Text style={styles.brand}>BELCANTO</Text>
      <Text style={styles.eyebrow}>{copy.eyebrow}</Text>
      <Text accessibilityRole="header" style={styles.title}>
        {item.fullName}
      </Text>
      <Text style={styles.subtitle}>{copy.description}</Text>

      <PremiumCard>
        <Text style={uiStyles.sectionTitle}>{copy.title}</Text>
        <Text style={[uiStyles.supporting, styles.meta]}>
          Поступление: {item.enrollmentReference}
        </Text>
        {item.invitationExpiresAt ? (
          <Text style={[uiStyles.supporting, styles.meta]}>
            Действует до {formatBelcantoDate(item.invitationExpiresAt)}
          </Text>
        ) : null}
      </PremiumCard>

      {canPublishFirstMinute ? (
        <View style={styles.section}>
          <Text style={uiStyles.cardTitle}>First Belcanto Minute</Text>
          <Text style={uiStyles.body}>
            Три коротких наблюдения — первая реальная ценность ученика внутри приложения.
          </Text>
          <View style={styles.form}>
            <PremiumTextField
              ref={whatWorkedRef}
              autoCapitalize="sentences"
              error={firstMinuteErrors.whatWorked}
              label="Что уже получилось"
              multiline
              onChangeText={(value) => {
                setWhatWorked(value);
                setFirstMinuteErrors((current) => ({ ...current, whatWorked: undefined }));
              }}
              placeholder="Конкретное наблюдение педагога"
              value={whatWorked}
            />
            <PremiumTextField
              ref={currentFocusRef}
              autoCapitalize="sentences"
              error={firstMinuteErrors.currentFocus}
              label="Фокус сейчас"
              multiline
              onChangeText={(value) => {
                setCurrentFocus(value);
                setFirstMinuteErrors((current) => ({ ...current, currentFocus: undefined }));
              }}
              placeholder="На чём сосредоточиться сейчас"
              value={currentFocus}
            />
            <PremiumTextField
              ref={nextStepRef}
              autoCapitalize="sentences"
              error={firstMinuteErrors.nextStep}
              label="Следующий шаг"
              multiline
              onChangeText={(value) => {
                setNextStep(value);
                setFirstMinuteErrors((current) => ({ ...current, nextStep: undefined }));
              }}
              placeholder="Понятное следующее действие"
              value={nextStep}
            />
          </View>
          <PrimaryButton
            busy={busyAction === "first-minute"}
            label="Опубликовать ориентир"
            onPress={() => void publishFirstMinute()}
          />
        </View>
      ) : null}

      {canIssue || canReissue || canRevoke ? (
        <View style={styles.section}>
          <Text style={uiStyles.cardTitle}>Персональное приглашение</Text>
          <Text style={uiStyles.body}>
            Только владелец открывает доступ. Ссылка одноразовая и ограничена по времени.
          </Text>
          {canIssue ? (
            <PrimaryButton
              busy={busyAction === "invitation"}
              label="Открыть доступ"
              onPress={() => void issueInvitation("issue")}
            />
          ) : null}
          {canReissue ? (
            <PrimaryButton
              busy={busyAction === "invitation"}
              label="Отправить новую ссылку"
              onPress={() => void issueInvitation("reissue")}
            />
          ) : null}
          {canRevoke ? (
            <SecondaryButton
              disabled={busyAction !== null}
              label={busyAction === "revoke" ? "Отзываем…" : "Отозвать приглашение"}
              onPress={() => void revokeInvitation()}
              tone="danger"
            />
          ) : null}
        </View>
      ) : null}

      {issuedInvitation ? (
        <FeatureCard>
          <Text style={styles.invitationEyebrow}>ССЫЛКА ГОТОВА</Text>
          <Text style={styles.invitationTitle}>Первая минута уже внутри приглашения</Text>
          <Text selectable style={styles.invitationLink}>
            {issuedInvitation.activationLink}
          </Text>
          <View style={styles.shareAction}>
            <PrimaryButton label="Отправить ученику" onPress={() => void shareInvitation()} />
          </View>
        </FeatureCard>
      ) : null}

      {actionError ? (
        <InlineNotice body={actionError} title="Действие не завершено" tone="error" />
      ) : null}
      <SecondaryButton label="К очереди" onPress={() => router.back()} />
    </PremiumScrollScreen>
  );
}

const styles = StyleSheet.create({
  centered: { gap: space.s4, justifyContent: "center", minHeight: 680 },
  content: { minHeight: 780 },
  stackGap: { gap: space.s3 },
  brand: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 13,
    letterSpacing: 2.4,
    lineHeight: 17,
  },
  eyebrow: {
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
  meta: { marginTop: space.s2 },
  section: { gap: space.s3, marginTop: space.s8 },
  form: { gap: space.s3 },
  invitationEyebrow: {
    color: semantic.textGold,
    fontFamily: "Onest_600SemiBold",
    fontSize: 10,
    letterSpacing: 1,
    lineHeight: 13,
  },
  invitationTitle: {
    color: semantic.textPrimary,
    fontFamily: "Onest_700Bold",
    fontSize: 19,
    lineHeight: 23,
    marginTop: space.s2,
  },
  invitationLink: {
    color: semantic.textSecondary,
    marginTop: space.s4,
    ...typeStyles.caption,
  },
  shareAction: { marginTop: space.s6 },
});
