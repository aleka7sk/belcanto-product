import { router, useLocalSearchParams } from "expo-router";
import { useMemo, useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { ApiTransportError, useApiClient } from "@/api";
import type { HomeworkAssignment, MediaObject } from "@/api/contracts";
import { createIntentIdempotency } from "@/controllers";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice, PremiumTextField, TextAction } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import { EventDetailCard } from "../../patterns/eventPatterns";
import { EvidenceTile } from "../../patterns/journalPatterns";
import {
  TaskRow,
  UploadProgressCard,
  uploadAccent,
} from "../../patterns/practicePatterns";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";
import { homeworkStatusLabel } from "./PracticeHubScreen";

/**
 * Homework detail (Figma STU-PRACTICE-02..09/15, flow E). One route for
 * the API-authorized viewers: the Student walks assigned → in progress
 * (task toggles) → submit → «на проверке» → feedback/resubmit →
 * accepted-with-evidence; the Teacher reviews the latest attempt with an
 * explicit needs_revision/accepted decision and may cancel with a
 * preserved reason. Upload bars render real uploadedBytes of attached
 * media — voice recording itself arrives with the recorder module.
 */

function mediaLine(media: MediaObject): string {
  const megabytes = (media.byteSize / (1024 * 1024)).toFixed(1).replace(".", ",");
  return `${media.kind} · ${megabytes} МБ`;
}

