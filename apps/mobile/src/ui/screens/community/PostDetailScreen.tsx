import { router, useLocalSearchParams } from "expo-router";
import { useMemo, useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { useApiClient } from "@/api";
import type { CommunityReportReason } from "@/api/contracts";
import { COMMUNITY_REPORT_REASONS } from "@/api/contracts";
import { createIntentIdempotency } from "@/controllers";
import { useMessage, type MessageKey } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice, PremiumTextField } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import {
  CommentBubble,
  CommunityPostCard,
  postKickerKey,
} from "../../patterns/communityPatterns";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * Post detail and thread (Figma COM-POST-01, COM-ANN-01). Safety
 * lives here too: report with a mandatory reason (COM-SAFE-02), block
 * with its honest scope statement (COM-SAFE-03), the author's own
 * removal, and the tombstone state for unavailable content
 * (COM-SAFE-05). Reporter identity is never revealed to the author.
 */

const REASON_KEYS: Record<CommunityReportReason, { title: MessageKey; body: MessageKey }> = {
  abuse: { title: "com.reason.abuse", body: "com.reason.abuseBody" },
  personal_data: { title: "com.reason.personal_data", body: "com.reason.personal_dataBody" },
  spam: { title: "com.reason.spam", body: "com.reason.spamBody" },
  other: { title: "com.reason.other", body: "com.reason.otherBody" },
};

type ReportTarget = { targetType: "post" | "comment"; targetId: string };

