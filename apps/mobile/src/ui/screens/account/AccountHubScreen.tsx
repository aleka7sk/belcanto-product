import { router } from "expo-router";
import { RefreshControl } from "react-native";

import { useApiClient } from "@/api";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice } from "../../components";
import {
  AccountScreenShell,
  ProfileHero,
  ScreenHeading,
  SettingsRow,
} from "../../patterns/accountPatterns";
import { semantic } from "../../tokens";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav, initialsOf, useAccountResource, useWorkingRole } from "./shared";

/** ACC-01 · Профиль и настройки (Figma 365:2). */
export function AccountHubScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { state } = useSession();
  const workingRole = useWorkingRole();
  const profile = useAccountResource((accessToken) => api.myProfile(accessToken));
  const twofa = useAccountResource((accessToken) => api.twofaStatus(accessToken));
  const isStudent = state.bootstrap?.studentId !== undefined;

  const roleLabel = workingRole ? message(`role.${workingRole}`) : "";
  const roleCount = profile.value?.roles.length ?? 0;
  const roleSubtitle =
    roleCount <= 1
      ? message("acc01.roleRow.only")
      : roleCount === 2
        ? message("acc01.roleRow.one")
        : message("acc01.roleRow.many", { count: roleCount - 1 });

  return (
    <AccountScreenShell
      navigation={<AccountNav />}
      refreshControl={
        <RefreshControl
          onRefresh={() => {
            void Promise.all([profile.reload(), twofa.reload()]);
          }}
          refreshing={profile.refreshing || twofa.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="account-hub"
    >
      <ScreenHeading
        eyebrow={message("acc01.eyebrow")}
        subtitle={message("acc01.subtitle")}
        title={message("acc01.title")}
      />
      {profile.error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={apiErrorMessage(profile.error)}
          onAction={() => void profile.reload()}
          title={message("acc01.title")}
        />
      ) : null}
      {profile.value ? (
        <ProfileHero
          context={message("common.roleContext", {
            role: roleLabel,
            tenant: profile.value.tenantName,
          })}
          initials={initialsOf(profile.value.fullName)}
          name={profile.value.fullName}
          status={message("acc01.verified")}
          statusTone="success"
        />
      ) : null}
      <SettingsRow
        emphasis
        onPress={() => router.push("/(protected)/account/role")}
        subtitle={roleSubtitle}
        tail={roleCount > 1 ? message("acc01.roleRow.tail") : undefined}
        testID="account-role-row"
        title={message("acc01.roleRow.title", { role: roleLabel })}
      />
      <SettingsRow
        onPress={() => router.push("/(protected)/activity")}
        subtitle={message("act.entry.subtitle")}
        testID="account-activity-row"
        title={message("act.entry.title")}
      />
      <SettingsRow
        onPress={() => router.push("/(protected)/activity/preferences")}
        subtitle={message("act.prefs.subtitle")}
        testID="account-notifications-row"
        title={message("act.prefs.title")}
      />
      {isStudent ? (
        <SettingsRow
          onPress={() => router.push("/(protected)/progress")}
          subtitle={message("growth.entry.subtitle")}
          testID="account-progress-row"
          title={message("growth.entry.title")}
        />
      ) : null}
      <SettingsRow
        onPress={() => router.push("/(protected)/account/personal")}
        subtitle={message("acc01.personal.subtitle")}
        title={message("acc01.personal.title")}
      />
      <SettingsRow
        onPress={() => router.push("/(protected)/account/security")}
        subtitle={message("acc01.security.subtitle")}
        tail={twofa.value?.enabled ? message("acc01.security.protected") : undefined}
        title={message("acc01.security.title")}
      />
      <SettingsRow
        onPress={() => router.push("/(protected)/account/privacy")}
        subtitle={message("acc01.privacy.subtitle")}
        title={message("acc01.privacy.title")}
      />
      <SettingsRow
        disabledReason={message("acc01.device.subtitle")}
        onPress={() => undefined}
        subtitle={message("acc01.device.subtitle")}
        title={message("acc01.device.title")}
      />
      <SettingsRow
        onPress={() => router.push("/(protected)/account/data")}
        subtitle={message("acc01.data.subtitle")}
        title={message("acc01.data.title")}
      />
    </AccountScreenShell>
  );
}
