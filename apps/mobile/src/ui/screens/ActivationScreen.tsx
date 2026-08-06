import { router } from "expo-router";
import { useEffect, useMemo, useState } from "react";
import { AccessibilityInfo, StyleSheet, Text, View } from "react-native";

import { useActivationLink } from "@/activation/useActivationLinkState";
import {
  ApiError,
  useApiClient,
  type ActivationPreview,
  type ActivationProgressView,
  type TwofaEnrollment,
} from "@/api";
import { createIntentIdempotency, prepareActivationPreview } from "@/controllers";
import { useMessage } from "@/i18n";
import { isSessionRestoring, useSession } from "@/session";
import {
  AmbientGlow,
  BrandBadge,
  InlineNotice,
  LoadingScreen,
  PremiumScrollScreen,
  PremiumTextField,
  PrimaryButton,
  SecondaryButton,
} from "../components";
import {
  AccountBanner,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../patterns/accountPatterns";
import { colors, fonts, radius, semantic, space, spacing, strokes, typeScale, typeStyles } from "../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../viewModels";

/**
 * Activation wizard (Figma AUTH-01..05 + AUTH-10, flow A). The server
 * owns step state: password → verified contact → optional 2FA → finish.
 * Re-opening an in-progress invitation resumes from the saved steps
 * instead of starting over (AUTH-10); the token itself never renders in
 * copy and expired links collapse into the invalid-link state.
 */

type WizardStep =
  | "loading"
  | "invitation"
  | "resume"
  | "password"
  | "contact"
  | "twofa"
  | "codes"
  | "done";

type PreviewState =
  | { status: "idle" | "loading" }
  | { status: "ready"; preview: ActivationPreview }
  | { status: "error"; message: string };

export function ActivationScreen() {
  const link = useActivationLink();
  const api = useApiClient();
  const message = useMessage();
  const { state: sessionState, signOut } = useSession();
  const [previewState, setPreviewState] = useState<PreviewState>({ status: "idle" });
  const [progress, setProgress] = useState<ActivationProgressView | null>(null);
  const [step, setStep] = useState<WizardStep>("loading");
  const [reloadVersion, setReloadVersion] = useState(0);

  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");
  const [passwordConfirmation, setPasswordConfirmation] = useState("");
  const [contactValue, setContactValue] = useState("");
  const [codeSent, setCodeSent] = useState(false);
  const [code, setCode] = useState("");
  const [enrollment, setEnrollment] = useState<TwofaEnrollment | null>(null);
  const [recoveryCodes, setRecoveryCodes] = useState<string[]>([]);
  const [busy, setBusy] = useState(false);
  const [requestError, setRequestError] = useState<string | null>(null);
  const [switchingAccount, setSwitchingAccount] = useState(false);
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
      void Promise.all([
        api.previewActivation(prepared.value, abortController.signal),
        api.activationProgress({ token: link.token }, abortController.signal),
      ])
        .then(([preview, loadedProgress]) => {
          setPreviewState({ status: "ready", preview });
          setProgress(loadedProgress);
          if (loadedProgress.completed) {
            setStep("done");
          } else if (
            loadedProgress.passwordSet ||
            loadedProgress.contactVerified ||
            loadedProgress.twofaEnrolled
          ) {
            setStep("resume");
          } else {
            setStep("invitation");
          }
        })
        .catch((error: unknown) => {
          if (!abortController.signal.aborted) {
            setPreviewState({ status: "error", message: apiErrorMessage(error) });
          }
        });
    });
    return () => abortController.abort();
    // A retry intentionally repeats the same safe read pair.
  }, [api, link, reloadVersion]);

  const refreshProgress = async (): Promise<ActivationProgressView | null> => {
    if (link.status !== "ready") return null;
    const next = await api.activationProgress({ token: link.token });
    setProgress(next);
    return next;
  };

  const fail = (error: unknown) => {
    if (error instanceof ApiError && error.code === "ACTIVATION_INVALID") {
      setPreviewState({
        status: "error",
        message: "Приглашение больше не действует. Попросите школу открыть доступ заново.",
      });
      return;
    }
    const text = apiErrorMessage(error);
    setRequestError(text);
    AccessibilityInfo.announceForAccessibility(text);
  };

  const run = async (operation: () => Promise<void>) => {
    setBusy(true);
    setRequestError(null);
    try {
      await operation();
    } catch (error) {
      fail(error);
    } finally {
      setBusy(false);
    }
  };

  const nextIncomplete = (view: ActivationProgressView | null): WizardStep => {
    if (view === null) return "password";
    if (!view.passwordSet) return "password";
    if (!view.contactVerified) return "contact";
    return "twofa";
  };

  const finish = async () => {
    if (link.status !== "ready") return;
    await api.finishActivation(
      { token: link.token, phone: phone.trim() },
      idempotency.key(),
    );
    idempotency.complete();
    setStep("done");
    AccessibilityInfo.announceForAccessibility(message("auth05.title"));
  };

  if (
    isSessionRestoring(sessionState) ||
    link.status === "loading" ||
    previewState.status === "loading" ||
    (link.status === "ready" && previewState.status === "idle")
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
        title={message("auth09.title")}
      />
    );
  }

  if (previewState.status === "error") {
    return (
      <ActivationUnavailable
        body={previewState.message}
        onRetry={() => setReloadVersion((value) => value + 1)}
        title={message("auth09.title")}
      />
    );
  }

  if (link.status !== "ready" || previewState.status !== "ready") {
    return <LoadingScreen label="Подготавливаем активацию" />;
  }

  const { preview } = previewState;
  const brand = (
    <View style={styles.brandRow}>
      <BrandBadge />
      <View>
        <Text style={styles.brandName}>{message("auth.brand")}</Text>
        <Text style={styles.brandTagline}>{message("auth.brandTagline")}</Text>
      </View>
    </View>
  );
  const errorNotice =
    requestError !== null ? (
      <InlineNotice body={requestError} title={message("common.retry")} tone="error" />
    ) : null;

  if (step === "done") {
    return (
      <PremiumScrollScreen contentStyle={styles.wizardContent}>
        <AmbientGlow />
        {brand}
        <ScreenHeading
          eyebrow={message("auth05.eyebrow")}
          subtitle={message("auth05.subtitle")}
          title={message("auth05.title")}
        />
        <StatusCard
          body={message(
            preview.kind === "student_activation"
              ? "auth05.card.body.student"
              : "auth05.card.body.staff",
          )}
          title={preview.displayName}
          tone="success"
        />
        <StatusRow
          status={message("auth05.entry.status")}
          subtitle={message("auth05.entry.body")}
          title={message("auth05.entry.title")}
          tone="info"
        />
        <AccountBanner
          body={message("auth05.banner.body")}
          title={message("auth05.banner.title")}
        />
        <BlockAction
          label={message("auth05.open")}
          onPress={() => router.replace("/sign-in")}
          testID="activation-open-today"
        />
      </PremiumScrollScreen>
    );
  }

  if (step === "resume") {
    const target = nextIncomplete(progress);
    const targetLabel = message(
      target === "password"
        ? "auth10.step.password"
        : target === "contact"
          ? "auth10.step.contact"
          : "auth10.step.twofa",
    );
    return (
      <PremiumScrollScreen contentStyle={styles.wizardContent}>
        <AmbientGlow />
        {brand}
        <ScreenHeading
          eyebrow={message("auth10.eyebrow")}
          subtitle={message("auth10.subtitle")}
          title={message("auth10.title")}
        />
        <StatusRow
          status={progress?.passwordSet ? message("auth10.statusDone") : message("auth10.statusNext")}
          subtitle={
            progress?.passwordSet
              ? message("auth10.done.password")
              : message("auth10.pending")
          }
          title={message("auth10.step.password")}
          tone={progress?.passwordSet ? "success" : "warning"}
        />
        <StatusRow
          status={
            progress?.contactVerified
              ? message("auth10.statusDone")
              : message("auth10.statusNext")
          }
          subtitle={
            progress?.contactVerified
              ? message("auth10.done.contact")
              : message("auth10.pending")
          }
          title={message("auth10.step.contact")}
          tone={progress?.contactVerified ? "success" : "warning"}
        />
        <StatusRow
          status={
            progress?.twofaEnrolled
              ? message("auth10.statusDone")
              : message("auth10.statusNext")
          }
          subtitle={
            progress?.twofaEnrolled
              ? message("auth10.done.twofa")
              : message("auth10.pending")
          }
          title={message("auth10.step.twofa")}
          tone={progress?.twofaEnrolled ? "success" : "muted"}
        />
        <AccountBanner
          body={message("auth10.banner.body")}
          title={message("auth10.banner.title")}
        />
        {progress?.passwordSet ? (
          <PremiumTextField
            keyboardType="phone-pad"
            label={message("auth02.phone")}
            helper={preview.maskedPhone}
            onChangeText={setPhone}
            testID="activation-resume-phone"
            value={phone}
          />
        ) : null}
        {errorNotice}
        <BlockAction
          busy={busy}
          disabled={progress?.passwordSet === true && phone.trim().length === 0}
          label={message("auth10.continue", { step: targetLabel })}
          onPress={() => setStep(target)}
          testID="activation-resume-continue"
        />
        <BlockAction
          kind="secondary"
          label={message("auth10.leave")}
          onPress={() => router.replace("/sign-in")}
        />
      </PremiumScrollScreen>
    );
  }

  if (step === "invitation") {
    return (
      <PremiumScrollScreen contentStyle={styles.wizardContent}>
        <AmbientGlow />
        {brand}
        <ScreenHeading
          eyebrow={message("auth01.eyebrow")}
          subtitle={message("auth01.subtitle")}
          title={message("auth01.title")}
        />
        <StatusCard
          body={message(`auth01.card.role.${preview.kind}` as const)}
          status={message("auth01.card.expires", {
            date: formatBelcantoDate(preview.expiresAt),
          })}
          title={preview.displayName}
          tone="warning"
        />
        <StatusRow
          status={message("auth01.next.status")}
          subtitle={message("auth01.next.body")}
          title={message("auth01.next.title")}
          tone="info"
        />
        <AccountBanner
          body={message("auth01.banner.body")}
          title={message("auth01.banner.title")}
        />
        <BlockAction
          label={message("auth01.continue")}
          onPress={() => setStep("password")}
          testID="activation-continue"
        />
        <BlockAction
          kind="secondary"
          label={message("auth01.notMine")}
          onPress={() => router.replace("/sign-in")}
        />
      </PremiumScrollScreen>
    );
  }

  if (step === "password") {
    const submitPassword = () =>
      run(async () => {
        if (password !== passwordConfirmation) {
          setRequestError("Пароли не совпадают.");
          return;
        }
        if (link.status !== "ready") return;
        await api.setActivationPassword({
          token: link.token,
          phone: phone.trim(),
          password,
        });
        await refreshProgress();
        setStep("contact");
      });
    return (
      <PremiumScrollScreen keyboardAware contentStyle={styles.wizardContent}>
        <AmbientGlow />
        {brand}
        <ScreenHeading
          eyebrow={message("auth02.eyebrow")}
          subtitle={message("auth02.subtitle")}
          title={message("auth02.title")}
        />
        <PremiumTextField
          autoComplete="tel"
          helper={preview.maskedPhone}
          keyboardType="phone-pad"
          label={message("auth02.phone")}
          onChangeText={setPhone}
          placeholder="+7 700 000 00 00"
          testID="activation-phone"
          textContentType="telephoneNumber"
          value={phone}
        />
        <PremiumTextField
          autoComplete="new-password"
          label={message("auth02.password")}
          onChangeText={setPassword}
          secureTextEntry
          testID="activation-password"
          textContentType="newPassword"
          value={password}
        />
        <PremiumTextField
          autoComplete="new-password"
          label={message("auth02.repeat")}
          onChangeText={setPasswordConfirmation}
          secureTextEntry
          testID="activation-password-repeat"
          textContentType="newPassword"
          value={passwordConfirmation}
        />
        <StatusCard
          body={message("auth02.card.body")}
          status={message("auth02.card.status")}
          title={message("auth02.card.title")}
          tone="warning"
        />
        {errorNotice}
        <BlockAction
          busy={busy}
          disabled={
            phone.trim().length === 0 ||
            password.length === 0 ||
            passwordConfirmation.length === 0
          }
          label={message("auth02.save")}
          onPress={() => void submitPassword()}
          testID="activation-save-password"
        />
        <BlockAction
          kind="secondary"
          label={message("auth02.back")}
          onPress={() => setStep("invitation")}
        />
      </PremiumScrollScreen>
    );
  }

  if (step === "contact") {
    const sendCode = () =>
      run(async () => {
        if (link.status !== "ready") return;
        await api.startActivationContact({
          token: link.token,
          kind: "email",
          value: contactValue.trim(),
        });
        setCodeSent(true);
        setCode("");
      });
    const confirmCode = () =>
      run(async () => {
        if (link.status !== "ready") return;
        await api.verifyActivationContact({ token: link.token, code: code.trim() });
        await refreshProgress();
        setStep("twofa");
      });
    return (
      <PremiumScrollScreen keyboardAware contentStyle={styles.wizardContent}>
        <AmbientGlow />
        {brand}
        <ScreenHeading
          eyebrow={message("auth03.eyebrow")}
          subtitle={
            codeSent
              ? message("auth03.subtitle.code", { contact: contactValue.trim() })
              : message("auth03.subtitle.entry")
          }
          title={message("auth03.title")}
        />
        {codeSent ? (
          <>
            <PremiumTextField
              autoComplete="one-time-code"
              keyboardType="number-pad"
              label={message("auth03.codeLabel")}
              onChangeText={setCode}
              testID="activation-contact-code"
              textContentType="oneTimeCode"
              value={code}
            />
            <StatusRow
              onPress={() => void sendCode()}
              status={message("auth03.resend.action")}
              subtitle={message("auth03.resend.body")}
              title={message("auth03.resend.title")}
              tone="info"
            />
          </>
        ) : (
          <PremiumTextField
            autoComplete="email"
            keyboardType="email-address"
            label={message("acc03.new.email")}
            onChangeText={setContactValue}
            testID="activation-contact-value"
            textContentType="emailAddress"
            value={contactValue}
          />
        )}
        <AccountBanner
          body={message("auth03.banner.body")}
          title={message("auth03.banner.title")}
        />
        {errorNotice}
        {codeSent ? (
          <>
            <BlockAction
              busy={busy}
              disabled={code.trim().length === 0}
              label={message("auth03.confirm")}
              onPress={() => void confirmCode()}
              testID="activation-contact-confirm"
            />
            <BlockAction
              kind="secondary"
              label={message("auth03.changeContact")}
              onPress={() => {
                setCodeSent(false);
                setCode("");
              }}
            />
          </>
        ) : (
          <BlockAction
            busy={busy}
            disabled={contactValue.trim().length === 0}
            label={message("acc03.sendCode")}
            onPress={() => void sendCode()}
            testID="activation-contact-send"
          />
        )}
      </PremiumScrollScreen>
    );
  }

  if (step === "codes") {
    return (
      <PremiumScrollScreen contentStyle={styles.wizardContent}>
        <AmbientGlow />
        {brand}
        <ScreenHeading
          eyebrow={message("auth04.eyebrow")}
          subtitle={message("auth04.banner.body")}
          title={message("acc06.codesTitle")}
        />
        <View style={styles.codesCard} testID="activation-recovery-codes">
          {recoveryCodes.map((recoveryCode) => (
            <Text key={recoveryCode} style={styles.codeText}>
              {recoveryCode}
            </Text>
          ))}
        </View>
        {errorNotice}
        <BlockAction
          busy={busy}
          label={message("acc06.codesDone")}
          onPress={() => void run(finish)}
          testID="activation-codes-done"
        />
      </PremiumScrollScreen>
    );
  }

  const startTwofa = () =>
    run(async () => {
      if (link.status !== "ready") return;
      const started = await api.startActivationTwofa({ token: link.token });
      setEnrollment(started);
      setCode("");
    });
  const confirmTwofa = () =>
    run(async () => {
      if (link.status !== "ready") return;
      const codes = await api.confirmActivationTwofa({
        token: link.token,
        code: code.trim(),
      });
      setRecoveryCodes(codes);
      await refreshProgress();
      setStep("codes");
    });
  return (
    <PremiumScrollScreen keyboardAware contentStyle={styles.wizardContent}>
      <AmbientGlow />
      {brand}
      <ScreenHeading
        eyebrow={message("auth04.eyebrow")}
        subtitle={message("auth04.subtitle")}
        title={message("auth04.title")}
      />
      <StatusCard
        body={message("auth04.card.body")}
        status={message("auth04.card.status")}
        title={message("auth04.card.title")}
        tone="info"
      />
      {enrollment !== null ? (
        <>
          <View style={styles.codesCard} testID="activation-twofa-secret">
            <Text style={styles.secretLabel}>{message("acc06.secretLabel")}</Text>
            <Text selectable style={styles.codeText}>
              {enrollment.secret}
            </Text>
          </View>
          <PremiumTextField
            autoComplete="one-time-code"
            keyboardType="number-pad"
            label={message("auth04.codeLabel")}
            onChangeText={setCode}
            testID="activation-twofa-code"
            textContentType="oneTimeCode"
            value={code}
          />
        </>
      ) : null}
      <AccountBanner
        body={message("auth04.banner.body")}
        title={message("auth04.banner.title")}
      />
      {errorNotice}
      {enrollment !== null ? (
        <BlockAction
          busy={busy}
          disabled={code.trim().length === 0}
          label={message("auth04.enable")}
          onPress={() => void confirmTwofa()}
          testID="activation-twofa-confirm"
        />
      ) : (
        <BlockAction
          busy={busy}
          label={message("auth04.enable")}
          onPress={() => void startTwofa()}
          testID="activation-twofa-start"
        />
      )}
      <BlockAction
        busy={busy}
        kind="secondary"
        label={message("auth04.later")}
        onPress={() => void run(finish)}
        testID="activation-twofa-later"
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
  wizardContent: { gap: space.s3, minHeight: 780 },
  centerBadge: { alignItems: "center", marginTop: spacing.xxl },
  centerTitle: {
    color: colors.textPrimary,
    fontFamily: fonts.extrabold,
    marginTop: spacing.loose,
    textAlign: "center",
    ...typeScale.screenTitle,
  },
  unavailableActions: { gap: spacing.md, marginTop: spacing.xxl },
  brandRow: {
    alignItems: "center",
    flexDirection: "row",
    gap: space.s3,
    marginBottom: space.s2,
    marginTop: spacing.md,
  },
  brandName: {
    color: semantic.textPrimary,
    fontFamily: "Onest_800ExtraBold",
    fontSize: 16,
    letterSpacing: 2,
    lineHeight: 22,
  },
  brandTagline: { color: semantic.textSecondary, ...typeStyles.caption },
  codesCard: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderDefault,
    borderRadius: radius.lg,
    borderWidth: strokes.hairline,
    gap: space.s2,
    padding: space.s4,
  },
  secretLabel: { color: semantic.textMuted, ...typeStyles.labelM },
  codeText: {
    color: semantic.textPrimary,
    fontFamily: "Onest_600SemiBold",
    fontSize: 15,
    letterSpacing: 1,
    lineHeight: 22,
  },
});
