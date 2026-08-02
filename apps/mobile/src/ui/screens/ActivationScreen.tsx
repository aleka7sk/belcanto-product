import { router } from "expo-router";
import { useEffect, useMemo, useRef, useState } from "react";
import {
  AccessibilityInfo,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";

import { useActivationLink } from "@/activation/useActivationLinkState";
import { useApiClient, type ActivationPreview } from "@/api";
import {
  createIntentIdempotency,
  prepareActivationPreview,
  prepareCompleteActivation,
  type CompleteActivationDraft,
} from "@/controllers";
import { isSessionRestoring, useSession } from "@/session";
import {
  AmbientGlow,
  BrandBadge,
  InlineNotice,
  LoadingScreen,
  PremiumCard,
  PremiumScrollScreen,
  PremiumTextField,
  PrimaryButton,
  SecondaryButton,
  TextAction,
  uiStyles,
} from "../components";
import { colors, fonts, spacing, typeScale } from "../tokens";
import {
  apiErrorMessage,
  formIssueMap,
  formatBelcantoDate,
} from "../viewModels";

type ActivationErrors = Partial<
  Record<keyof CompleteActivationDraft, string | undefined>
>;
type PreviewState =
  | { status: "idle" | "loading" }
  | { status: "ready"; preview: ActivationPreview }
  | { status: "error"; message: string };

const activationCopy: Record<
  ActivationPreview["kind"],
  { subtitle: string; success: string; identity: string }
> = {
  owner_bootstrap: {
    subtitle: "Завершите защищённую настройку аккаунта владельца",
    success: "Теперь войдите и откройте рабочее пространство владельца.",
    identity: "АККАУНТ ВЛАДЕЛЬЦА",
  },
  staff_activation: {
    subtitle: "Ваша рабочая учётная запись уже подготовлена школой",
    success: "Теперь войдите в рабочее пространство Belcanto.",
    identity: "СОТРУДНИК BELCANTO",
  },
  student_activation: {
    subtitle: "Ваш профиль уже подготовлен школой",
    success: "Теперь войдите и откройте свой первый учебный ориентир.",
    identity: "ЭТО ВЫ?",
  },
};

export function ActivationScreen() {
  const link = useActivationLink();
  const api = useApiClient();
  const { state: sessionState, signOut } = useSession();
  const [previewState, setPreviewState] = useState<PreviewState>({ status: "idle" });
  const [reloadVersion, setReloadVersion] = useState(0);
  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [passwordConfirmation, setPasswordConfirmation] = useState("");
  const [errors, setErrors] = useState<ActivationErrors>({});
  const [requestError, setRequestError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [completed, setCompleted] = useState(false);
  const [switchingAccount, setSwitchingAccount] = useState(false);
  const phoneRef = useRef<TextInput>(null);
  const passwordRef = useRef<TextInput>(null);
  const confirmationRef = useRef<TextInput>(null);
  const idempotency = useMemo(() => createIntentIdempotency(), []);

  useEffect(() => {
    if (link.status !== "ready") return;
    const abortController = new AbortController();
    queueMicrotask(() => {
      if (abortController.signal.aborted) return;
      const prepared = prepareActivationPreview({ token: link.token });
      if (!prepared.ok) {
        setPreviewState({
          status: "error",
          message: "Ссылка приглашения имеет неверный формат.",
        });
        return;
      }
      setPreviewState({ status: "loading" });
      void api
        .previewActivation(prepared.value, abortController.signal)
        .then((preview) => {
          setPreviewState({ status: "ready", preview });
        })
        .catch((error: unknown) => {
          if (!abortController.signal.aborted) {
            setPreviewState({ status: "error", message: apiErrorMessage(error) });
          }
        });
    });
    return () => abortController.abort();
    // A retry intentionally repeats the same safe preview operation.
  }, [api, link, reloadVersion]);

  if (
    isSessionRestoring(sessionState) ||
    link.status === "loading" ||
    previewState.status === "loading"
  ) {
    return <LoadingScreen label="Проверяем приглашение" />;
  }

  if (link.status === "ready" && sessionState.tokens !== null) {
    const switchAccount = async () => {
      setSwitchingAccount(true);
      try {
        await signOut();
      } finally {
        setSwitchingAccount(false);
      }
    };
    return (
      <PremiumScrollScreen contentStyle={styles.centeredContent}>
        <AmbientGlow />
        <View style={styles.centerBadge}>
          <BrandBadge large />
        </View>
        <Text accessibilityRole="header" style={styles.centerTitle}>
          Сменить аккаунт
        </Text>
        <InlineNotice
          body="Сейчас в приложении открыт другой аккаунт. Выйдите из него, чтобы безопасно активировать это персональное приглашение."
          title="Приглашение для другого входа"
        />
        <View style={styles.unavailableActions}>
          <PrimaryButton
            busy={switchingAccount}
            label="Выйти и продолжить"
            onPress={() => void switchAccount()}
          />
          <SecondaryButton label="Не сейчас" onPress={() => router.replace("/")} />
        </View>
      </PremiumScrollScreen>
    );
  }

  if (completed) {
    const successCopy =
      previewState.status === "ready"
        ? activationCopy[previewState.preview.kind].success
        : activationCopy.student_activation.success;
    return (
      <PremiumScrollScreen contentStyle={styles.centeredContent}>
        <AmbientGlow />
        <View style={styles.centerBadge}>
          <BrandBadge kind="success" large />
        </View>
        <Text accessibilityRole="header" style={styles.centerTitle}>
          Доступ активирован
        </Text>
        <Text style={styles.centerSubtitle}>
          Пароль сохранён. {successCopy}
        </Text>
        <View style={styles.successAction}>
          <PrimaryButton label="Войти" onPress={() => router.replace("/sign-in")} />
        </View>
      </PremiumScrollScreen>
    );
  }

  if (link.status === "absent") {
    return (
      <ActivationUnavailable
        body="Откройте персональную ссылку, которую вам отправила школа. Без неё доступ активировать нельзя."
        title="Нужно приглашение"
      />
    );
  }

  if (link.status === "invalid") {
    return (
      <ActivationUnavailable
        body="Ссылка повреждена или пришла не из разрешённого канала Belcanto. Попросите школу отправить приглашение заново."
        title="Приглашение не открылось"
      />
    );
  }

  if (previewState.status === "error") {
    return (
      <ActivationUnavailable
        body={previewState.message}
        onRetry={() => setReloadVersion((value) => value + 1)}
        title="Не удалось проверить приглашение"
      />
    );
  }

  if (link.status !== "ready" || previewState.status !== "ready") {
    return <LoadingScreen label="Подготавливаем активацию" />;
  }

  const submit = async () => {
    setRequestError(null);
    const result = prepareCompleteActivation(
      { token: link.token, phone, password, passwordConfirmation },
      idempotency.key(),
    );
    if (!result.ok) {
      const nextErrors = formIssueMap(result.issues);
      setErrors(nextErrors);
      const first = nextErrors.phone
        ? phoneRef
        : nextErrors.password
          ? passwordRef
          : confirmationRef;
      first.current?.focus();
      AccessibilityInfo.announceForAccessibility("Проверьте поля активации");
      return;
    }
    setErrors({});
    setSubmitting(true);
    try {
      await api.completeActivation(result.value.body, result.value.idempotencyKey);
      idempotency.complete();
      setCompleted(true);
      AccessibilityInfo.announceForAccessibility("Доступ активирован");
    } catch (error) {
      const message = apiErrorMessage(error);
      setRequestError(message);
      AccessibilityInfo.announceForAccessibility(message);
    } finally {
      setSubmitting(false);
    }
  };

  const { preview } = previewState;
  const copy = activationCopy[preview.kind];
  return (
    <PremiumScrollScreen keyboardAware contentStyle={styles.formContent}>
      <AmbientGlow />
      <View style={styles.centerBadge}>
        <BrandBadge kind="success" large />
      </View>
      <Text accessibilityRole="header" style={styles.centerTitle}>
        Добро пожаловать в Belcanto
      </Text>
      <Text style={styles.centerSubtitle}>{copy.subtitle}</Text>

      <PremiumCard style={styles.identityCard}>
        <Text style={styles.identityEyebrow}>{copy.identity}</Text>
        <Text style={uiStyles.sectionTitle}>{preview.displayName}</Text>
        <Text style={[uiStyles.supporting, styles.identityMeta]}>
          {preview.maskedPhone} · действует до {formatBelcantoDate(preview.expiresAt)}
        </Text>
      </PremiumCard>

      <View style={styles.form}>
        <PremiumTextField
          ref={phoneRef}
          autoComplete="tel"
          error={errors.phone}
          keyboardType="phone-pad"
          label="Телефон"
          onChangeText={(value) => {
            setPhone(value);
            setErrors((current) => ({ ...current, phone: undefined }));
          }}
          onSubmitEditing={() => passwordRef.current?.focus()}
          placeholder="+7 700 000 00 00"
          returnKeyType="next"
          textContentType="telephoneNumber"
          value={phone}
        />
        <PremiumTextField
          ref={passwordRef}
          autoComplete="new-password"
          error={errors.password}
          helper="Не менее 15 символов"
          label="Придумайте пароль"
          onChangeText={(value) => {
            setPassword(value);
            setErrors((current) => ({ ...current, password: undefined }));
          }}
          onSubmitEditing={() => confirmationRef.current?.focus()}
          placeholder="Новый пароль"
          returnKeyType="next"
          secureTextEntry
          textContentType="newPassword"
          value={password}
        />
        <PremiumTextField
          ref={confirmationRef}
          autoComplete="new-password"
          error={errors.passwordConfirmation}
          label="Повторите пароль"
          onChangeText={(value) => {
            setPasswordConfirmation(value);
            setErrors((current) => ({ ...current, passwordConfirmation: undefined }));
          }}
          onSubmitEditing={() => void submit()}
          placeholder="Повторите пароль"
          returnKeyType="done"
          secureTextEntry
          textContentType="newPassword"
          value={passwordConfirmation}
        />
      </View>

      {requestError ? (
        <View style={styles.requestNotice}>
          <InlineNotice body={requestError} title="Активация не завершена" tone="error" />
        </View>
      ) : null}
      <View style={styles.primaryAction}>
        <PrimaryButton
          busy={submitting}
          label="Активировать доступ"
          onPress={() => void submit()}
        />
      </View>
      <TextAction
        label="Уже активировали? Войти"
        onPress={() => router.replace("/sign-in")}
      />
    </PremiumScrollScreen>
  );
}

function ActivationUnavailable({
  title,
  body,
  onRetry,
}: {
  title: string;
  body: string;
  onRetry?: (() => void) | undefined;
}) {
  return (
    <PremiumScrollScreen contentStyle={styles.centeredContent}>
      <AmbientGlow />
      <View style={styles.centerBadge}>
        <BrandBadge large />
      </View>
      <Text accessibilityRole="header" style={styles.centerTitle}>
        {title}
      </Text>
      <InlineNotice body={body} title="Закрытый доступ Belcanto" tone="error" />
      <View style={styles.unavailableActions}>
        {onRetry ? <PrimaryButton label="Проверить снова" onPress={onRetry} /> : null}
        <SecondaryButton label="Перейти ко входу" onPress={() => router.replace("/sign-in")} />
      </View>
    </PremiumScrollScreen>
  );
}

const styles = StyleSheet.create({
  centeredContent: { justifyContent: "center", minHeight: 720 },
  formContent: { minHeight: 920 },
  centerBadge: { alignItems: "center", marginTop: spacing.xxl },
  centerTitle: {
    color: colors.textPrimary,
    fontFamily: fonts.extrabold,
    marginTop: spacing.loose,
    textAlign: "center",
    ...typeScale.screenTitle,
  },
  centerSubtitle: {
    color: colors.textSecondary,
    fontFamily: fonts.regular,
    marginTop: spacing.sm,
    textAlign: "center",
    ...typeScale.body,
  },
  successAction: { marginTop: spacing.loose },
  identityCard: { marginTop: spacing.section },
  identityEyebrow: {
    color: colors.textGold,
    fontFamily: fonts.semibold,
    marginBottom: spacing.sm,
    ...typeScale.eyebrow,
  },
  identityMeta: { marginTop: spacing.sm },
  form: { gap: spacing.field, marginTop: spacing.xxl },
  requestNotice: { marginTop: spacing.field },
  primaryAction: { marginTop: spacing.section },
  unavailableActions: { gap: spacing.md, marginTop: spacing.xxl },
});
