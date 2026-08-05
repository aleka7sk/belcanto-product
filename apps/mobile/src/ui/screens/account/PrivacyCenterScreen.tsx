import { router } from "expo-router";

import { useMessage } from "@/i18n";
import {
  AccountScreenShell,
  ScreenHeading,
  SettingsRow,
  StatusCard,
} from "../../patterns/accountPatterns";
import { AccountNav } from "./shared";

/** ACC-10 · Конфиденциальность (Figma 366:397). */
export function PrivacyCenterScreen() {
  const message = useMessage();
  return (
    <AccountScreenShell navigation={<AccountNav />} testID="account-privacy">
      <ScreenHeading
        eyebrow={message("acc10.eyebrow")}
        subtitle={message("acc10.subtitle")}
        title={message("acc10.title")}
      />
      <StatusCard
        body={message("acc10.card.body")}
        status={message("acc10.card.status")}
        title={message("acc10.card.title")}
        tone="success"
      />
      <SettingsRow
        onPress={() => router.push("/(protected)/account/privacy/visibility")}
        subtitle={message("acc10.visibility.subtitle")}
        title={message("acc10.visibility.title")}
      />
      <SettingsRow
        onPress={() => router.push("/(protected)/account/privacy/community")}
        subtitle={message("acc10.community.subtitle")}
        testID="privacy-community-row"
        title={message("acc10.community.title")}
      />
      <SettingsRow
        disabledReason={message("acc10.blocked.later")}
        onPress={() => undefined}
        subtitle={message("acc10.blocked.later")}
        title={message("acc10.blocked.title")}
      />
      <SettingsRow
        onPress={() => router.push("/(protected)/account/data")}
        subtitle={message("acc10.export.subtitle")}
        title={message("acc10.export.title")}
      />
      <SettingsRow
        onPress={() => router.push("/(protected)/account/data/deletion")}
        subtitle={message("acc10.deletion.subtitle")}
        title={message("acc10.deletion.title")}
      />
      <SettingsRow
        onPress={() => router.push("/(protected)/account/data/policies")}
        subtitle={message("acc10.legal.subtitle")}
        title={message("acc10.legal.title")}
      />
    </AccountScreenShell>
  );
}
