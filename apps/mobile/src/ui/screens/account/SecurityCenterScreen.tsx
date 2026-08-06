import { router } from "expo-router";
import { RefreshControl } from "react-native";

import { useApiClient } from "@/api";
import { useMessage } from "@/i18n";
import {
  AccountBanner,
  AccountScreenShell,
  ScreenHeading,
  SettingsRow,
  StatusCard,
} from "../../patterns/accountPatterns";
import { semantic } from "../../tokens";
import { AccountNav, useAccountResource } from "./shared";

/** ACC-05 · Безопасность (Figma 365:297). */
export function SecurityCenterScreen() {
  const message = useMessage();
  const api = useApiClient();
  const twofa = useAccountResource((accessToken) => api.twofaStatus(accessToken));
  const sessions = useAccountResource((accessToken) =>
    api.listMySessions(accessToken),
  );

  const protectedAccount = twofa.value?.enabled === true;
  const sessionCount = sessions.value?.length ?? 0;

  return (
    <AccountScreenShell
      navigation={<AccountNav />}
      refreshControl={
        <RefreshControl
          onRefresh={() => {
            void Promise.all([twofa.reload(), sessions.reload()]);
          }}
          refreshing={twofa.refreshing || sessions.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="account-security"
    >
      <ScreenHeading
        eyebrow={message("acc05.eyebrow")}
        subtitle={message("acc05.subtitle")}
        title={message("acc05.title")}
      />
      <StatusCard
        body={message(
          protectedAccount ? "acc05.summary.good.body" : "acc05.summary.basic.body",
        )}
        status={protectedAccount ? message("acc05.summary.status") : undefined}
        title={message(
          protectedAccount ? "acc05.summary.good.title" : "acc05.summary.basic.title",
        )}
        tone={protectedAccount ? "success" : "warning"}
      />
      <SettingsRow
        onPress={() => router.push("/recover")}
        subtitle={message("acc05.password.subtitle")}
        title={message("acc05.password.title")}
      />
      <SettingsRow
        emphasis
        onPress={() => router.push("/(protected)/account/security/twofa")}
        subtitle={message("acc05.twofa.subtitle")}
        tail={
          twofa.value === null
            ? undefined
            : message(twofa.value.enabled ? "acc05.twofa.on" : "acc05.twofa.off")
        }
        testID="security-twofa-row"
        title={message("acc05.twofa.title")}
      />
      <SettingsRow
        disabledReason={message("acc05.faceid.unavailable")}
        onPress={() => undefined}
        subtitle={message("acc05.faceid.subtitle")}
        title={message("acc05.faceid.title")}
      />
      <SettingsRow
        onPress={() => router.push("/(protected)/account/security/devices")}
        subtitle={
          sessionCount === 1
            ? message("acc05.devices.subtitle.one")
            : message("acc05.devices.subtitle.many", { count: sessionCount })
        }
        testID="security-devices-row"
        title={message("acc05.devices.title")}
      />
      <SettingsRow
        emphasis
        onPress={() => router.push("/(protected)/account/security/activity")}
        subtitle={message("acc05.activity.subtitle")}
        tail={message("acc05.activity.tail")}
        testID="security-activity-row"
        title={message("acc05.activity.title")}
      />
      <AccountBanner
        body={message("acc05.banner.body")}
        title={message("acc05.banner.title")}
      />
    </AccountScreenShell>
  );
}
