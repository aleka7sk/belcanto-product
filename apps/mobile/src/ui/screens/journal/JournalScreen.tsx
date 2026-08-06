import { router, useLocalSearchParams } from "expo-router";
import { useMemo, useState } from "react";
import { RefreshControl, StyleSheet, Text, View } from "react-native";

import { ApiError, ApiTransportError, useApiClient } from "@/api";
import type {
  JournalDraft,
  JournalEvidenceInput,
  JournalVersion,
  Lesson,
  LessonJournal,
} from "@/api/contracts";
import { createIntentIdempotency } from "@/controllers";
import { useMessage, type MessageFormatter } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice, PremiumTextField, TextAction } from "../../components";
import {
  AccountBanner,
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
} from "../../patterns/accountPatterns";
import { EventDetailCard } from "../../patterns/eventPatterns";
import { LessonRecapCard } from "../../patterns/journalPatterns";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * Lesson journal (Figma TCH-JOURNAL-03/04/08/10/13 + Lesson Recap
 * 309:20119, student side of flows C/H). One route serves every viewer the
 * API authorizes: the Lesson's Teacher edits draft → publish → correction,
 * managers read including the draft, the Student sees published versions
 * only. DEC-007 shows up twice — publishing snapshots an immutable version,
 * and a correction demands an explicit reason before the button unlocks.
 *
 * The 6-step wizard of Page 26 spans attendance, repertoire and homework;
 * this screen realizes its journal core (steps 2–3 and review/publish) on
 * the journal API. The remaining steps join with their own slices.
 */

const EMPTY_DRAFT: JournalDraft = { whatWorked: "", currentFocus: "", nextStep: "" };

type JournalValue = LessonJournal | "missing";
type LessonValue = Lesson | "unavailable";

interface EvidenceRow {
  area: string;
  note: string;
}

function lessonStartedNow(lesson: Lesson | null): boolean {
  return lesson !== null && new Date(lesson.startsAt).getTime() <= Date.now();
}

function draftSeed(journal: JournalValue | null): JournalDraft {
  if (journal === null || journal === "missing") return EMPTY_DRAFT;
  if (journal.draft !== undefined) return journal.draft;
  const current = journal.versions[0];
  if (current === undefined) return EMPTY_DRAFT;
  return {
    whatWorked: current.whatWorked,
    currentFocus: current.currentFocus,
    nextStep: current.nextStep,
  };
}

function versionSections(
  version: JournalVersion,
  message: MessageFormatter,
): { label: string; body: string }[] {
  return [
    { label: message("jrnl.recap.whatWorked"), body: version.whatWorked },
    { label: message("jrnl.recap.currentFocus"), body: version.currentFocus },
    { label: message("jrnl.recap.nextStep"), body: version.nextStep },
  ];
}

function VersionHistory({
  versions,
  message,
}: {
  versions: readonly JournalVersion[];
  message: MessageFormatter;
}) {
  if (versions.length <= 1) return null;
  return (
    <>
      <Text style={historyStyles.sectionTitle}>
        {message("jrnl.versions.title")}
      </Text>
      {versions.map((version) => (
        <StatusCard
          body={version.whatWorked}
          key={version.version}
          status={
            version.correctionNote !== undefined
              ? message("jrnl.version.correction", { note: version.correctionNote })
              : undefined
          }
          title={message("jrnl.version.label", {
            version: version.version,
            date: formatBelcantoDate(version.publishedAt),
          })}
          tone="warning"
        />
      ))}
    </>
  );
}

