import { useState } from "react";

import { useApiClient } from "@/api";
import type { NotificationCategory } from "@/api/contracts";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice } from "../../components";
import {
  AccountBanner,
  AccountScreenShell,
  ScreenHeading,
  ToggleRow,
} from "../../patterns/accountPatterns";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * Notification preferences (Figma ACT-03/04). One toggle per category
 * for the push channel; the in-app channel is always on («Всегда
 * включено») and every change persists immediately. Push delivery
 * itself ships with the device-permission modules — the stored choice
 * is honest data, not a live channel yet, and the banner says so.
 */
export function NotificationPreferencesScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const preferences = useAccountResource((accessToken) =>
    api.notificationPreferences(accessToken),
  );
  const [busyCategory, setBusyCategory] = useState<NotificationCategory | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const toggle = async (category: NotificationCategory, next: boolean) => {
    setActionError(null);
    setBusyCategory(category);
    try {
      await runAuthenticated((accessToken) =>
        api.updateNotificationPreference(accessToken, {
          category,
          pushEnabled: next,
        }),
      );
      await preferences.reload();
    } catch (cause) {
      setActionError(apiErrorMessage(cause));
      await preferences.reload();
    } finally {
      setBusyCategory(null);
    }
  };

  return (
    <AccountScreenShell navigation={<AccountNav />} testID="notification-preferences">
      <ScreenHeading
        eyebrow={message("act.prefs.eyebrow")}
        subtitle={message("act.prefs.subtitle")}
        title={message("act.prefs.title")}
      />
      {preferences.error !== null ? (
        <InlineNotice
          body={apiErrorMessage(preferences.error)}
          title={message("common.retry")}
          tone="error"
        />
      ) : null}
      {actionError !== null ? (
        <InlineNotice body={actionError} title={message("common.retry")} tone="error" />
      ) : null}
      {(preferences.value ?? []).map((preference) => (
        <ToggleRow
          key={preference.category}
          offLabel={message("act.prefs.push.off")}
          onLabel={message("act.prefs.push.on")}
          onToggle={
            busyCategory === null
              ? (next) => void toggle(preference.category, next)
              : undefined
          }
          subtitle={message(`act.prefs.category.${preference.category}.sub`)}
          testID={`preference-${preference.category}`}
          title={message(`act.category.${preference.category}`)}
          value={preference.pushEnabled}
        />
      ))}
      <AccountBanner
        body={message("act.prefs.push.pending")}
        title={message("act.prefs.inApp")}
      />
    </AccountScreenShell>
  );
}
