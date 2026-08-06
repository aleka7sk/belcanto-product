import { useEffect, useState } from "react";
import { RefreshControl } from "react-native";

import { useApiClient, type PrivacySettings } from "@/api";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice } from "../../components";
import {
  AccountBanner,
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  ToggleRow,
} from "../../patterns/accountPatterns";
import { semantic } from "../../tokens";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav, useAccountResource } from "./shared";

/**
 * ACC-12 · Community privacy (Figma 366:568). Four community toggles over
 * the versioned privacy settings; a stale version yields CONFLICT and the
 * screen reloads the server copy. Student↔Student direct messages stay
 * structurally unavailable (DEC-011) — the row is locked with the
 * design's explanation.
 */
export function CommunityPrivacyScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const settings = useAccountResource((accessToken) =>
    api.privacySettings(accessToken),
  );
  const [draft, setDraft] = useState<PrivacySettings | null>(null);
  const [busy, setBusy] = useState(false);
  const [feedback, setFeedback] = useState<{ kind: "ok" | "error"; text: string } | null>(null);

  const loaded = settings.value;
  useEffect(() => {
    let active = true;
    queueMicrotask(() => {
      if (active && loaded !== null) setDraft(loaded);
    });
    return () => {
      active = false;
    };
  }, [loaded]);

  const dirty =
    draft !== null &&
    settings.value !== null &&
    (draft.communityProfileVisible !== settings.value.communityProfileVisible ||
      draft.achievementsVisible !== settings.value.achievementsVisible ||
      draft.mentionsAllowed !== settings.value.mentionsAllowed);

  const save = async () => {
    if (draft === null) return;
    setBusy(true);
    setFeedback(null);
    try {
      await runAuthenticated((accessToken) =>
        api.updatePrivacySettings(accessToken, draft),
      );
      await settings.reload();
      setFeedback({ kind: "ok", text: message("acc12.saved") });
    } catch (cause) {
      const conflict =
        cause instanceof Error && "code" in cause && cause.code === "CONFLICT";
      setFeedback({
        kind: "error",
        text: conflict ? message("acc12.conflict") : apiErrorMessage(cause),
      });
      await settings.reload();
    } finally {
      setBusy(false);
    }
  };

  return (
    <AccountScreenShell
      navigation={<AccountNav />}
      refreshControl={
        <RefreshControl
          onRefresh={() => {
            void settings.reload();
          }}
          refreshing={settings.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="account-community-privacy"
    >
      <ScreenHeading
        eyebrow={message("acc12.eyebrow")}
        subtitle={message("acc12.subtitle")}
        title={message("acc12.title")}
      />
      {draft !== null ? (
        <>
          <ToggleRow
            offLabel={message("acc12.profile.off")}
            onLabel={message("acc12.profile.on")}
            onToggle={(next) =>
              setDraft({ ...draft, communityProfileVisible: next })
            }
            subtitle={message("acc12.profile.subtitle")}
            testID="privacy-profile-toggle"
            title={message("acc12.profile.title")}
            value={draft.communityProfileVisible}
          />
          <ToggleRow
            offLabel={message("common.disabled")}
            onLabel={message("common.enabled")}
            onToggle={(next) => setDraft({ ...draft, achievementsVisible: next })}
            subtitle={message("acc12.achievements.subtitle")}
            testID="privacy-achievements-toggle"
            title={message("acc12.achievements.title")}
            value={draft.achievementsVisible}
          />
          <ToggleRow
            lockedStatus={message("acc12.dm.locked")}
            offLabel={message("common.disabled")}
            onLabel={message("common.enabled")}
            subtitle={message("acc12.dm.subtitle")}
            testID="privacy-dm-locked"
            title={message("acc12.dm.title")}
            value={draft.staffMessagesAllowed}
          />
          <ToggleRow
            offLabel={message("common.disabled")}
            onLabel={message("common.enabled")}
            onToggle={(next) => setDraft({ ...draft, mentionsAllowed: next })}
            subtitle={message("acc12.mentions.subtitle")}
            testID="privacy-mentions-toggle"
            title={message("acc12.mentions.title")}
            value={draft.mentionsAllowed}
          />
        </>
      ) : null}
      <AccountBanner
        body={message("acc12.banner.body")}
        title={message("acc12.banner.title")}
      />
      {feedback !== null ? (
        <InlineNotice
          body={feedback.text}
          title={feedback.kind === "ok" ? message("acc12.saved") : message("common.retry")}
          tone={feedback.kind === "ok" ? "success" : "error"}
        />
      ) : null}
      <BlockAction
        busy={busy}
        disabled={!dirty}
        label={message("acc12.save")}
        onPress={() => void save()}
        testID="privacy-save"
      />
    </AccountScreenShell>
  );
}
