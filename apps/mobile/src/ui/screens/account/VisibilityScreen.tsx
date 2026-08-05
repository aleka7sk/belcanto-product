import { router } from "expo-router";

import { useMessage } from "@/i18n";
import {
  AccountBanner,
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusRow,
} from "../../patterns/accountPatterns";
import { AccountNav } from "./shared";

/**
 * ACC-11 · Кто видит данные (Figma 366:511). The visibility matrix is a
 * statement of the HOF permission model, not per-user settings — rows are
 * informational; the audit trail lives in the security feed.
 */
export function VisibilityScreen() {
  const message = useMessage();
  return (
    <AccountScreenShell navigation={<AccountNav />} testID="account-visibility">
      <ScreenHeading
        eyebrow={message("acc11.eyebrow")}
        subtitle={message("acc11.subtitle")}
        title={message("acc11.title")}
      />
      <StatusRow
        status={message("acc11.progress.note")}
        subtitle={message("acc11.progress.who")}
        title={message("acc11.progress.title")}
        tone="info"
      />
      <StatusRow
        status={message("acc11.homework.note")}
        subtitle={message("acc11.homework.who")}
        title={message("acc11.homework.title")}
        tone="info"
      />
      <StatusRow
        status={message("acc11.practice.note")}
        subtitle={message("acc11.practice.who")}
        title={message("acc11.practice.title")}
        tone="success"
      />
      <StatusRow
        status={message("acc11.contacts.note")}
        subtitle={message("acc11.contacts.who")}
        title={message("acc11.contacts.title")}
        tone="success"
      />
      <StatusRow
        status={message("acc11.group.note")}
        subtitle={message("acc11.group.who")}
        title={message("acc11.group.title")}
        tone="muted"
      />
      <AccountBanner
        body={message("acc11.banner.body")}
        title={message("acc11.banner.title")}
      />
      <BlockAction
        kind="secondary"
        label={message("acc11.auditLog")}
        onPress={() => router.push("/(protected)/account/security/activity")}
      />
    </AccountScreenShell>
  );
}
