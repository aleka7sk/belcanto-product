import { router, useLocalSearchParams } from "expo-router";
import { useMemo, useState } from "react";
import { RefreshControl, StyleSheet, Text } from "react-native";

import { useApiClient } from "@/api";
import type { Assessment } from "@/api/contracts";
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
  type StatusTone,
} from "../../patterns/accountPatterns";
import { semantic, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * Assessment detail (TCH-REVIEW-06/07 + the Student's read view).
 * Versions are immutable snapshots: the superseded original links its
 * replacement, withdrawal keeps the record with its reason. The author
 * corrects through supersede and publishes drafts from here.
 */

export function assessmentStatusTone(status: Assessment["status"]): StatusTone {
  switch (status) {
    case "published":
      return "success";
    case "draft":
      return "warning";
    case "superseded":
      return "muted";
    case "withdrawn":
      return "danger";
  }
}

export function assessmentStatusLabel(
  status: Assessment["status"],
  message: MessageFormatter,
): string {
  return message(`asmt.status.${status}`);
}

export function AssessmentDetailScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ assessmentId?: string }>();
  const assessmentId = typeof params.assessmentId === "string" ? params.assessmentId : "";
  const myAccountId = state.bootstrap?.accountId ?? "";

  const assessment = useAccountResource((accessToken) =>
    api.getAssessment(accessToken, assessmentId),
  );
  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [withdrawReason, setWithdrawReason] = useState("");
  const [withdrawOpen, setWithdrawOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  if (assessmentId === "") {
    return (
      <AccountScreenShell navigation={<AccountNav active="review" />} testID="assessment-guard">
        <InlineNotice
          body={message("asmt.compose.guardBody")}
          title={message("asmt.compose.guardTitle")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const view = assessment.value;
  const own = view !== null && view.author.accountId === myAccountId;

  const publish = async () => {
    if (view === null) return;
    setActionError(null);
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.publishAssessment(
          accessToken,
          view.id,
          { expectedVersion: view.version },
          idempotency.key(),
        ),
      );
      idempotency.complete();
      await assessment.reload();
    } catch (cause) {
      setActionError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  const withdraw = async () => {
    if (view === null || withdrawReason.trim() === "") return;
    setActionError(null);
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.withdrawAssessment(
          accessToken,
          view.id,
          { reason: withdrawReason.trim() },
          idempotency.key(),
        ),
      );
      idempotency.complete();
      setWithdrawOpen(false);
      setWithdrawReason("");
      await assessment.reload();
    } catch (cause) {
      setActionError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AccountScreenShell
      navigation={<AccountNav active="review" />}
      refreshControl={
        <RefreshControl
          onRefresh={() => void assessment.reload()}
          refreshing={assessment.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="assessment-detail"
    >
      {view === null ? (
        assessment.error !== null ? (
          <ErrorNotice
            actionLabel={message("common.retry")}
            body={apiErrorMessage(assessment.error)}
            onAction={() => void assessment.reload()}
            title={message("asmt.compose.eyebrow")}
          />
        ) : (
          <Text style={styles.muted}>{message("common.loading")}</Text>
        )
      ) : (
        <>
          <ScreenHeading
            eyebrow={message(`asmt.context.${view.contextType}`)}
            subtitle={`${view.author.fullName} · ${view.assessmentDate} · ${message(
              `asmt.visibility.${view.visibility}`,
            )}`}
            title={message(`asmt.type.${view.type === "self" ? "observation" : view.type}`)}
          />
          <StatusCard
            body={view.summary ?? message("asmt.detail.noSummary")}
            status={`${assessmentStatusLabel(view.status, message)}${
              view.confidence !== undefined
                ? ` · ${message("asmt.detail.confidence")}: ${message(`asmt.confidence.${view.confidence}`)}`
                : ""
            }`}
            title={message("asmt.detail.summary")}
            tone={assessmentStatusTone(view.status)}
          />
          {view.strengths !== undefined ? (
            <StatusCard
              body={view.strengths}
              title={message("asmt.detail.strengths")}
              tone="success"
            />
          ) : null}
          {view.developmentAreas !== undefined ? (
            <StatusCard
              body={view.developmentAreas}
              title={message("asmt.detail.developmentAreas")}
              tone="info"
            />
          ) : null}
          {view.recommendations !== undefined ? (
            <StatusCard
              body={view.recommendations}
              title={message("asmt.detail.recommendations")}
              tone="info"
            />
          ) : null}
          {view.areas !== undefined ? (
            <Text style={styles.areas}>
              {message("asmt.detail.areas")}: {view.areas}
            </Text>
          ) : null}
          {view.evidence.map((entry) => (
            <StatusRow
              key={entry.id}
              status={formatBelcantoDate(entry.addedAt)}
              subtitle={entry.note}
              testID={`evidence-${entry.id}`}
              title={message(`asmt.evidence.${entry.kind}`)}
              tone="muted"
            />
          ))}
          {view.status === "withdrawn" && view.withdrawalReason !== undefined ? (
            <StatusCard
              body={view.withdrawalReason}
              status={message("asmt.detail.withdrawnFooter")}
              title={message("asmt.status.withdrawn")}
              tone="danger"
            />
          ) : null}
          {view.status === "superseded" && view.supersededById !== undefined ? (
            <StatusRow
              onPress={() =>
                router.push({
                  pathname: "/(protected)/assessment/[assessmentId]",
                  params: { assessmentId: view.supersededById! },
                })
              }
              status={message("asmt.detail.openReplacement")}
              subtitle={message("asmt.detail.supersededBody")}
              testID="assessment-replacement"
              title={message("asmt.status.superseded")}
              tone="warning"
            />
          ) : null}
          {actionError !== null ? (
            <InlineNotice body={actionError} title={message("common.retry")} tone="error" />
          ) : null}
          {own && view.status === "draft" ? (
            <BlockAction
              busy={busy}
              label={message("asmt.detail.publish")}
              onPress={() => void publish()}
              testID="assessment-publish"
            />
          ) : null}
          {own && view.status === "published" ? (
            <BlockAction
              kind="secondary"
              label={message("asmt.detail.supersede")}
              onPress={() =>
                router.push({
                  pathname: "/(protected)/teacher/assessment/create",
                  params: { supersedeId: view.id },
                })
              }
              testID="assessment-supersede"
            />
          ) : null}
          {own && (view.status === "draft" || view.status === "published") ? (
            withdrawOpen ? (
              <>
                <PremiumTextField
                  label={message("asmt.detail.withdrawReason")}
                  multiline
                  onChangeText={setWithdrawReason}
                  placeholder={message("asmt.detail.withdrawPlaceholder")}
                  testID="assessment-withdraw-reason"
                  value={withdrawReason}
                />
                <BlockAction
                  busy={busy}
                  disabled={withdrawReason.trim() === ""}
                  kind="secondary"
                  label={message("asmt.detail.withdrawConfirm")}
                  onPress={() => void withdraw()}
                  testID="assessment-withdraw-confirm"
                />
              </>
            ) : (
              <BlockAction
                kind="secondary"
                label={message("asmt.detail.withdraw")}
                onPress={() => setWithdrawOpen(true)}
                testID="assessment-withdraw"
              />
            )
          ) : null}
        </>
      )}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  muted: { color: semantic.textSecondary, ...typeStyles.bodyS },
  areas: {
    color: semantic.textGold,
    ...typeStyles.labelM,
  },
});
