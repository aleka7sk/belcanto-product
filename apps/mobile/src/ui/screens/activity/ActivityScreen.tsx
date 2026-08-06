import { router } from "expo-router";
import { useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { useApiClient } from "@/api";
import type {
  ActivityEntry,
  IsoDateTime,
  NotificationCategory,
} from "@/api/contracts";
import { NOTIFICATION_CATEGORIES } from "@/api/contracts";
import { useMessage, type MessageFormatter, type MessageKey } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusRow,
} from "../../patterns/accountPatterns";
import { AreaChip } from "../../patterns/journalPatterns";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { almatyDayRange } from "../../teacherToday";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * Activity feed (Figma ACT-01/02). Entries come from the outbox worker;
 * filters never delete and never mark anything read (ACT-02 verbatim).
 * Every entry deep-links into the object it talks about.
 */

const KIND_KEYS: Record<string, MessageKey> = {
  JournalPublished: "act.kind.JournalPublished",
  JournalCorrected: "act.kind.JournalCorrected",
  HomeworkAssigned: "act.kind.HomeworkAssigned",
  HomeworkStarted: "act.kind.HomeworkStarted",
  HomeworkSubmitted: "act.kind.HomeworkSubmitted",
  HomeworkReviewed: "act.kind.HomeworkReviewed",
  HomeworkCompleted: "act.kind.HomeworkCompleted",
  HomeworkCancelled: "act.kind.HomeworkCancelled",
  GoalCompleted: "act.kind.GoalCompleted",
  AchievementAwarded: "act.kind.AchievementAwarded",
  SongStageChanged: "act.kind.SongStageChanged",
  AttendanceAbsenceRecorded: "act.kind.AttendanceAbsenceRecorded",
};

export function activityKindTitle(kind: string, message: MessageFormatter): string {
  const key = KIND_KEYS[kind];
  return message(key ?? "act.kind.unknown");
}

function payloadString(entry: ActivityEntry, key: string): string | null {
  const value = entry.payload[key];
  return typeof value === "string" && value !== "" ? value : null;
}

/** Deep link per kind; falls back to staying on the feed. */
export function openActivityTarget(entry: ActivityEntry): boolean {
  const homeworkId = payloadString(entry, "homeworkId");
  const occurrenceId = payloadString(entry, "occurrenceId");
  const studentId = payloadString(entry, "studentId");
  switch (entry.kind) {
    case "JournalPublished":
    case "JournalCorrected":
      if (occurrenceId !== null && studentId !== null) {
        router.push({
          pathname: "/(protected)/journal/[occurrenceId]/[studentId]",
          params: { occurrenceId, studentId },
        });
        return true;
      }
      return false;
    case "HomeworkAssigned":
    case "HomeworkStarted":
    case "HomeworkSubmitted":
    case "HomeworkReviewed":
    case "HomeworkCompleted":
    case "HomeworkCancelled":
      if (homeworkId !== null) {
        router.push({
          pathname: "/(protected)/practice/[homeworkId]",
          params: { homeworkId },
        });
        return true;
      }
      return false;
    case "GoalCompleted":
    case "AchievementAwarded":
      router.push("/(protected)/progress");
      return true;
    case "SongStageChanged":
      router.push("/(protected)/repertoire");
      return true;
    case "AttendanceAbsenceRecorded":
      if (occurrenceId !== null) {
        router.push({
          pathname: "/(protected)/teacher/lesson/[lessonId]",
          params: { lessonId: occurrenceId },
        });
        return true;
      }
      return false;
    default:
      return false;
  }
}

function splitByToday(entries: readonly ActivityEntry[]): {
  today: ActivityEntry[];
  earlier: ActivityEntry[];
} {
  const { from } = almatyDayRange(new Date());
  const today: ActivityEntry[] = [];
  const earlier: ActivityEntry[] = [];
  for (const entry of entries) {
    if (new Date(entry.occurredAt).getTime() >= from.getTime()) {
      today.push(entry);
    } else {
      earlier.push(entry);
    }
  }
  return { today, earlier };
}

function nowIso(): IsoDateTime {
  return new Date().toISOString() as IsoDateTime;
}

export function ActivityScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const feed = useAccountResource((accessToken) => api.activityFeed(accessToken));
  const [filter, setFilter] = useState<NotificationCategory | null>(null);
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const view = feed.value;
  const entries = (view?.entries ?? []).filter(
    (entry) => filter === null || entry.category === filter,
  );
  const { today, earlier } = splitByToday(entries);

  const markRead = async () => {
    setActionError(null);
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.markActivityRead(accessToken, { upTo: nowIso() }),
      );
      await feed.reload();
    } catch (cause) {
      setActionError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  const renderEntry = (entry: ActivityEntry) => (
    <StatusRow
      key={entry.id}
      onPress={() => {
        openActivityTarget(entry);
      }}
      status={
        entry.readAt === undefined
          ? message("act.new")
          : message(`act.category.${entry.category}`)
      }
      subtitle={formatBelcantoDate(entry.occurredAt)}
      testID={`activity-${entry.id}`}
      title={activityKindTitle(entry.kind, message)}
      tone={entry.readAt === undefined ? "info" : "muted"}
    />
  );

  return (
    <AccountScreenShell navigation={<AccountNav />} testID="activity-feed">
      <ScreenHeading
        eyebrow={message("act.eyebrow")}
        subtitle={
          view !== null && view.unreadCount > 0
            ? message("act.unread", { count: view.unreadCount })
            : message("act.subtitle")
        }
        title={message("act.title")}
      />
      {feed.error !== null ? (
        <InlineNotice
          body={apiErrorMessage(feed.error)}
          title={message("common.retry")}
          tone="error"
        />
      ) : null}
      <View style={styles.chips}>
        <AreaChip
          accent={semantic.accentViolet}
          active={filter === null}
          label={message("act.filter.all")}
          onPress={() => setFilter(null)}
          testID="activity-filter-all"
        />
        {NOTIFICATION_CATEGORIES.map((category) => (
          <AreaChip
            accent={category === "important" ? semantic.feedbackWarning : semantic.accentCyan}
            active={filter === category}
            key={category}
            label={message(`act.category.${category}`)}
            onPress={() =>
              setFilter((current) => (current === category ? null : category))
            }
            testID={`activity-filter-${category}`}
          />
        ))}
      </View>
      {view !== null && entries.length === 0 ? (
        <Text style={styles.empty}>{message("act.empty")}</Text>
      ) : null}
      {actionError !== null ? (
        <InlineNotice body={actionError} title={message("common.retry")} tone="error" />
      ) : null}
      {today.length > 0 ? (
        <Text style={styles.sectionTitle}>{message("act.today").toUpperCase()}</Text>
      ) : null}
      {today.map(renderEntry)}
      {earlier.length > 0 ? (
        <Text style={styles.sectionTitle}>{message("act.earlier").toUpperCase()}</Text>
      ) : null}
      {earlier.map(renderEntry)}
      {view !== null && view.unreadCount > 0 ? (
        <BlockAction
          busy={busy}
          kind="secondary"
          label={message("act.markRead")}
          onPress={() => void markRead()}
          testID="activity-mark-read"
        />
      ) : null}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  chips: { flexDirection: "row", flexWrap: "wrap", gap: space.s2 },
  sectionTitle: {
    color: semantic.textGold,
    marginTop: space.s2,
    ...typeStyles.labelM,
  },
  empty: { color: semantic.textSecondary, ...typeStyles.bodyS },
});
