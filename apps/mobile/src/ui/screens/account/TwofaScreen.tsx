import { router } from "expo-router";
import { useState } from "react";
import { RefreshControl, StyleSheet, Text, View } from "react-native";

import { useApiClient, type TwofaEnrollment } from "@/api";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice, PremiumTextField } from "../../components";
import {
  AccountBanner,
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import { radius, semantic, space, strokes, typeStyles } from "../../tokens";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav, useAccountResource } from "./shared";

/**
 * ACC-06 · Двухэтапная проверка (Figma 365:401). Enrollment walks the
 * design's three steps: connect the authenticator (secret shown once),
 * confirm with a 6-digit code, store the one-time recovery codes. When
 * 2FA is on, the same screen owns the disable path (password + code).
 */
export function TwofaScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const status = useAccountResource((accessToken) => api.twofaStatus(accessToken));

  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [enrollment, setEnrollment] = useState<TwofaEnrollment | null>(null);
  const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const start = async () => {
    setBusy(true);
    setError(null);
    try {
      const started = await runAuthenticated((accessToken) =>
        api.startTwofaEnrollment(accessToken, { currentPassword: password }),
      );
      setEnrollment(started);
      setPassword("");
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  const confirm = async () => {
    setBusy(true);
    setError(null);
    try {
      const codes = await runAuthenticated((accessToken) =>
        api.confirmTwofaEnrollment(accessToken, { code: code.trim() }),
      );
      setRecoveryCodes(codes);
      setEnrollment(null);
      setCode("");
      await status.reload();
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  const disable = async () => {
    setBusy(true);
    setError(null);
    try {
      await runAuthenticated((accessToken) =>
        api.disableTwofa(accessToken, {
          currentPassword: password,
          code: code.trim(),
        }),
      );
      setPassword("");
      setCode("");
      await status.reload();
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  const enabled = status.value?.enabled === true;

  return (
    <AccountScreenShell
      navigation={<AccountNav />}
      refreshControl={
        <RefreshControl
          onRefresh={() => {
            void status.reload();
          }}
          refreshing={status.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="account-twofa"
    >
      <ScreenHeading
        eyebrow={message("acc06.eyebrow")}
        subtitle={message("acc06.subtitle")}
        title={message("acc06.title")}
      />
      {recoveryCodes !== null ? (
        <>
          <StatusCard
            body={message("acc06.step3.body")}
            status={message("acc06.step3.status")}
            title={message("acc06.codesTitle")}
            tone="warning"
          />
          <View style={styles.codesCard} testID="twofa-recovery-codes">
            {recoveryCodes.map((recoveryCode) => (
              <Text key={recoveryCode} style={styles.code}>
                {recoveryCode}
              </Text>
            ))}
          </View>
          <BlockAction
            label={message("acc06.codesDone")}
            onPress={() => {
              setRecoveryCodes(null);
              router.back();
            }}
            testID="twofa-codes-done"
          />
        </>
      ) : enabled ? (
        <>
          <StatusCard
            body={message("acc06.enabled.body")}
            status={message("acc06.recovery.remaining", {
              count: status.value?.recoveryCodesRemaining ?? 0,
            })}
            title={message("acc06.enabled.title")}
            tone="success"
          />
          <PremiumTextField
            label={message("acc06.passwordLabel")}
            onChangeText={setPassword}
            secureTextEntry
            testID="twofa-password"
            value={password}
          />
          <PremiumTextField
            label={message("acc06.disableCode")}
            onChangeText={setCode}
            testID="twofa-code"
            value={code}
          />
          {error !== null ? (
            <InlineNotice body={error} title={message("common.retry")} tone="error" />
          ) : null}
          <BlockAction
            busy={busy}
            disabled={password.length === 0 || code.trim().length === 0}
            kind="secondary"
            label={message("acc06.disable")}
            onPress={() => void disable()}
            testID="twofa-disable"
          />
        </>
      ) : enrollment !== null ? (
        <>
          <StatusCard
            body={message("acc06.step1.body")}
            status={message("acc06.step1.status")}
            title={message("acc06.step1.title")}
            tone="info"
          />
          <View style={styles.codesCard} testID="twofa-secret">
            <Text style={styles.secretLabel}>{message("acc06.secretLabel")}</Text>
            <Text selectable style={styles.code}>
              {enrollment.secret}
            </Text>
          </View>
          <StatusRow
            status={message("acc06.step2.status")}
            subtitle={message("acc06.step2.body")}
            title={message("acc06.step2.title")}
            tone="warning"
          />
          <PremiumTextField
            keyboardType="number-pad"
            label={message("acc06.codeLabel")}
            onChangeText={setCode}
            testID="twofa-confirm-code"
            value={code}
          />
          {error !== null ? (
            <InlineNotice body={error} title={message("common.retry")} tone="error" />
          ) : null}
          <BlockAction
            busy={busy}
            disabled={code.trim().length === 0}
            label={message("acc06.enable")}
            onPress={() => void confirm()}
            testID="twofa-enable"
          />
        </>
      ) : (
        <>
          <StatusCard
            body={message("acc06.step1.body")}
            status={message("acc06.step1.status")}
            title={message("acc06.step1.title")}
            tone="info"
          />
          <StatusRow
            status={message("acc06.step2.status")}
            subtitle={message("acc06.step2.body")}
            title={message("acc06.step2.title")}
            tone="warning"
          />
          <StatusCard
            body={message("acc06.step3.body")}
            status={message("acc06.step3.status")}
            title={message("acc06.step3.title")}
            tone="warning"
          />
          <AccountBanner
            body={message("acc06.banner.body")}
            title={message("acc06.banner.title")}
            tone="success"
          />
          <PremiumTextField
            label={message("acc06.passwordLabel")}
            onChangeText={setPassword}
            secureTextEntry
            testID="twofa-start-password"
            value={password}
          />
          {error !== null ? (
            <InlineNotice body={error} title={message("common.retry")} tone="error" />
          ) : null}
          <BlockAction
            busy={busy}
            disabled={password.length === 0}
            label={message("acc06.start")}
            onPress={() => void start()}
            testID="twofa-start"
          />
        </>
      )}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  codesCard: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderDefault,
    borderRadius: radius.lg,
    borderWidth: strokes.hairline,
    gap: space.s2,
    padding: space.s4,
  },
  secretLabel: { color: semantic.textMuted, ...typeStyles.labelM },
  code: {
    color: semantic.textPrimary,
    fontFamily: "Onest_600SemiBold",
    fontSize: 15,
    letterSpacing: 1,
    lineHeight: 22,
  },
});
