import { router } from "expo-router";
import { useState } from "react";

import { useApiClient } from "@/api";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice, PremiumTextField } from "../../components";
import {
  AccountBanner,
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "./shared";

/** ACC-08 · Устройства и сеансы (Figma 365:517). */
export function DevicesScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const sessions = useAccountResource((accessToken) =>
    api.listMySessions(accessToken),
  );
  const [password, setPassword] = useState("");
  const [askPassword, setAskPassword] = useState(false);
  const [busy, setBusy] = useState(false);
  const [feedback, setFeedback] = useState<{ kind: "ok" | "error"; text: string } | null>(null);

  const current = sessions.value?.find((device) => device.current) ?? null;
  const others = sessions.value?.filter((device) => !device.current) ?? [];

  const revokeOthers = async () => {
    setBusy(true);
    setFeedback(null);
    try {
      const result = await runAuthenticated((accessToken) =>
        api.revokeOtherSessions(accessToken, { currentPassword: password }),
      );
      setPassword("");
      setAskPassword(false);
      setFeedback({
        kind: "ok",
        text: message("acc08.revoked", { count: result.revokedCount }),
      });
      await sessions.reload();
    } catch (cause) {
      setFeedback({ kind: "error", text: apiErrorMessage(cause) });
    } finally {
      setBusy(false);
    }
  };

  return (
    <AccountScreenShell navigation={<AccountNav />} testID="account-devices">
      <ScreenHeading
        eyebrow={message("acc08.eyebrow")}
        subtitle={message("acc08.subtitle")}
        title={message("acc08.title")}
      />
      {current !== null ? (
        <StatusCard
          body={[current.deviceLabel ?? message("acc08.unknownDevice"), current.platform]
            .filter(Boolean)
            .join(" · ")}
          status={message("acc08.current.status")}
          title={message("acc08.current.title")}
          tone="success"
        />
      ) : null}
      {others.map((device) => (
        <StatusRow
          key={device.sessionId}
          status={formatBelcantoDate(device.lastSeenAt ?? device.createdAt)}
          subtitle={device.platform ?? ""}
          testID={`device-${device.sessionId}`}
          title={device.deviceLabel ?? message("acc08.unknownDevice")}
          tone="info"
        />
      ))}
      <AccountBanner
        body={message("acc08.banner.body")}
        title={message("acc08.banner.title")}
        tone="warning"
      />
      {feedback !== null ? (
        <InlineNotice
          body={feedback.text}
          title={feedback.kind === "ok" ? message("acc08.revokeOthers") : message("common.retry")}
          tone={feedback.kind === "ok" ? "success" : "error"}
        />
      ) : null}
      {askPassword ? (
        <>
          <InlineNotice
            body={message("acc08.passwordPrompt")}
            title={message("acc08.revokeOthers")}
          />
          <PremiumTextField
            label={message("acc03.password")}
            onChangeText={setPassword}
            secureTextEntry
            testID="devices-password"
            value={password}
          />
          <BlockAction
            busy={busy}
            disabled={password.length === 0}
            label={message("acc08.revokeOthers")}
            onPress={() => void revokeOthers()}
            testID="devices-revoke-confirm"
          />
        </>
      ) : (
        <BlockAction
          disabled={others.length === 0}
          label={message("acc08.revokeOthers")}
          onPress={() => setAskPassword(true)}
          testID="devices-revoke"
        />
      )}
      <BlockAction
        kind="secondary"
        label={message("acc08.history")}
        onPress={() => router.push("/(protected)/account/security/activity")}
      />
    </AccountScreenShell>
  );
}
