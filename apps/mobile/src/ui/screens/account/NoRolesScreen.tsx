import { useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { BrandBadge, InlineNotice, SecondaryButton } from "../../components";
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

/**
 * AUTH-11 · Доступ готовится (Figma 377:1070). The account is active but
 * no role is granted yet — instead of an empty navigation shell the
 * screen explains the state and lets the person re-check or leave.
 */
export function NoRolesScreen() {
  const message = useMessage();
  const { retryBootstrap, signOut } = useSession();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = async () => {
    setBusy(true);
    setError(null);
    try {
      await retryBootstrap();
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AccountScreenShell testID="no-roles">
      <View style={styles.brandRow}>
        <BrandBadge />
        <View>
          <Text style={styles.brandName}>{message("auth.brand")}</Text>
          <Text style={styles.brandTagline}>{message("auth.brandTagline")}</Text>
        </View>
      </View>
      <ScreenHeading
        eyebrow={message("auth11.eyebrow")}
        subtitle={message("auth11.subtitle")}
        title={message("auth11.title")}
      />
      <StatusCard
        body={message("auth11.card.body")}
        status={message("auth11.card.status")}
        title={message("auth09.banner.body")}
        tone="info"
      />
      <StatusRow
        status={message("auth11.row.status")}
        subtitle={message("auth11.row.subtitle")}
        title={message("auth11.row.title")}
        tone="muted"
      />
      <AccountBanner
        body={message("auth11.banner.body")}
        title={message("auth11.banner.title")}
      />
      {error !== null ? (
        <InlineNotice body={error} title={message("common.retry")} tone="error" />
      ) : null}
      <BlockAction
        busy={busy}
        label={message("auth11.refresh")}
        onPress={() => void refresh()}
        testID="no-roles-refresh"
      />
      <BlockAction
        kind="secondary"
        label={message("auth11.contact")}
        onPress={() => void refresh()}
      />
      <SecondaryButton label="Выйти" onPress={() => void signOut()} />
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
