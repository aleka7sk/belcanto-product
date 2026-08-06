import { router, useLocalSearchParams } from "expo-router";
import { useMemo, useState } from "react";

import { useApiClient } from "@/api";
import type { CommunityReportReason } from "@/api/contracts";
import { COMMUNITY_REPORT_REASONS } from "@/api/contracts";
import { createIntentIdempotency } from "@/controllers";
import { useMessage, type MessageKey } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice, PremiumTextField } from "../../components";
import { domainAccent } from "../../domainAccent";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusRow,
} from "../../patterns/accountPatterns";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav } from "../account/shared";

/**
 * COM-SAFE-02 «Пожаловаться» (Figma 348:683) — its own screen, exactly
 * as storyboarded: one question, four reasons, one action. Reached from
 * the thread; never rendered inside it. Reporter identity is never
 * revealed to the author.
 */

const REASON_KEYS: Record<CommunityReportReason, { title: MessageKey; body: MessageKey }> = {
  abuse: { title: "com.reason.abuse", body: "com.reason.abuseBody" },
  personal_data: { title: "com.reason.personal_data", body: "com.reason.personal_dataBody" },
  spam: { title: "com.reason.spam", body: "com.reason.spamBody" },
  other: { title: "com.reason.other", body: "com.reason.otherBody" },
};

export function parseReportTarget(params: {
  targetType?: string | string[] | undefined;
  targetId?: string | string[] | undefined;
}): { targetType: "post" | "comment"; targetId: string } | null {
  const rawType = Array.isArray(params.targetType) ? params.targetType[0] : params.targetType;
  const rawId = Array.isArray(params.targetId) ? params.targetId[0] : params.targetId;
  if ((rawType !== "post" && rawType !== "comment") || rawId === undefined || rawId === "") {
    return null;
  }
  return { targetType: rawType, targetId: rawId };
}

export function ReportScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ targetType?: string; targetId?: string }>();
  const target = parseReportTarget(params);

  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [reason, setReason] = useState<CommunityReportReason | null>(null);
  const [note, setNote] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filed, setFiled] = useState(false);

  const send = async () => {
    if (target === null || reason === null) return;
    const trimmed = note.trim();
    setError(null);
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.reportCommunityContent(
          accessToken,
          {
            targetType: target.targetType,
            targetId: target.targetId,
            reason,
            ...(trimmed === "" ? {} : { note: trimmed }),
          },
          idempotency.key(),
        ),
      );
      idempotency.complete();
      setFiled(true);
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  if (target === null) {
    return (
      <AccountScreenShell navigation={<AccountNav active="community" />} testID="community-report-guard">
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

  if (filed) {
    return (
      <AccountScreenShell navigation={<AccountNav active="community" />} testID="community-report-filed">
        <ScreenHeading
          accent={domainAccent("community")}
          eyebrow={message("com.report.eyebrow")}
          subtitle={message("com.report.subtitle")}
          title={message("com.report.filedTitle")}
        />
        <InlineNotice
          body={message("com.report.filedBody")}
          title={message("com.report.filedTitle")}
          tone="success"
        />
        <BlockAction
          label={message("com.tombstone.back")}
          onPress={() => router.back()}
          testID="community-report-done"
        />
      </AccountScreenShell>
    );
  }

  return (
    <AccountScreenShell
      keyboardAware
      navigation={<AccountNav active="community" />}
      testID="community-report"
    >
      <ScreenHeading
        accent={domainAccent("community")}
        eyebrow={message("com.report.eyebrow")}
        subtitle={message("com.report.subtitle")}
        title={message("com.report.title")}
      />
      {COMMUNITY_REPORT_REASONS.map((entry) => (
        <StatusRow
          key={entry}
          onPress={() => setReason(entry)}
          status={
            reason === entry
              ? message("com.reason.selected")
              : message("com.reason.choose")
          }
          subtitle={message(REASON_KEYS[entry].body)}
          testID={`community-reason-${entry}`}
          title={message(REASON_KEYS[entry].title)}
          tone={reason === entry ? "info" : "muted"}
        />
      ))}
      {reason === "other" ? (
        <PremiumTextField
          label={message("com.report.noteLabel")}
          multiline
          onChangeText={setNote}
          placeholder={message("com.report.notePlaceholder")}
          testID="community-report-note"
          value={note}
        />
      ) : null}
      {error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={error}
          onAction={() => void send()}
          title={message("com.report.title")}
        />
      ) : null}
      <BlockAction
        busy={busy}
        disabled={reason === null || (reason === "other" && note.trim() === "")}
        label={
          reason === null
            ? message("com.report.pickFirst")
            : message("com.report.send")
        }
        onPress={() => void send()}
        testID="community-report-send"
      />
      <BlockAction
        kind="secondary"
        label={message("common.cancel")}
        onPress={() => router.back()}
        testID="community-report-cancel"
      />
    </AccountScreenShell>
  );
}
