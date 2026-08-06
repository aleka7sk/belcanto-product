import { useMemo, useState } from "react";
import { RefreshControl, StyleSheet, Text } from "react-native";

import { useApiClient } from "@/api";
import type { CommunityReport } from "@/api/contracts";
import { createIntentIdempotency } from "@/controllers";
import { useMessage, type MessageFormatter } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice, PremiumTextField } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import { semantic, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource, useWorkingRole } from "../account/shared";

/**
 * Moderation queue and review (Figma COM-MOD-01/02). Owner and
 * Administrator decide reports: hide the target with a reason or keep
 * it — either way the report becomes reviewed with the decision, the
 * reason, the moderator and the moment recorded forever. Moderator as
 * a delegated capability is a recorded future extension.
 */

export function reportReasonTitle(
  report: Pick<CommunityReport, "reason">,
  message: MessageFormatter,
): string {
  return message(`com.reason.${report.reason}`);
}

export function ModerationScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const workingRole = useWorkingRole();
  const moderator = workingRole === "Owner" || workingRole === "Administrator";

  const queue = useAccountResource((accessToken) => api.moderationQueue(accessToken));
  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [decisionReason, setDecisionReason] = useState("");
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  if (!moderator) {
    return (
      <AccountScreenShell navigation={<AccountNav active="community" />} testID="moderation-guard">
        <InlineNotice
          body={message("com.moderation.guardBody")}
          title={message("com.moderation.guardTitle")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const reports = queue.value ?? [];
  const selected = reports.find((report) => report.id === selectedId) ?? null;

  const decide = async (decision: "hidden" | "kept") => {
    if (selected === null) return;
    const reason = decisionReason.trim();
    if (reason === "") return;
    setActionError(null);
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.decideCommunityReport(
          accessToken,
          selected.id,
          { decision, decisionReason: reason },
          idempotency.key(),
        ),
      );
      idempotency.complete();
      setSelectedId(null);
      setDecisionReason("");
      await queue.reload();
    } catch (cause) {
      setActionError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AccountScreenShell
      navigation={<AccountNav active="community" />}
      refreshControl={
        <RefreshControl
          onRefresh={() => void queue.reload()}
          refreshing={queue.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="moderation-queue"
    >
      <ScreenHeading
        eyebrow={message("com.moderation.eyebrow")}
        subtitle={message("com.moderation.subtitle")}
        title={message("com.moderation.title")}
      />
      {queue.error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={apiErrorMessage(queue.error)}
          onAction={() => void queue.reload()}
          title={message("com.moderation.title")}
        />
      ) : null}
      {queue.value !== null && reports.length === 0 ? (
        <Text style={styles.muted}>{message("com.moderation.empty")}</Text>
      ) : null}
      {selected === null ? (
        reports.map((report) => (
          <StatusRow
            key={report.id}
            onPress={
              report.status === "new"
                ? () => {
                    setSelectedId(report.id);
                    setDecisionReason("");
                  }
                : undefined
            }
            status={
              report.status === "new"
                ? `${message("com.moderation.new")} · ${formatBelcantoDate(report.createdAt)}`
                : `${message("com.moderation.reviewed")} · ${message(
                    report.decision === "hidden"
                      ? "com.moderation.hidden"
                      : "com.moderation.kept",
                  )}`
            }
            subtitle={report.targetExcerpt ?? message("com.moderation.noExcerpt")}
            testID={`moderation-report-${report.id}`}
            title={reportReasonTitle(report, message)}
            tone={report.status === "new" ? "warning" : "muted"}
          />
        ))
      ) : (
        <>
          <StatusCard
            body={selected.targetExcerpt ?? message("com.moderation.noExcerpt")}
            status={`${reportReasonTitle(selected, message)}${
              selected.note !== undefined ? ` · ${selected.note}` : ""
            }`}
            title={message("com.moderation.reviewTitle")}
            tone="warning"
          />
          <PremiumTextField
            label={message("com.moderation.reasonLabel")}
            multiline
            onChangeText={setDecisionReason}
            placeholder={message("com.moderation.reasonPlaceholder")}
            testID="moderation-reason-input"
            value={decisionReason}
          />
          {actionError !== null ? (
            <InlineNotice body={actionError} title={message("common.retry")} tone="error" />
          ) : null}
          <BlockAction
            busy={busy}
            disabled={decisionReason.trim() === ""}
            label={message("com.moderation.hide")}
            onPress={() => void decide("hidden")}
            testID="moderation-hide"
          />
          <BlockAction
            busy={busy}
            disabled={decisionReason.trim() === ""}
            kind="secondary"
            label={message("com.moderation.keep")}
            onPress={() => void decide("kept")}
            testID="moderation-keep"
          />
          <BlockAction
            kind="secondary"
            label={message("common.cancel")}
            onPress={() => {
              setSelectedId(null);
              setDecisionReason("");
            }}
            testID="moderation-cancel"
          />
        </>
      )}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  muted: { color: semantic.textSecondary, ...typeStyles.bodyS },
});
