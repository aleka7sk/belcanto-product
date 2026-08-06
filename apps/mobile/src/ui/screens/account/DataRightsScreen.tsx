import { router } from "expo-router";
import { useState } from "react";
import { RefreshControl } from "react-native";

import { useApiClient, type DataExportRequest } from "@/api";
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
  type StatusTone,
} from "../../patterns/accountPatterns";
import { semantic } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "./shared";

/**
 * ACC-14 · Запрос копии данных + ACC-15 · Статус (Figma 366:731/366:782).
 * Export needs fresh re-authentication (HOF-12); while DEC-104 is open no
 * readiness deadline is promised — the banner says so explicitly.
 */
export function DataRightsScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const exports = useAccountResource((accessToken) =>
    api.listDataExports(accessToken),
  );
  const [askPassword, setAskPassword] = useState(false);
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const open = exports.value?.find(
    (request) => request.status === "requested" || request.status === "processing",
  );

  const request = async () => {
    setBusy(true);
    setError(null);
    try {
      await runAuthenticated((accessToken) =>
        api.createDataExport(accessToken, { currentPassword: password }),
      );
      setPassword("");
      setAskPassword(false);
      await exports.reload();
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  const statusTone = (status: DataExportRequest["status"]): StatusTone =>
    status === "ready"
      ? "success"
      : status === "requested" || status === "processing"
        ? "info"
        : status === "expired"
          ? "warning"
          : "muted";

  return (
    <AccountScreenShell
      navigation={<AccountNav />}
      refreshControl={
        <RefreshControl
          onRefresh={() => {
            void exports.reload();
          }}
          refreshing={exports.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="account-data-rights"
    >
      <ScreenHeading
        eyebrow={message("acc14.eyebrow")}
        subtitle={message("acc14.subtitle")}
        title={message("acc14.title")}
      />
      <StatusCard
        body={message("acc14.card.body")}
        status={message("acc14.card.status")}
        title={message("acc14.card.title")}
        tone="success"
      />
      <StatusRow
        status={message("acc14.format.status")}
        subtitle={message("acc14.format.subtitle")}
        title={message("acc14.format.title")}
        tone="info"
      />
      <StatusRow
        status={message("acc14.protect.status")}
        subtitle={message("acc14.protect.subtitle")}
        title={message("acc14.protect.title")}
        tone="warning"
      />
      <AccountBanner
        body={message("acc14.banner.body")}
        title={message("acc14.banner.title")}
      />
      {open !== undefined ? (
        <InlineNotice
          body={message("acc14.openExists")}
          title={message(`acc15.status.${open.status}` as const)}
        />
      ) : null}
      {error !== null ? (
        <InlineNotice body={error} title={message("common.retry")} tone="error" />
      ) : null}
      {askPassword && open === undefined ? (
        <>
          <PremiumTextField
            label={message("acc03.password")}
            onChangeText={setPassword}
            secureTextEntry
            testID="export-password"
            value={password}
          />
          <BlockAction
            busy={busy}
            disabled={password.length === 0}
            label={message("acc14.request")}
            onPress={() => void request()}
            testID="export-request-confirm"
          />
        </>
      ) : (
        <BlockAction
          disabled={open !== undefined}
          label={message("acc14.request")}
          onPress={() => setAskPassword(true)}
          testID="export-request"
        />
      )}
      <ScreenHeading
        eyebrow={message("acc15.eyebrow")}
        subtitle={message("acc15.subtitle")}
        title={message("acc15.title")}
      />
      {exports.value !== null && exports.value.length === 0 ? (
        <InlineNotice body={message("acc15.empty")} title={message("acc15.title")} />
      ) : null}
      {(exports.value ?? []).map((entry) => (
        <StatusRow
          key={entry.id}
          status={message(`acc15.status.${entry.status}` as const)}
          subtitle={message("acc15.requestedAt", {
            date: formatBelcantoDate(entry.requestedAt),
          })}
          testID={`export-${entry.id}`}
          title={entry.id}
          tone={statusTone(entry.status)}
        />
      ))}
      <AccountBanner
        body={message("acc15.banner.body")}
        title={message("acc15.banner.title")}
      />
      <BlockAction
        kind="secondary"
        label={message("acc10.deletion.title")}
        onPress={() => router.push("/(protected)/account/data/deletion")}
      />
    </AccountScreenShell>
  );
}
