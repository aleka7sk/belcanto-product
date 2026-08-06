import { router } from "expo-router";
import { useState } from "react";
import { RefreshControl } from "react-native";

import { ApiError, useApiClient, type DeletionRequest } from "@/api";
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
import { semantic } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "./shared";

/**
 * ACC-16 · Удаление аккаунта + ACC-17 · Запрос принят (Figma
 * 366:833/366:888). DEC-104 guard: no deletion date exists while the
 * retention decision is open, so the accepted state speaks only of the
 * reviewable request and cancellation stays available at any moment.
 */
export function DeletionScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const open = useAccountResource<DeletionRequest | null>(async (accessToken) => {
    try {
      return await api.deletionRequest(accessToken);
    } catch (cause) {
      if (cause instanceof ApiError && cause.code === "NOT_FOUND") return null;
      throw cause;
    }
  });
  const [askPassword, setAskPassword] = useState(false);
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [feedback, setFeedback] = useState<{ kind: "ok" | "error"; text: string } | null>(null);

  const create = async () => {
    setBusy(true);
    setFeedback(null);
    try {
      await runAuthenticated((accessToken) =>
        api.createDeletionRequest(accessToken, { currentPassword: password }),
      );
      setPassword("");
      setAskPassword(false);
      await open.reload();
    } catch (cause) {
      setFeedback({ kind: "error", text: apiErrorMessage(cause) });
    } finally {
      setBusy(false);
    }
  };

  const cancel = async () => {
    setBusy(true);
    setFeedback(null);
    try {
      await runAuthenticated((accessToken) =>
        api.cancelDeletionRequest(accessToken),
      );
      await open.reload();
      setFeedback({ kind: "ok", text: message("acc17.cancelled") });
    } catch (cause) {
      setFeedback({ kind: "error", text: apiErrorMessage(cause) });
    } finally {
      setBusy(false);
    }
  };

  if (open.value) {
    const request = open.value;
    return (
      <AccountScreenShell navigation={<AccountNav />} testID="account-deletion-status">
        <ScreenHeading
          eyebrow={message("acc17.eyebrow")}
          subtitle={message("acc17.subtitle")}
          title={message("acc17.title")}
        />
        <StatusCard
          body={message("acc17.card.body", {
            date: formatBelcantoDate(request.requestedAt),
            id: request.id,
          })}
          status={message("acc17.card.status")}
          title={message("acc17.card.title")}
          tone="warning"
        />
        <StatusRow
          status={message("acc17.review.status")}
          subtitle={message("acc17.review.subtitle")}
          title={message("acc17.review.title")}
          tone="info"
        />
        <StatusRow
          onPress={() => router.push("/(protected)/account/data")}
          status={message("acc17.export.status")}
          subtitle={message("acc17.export.subtitle")}
          title={message("acc17.export.title")}
          tone="info"
        />
        <AccountBanner
          body={message("acc17.banner.body")}
          title={message("acc17.banner.title")}
        />
        {feedback !== null ? (
          <InlineNotice
            body={feedback.text}
            title={feedback.kind === "ok" ? message("acc17.cancelled") : message("common.retry")}
            tone={feedback.kind === "ok" ? "success" : "error"}
          />
        ) : null}
        <BlockAction
          busy={busy}
          label={message("acc17.cancel")}
          onPress={() => void cancel()}
          testID="deletion-cancel"
        />
        <BlockAction
          kind="secondary"
          label={message("auth11.contact")}
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
            void open.reload();
          }}
          refreshing={open.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="account-deletion"
    >
      <ScreenHeading
        eyebrow={message("acc16.eyebrow")}
        subtitle={message("acc16.subtitle")}
        title={message("acc16.title")}
      />
      <StatusCard
        body={message("acc16.card.body")}
        status={message("acc16.card.status")}
        title={message("acc16.card.title")}
        tone="warning"
      />
      <StatusRow
        status={message("acc16.records.status")}
        subtitle={message("acc16.records.subtitle")}
        title={message("acc16.records.title")}
        tone="info"
      />
      <StatusRow
        status={message("acc16.community.status")}
        subtitle={message("acc16.community.subtitle")}
        title={message("acc16.community.title")}
        tone="info"
      />
      <StatusRow
        status={message("acc16.active.status")}
        subtitle={message("acc16.active.subtitle")}
        title={message("acc16.active.title")}
        tone="warning"
      />
      <AccountBanner
        body={message("acc16.banner.body")}
        title={message("acc16.banner.title")}
        tone="danger"
      />
      {feedback !== null ? (
        <InlineNotice body={feedback.text} title={message("common.retry")} tone="error" />
      ) : null}
      {askPassword ? (
        <>
          <PremiumTextField
            label={message("acc03.password")}
            onChangeText={setPassword}
            secureTextEntry
            testID="deletion-password"
            value={password}
          />
          <BlockAction
            busy={busy}
            disabled={password.length === 0}
            label={message("acc16.continue")}
            onPress={() => void create()}
            testID="deletion-confirm"
          />
        </>
      ) : (
        <BlockAction
          label={message("acc16.continue")}
          onPress={() => setAskPassword(true)}
          testID="deletion-continue"
        />
      )}
      <BlockAction
        kind="secondary"
        label={message("acc16.keep")}
        onPress={() => router.back()}
      />
    </AccountScreenShell>
  );
}
