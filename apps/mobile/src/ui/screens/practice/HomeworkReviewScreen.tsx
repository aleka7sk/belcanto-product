import { router, useLocalSearchParams } from "expo-router";
import { useMemo, useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { ApiTransportError, useApiClient } from "@/api";
import { createIntentIdempotency } from "@/controllers";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice, PremiumTextField, TextAction } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusRow,
} from "../../patterns/accountPatterns";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * Teacher review step — its own screen: the latest attempt, one
 * decision (needs revision / accepted), one comment, and — on
 * acceptance — the named-area evidence that unlocks progress (DEC-006).
 * The homework detail links here; the forms never stack inline again.
 */
export function HomeworkReviewScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ homeworkId?: string }>();
  const homeworkId = typeof params.homeworkId === "string" ? params.homeworkId : "";

  const homework = useAccountResource((accessToken) =>
    api.getHomework(accessToken, homeworkId),
  );
  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [decision, setDecision] = useState<"needs_revision" | "accepted">("needs_revision");
  const [body, setBody] = useState("");
  const [nextStep, setNextStep] = useState("");
  const [evidenceArea, setEvidenceArea] = useState("");
  const [evidenceNote, setEvidenceNote] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const view = homework.value;
  const isTeacher =
    view !== null && state.bootstrap !== null && view.teacher.accountId === state.bootstrap.accountId;
  const reviewable = isTeacher && view !== null && view.status === "submitted";
  const latestSubmission = view?.submissions[0];

  const send = async () => {
    if (view === null) return;
    setError(null);
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.reviewHomework(
          accessToken,
          view.id,
          {
            decision,
            body: body.trim(),
            ...(nextStep.trim() !== "" ? { nextStep: nextStep.trim() } : {}),
            ...(decision === "accepted" && evidenceArea.trim() !== ""
              ? { evidenceArea: evidenceArea.trim(), evidenceNote: evidenceNote.trim() }
              : {}),
            expectedVersion: view.version,
          },
          idempotency.key(),
        ),
      );
      idempotency.complete();
      router.back();
    } catch (cause) {
      if (!(cause instanceof ApiTransportError)) idempotency.abandon();
      setError(apiErrorMessage(cause));
      await homework.reload();
    } finally {
      setBusy(false);
    }
  };

  return (
    <AccountScreenShell
      keyboardAware
      navigation={<AccountNav active="review" />}
      testID="homework-review-screen"
    >
      {view === null ? (
        homework.error !== null ? (
          <ErrorNotice
            actionLabel={message("common.retry")}
            body={apiErrorMessage(homework.error)}
            onAction={() => void homework.reload()}
            title={message("prac.detail.title")}
          />
        ) : (
          <Text style={styles.muted}>{message("common.loading")}</Text>
        )
      ) : !reviewable ? (
        <>
          <InlineNotice
            body={message("prac.guard.body")}
            title={message("prac.guard.title")}
            tone="error"
          />
          <BlockAction
            kind="secondary"
            label={message("common.cancel")}
            onPress={() => router.back()}
          />
        </>
      ) : (
        <>
          <ScreenHeading
            eyebrow={message("prac.review.title")}
            subtitle={message("prac.review.bodyHint")}
            title={message("prac.review.open")}
          />
          {latestSubmission !== undefined ? (
            <StatusRow
              status={message("prac.attempt.media", { count: latestSubmission.media.length })}
              subtitle={latestSubmission.note ?? ""}
              title={message("prac.attempt.entry", {
                attempt: latestSubmission.attempt,
                date: formatBelcantoDate(latestSubmission.submittedAt),
              })}
              tone="info"
            />
          ) : null}
          <View style={styles.decisionRow}>
            <TextAction
              label={
                (decision === "needs_revision" ? "● " : "○ ") +
                message("prac.review.needsRevision")
              }
              onPress={() => setDecision("needs_revision")}
            />
            <TextAction
              label={
                (decision === "accepted" ? "● " : "○ ") + message("prac.review.accept")
              }
              onPress={() => setDecision("accepted")}
            />
          </View>
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("prac.review.body")}
            multiline
            onChangeText={setBody}
            placeholder={message("prac.review.bodyHint")}
            testID="homework-review-body"
            value={body}
          />
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("prac.review.nextStep")}
            multiline
            onChangeText={setNextStep}
            testID="homework-review-next"
            value={nextStep}
          />
          {decision === "accepted" ? (
            <>
              <PremiumTextField
                autoCapitalize="sentences"
                helper={message("prac.review.evidenceHint")}
                label={message("prac.review.evidenceArea")}
                onChangeText={setEvidenceArea}
                testID="homework-review-area"
                value={evidenceArea}
              />
              <PremiumTextField
                autoCapitalize="sentences"
                label={message("prac.review.evidenceNote")}
                multiline
                onChangeText={setEvidenceNote}
                testID="homework-review-evidence"
                value={evidenceNote}
              />
            </>
          ) : null}
          {error !== null ? (
            <ErrorNotice
              actionLabel={message("common.retry")}
              body={error}
              onAction={() => void send()}
              title={message("prac.review.open")}
            />
          ) : null}
          <BlockAction
            busy={busy}
            disabled={
              body.trim() === "" ||
              (evidenceArea.trim() === "") !== (evidenceNote.trim() === "")
            }
            label={
              decision === "accepted"
                ? message("prac.review.accept")
                : message("prac.review.submit")
            }
            onPress={() => void send()}
            testID="homework-review-submit"
          />
          <BlockAction
            kind="secondary"
            label={message("common.cancel")}
            onPress={() => router.back()}
            testID="homework-review-cancel"
          />
        </>
      )}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  decisionRow: {
    flexDirection: "row",
    gap: space.s4,
    justifyContent: "flex-start",
  },
  muted: { color: semantic.textSecondary, ...typeStyles.bodyS },
});
