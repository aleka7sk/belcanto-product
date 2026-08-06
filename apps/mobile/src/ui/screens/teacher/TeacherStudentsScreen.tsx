import { router, useLocalSearchParams } from "expo-router";
import { useState } from "react";
import { Pressable, StyleSheet, Text, View } from "react-native";

import { useApiClient } from "@/api";
import type { Assessment, StudentDirectoryItem } from "@/api/contracts";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice } from "../../components";
import {
  AccountScreenShell,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import { semantic, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource, useWorkingRole } from "../account/shared";

/**
 * TCH-STUDENTS-01/02 + TCH-REVIEW-01: the Teacher's roster with the
 * review queue. Statuses show organisational signals, never «strong»
 * and «weak» students (privacy invariant, verbatim from Page 27).
 * Analytics ships with the administrator workspace — the segment says
 * so honestly instead of pretending.
 */

type Segment = "students" | "review" | "analytics";

/** Review signal derived from real data only: latest published assessment. */
export function latestPublished(assessments: readonly Assessment[]): Assessment | null {
  for (const assessment of assessments) {
    if (assessment.status === "published") return assessment;
  }
  return null;
}

export function TeacherStudentsScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const workingRole = useWorkingRole();
  const params = useLocalSearchParams<{ segment?: string }>();
  const [segment, setSegment] = useState<Segment>(
    params.segment === "review" ? "review" : "students",
  );
  const [reviewByStudent, setReviewByStudent] = useState<
    Record<string, Assessment | null>
  >({});
  const [reviewError, setReviewError] = useState<string | null>(null);

  const students = useAccountResource((accessToken) => api.listStudents(accessToken, {}));

  if (workingRole !== "Teacher" && workingRole !== "Owner" && workingRole !== "Administrator") {
    return (
      <AccountScreenShell navigation={<AccountNav active="today" />} testID="teacher-students-guard">
        <InlineNotice
          body={message("asmt.guardBody")}
          title={message("asmt.guardTitle")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const roster = students.value ?? [];

  const loadReviewSignals = async () => {
    setReviewError(null);
    try {
      const entries = await runAuthenticated(async (accessToken) => {
        const result: Record<string, Assessment | null> = {};
        for (const item of roster) {
          const history = await api.listStudentAssessments(accessToken, item.studentId);
          result[item.studentId] = latestPublished(history);
        }
        return result;
      });
      setReviewByStudent(entries);
    } catch (cause) {
      setReviewError(apiErrorMessage(cause));
    }
  };

  const segments: { key: Segment; label: string }[] = [
    { key: "students", label: message("asmt.segment.students") },
    { key: "review", label: message("asmt.segment.review") },
    { key: "analytics", label: message("asmt.segment.analytics") },
  ];

  return (
    <AccountScreenShell navigation={<AccountNav active="today" />} testID="teacher-students">
      <ScreenHeading
        eyebrow={message("asmt.roster.eyebrow")}
        subtitle={message("asmt.roster.subtitle")}
        title={message("asmt.roster.title")}
      />
      <View style={styles.segments}>
        {segments.map((entry) => (
          <Pressable
            accessibilityLabel={entry.label}
            accessibilityRole="button"
            key={entry.key}
            onPress={() => {
              setSegment(entry.key);
              if (entry.key === "review" && Object.keys(reviewByStudent).length === 0) {
                void loadReviewSignals();
              }
            }}
            style={[styles.segment, segment === entry.key && styles.segmentActive]}
            testID={`teacher-segment-${entry.key}`}
          >
            <Text
              style={[styles.segmentLabel, segment === entry.key && styles.segmentLabelActive]}
            >
              {entry.label}
            </Text>
          </Pressable>
        ))}
      </View>
      {students.error !== null ? (
        <InlineNotice
          body={apiErrorMessage(students.error)}
          title={message("common.retry")}
          tone="error"
        />
      ) : null}
      {segment === "analytics" ? (
        <StatusCard
          body={message("asmt.analytics.body")}
          status={message("asmt.analytics.footer")}
          title={message("asmt.analytics.title")}
          tone="muted"
        />
      ) : segment === "review" ? (
        <>
          {reviewError !== null ? (
            <InlineNotice body={reviewError} title={message("common.retry")} tone="error" />
          ) : null}
          {roster.map((item: StudentDirectoryItem) => {
            const latest = reviewByStudent[item.studentId];
            return (
              <StatusRow
                key={item.studentId}
                onPress={() =>
                  router.push({
                    pathname: "/(protected)/teacher/assessment/create",
                    params: { studentId: item.studentId },
                  })
                }
                status={
                  latest === undefined
                    ? message("asmt.review.loading")
                    : latest === null
                      ? message("asmt.review.none")
                      : message("asmt.review.latest", {
                          date: formatBelcantoDate(latest.createdAt),
                        })
                }
                subtitle={message("asmt.review.action")}
                testID={`review-${item.studentId}`}
                title={item.fullName}
                tone={latest === null ? "warning" : "info"}
              />
            );
          })}
        </>
      ) : (
        <>
          {roster.map((item: StudentDirectoryItem) => (
            <StatusRow
              key={item.studentId}
              onPress={() =>
                router.push({
                  pathname: "/(protected)/student/[studentId]",
                  params: { studentId: item.studentId, staff: "1" },
                })
              }
              status={message("asmt.roster.open")}
              subtitle={item.primaryTeacher.fullName}
              testID={`roster-${item.studentId}`}
              title={item.fullName}
              tone="info"
            />
          ))}
          <StatusCard
            body={message("asmt.notRating.body")}
            status={message("asmt.notRating.footer")}
            title={message("asmt.notRating.title")}
            tone="muted"
          />
        </>
      )}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  segments: { flexDirection: "row", gap: 6 },
  segment: {
    alignItems: "center",
    backgroundColor: semantic.bgRaised,
    borderRadius: 14,
    flex: 1,
    justifyContent: "center",
    minHeight: 38,
  },
  segmentActive: { backgroundColor: semantic.bgAction },
  segmentLabel: {
    color: semantic.textSecondary,
    ...typeStyles.labelM,
  },
  segmentLabelActive: { color: semantic.textOnAction },
});