export function HomeworkDetailScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ homeworkId?: string }>();
  const homeworkId = typeof params.homeworkId === "string" ? params.homeworkId : "";

  const homework = useAccountResource((accessToken) =>
    api.getHomework(accessToken, homeworkId),
  );
  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [busy, setBusy] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [note, setNote] = useState("");
  const [reviewBody, setReviewBody] = useState("");
  const [reviewNextStep, setReviewNextStep] = useState("");
  const [reviewDecision, setReviewDecision] = useState<"needs_revision" | "accepted">(
    "needs_revision",
  );
  const [evidenceArea, setEvidenceArea] = useState("");
  const [evidenceNote, setEvidenceNote] = useState("");
  const [cancelling, setCancelling] = useState(false);
  const [cancelReason, setCancelReason] = useState("");

  const view = homework.value;
  const bootstrap = state.bootstrap;
  const isStudentSelf =
    view !== null && bootstrap?.studentId !== undefined && bootstrap.studentId === view.studentId;
  const isTeacher = view !== null && bootstrap !== null && view.teacher.accountId === bootstrap.accountId;

  const run = async (label: string, operation: (accessToken: string) => Promise<HomeworkAssignment>) => {
    setActionError(null);
    setBusy(label);
    try {
      await runAuthenticated(operation);
      idempotency.complete();
      setNote("");
      setReviewBody("");
      setReviewNextStep("");
      setEvidenceArea("");
      setEvidenceNote("");
      setCancelling(false);
      setCancelReason("");
      await homework.reload();
    } catch (cause) {
      if (!(cause instanceof ApiTransportError)) idempotency.abandon();
      setActionError(apiErrorMessage(cause));
      await homework.reload();
    } finally {
      setBusy(null);
    }
  };

  if (view === null) {
    return (
      <AccountScreenShell navigation={<AccountNav active="practice" />} testID="homework-loading">
        {homework.error !== null ? (
          <InlineNotice
            body={apiErrorMessage(homework.error)}
            title={message("common.retry")}
            tone="error"
          />
        ) : null}
      </AccountScreenShell>
    );
  }

  const doneTasks = view.tasks.filter((task) => task.status === "done").length;
  const latestFeedback = view.feedback[0];
  const acceptedFeedback = view.feedback.find((entry) => entry.decision === "accepted");
  const contextLine =
    view.dueAt !== undefined
      ? message("prac.detail.context", {
          teacher: view.teacher.fullName,
          date: formatBelcantoDate(view.dueAt),
        })
      : message("prac.detail.context.noDue", { teacher: view.teacher.fullName });
  const canSubmit =
    isStudentSelf && (view.status === "in_progress" || view.status === "reviewed");
  const uploadsInFlight = view.submissions.flatMap((submission, submissionIndex) =>
    submission.media
      .filter((media) => media.status === "pending" || media.status === "uploading")
      .map((media, mediaIndex) => ({
        media,
        title: message("prac.attempt.uploading", { attempt: submission.attempt }),
        accent: uploadAccent(submissionIndex + mediaIndex),
      })),
  );

  return (
    <AccountScreenShell navigation={<AccountNav active="practice" />} testID="homework-detail">
      <ScreenHeading
        eyebrow={
          isTeacher
            ? message("prac.detail.eyebrow.teacher")
            : message("prac.detail.eyebrow.student")
        }
        subtitle={`${contextLine} · ${homeworkStatusLabel(view.status, message)}`}
        title={message("prac.detail.title")}
      />
      <EventDetailCard
        accent={semantic.accentCyan}
        body={view.goal}
        status={view.readinessCriteria}
        statusColor={semantic.accentCyan}
        title={message("prac.goal.title")}
      />
      {view.tasks.length > 0 ? (
        <View style={styles.planCard}>
          <Text style={styles.planTitle}>{message("prac.plan.title")}</Text>
          {view.tasks.map((task) => (
            <TaskRow
              doneLabel={message("prac.task.done")}
              key={task.id}
              onToggle={
                isStudentSelf && view.status === "in_progress"
                  ? () =>
                      void run(`task-${task.id}`, (accessToken) =>
                        api.markHomeworkTask(accessToken, view.id, task.id, {
                          done: task.status !== "done",
                        }),
                      )
                  : undefined
              }
              pendingLabel={message("prac.task.pending")}
              task={task}
              testID={`homework-task-${task.id}`}
            />
          ))}
          <Text style={styles.planProgress}>
            {message("prac.plan.progress", { done: doneTasks, total: view.tasks.length })}
          </Text>
        </View>
      ) : null}
      {view.attachments.length > 0 ? (
        <StatusCard
          body={view.attachments.map(mediaLine).join(" · ")}
          status={message("prac.materials.ready")}
          title={message("prac.materials.title")}
          tone="success"
        />
      ) : null}
      {uploadsInFlight.map((upload) => (
        <UploadProgressCard
          accent={upload.accent}
          byteSize={upload.media.byteSize}
          key={upload.media.id}
          title={`${upload.title} · ${mediaLine(upload.media)}`}
          uploadedBytes={upload.media.uploadedBytes}
        />
      ))}

      {view.status === "cancelled" ? (
        <StatusCard
          body={message("prac.cancelled.reason", { reason: view.cancelReason ?? "" })}
          title={message("prac.cancelled.title")}
          tone="danger"
        />
      ) : null}
      {view.status === "expired" ? (
        <StatusCard
          body={message("prac.expired.body")}
          title={message("prac.expired.title")}
          tone="muted"
        />
      ) : null}
      {view.status === "submitted" && !isTeacher ? (
        <StatusCard
          body={message("prac.submitted.body")}
          status={message("prac.submitted.status")}
          title={message("prac.submitted.title")}
          tone="warning"
        />
      ) : null}
      {latestFeedback !== undefined && view.status === "reviewed" ? (
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
      {view.status === "completed" && acceptedFeedback !== undefined ? (
        <>
          <StatusCard
            body={acceptedFeedback.body}
            status={message("prac.accepted.body")}
            title={message("prac.accepted.title")}
            tone="success"
          />
          {acceptedFeedback.evidenceArea !== undefined &&
          acceptedFeedback.evidenceNote !== undefined ? (
            <EvidenceTile
              kind="teacherNote"
              note={acceptedFeedback.evidenceNote}
              sourceLine={`${acceptedFeedback.teacher.fullName} · ${formatBelcantoDate(acceptedFeedback.createdAt)}`}
              title={acceptedFeedback.evidenceArea}
              visibility={message("growth.evidence.visibility")}
            />
          ) : null}
          {isStudentSelf ? (
            <BlockAction
              kind="secondary"
              label={message("prac.accepted.action")}
              onPress={() => router.push("/(protected)/progress")}
            />
          ) : null}
        </>
      ) : null}

      {view.submissions.length > 0 ? (
        <>
          <Text style={styles.sectionTitle}>{message("prac.attempts.title")}</Text>
          {view.submissions.map((submission) => (
            <StatusRow
              key={submission.id}
              status={message("prac.attempt.media", { count: submission.media.length })}
              subtitle={submission.note ?? ""}
              title={message("prac.attempt.entry", {
                attempt: submission.attempt,
                date: formatBelcantoDate(submission.submittedAt),
              })}
              tone="info"
            />
          ))}
        </>
      ) : null}

      {actionError !== null ? (
        <InlineNotice body={actionError} title={message("common.retry")} tone="error" />
      ) : null}

      {isStudentSelf && view.status === "assigned" ? (
        <BlockAction
          busy={busy === "start"}
          label={message("prac.start")}
          onPress={() =>
            void run("start", (accessToken) =>
              api.startHomework(accessToken, view.id, idempotency.key()),
            )
          }
          testID="homework-start"
        />
      ) : null}
      {canSubmit ? (
        <>
          {view.status === "reviewed" ? (
            <ScreenHeading
              eyebrow={message("prac.resubmit.title")}
              subtitle={message("prac.resubmit.body")}
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
          <BlockAction
            busy={busy === "submit"}
            disabled={note.trim() === ""}
            label={message("prac.submit")}
            onPress={() =>
              void run("submit", (accessToken) =>
                api.submitHomework(
                  accessToken,
                  view.id,
                  { note: note.trim(), expectedVersion: view.version },
                  idempotency.key(),
                ),
              )
            }
            testID="homework-submit"
          />
        </>
      ) : null}

      {isTeacher && view.status === "submitted" ? (
        <>
          <ScreenHeading
            eyebrow={message("prac.review.title")}
            subtitle={message("prac.review.bodyHint")}
            title={message("prac.feedback.title")}
          />
          <View style={styles.decisionRow}>
            <TextAction
              label={
                (reviewDecision === "needs_revision" ? "● " : "○ ") +
                message("prac.review.needsRevision")
              }
              onPress={() => setReviewDecision("needs_revision")}
            />
            <TextAction
              label={
                (reviewDecision === "accepted" ? "● " : "○ ") +
                message("prac.review.accept")
              }
              onPress={() => setReviewDecision("accepted")}
            />
          </View>
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("prac.review.body")}
            multiline
            onChangeText={setReviewBody}
            placeholder={message("prac.review.bodyHint")}
            testID="homework-review-body"
            value={reviewBody}
          />
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("prac.review.nextStep")}
            multiline
            onChangeText={setReviewNextStep}
            testID="homework-review-next"
            value={reviewNextStep}
          />
          {reviewDecision === "accepted" ? (
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
          <BlockAction
            busy={busy === "review"}
            disabled={
              reviewBody.trim() === "" ||
              (evidenceArea.trim() === "") !== (evidenceNote.trim() === "")
            }
            label={
              reviewDecision === "accepted"
                ? message("prac.review.accept")
                : message("prac.review.submit")
            }
            onPress={() =>
              void run("review", (accessToken) =>
                api.reviewHomework(
                  accessToken,
                  view.id,
                  {
                    decision: reviewDecision,
                    body: reviewBody.trim(),
                    ...(reviewNextStep.trim() !== ""
                      ? { nextStep: reviewNextStep.trim() }
                      : {}),
                    ...(reviewDecision === "accepted" && evidenceArea.trim() !== ""
                      ? {
                          evidenceArea: evidenceArea.trim(),
                          evidenceNote: evidenceNote.trim(),
                        }
                      : {}),
                    expectedVersion: view.version,
                  },
                  idempotency.key(),
                ),
              )
            }
            testID="homework-review-submit"
          />
        </>
      ) : null}

      {isTeacher &&
      ["draft", "assigned", "in_progress", "submitted", "reviewed"].includes(view.status) ? (
        cancelling ? (
          <>
            <PremiumTextField
              autoCapitalize="sentences"
              label={message("prac.cancel.reason")}
              multiline
              onChangeText={setCancelReason}
              testID="homework-cancel-reason"
              value={cancelReason}
            />
            <BlockAction
              busy={busy === "cancel"}
              disabled={cancelReason.trim() === ""}
              kind="secondary"
              label={message("prac.cancel.confirm")}
              onPress={() =>
                void run("cancel", (accessToken) =>
                  api.cancelHomework(
                    accessToken,
                    view.id,
                    { reason: cancelReason.trim() },
                    idempotency.key(),
                  ),
                )
              }
              testID="homework-cancel-confirm"
            />
            <BlockAction
              kind="secondary"
              label={message("common.cancel")}
              onPress={() => setCancelling(false)}
            />
          </>
        ) : (
          <BlockAction
            kind="secondary"
            label={message("prac.cancel.action")}
            onPress={() => setCancelling(true)}
            testID="homework-cancel-open"
          />
        )
      ) : null}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  planCard: {
    backgroundColor: semantic.bgSurface,
    borderColor: semantic.borderDefault,
    borderRadius: 20,
    borderWidth: 1,
    gap: 10,
    padding: space.s4,
  },
  planTitle: { color: semantic.textPrimary, ...typeStyles.headingM },
  planProgress: { color: semantic.textAccent, ...typeStyles.labelM },
  sectionTitle: {
    color: semantic.textPrimary,
    marginTop: space.s2,
    ...typeStyles.headingM,
  },
  metaLine: { color: semantic.textMuted, ...typeStyles.caption },
  decisionRow: {
    flexDirection: "row",
    gap: space.s4,
    justifyContent: "flex-start",
  },
});
