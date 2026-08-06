import { router, useLocalSearchParams } from "expo-router";
import { useMemo, useState } from "react";

import {
  useApiClient,
  type IsoDateTime,
  type Lesson,
  type RescheduleRequest,
} from "@/api";
import { createIntentIdempotency } from "@/controllers";
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
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * STU-SCHEDULE-12..15 · Запрос переноса/отмены (Figma 378:803..958).
 * The schedule never changes before the school decides; the proposed
 * time is a preference, not a booking. Candidate windows derive from
 * the lesson's own time — the administrator confirms the final slot.
 */
export function RescheduleRequestScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ occurrenceId?: string }>();
  const occurrenceId = typeof params.occurrenceId === "string" ? params.occurrenceId : "";
  const lesson = useAccountResource((accessToken) =>
    api.getLesson(accessToken, occurrenceId),
  );
  const requests = useAccountResource((accessToken) =>
    api.listRescheduleRequests(accessToken),
  );
  const idempotency = useMemo(() => createIntentIdempotency(), []);

  const [kind, setKind] = useState<"reschedule" | "cancellation">("reschedule");
  const [reason, setReason] = useState("");
  const [selectedOption, setSelectedOption] = useState(0);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sent, setSent] = useState<RescheduleRequest | null>(null);

  const options = useMemo(() => {
    if (lesson.value === null) return [];
    const base = new Date(lesson.value.startsAt).getTime();
    const day = 24 * 3600 * 1000;
    return [base + day, base + day + 3600 * 1000, base + 3 * day].map(
      (timestamp) => new Date(timestamp),
    );
  }, [lesson.value]);

  const myRequest =
    sent ??
    (requests.value ?? []).find(
      (request) => request.occurrenceId === occurrenceId && request.status === "pending",
    ) ??
    null;

  const submit = async () => {
    setBusy(true);
    setError(null);
    try {
      const proposed = options[selectedOption];
      const created = await runAuthenticated((accessToken) =>
        api.createRescheduleRequest(
          accessToken,
          {
            occurrenceId,
            kind,
            reason: reason.trim(),
            ...(kind === "reschedule" && proposed !== undefined
              ? { proposedStartsAt: proposed.toISOString() as IsoDateTime }
              : {}),
          },
          idempotency.key(),
        ),
      );
      idempotency.complete();
      setSent(created);
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  const withdraw = async (requestId: string) => {
    setBusy(true);
    setError(null);
    try {
      await runAuthenticated((accessToken) =>
        api.withdrawRescheduleRequest(accessToken, requestId),
      );
      setSent(null);
      await requests.reload();
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  const view: Lesson | null = lesson.value;

  if (myRequest !== null) {
    return (
      <AccountScreenShell navigation={<AccountNav active="schedule" />} testID="reschedule-sent">
        <ScreenHeading
          eyebrow={message("resched.sent.eyebrow")}
          subtitle={message("resched.sent.subtitle")}
          title={message("resched.sent.title")}
        />
        <StatusCard
          body={message("resched.sent.card.body", {
            id: myRequest.id,
            date: formatBelcantoDate(myRequest.createdAt),
          })}
          status={message(`resched.status.${myRequest.status}` as const)}
          title={message("resched.sent.card.title")}
          tone="info"
        />
        {view !== null ? (
          <StatusRow
            status={message("resched.status.pending")}
            subtitle={formatBelcantoDate(view.startsAt)}
            title={view.title}
            tone="muted"
          />
        ) : null}
        {myRequest.proposedStartsAt !== undefined ? (
          <StatusRow
            status={message("resched.preference.selected")}
            subtitle={formatBelcantoDate(myRequest.proposedStartsAt)}
            title={message("resched.preference.title")}
            tone="info"
          />
        ) : null}
        {error !== null ? (
          <InlineNotice body={error} title={message("common.retry")} tone="error" />
        ) : null}
        <BlockAction
          label={message("resched.openSchedule")}
          onPress={() => router.replace("/(protected)/schedule")}
        />
        <BlockAction
          busy={busy}
          kind="secondary"
          label={message("resched.withdraw")}
          onPress={() => void withdraw(myRequest.id)}
          testID="reschedule-withdraw"
        />
      </AccountScreenShell>
    );
  }

  return (
    <AccountScreenShell navigation={<AccountNav active="schedule" />} testID="reschedule-request">
      <ScreenHeading
        eyebrow={message("resched.eyebrow")}
        subtitle={message("resched.subtitle")}
        title={message("resched.title")}
      />
      {view !== null ? (
        <StatusCard
          body={formatBelcantoDate(view.startsAt)}
          status={view.teacher.fullName}
          title={view.title}
          tone="info"
        />
      ) : null}
      <StatusRow
        onPress={() => setKind("reschedule")}
        status={kind === "reschedule" ? message("resched.preference.selected") : undefined}
        subtitle={message("resched.preference.subtitle")}
        testID="reschedule-kind-move"
        title={message("resched.kind.reschedule")}
        tone={kind === "reschedule" ? "success" : "muted"}
      />
      <StatusRow
        onPress={() => setKind("cancellation")}
        status={kind === "cancellation" ? message("resched.preference.selected") : undefined}
        subtitle={message("resched.subtitle")}
        testID="reschedule-kind-cancel"
        title={message("resched.kind.cancellation")}
        tone={kind === "cancellation" ? "warning" : "muted"}
      />
      <PremiumTextField
        label={message("resched.reason.label")}
        multiline
        onChangeText={setReason}
        testID="reschedule-reason"
        value={reason}
      />
      {kind === "reschedule"
        ? options.map((option, index) => (
            <StatusRow
              key={option.toISOString()}
              onPress={() => setSelectedOption(index)}
              status={
                index === selectedOption
                  ? message("resched.preference.selected")
                  : message("resched.preference.option")
              }
              subtitle={formatBelcantoDate(option.toISOString())}
              testID={`reschedule-option-${index}`}
              title={message("resched.preference.title")}
              tone={index === selectedOption ? "success" : "muted"}
            />
          ))
        : null}
      <AccountBanner
        body={message("resched.banner.body")}
        title={message("resched.banner.title")}
      />
      {kind === "reschedule" ? (
        <AccountBanner
          body={message("resched.reserve.banner.body")}
          title={message("resched.reserve.banner.title")}
          tone="warning"
        />
      ) : null}
      {error !== null ? (
        <InlineNotice body={error} title={message("common.retry")} tone="error" />
      ) : null}
      <BlockAction
        busy={busy}
        disabled={reason.trim().length === 0}
        label={message("resched.submit")}
        onPress={() => void submit()}
        testID="reschedule-submit"
      />
      <BlockAction
        kind="secondary"
        label={message("resched.keep")}
        onPress={() => router.back()}
      />
    </AccountScreenShell>
  );
}
