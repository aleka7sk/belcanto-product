import { router, useLocalSearchParams } from "expo-router";
import { RefreshControl } from "react-native";

import { useApiClient } from "@/api";
import type { HomeworkAssignment, HomeworkStatus } from "@/api/contracts";
import { useMessage, type MessageFormatter } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import { HOMEWORK_STATUS_TONE } from "../../patterns/practicePatterns";
import { domainAccent } from "../../domainAccent";
import { semantic } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * Practice hub (Figma STU-PRACTICE-01/16, flow E). Lists the Student's
 * homework newest-first; staff open the same hub for a Student via the
 * studentId parameter. Empty state deliberately promises nothing that
 * is not built — no streaks, no repertoire.
 */

export function homeworkStatusLabel(
  status: HomeworkStatus,
  message: MessageFormatter,
): string {
  return message(`prac.status.${status}`);
}

export function PracticeHubScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { state } = useSession();
  const params = useLocalSearchParams<{ studentId?: string }>();
  const paramStudentId = typeof params.studentId === "string" ? params.studentId : null;
  const studentId = paramStudentId ?? state.bootstrap?.studentId ?? null;

  const homework = useAccountResource((accessToken) =>
    studentId === null
      ? Promise.resolve<HomeworkAssignment[]>([])
      : api.listStudentHomework(accessToken, studentId),
  );

  if (studentId === null) {
    return (
      <AccountScreenShell navigation={<AccountNav />} testID="practice-guard">
        <InlineNotice
          body={message("prac.guard.body")}
          title={message("prac.guard.title")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const list = homework.value ?? [];
  const loading = homework.value === null && homework.error === null;
  const latest = list[0];
  const empty = homework.value !== null && list.length === 0;

  const openHomework = (assignment: HomeworkAssignment) =>
    router.push({
      pathname: "/(protected)/practice/[homeworkId]",
      params: { homeworkId: assignment.id },
    });

  return (
    <AccountScreenShell
      navigation={<AccountNav active="practice" />}
      refreshControl={
        <RefreshControl
          onRefresh={() => void homework.reload()}
          refreshing={homework.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="practice-hub"
    >
      <ScreenHeading
        accent={domainAccent("practice")}
        eyebrow={message("prac.eyebrow")}
        subtitle={
          loading
            ? message("common.loading")
            : empty || latest === undefined
              ? message("prac.empty.body")
              : message("prac.hub.subtitle", {
                  count: list.length,
                  date: formatBelcantoDate(latest.updatedAt),
                })
        }
        title={empty ? message("prac.empty.title") : message("prac.hub.title")}
      />
      {homework.error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={apiErrorMessage(homework.error)}
          onAction={() => void homework.reload()}
          title={message("prac.hub.title")}
        />
      ) : null}
      {empty ? (
        <>
          <StatusCard
            body={message("prac.empty.card.body")}
            status={message("prac.empty.card.status")}
            title={message("prac.empty.card.title")}
            tone="info"
          />
          {paramStudentId === null ? (
            <BlockAction
              kind="secondary"
              label={message("prac.empty.action")}
              onPress={() => router.push("/(protected)/progress")}
            />
          ) : null}
        </>
      ) : null}
      <BlockAction
        kind="secondary"
        label={message("rep.open")}
        onPress={() =>
          router.push({
            pathname: "/(protected)/repertoire",
            ...(paramStudentId !== null
              ? { params: { studentId: paramStudentId } }
              : {}),
          })
        }
        testID="practice-open-repertoire"
      />
      {empty ? null : (
        list.map((assignment) => (
          <StatusRow
            key={assignment.id}
            onPress={() => openHomework(assignment)}
            status={homeworkStatusLabel(assignment.status, message)}
            subtitle={
              assignment.dueAt !== undefined
                ? message("prac.hub.entry", {
                    teacher: assignment.teacher.fullName,
                    date: formatBelcantoDate(assignment.dueAt),
                  })
                : message("prac.hub.entry.noDue", {
                    teacher: assignment.teacher.fullName,
                  })
            }
            testID={`practice-homework-${assignment.id}`}
            title={assignment.goal}
            tone={HOMEWORK_STATUS_TONE[assignment.status]}
          />
        ))
      )}
    </AccountScreenShell>
  );
}
