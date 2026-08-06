import { router, useLocalSearchParams } from "expo-router";
import { useMemo, useState } from "react";
import { StyleSheet, Text } from "react-native";

import { ApiTransportError, useApiClient } from "@/api";
import { createIntentIdempotency } from "@/controllers";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice, PremiumTextField } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
} from "../../patterns/accountPatterns";
import { EventDetailCard } from "../../patterns/eventPatterns";
import { semantic, typeStyles } from "../../tokens";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * STU-PRACTICE submit step — its own screen: one attempt, one note, one
 * action. When the homework is in «Доработка», the teacher's latest
 * feedback frames the retry. Voice recording and file upload arrive
 * with the device-module slice; the note says so honestly.
 */
export function HomeworkSubmitScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ homeworkId?: string }>();
  const homeworkId = typeof params.homeworkId === "string" ? params.homeworkId : "";

  const homework = useAccountResource((accessToken) =>
    api.getHomework(accessToken, homeworkId),
  );
  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [note, setNote] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const view = homework.value;
  const isStudentSelf =
    view !== null &&
    state.bootstrap?.studentId !== undefined &&
    state.bootstrap.studentId === view.studentId;
  const canSubmit =
    isStudentSelf && view !== null && (view.status === "in_progress" || view.status === "reviewed");
  const latestFeedback = view?.feedback[0];

  const submit = async () => {
    if (view === null) return;
    setError(null);
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.submitHomework(
          accessToken,
          view.id,
          { note: note.trim(), expectedVersion: view.version },
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
      navigation={<AccountNav active="practice" />}
      testID="homework-submit-screen"
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
      ) : !canSubmit ? (
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
            eyebrow={
              view.status === "reviewed"
                ? message("prac.resubmit.title")
                : message("prac.eyebrow")
            }
            subtitle={
              view.status === "reviewed"
                ? message("prac.resubmit.body")
                : message("prac.submit.noteHint")
            }
            title={message("prac.submit.open")}
          />
          {view.status === "reviewed" && latestFeedback !== undefined ? (
            <EventDetailCard
              accent={semantic.feedbackWarning}
              body={latestFeedback.body}
              status={
                latestFeedback.nextStep !== undefined
                  ? message("prac.feedback.nextStep", { step: latestFeedback.nextStep })
                  : undefined
              }
              statusColor={semantic.textGold}
              title={message("prac.feedback.title")}
            />
          ) : null}
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("prac.submit.note")}
            multiline
            onChangeText={setNote}
            placeholder={message("prac.submit.noteHint")}
            testID="homework-note"
            value={note}
          />
          <Text style={styles.metaLine}>{message("prac.submit.recorderNote")}</Text>
          {error !== null ? (
            <ErrorNotice
              actionLabel={message("common.retry")}
              body={error}
              onAction={() => void submit()}
              title={message("prac.submit.open")}
            />
          ) : null}
          <BlockAction
            busy={busy}
            disabled={note.trim() === ""}
            label={message("prac.submit")}
            onPress={() => void submit()}
            testID="homework-submit"
          />
          <BlockAction
            kind="secondary"
            label={message("common.cancel")}
            onPress={() => router.back()}
            testID="homework-submit-cancel"
          />
        </>
      )}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  metaLine: { color: semantic.textMuted, ...typeStyles.caption },
  muted: { color: semantic.textSecondary, ...typeStyles.bodyS },
});
