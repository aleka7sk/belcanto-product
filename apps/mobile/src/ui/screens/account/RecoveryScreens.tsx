import { router, useLocalSearchParams } from "expo-router";
import { useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { ApiError, useApiClient } from "@/api";
import { useMessage } from "@/i18n";
import { BrandBadge, InlineNotice, PremiumTextField } from "../../components";
import {
  AccountBanner,
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage } from "../../viewModels";

/** Shared Belcanto brand header for the AUTH screens (Figma 377:8xx). */
function AuthBrand() {
  const message = useMessage();
  return (
    <View style={styles.brandRow}>
      <BrandBadge />
      <View>
        <Text style={styles.brandName}>{message("auth.brand")}</Text>
        <Text style={styles.brandTagline}>{message("auth.brandTagline")}</Text>
      </View>
    </View>
  );
}

/**
 * AUTH-07 · Восстановить доступ (Figma 377:963). The response never
 * reveals whether the contact belongs to an account; recovery is
 * phone-based today, so an email-looking value gets the honest format
 * hint instead of a fake success.
 */
export function RecoveryRequestScreen() {
  const message = useMessage();
  const api = useApiClient();
  const [contact, setContact] = useState("");
  const [busy, setBusy] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submit = async () => {
    setBusy(true);
    setError(null);
    try {
      await api.requestPasswordReset({ phone: contact.trim() });
      setSent(true);
    } catch (cause) {
      const invalid = cause instanceof ApiError && cause.code === "INVALID_INPUT";
      setError(invalid ? message("auth07.phoneOnly") : apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AccountScreenShell testID="recovery-request">
      <AuthBrand />
      <ScreenHeading
        eyebrow={message("auth07.eyebrow")}
        subtitle={message("auth07.subtitle")}
        title={message("auth07.title")}
      />
      <PremiumTextField
        keyboardType="phone-pad"
        label={message("auth07.input")}
        onChangeText={setContact}
        testID="recovery-contact"
        value={contact}
      />
      <StatusCard
        body={message("auth07.card.body")}
        status={message("auth07.card.status")}
        title={message("auth07.card.title")}
        tone="info"
      />
      {sent ? (
        <InlineNotice
          body={message("auth07.sent")}
          title={message("auth07.submit")}
          tone="success"
        />
      ) : null}
      {error !== null ? (
        <InlineNotice body={error} title={message("common.retry")} tone="error" />
      ) : null}
      <BlockAction
        busy={busy}
        disabled={contact.trim().length === 0}
        label={message("auth07.submit")}
        onPress={() => void submit()}
        testID="recovery-submit"
      />
      <BlockAction
        kind="secondary"
        label={message("auth07.backToSignIn")}
        onPress={() => router.replace("/sign-in")}
      />
    </AccountScreenShell>
  );
}

/**
 * AUTH-08 · Новый пароль + AUTH-09 · Ссылка недействительна (Figma
 * 377:987 / 377:1014). Completing the reset rotates the credential and
 * revokes every session family server-side — the sessions row states
 * that as a fact, not an option.
 */
export function RecoveryCompleteScreen() {
  const message = useMessage();
  const api = useApiClient();
  const params = useLocalSearchParams<{ token?: string }>();
  const token = typeof params.token === "string" ? params.token : "";
  const [password, setPassword] = useState("");
  const [repeat, setRepeat] = useState("");
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [invalidLink, setInvalidLink] = useState(token.length === 0);
  const [error, setError] = useState<string | null>(null);

  const submit = async () => {
    if (password !== repeat) {
      setError(message("auth08.mismatch"));
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await api.completePasswordReset({ token, newPassword: password });
      setDone(true);
    } catch (cause) {
      if (cause instanceof ApiError && cause.code === "UNAUTHENTICATED") {
        setInvalidLink(true);
      } else {
        setError(apiErrorMessage(cause));
      }
    } finally {
      setBusy(false);
    }
  };

  if (invalidLink) {
    return (
      <AccountScreenShell testID="recovery-invalid">
        <AuthBrand />
        <ScreenHeading
          eyebrow={message("auth09.eyebrow")}
          subtitle={message("auth09.subtitle")}
          title={message("auth09.title")}
        />
        <StatusCard
          body={message("auth09.card.body")}
          status={message("auth09.card.status")}
          title={message("auth09.card.title")}
          tone="muted"
        />
        <BlockAction
          label={message("auth09.goToSignIn")}
          onPress={() => router.replace("/sign-in")}
          testID="recovery-invalid-signin"
        />
        <BlockAction
          kind="secondary"
          label={message("auth09.requestNew")}
          onPress={() => router.replace("/recover")}
        />
        <AccountBanner
          body={message("auth09.banner.body")}
          title={message("auth09.banner.title")}
        />
      </AccountScreenShell>
    );
  }

  return (
    <AccountScreenShell testID="recovery-complete">
      <AuthBrand />
      <ScreenHeading
        eyebrow={message("auth08.eyebrow")}
        subtitle={message("auth08.subtitle")}
        title={message("auth08.title")}
      />
      <PremiumTextField
        label={message("auth08.newPassword")}
        onChangeText={setPassword}
        secureTextEntry
        testID="recovery-password"
        value={password}
      />
      <PremiumTextField
        label={message("auth08.repeatPassword")}
        onChangeText={setRepeat}
        secureTextEntry
        testID="recovery-repeat"
        value={repeat}
      />
      <StatusRow
        status={message("auth08.sessions.status")}
        subtitle={message("auth08.sessions.subtitle")}
        title={message("auth08.sessions.title")}
        tone="success"
      />
      {done ? (
        <InlineNotice
          body={message("auth08.done")}
          title={message("auth08.title")}
          tone="success"
        />
      ) : null}
      {error !== null ? (
        <InlineNotice body={error} title={message("common.retry")} tone="error" />
      ) : null}
      {done ? (
        <BlockAction
          label={message("auth09.goToSignIn")}
          onPress={() => router.replace("/sign-in")}
          testID="recovery-done-signin"
        />
      ) : (
        <BlockAction
          busy={busy}
          disabled={password.length === 0 || repeat.length === 0}
          label={message("auth08.submit")}
          onPress={() => void submit()}
          testID="recovery-save"
        />
      )}
      <BlockAction
        kind="secondary"
        label={message("auth08.cancel")}
        onPress={() => router.replace("/sign-in")}
      />
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  brandRow: {
    alignItems: "center",
    flexDirection: "row",
    gap: space.s3,
    marginBottom: space.s2,
  },
  brandName: {
    color: semantic.textPrimary,
    fontFamily: "Onest_800ExtraBold",
    fontSize: 16,
    letterSpacing: 2,
    lineHeight: 22,
  },
  brandTagline: { color: semantic.textSecondary, ...typeStyles.caption },
});