export function PostDetailScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ postId?: string }>();
  const postId = typeof params.postId === "string" ? params.postId : "";
  const myAccountId = state.bootstrap?.accountId ?? "";

  const post = useAccountResource((accessToken) =>
    api.getCommunityPost(accessToken, postId),
  );
  const blocks = useAccountResource((accessToken) =>
    api.blockedCommunityMembers(accessToken),
  );

  const commentIdempotency = useMemo(() => createIntentIdempotency(), []);
  const actionIdempotency = useMemo(() => createIntentIdempotency(), []);
  const [reply, setReply] = useState("");
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const [reportTarget, setReportTarget] = useState<ReportTarget | null>(null);
  const [reportReason, setReportReason] = useState<CommunityReportReason | null>(null);
  const [reportNote, setReportNote] = useState("");
  const [reportFiled, setReportFiled] = useState(false);

  if (postId === "") {
    return (
      <AccountScreenShell navigation={<AccountNav active="community" />} testID="community-post-guard">
        <InlineNotice
          body={message("com.guard.body")}
          title={message("com.guard.title")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const view = post.value;
  const own = view !== null && view.author.accountId === myAccountId;
  const tombstone = view !== null && view.status !== "published" && view.body === undefined;
  const authorBlocked =
    view !== null &&
    view.author.accountId !== "" &&
    (blocks.value?.blocked ?? []).includes(view.author.accountId);

  const run = async (action: (accessToken: string) => Promise<unknown>) => {
    setActionError(null);
    setBusy(true);
    try {
      await runAuthenticated(action);
      await Promise.all([post.reload(), blocks.reload()]);
      return true;
    } catch (cause) {
      setActionError(apiErrorMessage(cause));
      return false;
    } finally {
      setBusy(false);
    }
  };

  const sendReply = async () => {
    const body = reply.trim();
    if (body === "" || view === null) return;
    const key = commentIdempotency.key();
    const done = await run((accessToken) =>
      api.addCommunityComment(accessToken, view.id, { body }, key),
    );
    if (done) {
      commentIdempotency.complete();
      setReply("");
    }
  };

  const removeOwn = async () => {
    if (view === null) return;
    const done = await run((accessToken) =>
      api.removeCommunityContent(
        accessToken,
        { targetType: "post", targetId: view.id },
        actionIdempotency.key(),
      ),
    );
    if (done) actionIdempotency.complete();
  };

  const fileReport = async () => {
    if (reportTarget === null || reportReason === null) return;
    const note = reportNote.trim();
    const done = await run((accessToken) =>
      api.reportCommunityContent(
        accessToken,
        {
          targetType: reportTarget.targetType,
          targetId: reportTarget.targetId,
          reason: reportReason,
          ...(note === "" ? {} : { note }),
        },
        actionIdempotency.key(),
      ),
    );
    if (done) {
      actionIdempotency.complete();
      setReportFiled(true);
      setReportTarget(null);
      setReportReason(null);
      setReportNote("");
    }
  };

  const toggleBlockAuthor = async () => {
    if (view === null || view.author.accountId === "") return;
    await run((accessToken) =>
      api.blockCommunityMember(accessToken, {
        accountId: view.author.accountId,
        blocked: !authorBlocked,
      }),
    );
  };

  return (
    <AccountScreenShell navigation={<AccountNav active="community" />} testID="community-post">
      {view === null ? (
        post.error !== null ? (
          <InlineNotice
            body={apiErrorMessage(post.error)}
            title={message("common.retry")}
            tone="error"
          />
        ) : (
          <Text style={styles.muted}>{message("common.loading")}</Text>
        )
      ) : tombstone ? (
        <>
          <ScreenHeading
            eyebrow={message("com.tombstone.kicker")}
            subtitle={message("com.tombstone.subtitle")}
            title={message("com.tombstone.title")}
          />
          <StatusCard
            body={message("com.tombstone.keptBody")}
            status={message("com.tombstone.keptFooter")}
            title={message("com.tombstone.keptTitle")}
            tone="info"
          />
          <StatusCard
            body={message("com.tombstone.optionsBody")}
            status={message("com.tombstone.optionsFooter")}
            title={message("com.tombstone.optionsTitle")}
            tone="muted"
          />
          <BlockAction
            label={message("com.tombstone.back")}
            onPress={() => router.back()}
            testID="community-tombstone-back"
          />
        </>
      ) : (
        <>
          <ScreenHeading
            eyebrow={message(postKickerKey(view))}
            subtitle={`${view.author.fullName} · ${formatBelcantoDate(view.createdAt)} · ${message(
              view.audience === "staff" ? "com.audience.staff" : "com.audience.school",
            )}`}
            title={view.title ?? message("com.thread.title")}
          />
          <CommunityPostCard message={message} post={view} testID="community-post-card" />
          {(view.comments ?? []).map((comment) => (
            <CommentBubble
              comment={comment}
              key={comment.id}
              message={message}
              onReport={
                comment.author.accountId !== "" && comment.author.accountId !== myAccountId
                  ? () => {
                      setReportFiled(false);
                      setReportTarget({ targetType: "comment", targetId: comment.id });
                    }
                  : undefined
              }
              own={comment.author.accountId === myAccountId}
              testID={`community-comment-${comment.id}`}
            />
          ))}
          {view.status === "published" && view.commentsEnabled ? (
            <>
              <PremiumTextField
                label={message("com.reply.label")}
                multiline
                onChangeText={setReply}
                placeholder={message("com.reply.placeholder")}
                testID="community-reply-input"
                value={reply}
              />
              <BlockAction
                busy={busy}
                disabled={reply.trim() === ""}
                label={message("com.reply.send")}
                onPress={() => void sendReply()}
                testID="community-reply-send"
              />
            </>
          ) : view.status === "published" ? (
            <StatusCard
              body={message("com.commentsOffBody")}
              status={message("com.commentsOffFooter")}
              title={message("com.commentsOff")}
              tone="muted"
            />
          ) : (
            <StatusCard
              body={message("com.status.moderatorNote")}
              status={message(`com.status.${view.status}`)}
              title={message("com.tombstone.title")}
              tone="warning"
            />
          )}
          {actionError !== null ? (
            <InlineNotice body={actionError} title={message("common.retry")} tone="error" />
          ) : null}
          {reportFiled ? (
            <InlineNotice
              body={message("com.report.filedBody")}
              title={message("com.report.filedTitle")}
              tone="success"
            />
          ) : null}
          {own && view.status === "published" ? (
            <BlockAction
              busy={busy}
              kind="secondary"
              label={message("com.remove.action")}
              onPress={() => void removeOwn()}
              testID="community-remove-own"
            />
          ) : null}
          {!own && view.author.accountId !== "" ? (
            reportTarget === null ? (
              <View style={styles.safetyRow}>
                <BlockAction
                  kind="secondary"
                  label={message("com.report.action")}
                  onPress={() => {
                    setReportFiled(false);
                    setReportTarget({ targetType: "post", targetId: view.id });
                  }}
                  testID="community-report-open"
                />
                <StatusCard
                  body={message("com.block.body")}
                  status={message("com.block.footer")}
                  title={message("com.block.title")}
                  tone="muted"
                />
                <BlockAction
                  busy={busy}
                  kind="secondary"
                  label={message(authorBlocked ? "com.block.undo" : "com.block.action")}
                  onPress={() => void toggleBlockAuthor()}
                  testID="community-block-author"
                />
              </View>
            ) : (
              <>
                <ScreenHeading
                  eyebrow={message("com.report.eyebrow")}
                  subtitle={message("com.report.subtitle")}
                  title={message("com.report.title")}
                />
                {COMMUNITY_REPORT_REASONS.map((reason) => (
                  <StatusRow
                    key={reason}
                    onPress={() => setReportReason(reason)}
                    status={
                      reportReason === reason
                        ? message("com.reason.selected")
                        : message("com.reason.choose")
                    }
                    subtitle={message(REASON_KEYS[reason].body)}
                    testID={`community-reason-${reason}`}
                    title={message(REASON_KEYS[reason].title)}
                    tone={reportReason === reason ? "info" : "muted"}
                  />
                ))}
                {reportReason === "other" ? (
                  <PremiumTextField
                    label={message("com.report.noteLabel")}
                    multiline
                    onChangeText={setReportNote}
                    placeholder={message("com.report.notePlaceholder")}
                    testID="community-report-note"
                    value={reportNote}
                  />
                ) : null}
                <BlockAction
                  busy={busy}
                  disabled={
                    reportReason === null ||
                    (reportReason === "other" && reportNote.trim() === "")
                  }
                  label={
                    reportReason === null
                      ? message("com.report.pickFirst")
                      : message("com.report.send")
                  }
                  onPress={() => void fileReport()}
                  testID="community-report-send"
                />
                <BlockAction
                  kind="secondary"
                  label={message("common.cancel")}
                  onPress={() => {
                    setReportTarget(null);
                    setReportReason(null);
                    setReportNote("");
                  }}
                  testID="community-report-cancel"
                />
              </>
            )
          ) : null}
        </>
      )}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  muted: { color: semantic.textSecondary, ...typeStyles.bodyS },
  safetyRow: { gap: space.s3 },
});
