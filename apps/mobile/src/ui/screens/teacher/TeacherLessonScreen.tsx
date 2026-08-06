import { router, useLocalSearchParams } from "expo-router";
import { Fragment, type ReactNode } from "react";
import { RefreshControl, StyleSheet, Text } from "react-native";

import { ApiError, useApiClient } from "@/api";
import type {
  HomeworkAssignment,
  Lesson,
  LessonJournal,
  ProgressEvidence,
} from "@/api/contracts";
import { useMessage, type MessageFormatter } from "@/i18n";
import { ErrorNotice } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import { GrowthSignal, LessonRecapCard } from "../../patterns/journalPatterns";
import { HOMEWORK_STATUS_TONE } from "../../patterns/practicePatterns";
import { lessonAgendaState } from "../../teacherToday";
import { semantic, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";
import { homeworkStatusLabel } from "../practice/PracticeHubScreen";
import { AttendanceBlock } from "./AttendanceBlock";

/**
 * Teacher lesson context (Figma TCH-LESSON-01/02). Before, during and
 * after a Lesson the Teacher sees, per participant: the latest journal
 * state, the latest homework with its lifecycle status and the latest
 * named-area evidence (DEC-006). A group keeps every student's context
 * separate and says so — nothing here is ever published to the whole
 * group. The design's private pedagogical note waits for its own model.
 */

interface ParticipantContext {
  studentId: string;
  fullName: string;
  journal: LessonJournal | null;
  latestHomework: HomeworkAssignment | null;
  latestEvidence: ProgressEvidence | null;
}

interface LessonContextView {
  lesson: Lesson;
  participants: ParticipantContext[];
}

function lessonPhaseKey(
  lesson: Lesson,
): "tch.lesson.eyebrow.before" | "tch.lesson.eyebrow.now" | "tch.lesson.eyebrow.after" {
  const state = lessonAgendaState(lesson, Date.now());
  if (state === "now") return "tch.lesson.eyebrow.now";
  if (state === "completed") return "tch.lesson.eyebrow.after";
  return "tch.lesson.eyebrow.before";
}

function formatLine(lesson: Lesson, message: MessageFormatter): string {
  return [
    lesson.students.length > 1
      ? message("tch.format.group", { count: lesson.students.length })
      : message("tch.format.individual"),
    formatBelcantoDate(lesson.startsAt),
    lesson.location,
  ]
    .filter((part): part is string => part !== undefined && part !== "")
    .join(" · ");
}

function ParticipantBlocks({
  participant,
  occurrenceId,
  message,
}: {
  participant: ParticipantContext;
  occurrenceId: string;
  message: MessageFormatter;
}): ReactNode {
  const currentVersion = participant.journal?.versions[0];
  return (
    <Fragment>
      {currentVersion !== undefined ? (
        <LessonRecapCard
          eyebrow={message("jrnl.recap.published")}
          sections={[
            { label: message("jrnl.recap.whatWorked"), body: currentVersion.whatWorked },
            {
              label: message("jrnl.recap.currentFocus"),
              body: currentVersion.currentFocus,
            },
            { label: message("jrnl.recap.nextStep"), body: currentVersion.nextStep },
          ]}
          testID={`teacher-lesson-recap-${participant.studentId}`}
          title={message("jrnl.published.subtitle", {
            date: formatBelcantoDate(currentVersion.publishedAt),
            version: currentVersion.version,
          })}
          tone="published"
        />
      ) : (
        <StatusCard
          body={message("tch.lesson.recap.none")}
          title={message("tch.lesson.recap.title")}
        />
      )}
      {participant.latestHomework !== null ? (
        <StatusRow
          onPress={() =>
            router.push({
              pathname: "/(protected)/practice/[homeworkId]",
              params: { homeworkId: participant.latestHomework!.id },
            })
          }
          status={homeworkStatusLabel(participant.latestHomework.status, message)}
          subtitle={participant.latestHomework.goal}
          testID={`teacher-lesson-homework-${participant.studentId}`}
          title={message("tch.lesson.homework.title")}
          tone={HOMEWORK_STATUS_TONE[participant.latestHomework.status]}
        />
      ) : (
        <StatusCard
          body={message("tch.lesson.homework.none")}
          title={message("tch.lesson.homework.title")}
          tone="muted"
        />
      )}
      {participant.latestEvidence !== null ? (
        <GrowthSignal
          body={participant.latestEvidence.note}
          kind={message("tch.lesson.evidence.title")}
          supporting={message("growth.signal.source", {
            date: formatBelcantoDate(participant.latestEvidence.recordedAt),
          })}
          testID={`teacher-lesson-evidence-${participant.studentId}`}
          title={participant.latestEvidence.area}
        />
      ) : null}
      <BlockAction
        label={message("tch.lesson.openJournal")}
        onPress={() =>
          router.push({
            pathname: "/(protected)/journal/[occurrenceId]/[studentId]",
            params: { occurrenceId, studentId: participant.studentId },
          })
        }
        testID={`teacher-lesson-journal-${participant.studentId}`}
      />
      <BlockAction
        kind="secondary"
        label={message("rep.open")}
        onPress={() =>
          router.push({
            pathname: "/(protected)/repertoire",
            params: { studentId: participant.studentId },
          })
        }
        testID={`teacher-lesson-repertoire-${participant.studentId}`}
      />
      <BlockAction
        kind="secondary"
        label={message("growth.entry.title")}
        onPress={() =>
          router.push({
            pathname: "/(protected)/progress",
            params: { studentId: participant.studentId },
          })
        }
        testID={`teacher-lesson-growth-${participant.studentId}`}
      />
    </Fragment>
  );
}

export function TeacherLessonScreen() {
  const message = useMessage();
  const api = useApiClient();
  const params = useLocalSearchParams<{ lessonId?: string }>();
  const lessonId = typeof params.lessonId === "string" ? params.lessonId : "";

  const context = useAccountResource<LessonContextView>(async (accessToken) => {
    const lesson = await api.getLesson(accessToken, lessonId);
    const participants = await Promise.all(
      lesson.students.map(async (student): Promise<ParticipantContext> => {
        const [journal, homework, evidence] = await Promise.all([
          api
            .getJournal(accessToken, lesson.id, student.studentId)
            .catch((cause: unknown) => {
              if (cause instanceof ApiError && cause.code === "NOT_FOUND") return null;
              throw cause;
            }),
          api.listStudentHomework(accessToken, student.studentId),
          api.listProgressEvidence(accessToken, student.studentId),
        ]);
        return {
          studentId: student.studentId,
          fullName: student.fullName,
          journal,
          latestHomework: homework[0] ?? null,
          latestEvidence: evidence[0] ?? null,
        };
      }),
    );
    return { lesson, participants };
  });

  const view = context.value;
  if (view === null) {
    return (
      <AccountScreenShell navigation={<AccountNav active="today" />} testID="teacher-lesson-loading">
        {context.error !== null ? (
          <ErrorNotice
            actionLabel={message("common.retry")}
            body={apiErrorMessage(context.error)}
            onAction={() => void context.reload()}
            title={message("tch.today.title")}
          />
        ) : null}
      </AccountScreenShell>
    );
  }

  const { lesson, participants } = view;
  const group = participants.length > 1;

  return (
    <AccountScreenShell
      navigation={<AccountNav active="today" />}
      refreshControl={
        <RefreshControl
          onRefresh={() => void context.reload()}
          refreshing={context.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="teacher-lesson"
    >
      <ScreenHeading
        eyebrow={message(lessonPhaseKey(lesson))}
        subtitle={formatLine(lesson, message)}
        title={lesson.title}
      />
      {group ? (
        <>
          <Text style={styles.sectionTitle}>{message("tch.lesson.roster.title")}</Text>
          {participants.map((participant) => (
            <StatusRow
              key={participant.studentId}
              onPress={() =>
                router.push({
                  pathname: "/(protected)/journal/[occurrenceId]/[studentId]",
                  params: { occurrenceId: lesson.id, studentId: participant.studentId },
                })
              }
              status={
                participant.latestHomework !== null
                  ? message("tch.lesson.homework.latest", {
                      status: homeworkStatusLabel(
                        participant.latestHomework.status,
                        message,
                      ),
                    })
                  : message("tch.lesson.homework.none")
              }
              subtitle={message("tch.lesson.openJournal")}
              testID={`teacher-lesson-roster-${participant.studentId}`}
              title={participant.fullName}
              tone={
                participant.latestHomework !== null
                  ? HOMEWORK_STATUS_TONE[participant.latestHomework.status]
                  : "muted"
              }
            />
          ))}
          <StatusCard
            body={message("tch.lesson.privacy.body")}
            status={message("tch.lesson.privacy.status")}
            title={message("tch.lesson.privacy.title")}
            tone="info"
          />
        </>
      ) : (
        participants.map((participant) => (
          <ParticipantBlocks
            key={participant.studentId}
            message={message}
            occurrenceId={lesson.id}
            participant={participant}
          />
        ))
      )}
      <AttendanceBlock lessonId={lesson.id} students={lesson.students} />
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  sectionTitle: { color: semantic.textPrimary, ...typeStyles.headingM },
});
