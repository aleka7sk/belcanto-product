import { router, useLocalSearchParams } from "expo-router";
import { useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { useApiClient } from "@/api";
import type { LessonJournal } from "@/api/contracts";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import {
  AreaChip,
  EvidenceTile,
  GrowthSignal,
  groupEvidenceByArea,
} from "../../patterns/journalPatterns";
import { AchievementsSection, AssessmentsSection, GoalSection } from "./GrowthSections";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * Student progress (Figma STU-GROWTH-01/06/09/10, flows D/I). DEC-006 is
 * the spine: development renders as teacher-confirmed evidence grouped by
 * named areas and as the journal history — never as a single score, and
 * the empty state deliberately shows no zeroed chart.
 */
export function StudentProgressScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { state } = useSession();
  const params = useLocalSearchParams<{ studentId?: string }>();
  const paramStudentId = typeof params.studentId === "string" ? params.studentId : null;
  const studentId = paramStudentId ?? state.bootstrap?.studentId ?? null;
  const roles = state.bootstrap?.roles ?? [];
  const canLead = roles.includes("Teacher") || roles.includes("Administrator");
  const canManageCatalog = roles.includes("Owner") || roles.includes("Administrator");
  const [selectedArea, setSelectedArea] = useState<string | null>(null);

  const evidence = useAccountResource((accessToken) =>
    api.listProgressEvidence(accessToken, studentId ?? ""),
  );
  const journals = useAccountResource((accessToken) =>
    api.listStudentJournals(accessToken, studentId ?? ""),
  );

  if (studentId === null) {
    return (
      <AccountScreenShell navigation={<AccountNav />} testID="progress-guard">
        <InlineNotice
          body={message("growth.guard.body")}
          title={message("growth.guard.title")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const entries = evidence.value ?? [];
  const journalList = journals.value ?? [];
  const loading = evidence.value === null || journals.value === null;
  const groups = groupEvidenceByArea(entries);
  const latest = entries[0];
  const visible =
    selectedArea === null
      ? entries
      : (groups.find((group) => group.area === selectedArea)?.entries ?? []);

  const journalOf = (sourceId: string): LessonJournal | undefined => {
    const journalId = sourceId.split(":")[0];
    return journalList.find((journal) => journal.id === journalId);
  };
  const openJournal = (journal: LessonJournal) =>
    router.push({
      pathname: "/(protected)/journal/[occurrenceId]/[studentId]",
      params: { occurrenceId: journal.occurrenceId, studentId: journal.studentId },
    });

  const empty = !loading && entries.length === 0 && journalList.length === 0;
  const loadError = evidence.error ?? journals.error;

  return (
    <AccountScreenShell navigation={<AccountNav />} testID="student-progress">
      <ScreenHeading
        eyebrow={message("growth.eyebrow")}
        subtitle={
          empty || latest === undefined
            ? message("growth.empty.body")
            : message("growth.subtitle", {
                count: entries.length,
                date: formatBelcantoDate(latest.recordedAt),
              })
        }
        title={empty ? message("growth.empty.title") : message("growth.title")}
      />
      {loadError !== null && loadError !== undefined ? (
        <InlineNotice
          body={apiErrorMessage(loadError)}
          title={message("common.retry")}
          tone="error"
        />
      ) : null}
      {empty ? (
        <>
          <StatusCard
            body={message("growth.empty.card.body")}
            status={message("growth.empty.card.status")}
            title={message("growth.empty.card.title")}
            tone="info"
          />
          <BlockAction
            kind="secondary"
            label={message("growth.empty.action")}
            onPress={() => router.replace("/(protected)")}
          />
        </>
      ) : (
        <>
          {latest !== undefined ? (
            <GrowthSignal
              body={latest.note}
              kind={message("growth.signal.kind")}
              state={message("growth.signal.state")}
              supporting={message("growth.signal.source", {
                date: formatBelcantoDate(latest.recordedAt),
              })}
              testID="progress-latest-signal"
              title={latest.area}
            />
          ) : null}
          {groups.length > 0 ? (
            <>
              <Text style={styles.sectionTitle}>{message("growth.areas.title")}</Text>
              <View style={styles.chips}>
                <AreaChip
                  accent={semantic.accentViolet}
                  active={selectedArea === null}
                  label={message("growth.areas.all")}
                  onPress={() => setSelectedArea(null)}
                  testID="progress-area-all"
                />
                {groups.map((group) => (
                  <AreaChip
                    accent={group.accent}
                    active={selectedArea === group.area}
                    key={group.area}
                    label={`${group.area} · ${group.entries.length}`}
                    onPress={() =>
                      setSelectedArea((current) =>
                        current === group.area ? null : group.area,
                      )
                    }
                  />
                ))}
              </View>
            </>
          ) : null}
          {visible.length > 0 ? (
            <Text style={styles.sectionTitle}>
              {message("growth.evidence.title")}
            </Text>
          ) : null}
          {visible.map((entry) => {
            const journal = journalOf(entry.sourceId);
            return (
              <EvidenceTile
                key={entry.id}
                kind="teacherNote"
                note={entry.note}
                onPress={journal !== undefined ? () => openJournal(journal) : undefined}
                sourceLine={message("growth.evidence.source", {
                  date: formatBelcantoDate(entry.recordedAt),
                })}
                title={entry.area}
                visibility={message("growth.evidence.visibility")}
              />
            );
          })}
          {journalList.length > 0 ? (
            <>
              <Text style={styles.sectionTitle}>
                {message("growth.history.title")}
              </Text>
              <Text style={styles.sectionSubtitle}>
                {message("growth.history.subtitle")}
              </Text>
              {journalList.map((journal) => {
                const current = journal.versions[0];
                if (current === undefined) return null;
                return (
                  <StatusRow
                    key={journal.id}
                    onPress={() => openJournal(journal)}
                    status={
                      current.version > 1
                        ? message("growth.history.corrected", {
                            version: current.version,
                          })
                        : message("growth.history.version", {
                            version: current.version,
                          })
                    }
                    subtitle={current.currentFocus}
                    testID={`progress-journal-${journal.id}`}
                    title={message("growth.history.entry", {
                      date: formatBelcantoDate(current.publishedAt),
                    })}
                    tone="info"
                  />
                );
              })}
            </>
          ) : null}
        </>
      )}
      <AssessmentsSection studentId={studentId} />
      <GoalSection canLead={canLead} studentId={studentId} />
      <AchievementsSection
        canLead={canLead}
        canManageCatalog={canManageCatalog}
        studentId={studentId}
      />
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  sectionTitle: {
    color: semantic.textPrimary,
    marginTop: space.s2,
    ...typeStyles.headingM,
  },
  sectionSubtitle: { color: semantic.textSecondary, ...typeStyles.caption },
  chips: { flexDirection: "row", flexWrap: "wrap", gap: space.s2 },
});
