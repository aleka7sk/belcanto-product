import { router, useLocalSearchParams } from "expo-router";
import { useMemo, useState } from "react";
import { RefreshControl, StyleSheet, Text } from "react-native";

import { useApiClient } from "@/api";
import { createIntentIdempotency } from "@/controllers";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice, PremiumTextField } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  SettingsRow,
  StatusCard,
} from "../../patterns/accountPatterns";
import {
  CommentBubble,
  CommunityPostCard,
  postKickerKey,
} from "../../patterns/communityPatterns";
import { domainAccent } from "../../domainAccent";
import { semantic, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * COM-POST-01 «Post detail & thread» (Figma 347:151) — the thread and
 * nothing else, exactly as storyboarded. Safety is one tap away, on its
 * own screens: reporting (COM-SAFE-02, /community/report) and blocking
 * (COM-SAFE-03, /community/safety). Tombstones keep the mark and the
 * moment (COM-SAFE-05).
 */

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

  const commentIdempotency = useMemo(() => createIntentIdempotency(), []);
  const removeIdempotency = useMemo(() => createIntentIdempotency(), []);
  const [reply, setReply] = useState("");
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

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

  const run = async (action: (accessToken: string) => Promise<unknown>) => {
    setActionError(null);
    setBusy(true);
    try {
      await runAuthenticated(action);
      await post.reload();
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
        removeIdempotency.key(),
      ),
    );
    if (done) removeIdempotency.complete();
  };

  const openReport = (targetType: "post" | "comment", targetId: string) =>
    router.push({
      pathname: "/(protected)/community/report",
      params: { targetType, targetId },
    });

  return (
    <AccountScreenShell
      keyboardAware
      navigation={<AccountNav active="community" />}
      refreshControl={
        <RefreshControl
          onRefresh={() => void post.reload()}
          refreshing={post.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="community-post"
    >
      {view === null ? (
        post.error !== null ? (
          <ErrorNotice
            actionLabel={message("common.retry")}
            body={apiErrorMessage(post.error)}
            onAction={() => void post.reload()}
            title={message("com.thread.title")}
          />
        ) : (
          <Text style={styles.muted}>{message("common.loading")}</Text>
        )
      ) : tombstone ? (
        <>
          <ScreenHeading
            accent={domainAccent("community")}
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
            accent={domainAccent("community")}
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
                  ? () => openReport("comment", comment.id)
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
            <>
              <SettingsRow
                icon="settings"
                onPress={() => openReport("post", view.id)}
                subtitle={message("com.report.subtitle")}
                testID="community-report-open"
                title={message("com.report.action")}
              />
              <SettingsRow
                icon="users"
                onPress={() =>
                  router.push({
                    pathname: "/(protected)/community/safety",
                    params: { accountId: view.author.accountId, name: view.author.fullName },
                  })
                }
                subtitle={message("com.block.footer")}
                testID="community-safety-open"
                title={message("com.safety.openActions")}
              />
            </>
          ) : null}
        </>
      )}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  muted: { color: semantic.textSecondary, ...typeStyles.bodyS },
});
