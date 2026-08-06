import { router, useLocalSearchParams } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import { RefreshControl, StyleSheet, View } from "react-native";

import { useApiClient } from "@/api";
import type { Assessment, StudentDirectoryItem } from "@/api/contracts";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice } from "../../components";
import {
  AccountScreenShell,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import { ScreenList } from "../../screen";
import { SegmentedControl } from "../../segmentedControl";
import { domainAccent } from "../../domainAccent";
import { semantic, space } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource, useWorkingRole } from "../account/shared";

/**
 * TCH-STUDENTS-01/02 + TCH-REVIEW-01: the Teacher's roster with the
 * review queue. Statuses show organisational signals, never «strong»
 * and «weak» students (privacy invariant, verbatim from Page 27).
 * Analytics ships with the administrator workspace — the segment says
 * so honestly instead of pretending. The active segment lives in the
 * URL, so the «Ревью» tab both opens and switches it.
 */

type Segment = "students" | "review" | "analytics";

export function parseStudentsSegment(raw: string | string[] | undefined): Segment {
  const value = Array.isArray(raw) ? raw[0] : raw;
  if (value === "review" || value === "analytics") return value;
  return "students";
}

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
  const segment = parseStudentsSegment(params.segment);
  const [reviewByStudent, setReviewByStudent] = useState<
    Record<string, Assessment | null>
  >({});
  const [reviewLoaded, setReviewLoaded] = useState(false);
  const [reviewError, setReviewError] = useState<string | null>(null);

  const students = useAccountResource((accessToken) => api.listStudents(accessToken, {}));
  const roster = students.value;

  const loadReviewSignals = useCallback(async () => {
    if (roster === null) return;
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
      setReviewLoaded(true);
    } catch (cause) {
      setReviewError(apiErrorMessage(cause));
    }
  }, [api, roster, runAuthenticated]);

  useEffect(() => {
    if (segment !== "review" || reviewLoaded || roster === null) return;
    let active = true;
    queueMicrotask(() => {
      if (active) void loadReviewSignals();
    });
    return () => {
      active = false;
    };
  }, [segment, reviewLoaded, roster, loadReviewSignals]);

  if (workingRole !== "Teacher" && workingRole !== "Owner" && workingRole !== "Administrator") {
    return (
      <AccountScreenShell navigation={<AccountNav active="students" />} testID="teacher-students-guard">
        <InlineNotice
          body={message("asmt.guardBody")}
          title={message("asmt.guardTitle")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const list = roster ?? [];

  const header = (
    <View style={styles.header}>
      <ScreenHeading
        accent={domainAccent("teacher")}
        eyebrow={message("asmt.roster.eyebrow")}
        subtitle={message("asmt.roster.subtitle")}
        title={message("asmt.roster.title")}
      />
      <SegmentedControl<Segment>
        accessibilityLabel={message("asmt.roster.title")}
        activeColor={domainAccent("teacher")}
        active={segment}
        items={[
          { key: "students", label: message("asmt.segment.students") },
          { key: "review", label: message("asmt.segment.review") },
          { key: "analytics", label: message("asmt.segment.analytics") },
        ]}
        onSelect={(key) => router.setParams({ segment: key })}
        testIDPrefix="teacher-segment"
      />
      {students.error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={apiErrorMessage(students.error)}
          onAction={() => void students.reload()}
          title={message("asmt.roster.title")}
        />
      ) : null}
      {segment === "analytics" ? (
        <StatusCard
          body={message("asmt.analytics.body")}
          status={message("asmt.analytics.footer")}
          title={message("asmt.analytics.title")}
          tone="muted"
        />
      ) : null}
      {segment === "review" && reviewError !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={reviewError}
          onAction={() => void loadReviewSignals()}
          title={message("asmt.segment.review")}
        />
      ) : null}
    </View>
  );

  return (
    <ScreenList<StudentDirectoryItem>
      data={segment === "analytics" ? [] : list}
      keyExtractor={(item) => item.studentId}
      ListFooterComponent={
        segment === "students" ? (
          <View style={styles.footer}>
            <StatusCard
              body={message("asmt.notRating.body")}
              status={message("asmt.notRating.footer")}
              title={message("asmt.notRating.title")}
              tone="muted"
            />
          </View>
        ) : null
      }
      ListHeaderComponent={header}
      navigation={<AccountNav active={segment === "review" ? "review" : "students"} />}
      refreshControl={
        <RefreshControl
          onRefresh={() => void students.reload()}
          refreshing={students.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      renderItem={({ item }) => {
        if (segment === "review") {
          const latest = reviewByStudent[item.studentId];
          return (
            <StatusRow
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
        }
        return (
          <StatusRow
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
        );
      }}
      testID="teacher-students"
    />
  );
}

const styles = StyleSheet.create({
  footer: { marginTop: space.s2 },
  header: { gap: space.s3 },
});
