import { router } from "expo-router";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  AccessibilityInfo,
  Pressable,
  StyleSheet,
  Text,
  View,
} from "react-native";

import { canDelegateStudentOnboarding } from "@/access";
import { useApiClient, type StaffMember } from "@/api";
import {
  createIntentIdempotency,
  prepareGrantDelegation,
  prepareRevokeDelegation,
} from "@/controllers";
import { useSession } from "@/session";
import {
  AmbientGlow,
  InlineNotice,
  PremiumScrollScreen,
  PremiumTextField,
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
import { apiErrorMessage, formIssueMap, formatBelcantoDate } from "../viewModels";

type DelegationFieldErrors = Partial<
  Record<"reason" | "currentPassword", string | undefined>
>;

export function AccessDelegationScreen() {
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const bootstrap = state.bootstrap;
  const [administrators, setAdministrators] = useState<StaffMember[]>([]);
  const [selectedId, setSelectedId] = useState("");
  const [reason, setReason] = useState("Доступ к приёму новых учеников");
  const [currentPassword, setCurrentPassword] = useState("");
  const [fieldErrors, setFieldErrors] = useState<DelegationFieldErrors>({});
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const idempotency = useMemo(() => createIntentIdempotency(), []);

  const load = useCallback(async () => {
    if (
      bootstrap === null ||
      !canDelegateStudentOnboarding(bootstrap)
    ) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const staff = await runAuthenticated((accessToken) =>
        api.listStaff(accessToken, "Administrator"),
      );
      setAdministrators(staff);
      setSelectedId((current) =>
        staff.some((member) => member.accountId === current)
          ? current
          : (staff[0]?.accountId ?? ""),
      );
    } catch (loadError) {
      setError(apiErrorMessage(loadError));
    } finally {
      setLoading(false);
    }
  }, [api, bootstrap, runAuthenticated]);

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
  if (!canDelegateStudentOnboarding(bootstrap)) {
    return (
      <PremiumScrollScreen contentStyle={styles.centered}>
        <InlineNotice
          body="Выдавать и отзывать доступ суперадминистратора может только владелец."
          title="Нет разрешения"
          tone="error"
        />
        <SecondaryButton label="Назад" onPress={() => router.back()} />
      </PremiumScrollScreen>
    );
  }

  const selected = administrators.find((member) => member.accountId === selectedId);
  const active = selected?.onboardingDelegationId !== undefined;

  const submit = async () => {
    if (selected === undefined) return;
    setError(null);
    setSuccess(null);
    const key = idempotency.key();
    if (active) {
      const result = prepareRevokeDelegation(
        {
          delegationId: selected.onboardingDelegationId!,
          reason,
          currentPassword,
        },
        key,
      );
      if (!result.ok) {
        const mapped = formIssueMap(result.issues);
        setFieldErrors({
          reason: mapped.reason,
          currentPassword: mapped.currentPassword,
        });
        AccessibilityInfo.announceForAccessibility("Проверьте подтверждение доступа");
        return;
      }
      setBusy(true);
      try {
        await runAuthenticated((accessToken) =>
          api.revokeDelegation(
            accessToken,
            result.value.delegationId,
            result.value.body,
            result.value.idempotencyKey,
          ),
        );
        idempotency.complete();
        setCurrentPassword("");
        setSuccess("Доступ суперадминистратора отозван.");
        await load();
      } catch (submitError) {
        setError(apiErrorMessage(submitError));
      } finally {
        setBusy(false);
      }
      return;
    }

    const result = prepareGrantDelegation(
      {
        administratorAccountId: selected.accountId,
        reason,
        expiresAt: "",
        currentPassword,
      },
      key,
    );
    if (!result.ok) {
      const mapped = formIssueMap(result.issues);
      setFieldErrors({
        reason: mapped.reason,
        currentPassword: mapped.currentPassword,
      });
      AccessibilityInfo.announceForAccessibility("Проверьте подтверждение доступа");
      return;
    }
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.grantDelegation(
          accessToken,
          result.value.body,
          result.value.idempotencyKey,
        ),
      );
      idempotency.complete();
      setCurrentPassword("");
      setSuccess("Администратор получил ограниченный доступ суперадминистратора.");
      await load();
    } catch (submitError) {
      setError(apiErrorMessage(submitError));
    } finally {
      setBusy(false);
    }
  };

  return (
    <PremiumScrollScreen keyboardAware contentStyle={styles.content}>
      <AmbientGlow />
      <Text style={styles.brand}>BELCANTO</Text>
      <Text style={styles.eyebrow}>ДОСТУП ВЛАДЕЛЬЦА</Text>
      <Text accessibilityRole="header" style={styles.title}>
        Суперадминистратор
      </Text>
      <Text style={styles.subtitle}>
        Это не новая роль. Вы даёте администратору только право создавать
        учеников и видеть их путь к активации.
      </Text>

      {loading ? <Text style={uiStyles.body}>Загружаем администраторов…</Text> : null}
      <View accessibilityRole="radiogroup" style={styles.adminList}>
        {administrators.map((administrator) => {
          const selectedState = administrator.accountId === selectedId;
          const delegated = administrator.onboardingDelegationId !== undefined;
          return (
            <Pressable
              accessibilityLabel={`${administrator.fullName}. ${delegated ? "Доступ выдан" : "Обычный администратор"}`}
              accessibilityRole="radio"
              accessibilityState={{ checked: selectedState }}
              key={administrator.accountId}
              onPress={() => {
                setSelectedId(administrator.accountId);
                setError(null);
                setSuccess(null);
              }}
              style={({ pressed }) => [
                styles.adminCard,
                selectedState && styles.adminSelected,
                pressed && styles.adminPressed,
              ]}
            >
              <View style={styles.adminCopy}>
                <Text style={styles.adminName}>{administrator.fullName}</Text>
                <Text style={uiStyles.supporting}>
                  {delegated ? "Суперадминистратор" : "Администратор"}
                </Text>
                {administrator.onboardingDelegationExpiresAt ? (
                  <Text style={styles.expiry}>
                    До {formatBelcantoDate(administrator.onboardingDelegationExpiresAt)}
                  </Text>
                ) : null}
              </View>
              <View style={[styles.statusDot, delegated && styles.statusDotActive]} />
            </Pressable>
          );
        })}
      </View>

      {!loading && administrators.length === 0 ? (
        <InlineNotice
          body="Сначала активируйте учётную запись сотрудника с ролью Administrator через staff-bootstrap."
          title="Нет администраторов"
        />
      ) : null}

      {selected ? (
        <View style={styles.form}>
          <PremiumTextField
            autoCapitalize="sentences"
            error={fieldErrors.reason}
            label="Причина"
            onChangeText={(value) => {
              setReason(value);
              setFieldErrors((current) => ({ ...current, reason: undefined }));
            }}
            placeholder="Зачем нужен доступ"
            value={reason}
          />
          <PremiumTextField
            autoComplete="current-password"
            error={fieldErrors.currentPassword}
            label="Ваш текущий пароль"
            onChangeText={(value) => {
              setCurrentPassword(value);
              setFieldErrors((current) => ({ ...current, currentPassword: undefined }));
            }}
            placeholder="Подтвердите решение"
            secureTextEntry
            textContentType="password"
            value={currentPassword}
          />
        </View>
      ) : null}

      {error ? <InlineNotice body={error} title="Доступ не изменён" tone="error" /> : null}
      {success ? <InlineNotice body={success} title="Готово" tone="success" /> : null}
      <View style={styles.actions}>
        {selected ? (
          <PrimaryButton
            busy={busy}
            label={active ? "Отозвать доступ" : "Выдать доступ"}
            onPress={() => void submit()}
          />
        ) : null}
        <SecondaryButton label="Назад" onPress={() => router.back()} />
      </View>
    </PremiumScrollScreen>
  );
}

