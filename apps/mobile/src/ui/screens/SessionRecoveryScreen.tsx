import { router } from "expo-router";
import { useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { useSession } from "@/session";
import {
  AmbientGlow,
  BrandBadge,
  InlineNotice,
  PremiumScrollScreen,
  PrimaryButton,
  SecondaryButton,
} from "../components";
import { colors, fonts, spacing, typeScale } from "../tokens";
import { apiErrorMessage } from "../viewModels";

export function SessionRecoveryScreen() {
  const { retryBootstrap, signOut, state } = useSession();
  const [busy, setBusy] = useState(false);
  const [retryError, setRetryError] = useState<string | null>(null);

  const retry = async () => {
    setBusy(true);
    setRetryError(null);
    try {
      await retryBootstrap();
    } catch (error) {
      setRetryError(apiErrorMessage(error));
    } finally {
      setBusy(false);
    }
  };

  const leave = async () => {
    await signOut();
    router.replace("/");
  };

  const originalError = state.failure?.error;
  return (
    <PremiumScrollScreen contentStyle={styles.content}>
      <AmbientGlow />
      <View style={styles.badge}>
        <BrandBadge />
      </View>
      <Text accessibilityRole="header" style={styles.title}>
        Пространство не открылось
      </Text>
      <Text style={styles.subtitle}>
        Вход сохранён, но персональные данные и права пока не удалось загрузить.
      </Text>
      <InlineNotice
        body={retryError ?? apiErrorMessage(originalError)}
        title="Можно безопасно повторить"
        tone="error"
      />
      <View style={styles.actions}>
        <PrimaryButton
          busy={busy}
          label="Попробовать снова"
          onPress={() => void retry()}
        />
        <SecondaryButton label="Выйти" onPress={() => void leave()} />
      </View>
    </PremiumScrollScreen>
  );
}

const styles = StyleSheet.create({
  content: { justifyContent: "center", minHeight: 720 },
  badge: { alignItems: "center" },
  title: {
    color: colors.textPrimary,
    fontFamily: fonts.extrabold,
    marginTop: spacing.loose,
    textAlign: "center",
    ...typeScale.screenTitle,
  },
  subtitle: {
    color: colors.textSecondary,
    fontFamily: fonts.regular,
    marginBottom: spacing.section,
    marginTop: spacing.sm,
    textAlign: "center",
    ...typeScale.body,
  },
  actions: { gap: spacing.md, marginTop: spacing.section },
});
