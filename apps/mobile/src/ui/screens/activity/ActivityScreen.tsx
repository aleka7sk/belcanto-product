import { router } from "expo-router";
import { useState } from "react";
import { RefreshControl, StyleSheet, Text, View } from "react-native";

import { useApiClient } from "@/api";
import type {
  ActivityEntry,
  IsoDateTime,
  NotificationCategory,
} from "@/api/contracts";
import { NOTIFICATION_CATEGORIES } from "@/api/contracts";
import { useMessage, type MessageFormatter, type MessageKey } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice } from "../../components";
import {
  BlockAction,
  ScreenHeading,
  StatusRow,
} from "../../patterns/accountPatterns";
import { AreaChip } from "../../patterns/journalPatterns";
import { ScreenList } from "../../screen";
import { domainAccent } from "../../domainAccent";
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
  CommunityPostPublished: "act.kind.CommunityPostPublished",
  CommunityAnnouncementPublished: "act.kind.CommunityAnnouncementPublished",
  CommunityReportFiled: "act.kind.CommunityReportFiled",
  AssessmentPublished: "act.kind.AssessmentPublished",
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
    case "CommunityPostPublished":
    case "CommunityAnnouncementPublished": {
      const postId = payloadString(entry, "postId");
      if (postId !== null) {
        router.push({
          pathname: "/(protected)/community/[postId]",
          params: { postId },
        });
        return true;
      }
      return false;
    }
    case "CommunityReportFiled":
      router.push("/(protected)/community/moderation");
      return true;
    case "AssessmentPublished": {
      const assessmentId = payloadString(entry, "assessmentId");
      if (assessmentId !== null) {
        router.push({
          pathname: "/(protected)/assessment/[assessmentId]",
          params: { assessmentId },
        });
        return true;
      }
      return false;
    }
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

export type ActivityRow =
  | { rowKind: "section"; id: string; labelKey: "act.today" | "act.earlier" }
  | { rowKind: "entry"; id: string; entry: ActivityEntry };

/** Flatten the two day sections into one virtualizable row list. */
export function buildActivityRows(
  today: readonly ActivityEntry[],
  earlier: readonly ActivityEntry[],
): ActivityRow[] {
  const rows: ActivityRow[] = [];
  if (today.length > 0) {
    rows.push({ rowKind: "section", id: "section-today", labelKey: "act.today" });
    for (const entry of today) rows.push({ rowKind: "entry", id: entry.id, entry });
  }
  if (earlier.length > 0) {
    rows.push({ rowKind: "section", id: "section-earlier", labelKey: "act.earlier" });
    for (const entry of earlier) rows.push({ rowKind: "entry", id: entry.id, entry });
  }
  return rows;
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

  const rows = buildActivityRows(today, earlier);

  return (
    <ScreenList<ActivityRow>
      data={rows}
      keyExtractor={(row) => row.id}
      ListEmptyComponent={
        view !== null ? (
          <Text style={styles.empty}>{message("act.empty")}</Text>
        ) : (
          <Text style={styles.empty}>{message("common.loading")}</Text>
        )
      }
      ListFooterComponent={
        view !== null && view.unreadCount > 0 ? (
          <BlockAction
            busy={busy}
            kind="secondary"
            label={message("act.markRead")}
            onPress={() => void markRead()}
            testID="activity-mark-read"
          />
        ) : null
      }
      ListHeaderComponent={
        <View style={styles.header}>
          <ScreenHeading
            accent={domainAccent("activity")}
            eyebrow={message("act.eyebrow")}
            subtitle={
              view !== null && view.unreadCount > 0
                ? message("act.unread", { count: view.unreadCount })
                : message("act.subtitle")
            }
            title={message("act.title")}
          />
          {feed.error !== null ? (
            <ErrorNotice
              actionLabel={message("common.retry")}
              body={apiErrorMessage(feed.error)}
              onAction={() => void feed.reload()}
              title={message("act.title")}
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
          {actionError !== null ? (
            <InlineNotice body={actionError} title={message("common.retry")} tone="error" />
          ) : null}
        </View>
      }
      navigation={<AccountNav />}
      refreshControl={
        <RefreshControl
          onRefresh={() => void feed.reload()}
          refreshing={feed.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      renderItem={({ item }) =>
        item.rowKind === "section" ? (
          <Text style={styles.sectionTitle}>{message(item.labelKey).toUpperCase()}</Text>
        ) : (
          <StatusRow
            onPress={() => {
              openActivityTarget(item.entry);
            }}
            status={
              item.entry.readAt === undefined
                ? message("act.new")
                : message(`act.category.${item.entry.category}`)
            }
            subtitle={formatBelcantoDate(item.entry.occurredAt)}
            testID={`activity-${item.entry.id}`}
            title={activityKindTitle(item.entry.kind, message)}
            tone={item.entry.readAt === undefined ? "info" : "muted"}
          />
        )
      }
      testID="activity-feed"
    />
  );
}

const styles = StyleSheet.create({
  header: { gap: space.s3 },
  chips: { flexDirection: "row", flexWrap: "wrap", gap: space.s2 },
  sectionTitle: {
    color: semantic.textGold,
    marginTop: space.s2,
    ...typeStyles.labelM,
  },
  empty: { color: semantic.textSecondary, ...typeStyles.bodyS },
});
