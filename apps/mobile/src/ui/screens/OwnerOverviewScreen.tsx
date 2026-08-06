import { router, useLocalSearchParams } from "expo-router";
import { StyleSheet, Text, View } from "react-native";

import { useApiClient } from "@/api";
import { useMessage } from "@/i18n";
import { ErrorNotice, InlineNotice } from "../components";
import {
  AccountScreenShell,
  StatusCard,
  StatusRow,
} from "../patterns/accountPatterns";
import { PremiumContextHero } from "../patterns/premiumContextHero";
import { domainAccent } from "../domainAccent";
import { SegmentedControl } from "../segmentedControl";
import { semantic, space, typeStyles } from "../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../viewModels";
import { AccountNav, useAccountResource, useWorkingRole } from "./account/shared";

/**
 * Owner overview and governance (Figma OWN-01/17/18). Every number is
 * a derived count with its source named; analytics with defined
 * methodologies (OWN-03..14) ship as their own slice once metric
 * definitions are approved — the segment says so instead of showing
 * invented trends (OWN-23 principle: маленькая выборка не превращается
 * в уверенный вывод).
 */

type Segment = "overview" | "governance" | "analytics";

export function parseOverviewSegment(raw: string | string[] | undefined): Segment {
  const value = Array.isArray(raw) ? raw[0] : raw;
  if (value === "analytics" || value === "governance") return value;
  return "overview";
}

export function OwnerOverviewScreen() {
  const message = useMessage();
  const api = useApiClient();
  const workingRole = useWorkingRole();
  const params = useLocalSearchParams<{ segment?: string }>();
  const segment = parseOverviewSegment(params.segment);

  const summary = useAccountResource((accessToken) => api.operationsSummary(accessToken));
  const policies = useAccountResource((accessToken) => api.listPolicies(accessToken));

  if (workingRole !== "Owner" && workingRole !== "Administrator") {
    return (
      <AccountScreenShell navigation={<AccountNav active="overview" />} testID="overview-guard">
        <InlineNotice
          body={message("own.guardBody")}
          title={message("own.guardTitle")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const view = summary.value;

  return (
    <AccountScreenShell
      navigation={
        <AccountNav active={segment === "analytics" ? "analytics" : "overview"} />
      }
      testID="owner-overview"
    >
      <PremiumContextHero
        body={message("own.hero.body")}
        eyebrow={message("own.hero.eyebrow")}
        {...(view === null
          ? {}
          : {
              metric: message("own.hero.metric", {
                students: view.activeStudents,
                lessons: view.lessonsToday,
              }),
            })}
        role="Owner"
        title={message("own.hero.title")}
      />
      <SegmentedControl<Segment>
        activeColor={domainAccent("overview")}
        accessibilityLabel={message("own.hero.title")}
        active={segment}
        items={[
          { key: "overview", label: message("own.segment.overview") },
          { key: "governance", label: message("own.segment.governance") },
          { key: "analytics", label: message("own.segment.analytics") },
        ]}
        onSelect={(key) => router.setParams({ segment: key })}
        testIDPrefix="owner-segment"
      />
      {summary.error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={apiErrorMessage(summary.error)}
          onAction={() => void summary.reload()}
          title={message("own.hero.title")}
        />
      ) : null}
      {segment === "analytics" ? (
        <>
          <StatusCard
            body={message("own.analytics.body")}
            status={message("own.analytics.footer")}
            title={message("own.analytics.title")}
            tone="muted"
          />
          <StatusCard
            body={message("own.analytics.principleBody")}
            status={message("own.analytics.principleFooter")}
            title={message("own.analytics.principleTitle")}
            tone="info"
          />
        </>
      ) : segment === "governance" ? (
        <>
          <Text style={styles.section}>{message("own.policies.title").toUpperCase()}</Text>
          {(policies.value ?? []).map((policy) => (
            <StatusRow
              key={policy.id}
              status={
                policy.acceptedAt !== undefined
                  ? message("own.policies.accepted")
                  : message("own.policies.effective", {
                      date: formatBelcantoDate(policy.effectiveFrom),
                    })
              }
              subtitle={`${policy.kind} · v${policy.version}`}
              testID={`policy-${policy.id}`}
              title={policy.title}
              tone="info"
            />
          ))}
          <StatusCard
            body={message("own.policies.historyBody")}
            status={message("own.policies.historyFooter")}
            title={message("own.policies.historyTitle")}
            tone="muted"
          />
          <StatusRow
            onPress={() => router.push("/(protected)/access")}
            subtitle={message("own.team.body")}
            testID="owner-team"
            title={message("own.team.title")}
            tone="info"
          />
          <StatusRow
            onPress={() => router.push("/(protected)/account/security")}
            subtitle={message("own.audit.body")}
            testID="owner-audit"
            title={message("own.audit.title")}
            tone="info"
          />
        </>
      ) : view !== null ? (
        <>
          <Text style={styles.section}>{message("own.metrics.title").toUpperCase()}</Text>
          <View style={styles.metricsRow}>
            <View style={styles.metric}>
              <Text style={styles.metricValue}>{view.activeStudents}</Text>
              <Text style={styles.metricLabel}>{message("own.metric.students")}</Text>
            </View>
            <View style={styles.metric}>
              <Text style={styles.metricValue}>{view.activeSeries}</Text>
              <Text style={styles.metricLabel}>{message("own.metric.series")}</Text>
            </View>
          </View>
          <View style={styles.metricsRow}>
            <View style={styles.metric}>
              <Text style={styles.metricValue}>{view.lessonsToday}</Text>
              <Text style={styles.metricLabel}>{message("own.metric.lessonsToday")}</Text>
            </View>
            <View style={styles.metric}>
              <Text style={styles.metricValue}>{view.upcomingEventsWeek}</Text>
              <Text style={styles.metricLabel}>{message("own.metric.eventsWeek")}</Text>
            </View>
          </View>
          <StatusCard
            body={message("own.quality.body", {
              attendance: view.pastLessonsNoAttendance,
              journals: view.draftJournals,
            })}
            status={message("own.quality.footer")}
            title={message("own.quality.title")}
            tone={
              view.pastLessonsNoAttendance + view.draftJournals > 0 ? "warning" : "success"
            }
          />
          <StatusRow
            onPress={() => router.push("/(protected)/operations")}
            subtitle={message("own.operations.body")}
            testID="owner-operations"
            title={message("own.operations.title")}
            tone="info"
          />
          <StatusCard
            body={message("own.notRating.body")}
            status={message("own.notRating.footer")}
            title={message("own.notRating.title")}
            tone="muted"
          />
        </>
      ) : (
        <Text style={styles.muted}>{message("common.loading")}</Text>
      )}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  section: { color: semantic.textGold, ...typeStyles.labelM },
  metricsRow: { flexDirection: "row", gap: space.s3 },
  metric: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderGlass,
    borderRadius: 18,
    borderWidth: 1,
    flex: 1,
    gap: space.s1,
    padding: space.s4,
  },
  metricValue: { color: semantic.textPrimary, ...typeStyles.metricXl },
  metricLabel: { color: semantic.textSecondary, ...typeStyles.labelM },
  muted: { color: semantic.textSecondary, ...typeStyles.bodyS },
});
