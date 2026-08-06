import { router } from "expo-router";
import { RefreshControl } from "react-native";

import { ROLE_PRIORITY } from "@/access/activeRole";
import { useApiClient } from "@/api";
import { useMessage } from "@/i18n";
import {
  AccountBanner,
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
} from "../../patterns/accountPatterns";
import { semantic } from "../../tokens";
import { AccountNav, useAccountResource } from "./shared";

/**
 * ACC-19 · Мои роли и доступ (Figma 366:996). Read-only: roles are
 * granted and revoked by Admin/Owner and every change lands in the audit
 * feed. Roles the account does not hold are omitted, not teased — the
 * List Row permission rule from the component contract.
 */
export function RolesAccessScreen() {
  const message = useMessage();
  const api = useApiClient();
  const profile = useAccountResource((accessToken) => api.myProfile(accessToken));
  const roles = profile.value?.roles ?? [];
  const tenant = profile.value?.tenantName ?? "";

  return (
    <AccountScreenShell
      navigation={<AccountNav />}
      refreshControl={
        <RefreshControl
          onRefresh={() => {
            void profile.reload();
          }}
          refreshing={profile.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="account-roles"
    >
      <ScreenHeading
        eyebrow={message("acc19.eyebrow")}
        subtitle={message("acc19.subtitle")}
        title={message("acc19.title")}
      />
      {ROLE_PRIORITY.filter((role) => roles.includes(role)).map((role) => (
        <StatusCard
          key={role}
          body={
            role === "Student"
              ? message("acc19.roleContext", { tenant })
              : message("acc19.staffContext", { tenant })
          }
          status={
            role === "Teacher" || role === "Administrator" || role === "Owner"
              ? message("acc19.active2fa")
              : message("acc19.active")
          }
          title={message(`role.${role}`)}
          tone="success"
        />
      ))}
      <AccountBanner
        body={message("acc19.banner.body")}
        title={message("acc19.banner.title")}
      />
      <BlockAction
        disabled={roles.length <= 1}
        label={message("acc19.switch")}
        onPress={() => router.push("/(protected)/account/role")}
        testID="roles-switch"
      />
    </AccountScreenShell>
  );
}
