import { router } from "expo-router";
import { RefreshControl, StyleSheet, Text } from "react-native";

import { useApiClient } from "@/api";
import { useMessage } from "@/i18n";
import { ErrorNotice, InlineNotice } from "../components";
import {
  AccountScreenShell,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../patterns/accountPatterns";
import { semantic, typeStyles } from "../tokens";
import { apiErrorMessage } from "../viewModels";
import { AccountNav, useAccountResource, useWorkingRole } from "./account/shared";

/**
 * Administrator operations home (Figma AOP-01/02): «Сегодня — всё в
 * порядке» or the attention list, every signal derived from stored
 * rows. Deep operational flows (conflict centre, occurrence-scope
 * editing, event participant management) are recorded future slices —
 * this screen links only what exists.
 */

export function OperationsScreen() {
  const message = useMessage();
  const api = useApiClient();
  const workingRole = useWorkingRole();
  const manager = workingRole === "Owner" || workingRole === "Administrator";
  const summary = useAccountResource((accessToken) => api.operationsSummary(accessToken));

  if (!manager) {
    return (
      <AccountScreenShell navigation={<AccountNav active="operations" />} testID="operations-guard">
        <InlineNotice
          body={message("adm.guardBody")}
          title={message("adm.guardTitle")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const view = summary.value;
  const attention =
    view !== null &&
    (view.pendingReschedules > 0 ||
      view.newCommunityReports > 0 ||
      view.pastLessonsNoAttendance > 0 ||
      view.pendingDeletionRequests > 0);

  return (
    <AccountScreenShell
      navigation={<AccountNav active="operations" />}
      refreshControl={
        <RefreshControl
          onRefresh={() => void summary.reload()}
          refreshing={summary.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="operations-home"
    >
      <ScreenHeading
        eyebrow={message("adm.eyebrow")}
        subtitle={
          view === null
            ? message("common.loading")
            : attention
              ? message("adm.subtitle.attention")
              : message("adm.subtitle.calm")
        }
        title={message("adm.title")}
      />
      {summary.error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={apiErrorMessage(summary.error)}
          onAction={() => void summary.reload()}
          title={message("adm.title")}
        />
      ) : null}
      {view !== null ? (
        <>
          <Text style={styles.section}>{message("adm.section.today").toUpperCase()}</Text>
          <StatusCard
            body={message("adm.today.body", {
              lessons: view.lessonsToday,
              events: view.upcomingEventsWeek,
            })}
            status={message("adm.today.footer", { students: view.activeStudents })}
            title={message("adm.today.title")}
            tone="info"
          />
          <Text style={styles.section}>{message("adm.section.signals").toUpperCase()}</Text>
          <StatusRow
            onPress={() => router.push("/(protected)/lessons")}
            status={String(view.pastLessonsNoAttendance)}
            subtitle={message("adm.signal.attendanceBody")}
            testID="ops-attendance"
            title={message("adm.signal.attendance")}
            tone={view.pastLessonsNoAttendance > 0 ? "warning" : "success"}
          />
          <StatusRow
            status={String(view.draftJournals)}
            subtitle={message("adm.signal.journalsBody")}
            testID="ops-journals"
            title={message("adm.signal.journals")}
            tone={view.draftJournals > 0 ? "warning" : "success"}
          />
          <StatusRow
            onPress={() => router.push("/(protected)/community/moderation")}
            status={String(view.newCommunityReports)}
            subtitle={message("adm.signal.reportsBody")}
            testID="ops-reports"
            title={message("adm.signal.reports")}
            tone={view.newCommunityReports > 0 ? "warning" : "success"}
          />
          <StatusRow
            status={String(view.pendingReschedules)}
            subtitle={message("adm.signal.reschedulesBody")}
            testID="ops-reschedules"
            title={message("adm.signal.reschedules")}
            tone={view.pendingReschedules > 0 ? "warning" : "success"}
          />
          <StatusRow
            status={String(view.pendingDeletionRequests)}
            subtitle={message("adm.signal.deletionsBody")}
            testID="ops-deletions"
            title={message("adm.signal.deletions")}
            tone={view.pendingDeletionRequests > 0 ? "warning" : "muted"}
          />
          <Text style={styles.section}>{message("adm.section.manage").toUpperCase()}</Text>
          <StatusRow
            onPress={() => router.push("/(protected)/series")}
            status={message("adm.manage.seriesCount", { count: view.activeSeries })}
            subtitle={message("adm.manage.seriesBody")}
            testID="ops-series"
            title={message("adm.manage.series")}
            tone="info"
          />
          <StatusRow
            onPress={() => router.push("/(protected)/lessons")}
            subtitle={message("adm.manage.lessonsBody")}
            testID="ops-lessons"
            title={message("adm.manage.lessons")}
            tone="info"
          />
          <StatusRow
            onPress={() => router.push("/(protected)/create-student")}
            subtitle={message("adm.manage.studentsBody")}
            testID="ops-students"
            title={message("adm.manage.students")}
            tone="info"
          />
          <StatusCard
            body={message("adm.future.body")}
            status={message("adm.future.footer")}
            title={message("adm.future.title")}
            tone="muted"
          />
        </>
      ) : null}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  section: { color: semantic.textGold, ...typeStyles.labelM },
});