const styles = StyleSheet.create({
  centered: { gap: spacing.lg, justifyContent: "center", minHeight: 680 },
  content: { minHeight: 800 },
  brand: {
    color: colors.textPrimary,
    fontFamily: fonts.bold,
    ...typeScale.brand,
  },
  eyebrow: {
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
  adminList: { gap: spacing.sm },
  adminCard: {
    alignItems: "center",
    backgroundColor: colors.raised,
    borderColor: colors.borderGlass,
    borderRadius: radii.compactCard,
    borderWidth: metrics.borderWidth,
    flexDirection: "row",
    minHeight: 76,
    padding: spacing.lg,
  },
  adminSelected: { borderColor: colors.violet },
  adminPressed: { backgroundColor: colors.surface },
  adminCopy: { flex: 1 },
  adminName: {
    color: colors.textPrimary,
    fontFamily: fonts.semibold,
    marginBottom: spacing.xs,
    ...typeScale.body,
  },
  expiry: {
    color: colors.textMuted,
    fontFamily: fonts.regular,
    marginTop: spacing.xs,
    ...typeScale.label,
  },
  statusDot: {
    backgroundColor: colors.border,
    borderRadius: 6,
    height: 12,
    width: 12,
  },
  statusDotActive: { backgroundColor: colors.cyan },
  form: { gap: spacing.md, marginTop: spacing.xxl },
  actions: { gap: spacing.md, marginTop: spacing.xxl },
});
