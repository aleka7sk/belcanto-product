import { router, useLocalSearchParams } from "expo-router";
import { useState } from "react";

import { useApiClient } from "@/api";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
} from "../../patterns/accountPatterns";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * COM-SAFE-03 «Ограничить взаимодействие» (Figma 348:757) — its own
 * screen. Blocking is reversible and independent from reporting; the
 * scope statement is verbatim and honest (moderated shared spaces may
 * keep the member's messages visible). Chat muting arrives with the
 * chats slice — the screen ships what exists.
 */

export function SafetyScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ accountId?: string; name?: string }>();
  const accountId = typeof params.accountId === "string" ? params.accountId : "";
  const memberName = typeof params.name === "string" ? params.name : "";

  const blocks = useAccountResource((accessToken) =>
    api.blockedCommunityMembers(accessToken),
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const blocked = (blocks.value?.blocked ?? []).includes(accountId);

  const toggle = async () => {
    setError(null);
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.blockCommunityMember(accessToken, { accountId, blocked: !blocked }),
      );
      await blocks.reload();
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  if (accountId === "") {
    return (
      <AccountScreenShell navigation={<AccountNav active="community" />} testID="community-safety-guard">
        <InlineNotice
          body={message("com.guard.body")}
          title={message("com.guard.title")}
          tone="error"
        />
        <BlockAction
          kind="secondary"
          label={message("com.tombstone.back")}
          onPress={() => router.back()}
        />
      </AccountScreenShell>
    );
  }

  return (
    <AccountScreenShell navigation={<AccountNav active="community" />} testID="community-safety">
      <ScreenHeading
        eyebrow={message("com.safety.eyebrow")}
        subtitle={message("com.safety.subtitle")}
        title={message("com.safety.title")}
      />
      <StatusCard
        body={message("com.block.body")}
        status={
          blocked
            ? message("com.safety.blockedStatus")
            : message("com.block.footer")
        }
        title={
          memberName === ""
            ? message("com.block.title")
            : `${message("com.safety.memberContext")} · ${memberName}`
        }
        tone={blocked ? "warning" : "muted"}
      />
      {blocks.error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={apiErrorMessage(blocks.error)}
          onAction={() => void blocks.reload()}
          title={message("com.safety.title")}
        />
      ) : null}
      {error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={error}
          onAction={() => void toggle()}
          title={message("com.safety.title")}
        />
      ) : null}
      <BlockAction
        busy={busy}
        disabled={blocks.value === null}
        label={message(blocked ? "com.block.undo" : "com.block.action")}
        onPress={() => void toggle()}
        testID="community-safety-toggle"
      />
      <BlockAction
        kind="secondary"
        label={message("com.tombstone.back")}
        onPress={() => router.back()}
        testID="community-safety-back"
      />
    </AccountScreenShell>
  );
}
