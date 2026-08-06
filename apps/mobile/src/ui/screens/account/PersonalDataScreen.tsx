import { router } from "expo-router";
import { useEffect, useState } from "react";
import { RefreshControl } from "react-native";

import { useApiClient } from "@/api";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice, PremiumTextField } from "../../components";
import {
  AccountBanner,
  AccountScreenShell,
  BlockAction,
  ProfileHero,
  ScreenHeading,
  StatusRow,
} from "../../patterns/accountPatterns";
import { semantic } from "../../tokens";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav, initialsOf, useAccountResource, useWorkingRole } from "./shared";

/** ACC-02 · Личные данные (Figma 365:136). */
export function PersonalDataScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const workingRole = useWorkingRole();
  const profile = useAccountResource((accessToken) => api.myProfile(accessToken));
  const contacts = useAccountResource((accessToken) =>
    api.listMyContacts(accessToken),
  );
  const [fullName, setFullName] = useState("");
  const [saving, setSaving] = useState(false);
  const [feedback, setFeedback] = useState<{ kind: "ok" | "error"; text: string } | null>(null);

  const profileName = profile.value?.fullName;
  useEffect(() => {
    let active = true;
    queueMicrotask(() => {
      if (active && profileName !== undefined) setFullName(profileName);
    });
    return () => {
      active = false;
    };
  }, [profileName]);

  const email = contacts.value?.find((contact) => contact.kind === "email");
  const dirty = profile.value !== null && fullName.trim() !== profile.value.fullName;

  const save = async () => {
    setSaving(true);
    setFeedback(null);
    try {
      await runAuthenticated((accessToken) =>
        api.updateMyProfile(accessToken, { fullName: fullName.trim() }),
      );
      await profile.reload();
      setFeedback({ kind: "ok", text: message("acc02.saved") });
    } catch (cause) {
      setFeedback({ kind: "error", text: apiErrorMessage(cause) });
    } finally {
      setSaving(false);
    }
  };

  return (
    <AccountScreenShell
      navigation={<AccountNav />}
      refreshControl={
        <RefreshControl
          onRefresh={() => {
            void Promise.all([profile.reload(), contacts.reload()]);
          }}
          refreshing={profile.refreshing || contacts.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="account-personal"
    >
      <ScreenHeading
        eyebrow={message("acc02.eyebrow")}
        subtitle={message("acc02.subtitle")}
        title={message("acc02.title")}
      />
      {profile.value ? (
        <ProfileHero
          context={message("common.roleContext", {
            role: workingRole ? message(`role.${workingRole}`) : "",
            tenant: profile.value.tenantName,
          })}
          initials={initialsOf(profile.value.fullName)}
          name={profile.value.fullName}
          status={message("acc01.verified")}
        />
      ) : null}
      <PremiumTextField
        label={message("acc02.fullName")}
        onChangeText={setFullName}
        testID="personal-full-name"
        value={fullName}
        autoCapitalize="words"
      />
      <StatusRow
        onPress={() =>
          router.push({
            pathname: "/(protected)/account/contact-change",
            params: { kind: "email" },
          })
        }
        status={email ? message("acc03.current.confirmed") : message("acc02.emailMissing")}
        subtitle={email?.value ?? message("acc03.current.none")}
        title={message("acc02.email")}
        tone={email ? "success" : "muted"}
      />
      <StatusRow
        status={profile.value?.phone ? message("acc03.current.confirmed") : undefined}
        subtitle={profile.value?.phone ?? message("acc03.current.none")}
        title={message("acc02.phone")}
        tone="success"
      />
      <AccountBanner
        body={message("acc02.banner.body")}
        title={message("acc02.banner.title")}
      />
      {feedback ? (
        <InlineNotice
          body={feedback.text}
          title={feedback.kind === "ok" ? message("acc02.saved") : message("common.retry")}
          tone={feedback.kind === "ok" ? "success" : "error"}
        />
      ) : null}
      <BlockAction
        busy={saving}
        disabled={!dirty || fullName.trim().length === 0}
        label={message("common.save")}
        onPress={() => void save()}
        testID="personal-save"
      />
    </AccountScreenShell>
  );
}
