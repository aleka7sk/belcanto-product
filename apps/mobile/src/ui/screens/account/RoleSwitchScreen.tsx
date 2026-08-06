import { router } from "expo-router";
import { useState } from "react";
import { RefreshControl } from "react-native";

import { ROLE_PRIORITY, resolveActiveRole, useActiveRole } from "@/access/activeRole";
import { useApiClient, type Role } from "@/api";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import {
  AccountBanner,
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import { semantic } from "../../tokens";
import { AccountNav, useAccountResource, useWorkingRole } from "./shared";

/**
 * ACC-04 · Переключить роль (Figma 365:242) + ACC-20 · Доступ к роли
 * закрыт (366:1049). Switching replaces the navigation context; when the
 * previously chosen role is gone from the bootstrap set the screen
 * explains what is preserved and offers the remaining role.
 */
export function RoleSwitchScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { state } = useSession();
  const { preferredRole, setPreferredRole, roleRevoked } = useActiveRole();
  const workingRole = useWorkingRole();
  const profile = useAccountResource((accessToken) => api.myProfile(accessToken));
  const roles = state.bootstrap?.roles ?? [];
  const revoked = roleRevoked(roles);
  const [selected, setSelected] = useState<Role | null>(null);

  const target =
    selected ??
    ROLE_PRIORITY.find((role) => roles.includes(role) && role !== workingRole) ??
    null;

  if (revoked && preferredRole !== null) {
    const fallback = resolveActiveRole(roles, null);
    return (
      <AccountScreenShell navigation={<AccountNav />} testID="account-role-revoked">
        <ScreenHeading
          eyebrow={message("acc20.eyebrow")}
          subtitle={message("acc20.subtitle", {
            role: message(`role.${preferredRole}`),
          })}
          title={message("acc20.title")}
        />
        <StatusCard
          body={message("acc20.card.body")}
          status={message("acc20.card.status")}
          title={message("acc20.card.title")}
          tone="warning"
        />
        {fallback !== null ? (
          <StatusRow
            status={message("acc19.active")}
            subtitle={message("common.roleContext", {
              role: message(`role.${fallback}`),
              tenant: profile.value?.tenantName ?? "",
            })}
            title={message("acc20.available.title")}
            tone="success"
          />
        ) : null}
        <StatusRow
          status={message("acc20.nav.status")}
          subtitle={message("acc20.nav.subtitle")}
          title={message("acc20.nav.title")}
          tone="info"
        />
        <AccountBanner
          body={message("acc20.banner.body")}
          title={message("acc20.banner.title")}
          tone="warning"
        />
        {fallback !== null ? (
          <BlockAction
            label={message("acc20.goTo", { role: message(`role.${fallback}`) })}
            onPress={() => {
              setPreferredRole(fallback);
              router.replace("/(protected)");
            }}
            testID="role-revoked-fallback"
          />
        ) : null}
        <BlockAction
          kind="secondary"
          label={message("acc20.contactAdmin")}
          onPress={() => router.push("/(protected)/account")}
        />
      </AccountScreenShell>
    );
  }

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
      testID="account-role-switch"
    >
      <ScreenHeading
        eyebrow={message("acc04.eyebrow")}
        subtitle={message("acc04.subtitle")}
        title={message("acc04.title")}
      />
      {workingRole !== null ? (
        <StatusCard
          body={message(`acc04.desc.${workingRole}`)}
          status={profile.value?.tenantName ?? ""}
          title={message("acc04.current.title")}
          tone="success"
        />
      ) : null}
      {ROLE_PRIORITY.filter((role) => roles.includes(role)).map((role) => (
        <StatusRow
          key={role}
          onPress={role === workingRole ? undefined : () => setSelected(role)}
          status={
            role === workingRole
              ? message("acc04.currentRole")
              : message("acc04.available")
          }
          subtitle={message(`acc04.desc.${role}`)}
          testID={`role-option-${role}`}
          title={message(`role.${role}`)}
          tone={role === workingRole ? "success" : "info"}
        />
      ))}
      <AccountBanner
        body={message("acc04.banner.body")}
        title={message("acc04.banner.title")}
        tone="warning"
      />
      {target !== null && workingRole !== null ? (
        <BlockAction
          label={message("acc04.switchTo", { role: message(`role.${target}`) })}
          onPress={() => {
            setPreferredRole(target);
            router.replace("/(protected)");
          }}
          testID="role-switch-confirm"
        />
      ) : null}
      {workingRole !== null ? (
        <BlockAction
          kind="secondary"
          label={message("acc04.stay", { role: message(`role.${workingRole}`) })}
          onPress={() => router.back()}
        />
      ) : null}
    </AccountScreenShell>
  );
}