export function JournalScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ occurrenceId?: string; studentId?: string }>();
  const occurrenceId = typeof params.occurrenceId === "string" ? params.occurrenceId : "";
  const studentId = typeof params.studentId === "string" ? params.studentId : "";

  const journal = useAccountResource<JournalValue>(async (accessToken) => {
    try {
      return await api.getJournal(accessToken, occurrenceId, studentId);
    } catch (cause) {
      if (cause instanceof ApiError && cause.code === "NOT_FOUND") return "missing";
      throw cause;
    }
  });
  const lesson = useAccountResource<LessonValue>(async (accessToken) => {
    try {
      return await api.getLesson(accessToken, occurrenceId);
    } catch (cause) {
      if (
        cause instanceof ApiError &&
        (cause.code === "NOT_FOUND" || cause.code === "FORBIDDEN")
      ) {
        return "unavailable";
      }
      throw cause;
    }
  });

  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [overrides, setOverrides] = useState<Partial<JournalDraft>>({});
  const [evidenceRows, setEvidenceRows] = useState<EvidenceRow[]>([]);
  const [correctionNote, setCorrectionNote] = useState("");
  const [correcting, setCorrecting] = useState(false);
  const [saving, setSaving] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const bootstrap = state.bootstrap;
  const accountId = bootstrap?.accountId ?? "";
  const roles = bootstrap?.roles ?? [];
  const isSelf = bootstrap?.studentId !== undefined && bootstrap.studentId === studentId;
  const manager = roles.includes("Owner") || roles.includes("Administrator");

  const journalValue = journal.value;
  const lessonValue = lesson.value;
  const loadedJournal = journalValue !== null && journalValue !== "missing" ? journalValue : null;
  const loadedLesson = lessonValue !== null && lessonValue !== "unavailable" ? lessonValue : null;
  const teacherOf =
    (loadedJournal !== null && loadedJournal.teacher.accountId === accountId) ||
    (loadedLesson !== null && loadedLesson.teacher.accountId === accountId);

  const navActive = isSelf ? "profile" : "schedule";
  if (journalValue === null) {
    return (
      <AccountScreenShell navigation={<AccountNav active={navActive} />} testID="journal-loading">
        {journal.error !== null ? (
          <ErrorNotice
            actionLabel={message("common.retry")}
            body={apiErrorMessage(journal.error)}
            onAction={() => void journal.reload()}
            title={message("jrnl.eyebrow")}
          />
        ) : null}
      </AccountScreenShell>
    );
  }

  const studentName =
    loadedLesson?.students.find((student) => student.studentId === studentId)?.fullName ??
    null;
  const contextLine = [
    studentName,
    loadedLesson !== null ? formatBelcantoDate(loadedLesson.startsAt) : null,
  ]
    .filter((part): part is string => part !== null)
    .join(" · ");
  const teacherLine =
    loadedJournal !== null
      ? message("jrnl.teacher.line", { name: loadedJournal.teacher.fullName })
      : loadedLesson !== null
        ? message("jrnl.teacher.line", { name: loadedLesson.teacher.fullName })
        : "";
  const currentVersion = loadedJournal?.versions[0];

  // -------- Student: published versions only (recap + history) --------
  if (isSelf) {
    return (
      <AccountScreenShell navigation={<AccountNav active="profile" />} testID="journal-student">
        <ScreenHeading
          eyebrow={message("jrnl.eyebrow")}
          subtitle={[contextLine, teacherLine].filter(Boolean).join(" · ") || teacherLine}
          title={loadedLesson?.title ?? message("jrnl.published.title")}
        />
        {loadedJournal === null || currentVersion === undefined ? (
          <StatusCard
            body={message("jrnl.student.pending.body")}
            title={message("jrnl.student.pending.title")}
          />
        ) : (
          <>
            <LessonRecapCard
              eyebrow={message("jrnl.recap.published")}
              sections={versionSections(currentVersion, message)}
              testID="journal-student-recap"
              title={message("jrnl.published.subtitle", {
                date: formatBelcantoDate(currentVersion.publishedAt),
                version: currentVersion.version,
              })}
              tone="published"
            />
            <VersionHistory message={message} versions={loadedJournal.versions} />
          </>
        )}
      </AccountScreenShell>
    );
  }

  // -------- Manager (Owner/Administrator): read-only incl. draft --------
  if (!teacherOf && manager) {
    return (
      <AccountScreenShell navigation={<AccountNav active="schedule" />} testID="journal-manager">
        <ScreenHeading
          eyebrow={message("jrnl.eyebrow")}
          subtitle={[contextLine, teacherLine].filter(Boolean).join(" · ") || teacherLine}
          title={loadedLesson?.title ?? message("jrnl.editor.title")}
        />
        {loadedJournal === null ? (
          <StatusCard
            body={message("jrnl.student.pending.body")}
            title={message("jrnl.student.pending.title")}
          />
        ) : (
          <>
            {loadedJournal.draft !== undefined ? (
              <LessonRecapCard
                eyebrow={message("jrnl.recap.draft")}
                sections={[
                  { label: message("jrnl.recap.whatWorked"), body: loadedJournal.draft.whatWorked },
                  { label: message("jrnl.recap.currentFocus"), body: loadedJournal.draft.currentFocus },
                  { label: message("jrnl.recap.nextStep"), body: loadedJournal.draft.nextStep },
                ]}
                title={contextLine || message("jrnl.editor.title")}
                tone="draft"
              />
            ) : null}
            {currentVersion !== undefined ? (
              <LessonRecapCard
                eyebrow={message("jrnl.recap.published")}
                sections={versionSections(currentVersion, message)}
                title={message("jrnl.published.subtitle", {
                  date: formatBelcantoDate(currentVersion.publishedAt),
                  version: currentVersion.version,
                })}
                tone="published"
              />
            ) : null}
            <VersionHistory message={message} versions={loadedJournal.versions} />
          </>
        )}
      </AccountScreenShell>
    );
  }

  if (!teacherOf) {
    return (
      <AccountScreenShell navigation={<AccountNav active={navActive} />} testID="journal-guard">
        <InlineNotice
          body={message("jrnl.guard.body")}
          title={message("jrnl.guard.title")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  // -------- Teacher: draft → publish → correction --------
  const seed = draftSeed(journalValue);
  const draft: JournalDraft = {
    whatWorked: overrides.whatWorked ?? seed.whatWorked,
    currentFocus: overrides.currentFocus ?? seed.currentFocus,
    nextStep: overrides.nextStep ?? seed.nextStep,
  };
  const dirty =
    draft.whatWorked !== seed.whatWorked ||
    draft.currentFocus !== seed.currentFocus ||
    draft.nextStep !== seed.nextStep;
  const draftComplete =
    draft.whatWorked.trim() !== "" &&
    draft.currentFocus.trim() !== "" &&
    draft.nextStep.trim() !== "";
  const draftSaved = loadedJournal?.draft !== undefined;
  const published = currentVersion !== undefined;
  const correctionNeeded = published;
  const payloadEvidence: JournalEvidenceInput[] = evidenceRows
    .map((row) => ({ area: row.area.trim(), note: row.note.trim() }))
    .filter((row) => row.area !== "" && row.note !== "");
  const evidenceIncomplete = evidenceRows.some(
    (row) => (row.area.trim() === "") !== (row.note.trim() === ""),
  );
  const lessonStarted = lessonStartedNow(loadedLesson);
  const canPublish =
    draftSaved &&
    !dirty &&
    !evidenceIncomplete &&
    (loadedLesson === null || lessonStarted) &&
    (!correctionNeeded || correctionNote.trim() !== "");

  const setField = (field: keyof JournalDraft) => (value: string) =>
    setOverrides((current) => ({ ...current, [field]: value }));

  const saveDraft = async () => {
    setActionError(null);
    setSaving(true);
    try {
      await runAuthenticated((accessToken) =>
        api.saveJournalDraft(accessToken, {
          occurrenceId,
          studentId,
          whatWorked: draft.whatWorked.trim(),
          currentFocus: draft.currentFocus.trim(),
          nextStep: draft.nextStep.trim(),
        }),
      );
      setOverrides({});
      await journal.reload();
    } catch (cause) {
      setActionError(apiErrorMessage(cause));
    } finally {
      setSaving(false);
    }
  };

  const publish = async () => {
    setActionError(null);
    setPublishing(true);
    try {
      await runAuthenticated((accessToken) =>
        api.publishJournal(
          accessToken,
          {
            occurrenceId,
            studentId,
            ...(correctionNeeded ? { correctionNote: correctionNote.trim() } : {}),
            ...(payloadEvidence.length > 0 ? { evidence: payloadEvidence } : {}),
          },
          idempotency.key(),
        ),
      );
      idempotency.complete();
      setCorrecting(false);
      setCorrectionNote("");
      setEvidenceRows([]);
      await journal.reload();
    } catch (cause) {
      if (!(cause instanceof ApiTransportError)) idempotency.abandon();
      setActionError(apiErrorMessage(cause));
      await journal.reload();
    } finally {
      setPublishing(false);
    }
  };

  // Published summary with the correction entry point (TCH-JOURNAL-10/13).
  if (published && !correcting) {
    return (
      <AccountScreenShell navigation={<AccountNav active="schedule" />} testID="journal-published">
        <ScreenHeading
          eyebrow={message("jrnl.published.eyebrow")}
          subtitle={contextLine || teacherLine}
          title={message("jrnl.published.title")}
        />
        <LessonRecapCard
          eyebrow={message("jrnl.recap.published")}
          sections={versionSections(currentVersion, message)}
          testID="journal-published-recap"
          title={message("jrnl.published.subtitle", {
            date: formatBelcantoDate(currentVersion.publishedAt),
            version: currentVersion.version,
          })}
          tone="published"
        />
        <VersionHistory message={message} versions={loadedJournal?.versions ?? []} />
        <BlockAction
          kind="secondary"
          label={message("jrnl.correct.action")}
          onPress={() => setCorrecting(true)}
          testID="journal-correct-open"
        />
        <BlockAction
          kind="secondary"
          label={message("prac.create.title")}
          onPress={() =>
            router.push({
              pathname: "/(protected)/practice/create",
              params: { occurrenceId, studentId },
            })
          }
          testID="journal-create-homework"
        />
        <BlockAction
          kind="secondary"
          label={message("prac.detail.eyebrow.teacher")}
          onPress={() =>
            router.push({
              pathname: "/(protected)/practice",
              params: { studentId },
            })
          }
          testID="journal-open-homework"
        />
      </AccountScreenShell>
    );
  }

  return (
    <AccountScreenShell
      navigation={<AccountNav active="schedule" />}
      refreshControl={
        <RefreshControl
          onRefresh={() => {
            void Promise.all([journal.reload(), lesson.reload()]);
          }}
          refreshing={journal.refreshing || lesson.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="journal-editor"
    >
      <ScreenHeading
        eyebrow={correcting ? message("jrnl.correct.eyebrow") : message("jrnl.eyebrow")}
        subtitle={
          correcting
            ? message("jrnl.correct.subtitle")
            : contextLine || message("jrnl.editor.subtitle")
        }
        title={
          correcting
            ? message("jrnl.correct.title", {
                version: (loadedJournal?.currentVersion ?? 0) + 1,
              })
            : message("jrnl.editor.title")
        }
      />
      {correcting ? (
        <PremiumTextField
          autoCapitalize="sentences"
          helper={message("jrnl.correct.noteHelper")}
          label={message("jrnl.correct.note")}
          multiline
          onChangeText={setCorrectionNote}
          testID="journal-correction-note"
          value={correctionNote}
        />
      ) : null}
      <PremiumTextField
        autoCapitalize="sentences"
        helper={message("jrnl.field.visible")}
        label={message("jrnl.field.whatWorked")}
        multiline
        onChangeText={setField("whatWorked")}
        testID="journal-what-worked"
        value={draft.whatWorked}
      />
      <PremiumTextField
        autoCapitalize="sentences"
        helper={message("jrnl.field.visible")}
        label={message("jrnl.field.currentFocus")}
        multiline
        onChangeText={setField("currentFocus")}
        testID="journal-current-focus"
        value={draft.currentFocus}
      />
      <PremiumTextField
        autoCapitalize="sentences"
        helper={message("jrnl.field.visible")}
        label={message("jrnl.field.nextStep")}
        multiline
        onChangeText={setField("nextStep")}
        testID="journal-next-step"
        value={draft.nextStep}
      />
      <BlockAction
        busy={saving}
        disabled={!draftComplete || (!dirty && draftSaved)}
        kind="secondary"
        label={message("jrnl.saveDraft")}
        onPress={() => void saveDraft()}
        testID="journal-save-draft"
      />
      <Text style={styles.metaLine}>
        {draftSaved && !dirty && loadedJournal !== null
          ? message("jrnl.draft.saved", {
              date: formatBelcantoDate(loadedJournal.updatedAt),
            })
          : message("jrnl.draft.privacy")}
      </Text>

      <AccountBanner
        body={message("jrnl.evidence.subtitle")}
        title={message("jrnl.evidence.title")}
      />
      {evidenceRows.map((row, index) => (
        <View key={index} style={styles.evidenceRow}>
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("jrnl.evidence.area")}
            onChangeText={(value) =>
              setEvidenceRows((current) =>
                current.map((entry, at) =>
                  at === index ? { ...entry, area: value } : entry,
                ),
              )
            }
            testID={`journal-evidence-area-${index}`}
            value={row.area}
          />
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("jrnl.evidence.note")}
            multiline
            onChangeText={(value) =>
              setEvidenceRows((current) =>
                current.map((entry, at) =>
                  at === index ? { ...entry, note: value } : entry,
                ),
              )
            }
            placeholder={message("jrnl.evidence.noteHint")}
            testID={`journal-evidence-note-${index}`}
            value={row.note}
          />
          <TextAction
            align="right"
            label={message("jrnl.evidence.remove")}
            onPress={() =>
              setEvidenceRows((current) => current.filter((_, at) => at !== index))
            }
          />
        </View>
      ))}
      {evidenceRows.length < 10 ? (
        <BlockAction
          kind="secondary"
          label={message("jrnl.evidence.add")}
          onPress={() =>
            setEvidenceRows((current) => [...current, { area: "", note: "" }])
          }
          testID="journal-evidence-add"
        />
      ) : (
        <Text style={styles.metaLine}>{message("jrnl.evidence.limit")}</Text>
      )}

      {draftSaved && loadedJournal?.draft !== undefined ? (
        <>
          <ScreenHeading
            eyebrow={message("jrnl.review.eyebrow")}
            subtitle={contextLine || message("jrnl.editor.subtitle")}
            title={message("jrnl.review.title")}
          />
          <LessonRecapCard
            chips={
              payloadEvidence.length > 0
                ? [message("jrnl.recap.evidenceChip", { count: payloadEvidence.length })]
                : undefined
            }
            eyebrow={message("jrnl.recap.draft")}
            sections={[
              {
                label: message("jrnl.recap.whatWorked"),
                body: loadedJournal.draft.whatWorked,
              },
              {
                label: message("jrnl.recap.currentFocus"),
                body: loadedJournal.draft.currentFocus,
              },
              {
                label: message("jrnl.recap.nextStep"),
                body: loadedJournal.draft.nextStep,
              },
            ]}
            testID="journal-draft-recap"
            title={contextLine || message("jrnl.review.title")}
            tone="draft"
          />
        </>
      ) : null}
      {loadedLesson !== null && !lessonStarted ? (
        <StatusCard
          body={formatBelcantoDate(loadedLesson.startsAt)}
          status={message("jrnl.draft.privacy")}
          title={message("jrnl.publish.gate")}
          tone="warning"
        />
      ) : null}
      <EventDetailCard
        accent={semantic.accentCyan}
        body={message("jrnl.publish.updates.body")}
        status={message("jrnl.publish.updates.status")}
        statusColor={semantic.accentCyan}
        title={message("jrnl.publish.updates.title")}
      />
      {actionError !== null ? (
        <InlineNotice body={actionError} title={message("common.retry")} tone="error" />
      ) : null}
      <BlockAction
        busy={publishing}
        disabled={!canPublish}
        label={correcting ? message("jrnl.correct.publish") : message("jrnl.publish")}
        onPress={() => void publish()}
        testID="journal-publish"
      />
      {correcting && correctionNote.trim() === "" ? (
        <Text style={styles.metaLine}>{message("jrnl.correct.required")}</Text>
      ) : null}
      {correcting ? (
        <BlockAction
          kind="secondary"
          label={message("common.cancel")}
          onPress={() => {
            setCorrecting(false);
            setCorrectionNote("");
            setOverrides({});
          }}
        />
      ) : null}
    </AccountScreenShell>
  );
}

const historyStyles = StyleSheet.create({
  sectionTitle: {
    color: semantic.textPrimary,
    marginTop: space.s2,
    ...typeStyles.headingM,
  },
});

const styles = StyleSheet.create({
  metaLine: { color: semantic.textMuted, ...typeStyles.caption },
  evidenceRow: { gap: space.s3 },
});
