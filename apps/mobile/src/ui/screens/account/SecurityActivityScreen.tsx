import { router } from "expo-router";
import { useState } from "react";

import { useApiClient, type SecurityEvent } from "@/api";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice } from "../../components";
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

/**
 * ACC-09 · Новый вход (Figma 365:568) on top of the security feed. The
 * newest sign-in event is reviewed inline («Это были вы?»); the rest of
 * the feed lists below with cursor pagination.
 */
export function SecurityActivityScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const feed = useAccountResource((accessToken) =>
    api.listSecurityEvents(accessToken, {}),
  );
  const [extra, setExtra] = useState<SecurityEvent[]>([]);
  const [cursorBusy, setCursorBusy] = useState(false);
  const [reviewed, setReviewed] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const events = [...(feed.value?.events ?? []), ...extra];
  const newestSignIn =
    !reviewed && events.length > 0 && events[0]?.action === "SessionCreated"
      ? events[0]
      : null;
  const nextCursor =
    extra.length > 0 ? undefined : feed.value?.nextCursor;

  const loadMore = async () => {
    const cursor = feed.value?.nextCursor;
    if (cursor === undefined) return;
    setCursorBusy(true);
    try {
      const page = await runAuthenticated((accessToken) =>
        api.listSecurityEvents(accessToken, { cursor }),
      );
      setExtra((current) => [...current, ...page.events]);
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setCursorBusy(false);
    }
  };

  const eventTone = (event: SecurityEvent) =>
    event.decision === "deny"
      ? "danger"
      : event.action === "RefreshTokenReuseDetected"
        ? "warning"
        : "info";

  return (
    <AccountScreenShell navigation={<AccountNav />} testID="account-activity">
      <ScreenHeading
        eyebrow={message("acc09.eyebrow")}
        subtitle={newestSignIn ? message("acc09.subtitle") : message("acc09.listSubtitle")}
        title={newestSignIn ? message("acc09.title") : message("acc09.listTitle")}
      />
      {error !== null || feed.error !== null ? (
        <InlineNotice
          body={error ?? apiErrorMessage(feed.error)}
          title={message("common.retry")}
          tone="error"
        />
      ) : null}
      {newestSignIn !== null ? (
        <>
          <StatusCard
            body={formatBelcantoDate(newestSignIn.recordedAt)}
            status={message(
              newestSignIn.decision === "allow"
                ? "acc09.decision.allow"
                : "acc09.decision.deny",
            )}
            title={message("acc09.event.card.title")}
            tone="warning"
          />
          <AccountBanner
            body={message("acc09.banner.body")}
            title={message("acc09.banner.title")}
            tone="danger"
          />
          <BlockAction
            label={message("acc09.confirm")}
            onPress={() => setReviewed(true)}
            testID="activity-confirm"
          />
          <BlockAction
            kind="secondary"
            label={message("acc09.protect")}
            onPress={() => {
              setReviewed(true);
              router.push("/(protected)/account/security/devices");
            }}
            testID="activity-protect"
          />
        </>
      ) : null}
      {events.length === 0 && !feed.loading ? (
        <InlineNotice body={message("acc09.empty")} title={message("acc09.listTitle")} />
      ) : null}
      {events.map((event) => (
        <StatusRow
          key={event.id}
          status={message(
            event.decision === "allow" ? "acc09.decision.allow" : "acc09.decision.deny",
          )}
          subtitle={formatBelcantoDate(event.recordedAt)}
          testID={`activity-event-${event.id}`}
          title={message(`action.${event.action}`)}
          tone={eventTone(event)}
        />
      ))}
      {nextCursor !== undefined ? (
        <BlockAction
          busy={cursorBusy}
          kind="secondary"
          label={message("acc09.loadMore")}
          onPress={() => void loadMore()}
          testID="activity-load-more"
        />
      ) : null}
    </AccountScreenShell>
  );
}
