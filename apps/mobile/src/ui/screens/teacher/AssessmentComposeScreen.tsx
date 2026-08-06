import { router, useLocalSearchParams } from "expo-router";
import { useMemo, useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { useApiClient } from "@/api";
import type {
  AssessmentContentRequest,
  AssessmentContext,
  AssessmentType,
  AssessmentVisibility,
} from "@/api/contracts";
import { createIntentIdempotency } from "@/controllers";
import { useMessage, type MessageKey } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice, PremiumTextField } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
} from "../../patterns/accountPatterns";
import { AreaChip } from "../../patterns/journalPatterns";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav } from "../account/shared";

/**
 * Assessment composer (TCH-REVIEW-02/05/06). The system never decides:
 * the Teacher writes the observation, grounds it in evidence and
 * confirms the publication. In supersede mode the same form produces
 * the correcting version of a published assessment. Statuses describe
 * the observation in context — never the person (domain/assessment.md
 * core principle).
 */

const TYPE_KEYS: Record<Exclude<AssessmentType, "self">, MessageKey> = {
  observation: "asmt.type.observation",
  diagnostic: "asmt.type.diagnostic",
  formative: "asmt.type.formative",
  summative: "asmt.type.summative",
};

const CONTEXT_KEYS: Record<AssessmentContext, MessageKey> = {
  lesson: "asmt.context.lesson",
  homework_review: "asmt.context.homework_review",
  repertoire_practice: "asmt.context.repertoire_practice",
  concert_preparation: "asmt.context.concert_preparation",
  concert_performance: "asmt.context.concert_performance",
  diagnostic_session: "asmt.context.diagnostic_session",
  periodic_review: "asmt.context.periodic_review",
  teacher_observation: "asmt.context.teacher_observation",
};

const VISIBILITY_KEYS: Record<AssessmentVisibility, MessageKey> = {
  student_visible: "asmt.visibility.student_visible",
  staff_visible: "asmt.visibility.staff_visible",
  teacher_only: "asmt.visibility.teacher_only",
  owner_analytics: "asmt.visibility.owner_analytics",
};

function todayAlmaty(): string {
  const parts = new Intl.DateTimeFormat("en-CA", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    timeZone: "Asia/Almaty",
  }).format(new Date());
  return parts;
}

export function AssessmentComposeScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ studentId?: string; supersedeId?: string }>();
  const studentId = typeof params.studentId === "string" ? params.studentId : "";
  const supersedeId = typeof params.supersedeId === "string" ? params.supersedeId : "";

  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [type, setType] = useState<Exclude<AssessmentType, "self">>("formative");
  const [contextType, setContextType] = useState<AssessmentContext>("lesson");
  const [visibility, setVisibility] = useState<AssessmentVisibility>("student_visible");
  const [confidence, setConfidence] = useState<"" | "low" | "medium" | "high">("");
  const [summary, setSummary] = useState("");
  const [strengths, setStrengths] = useState("");
  const [developmentAreas, setDevelopmentAreas] = useState("");
  const [recommendations, setRecommendations] = useState("");
  const [areas, setAreas] = useState("");
  const [evidenceNote, setEvidenceNote] = useState("");
  const [busy, setBusy] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  if (studentId === "" && supersedeId === "") {
    return (
      <AccountScreenShell navigation={<AccountNav active="today" />} testID="assessment-compose-guard">
        <InlineNotice
          body={message("asmt.compose.guardBody")}
          title={message("asmt.compose.guardTitle")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const content = (): AssessmentContentRequest => ({
    type,
    contextType,
    assessmentDate: todayAlmaty(),
    visibility,
    ...(summary.trim() === "" ? {} : { summary: summary.trim() }),
    ...(strengths.trim() === "" ? {} : { strengths: strengths.trim() }),
    ...(developmentAreas.trim() === "" ? {} : { developmentAreas: developmentAreas.trim() }),
    ...(recommendations.trim() === "" ? {} : { recommendations: recommendations.trim() }),
    ...(areas.trim() === "" ? {} : { areas: areas.trim() }),
    ...(confidence === "" ? {} : { confidence }),
  });

  const submit = async (publish: boolean) => {
    setFormError(null);
    setBusy(true);
    try {
      if (supersedeId !== "") {
        const chain = await runAuthenticated((accessToken) =>
          api.supersedeAssessment(accessToken, supersedeId, content(), idempotency.key()),
        );
        idempotency.complete();
        router.replace({
          pathname: "/(protected)/assessment/[assessmentId]",
          params: { assessmentId: chain[1]!.id },
        });
        return;
      }
      const outcome = await runAuthenticated(async (accessToken) => {
        let draft = await api.createAssessment(
          accessToken,
          studentId,
          content(),
          idempotency.key(),
        );
        const note = evidenceNote.trim();
        if (note !== "") {
          draft = await api.addAssessmentEvidence(
            accessToken,
            draft.id,
            { kind: "observation", note },
            `${idempotency.key()}-evd`,
          );
        }
        if (publish) {
          draft = await api.publishAssessment(
            accessToken,
            draft.id,
            { expectedVersion: draft.version },
            `${idempotency.key()}-pub`,
          );
        }
        return draft;
      });
      idempotency.complete();
      router.replace({
        pathname: "/(protected)/assessment/[assessmentId]",
        params: { assessmentId: outcome.id },
      });
    } catch (cause) {
      setFormError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  const substanceMissing =
    summary.trim() === "" ||
    (strengths.trim() === "" &&
      developmentAreas.trim() === "" &&
      recommendations.trim() === "" &&
      evidenceNote.trim() === "");

  return (
    <AccountScreenShell navigation={<AccountNav active="today" />} testID="assessment-compose">
      <ScreenHeading
        eyebrow={message(supersedeId !== "" ? "asmt.compose.supersedeEyebrow" : "asmt.compose.eyebrow")}
        subtitle={message("asmt.compose.subtitle")}
        title={message(supersedeId !== "" ? "asmt.compose.supersedeTitle" : "asmt.compose.title")}
      />
      <Text style={styles.groupLabel}>{message("asmt.compose.typeLabel")}</Text>
      <View style={styles.chips}>
        {(Object.keys(TYPE_KEYS) as Exclude<AssessmentType, "self">[]).map((key) => (
          <AreaChip
            accent={semantic.accentViolet}
            active={type === key}
            key={key}
            label={message(TYPE_KEYS[key])}
            onPress={() => setType(key)}
            testID={`asmt-type-${key}`}
          />
        ))}
      </View>
      <Text style={styles.groupLabel}>{message("asmt.compose.contextLabel")}</Text>
      <View style={styles.chips}>
        {(Object.keys(CONTEXT_KEYS) as AssessmentContext[]).map((key) => (
          <AreaChip
            accent={semantic.accentCyan}
            active={contextType === key}
            key={key}
            label={message(CONTEXT_KEYS[key])}
            onPress={() => setContextType(key)}
            testID={`asmt-context-${key}`}
          />
        ))}
      </View>
      <PremiumTextField
        label={message("asmt.compose.summary")}
        multiline
        onChangeText={setSummary}
        placeholder={message("asmt.compose.summaryPlaceholder")}
        testID="asmt-summary"
        value={summary}
      />
      <PremiumTextField
        label={message("asmt.compose.strengths")}
        multiline
        onChangeText={setStrengths}
        placeholder={message("asmt.compose.strengthsPlaceholder")}
        testID="asmt-strengths"
        value={strengths}
      />
      <PremiumTextField
        label={message("asmt.compose.developmentAreas")}
        multiline
        onChangeText={setDevelopmentAreas}
        placeholder={message("asmt.compose.developmentAreasPlaceholder")}
        testID="asmt-development"
        value={developmentAreas}
      />
      <PremiumTextField
        label={message("asmt.compose.recommendations")}
        multiline
        onChangeText={setRecommendations}
        placeholder={message("asmt.compose.recommendationsPlaceholder")}
        testID="asmt-recommendations"
        value={recommendations}
      />
      <PremiumTextField
        label={message("asmt.compose.areas")}
        onChangeText={setAreas}
        placeholder={message("asmt.compose.areasPlaceholder")}
        testID="asmt-areas"
        value={areas}
      />
      {supersedeId === "" ? (
        <PremiumTextField
          label={message("asmt.compose.evidence")}
          multiline
          onChangeText={setEvidenceNote}
          placeholder={message("asmt.compose.evidencePlaceholder")}
          testID="asmt-evidence"
          value={evidenceNote}
        />
      ) : null}
      <Text style={styles.groupLabel}>{message("asmt.compose.confidenceLabel")}</Text>
      <View style={styles.chips}>
        {(["low", "medium", "high"] as const).map((key) => (
          <AreaChip
            accent={semantic.accentGold}
            active={confidence === key}
            key={key}
            label={message(`asmt.confidence.${key}`)}
            onPress={() => setConfidence((current) => (current === key ? "" : key))}
            testID={`asmt-confidence-${key}`}
          />
        ))}
      </View>
      <Text style={styles.groupLabel}>{message("asmt.compose.visibilityLabel")}</Text>
      <View style={styles.chips}>
        {(Object.keys(VISIBILITY_KEYS) as AssessmentVisibility[]).map((key) => (
          <AreaChip
            accent={semantic.accentMagenta}
            active={visibility === key}
            key={key}
            label={message(VISIBILITY_KEYS[key])}
            onPress={() => setVisibility(key)}
            testID={`asmt-visibility-${key}`}
          />
        ))}
      </View>
      <StatusCard
        body={message("asmt.compose.principleBody")}
        status={message("asmt.compose.principleFooter")}
        title={message("asmt.compose.principleTitle")}
        tone="info"
      />
      {formError !== null ? (
        <InlineNotice body={formError} title={message("common.retry")} tone="error" />
      ) : null}
      <BlockAction
        busy={busy}
        disabled={substanceMissing}
        label={message(supersedeId !== "" ? "asmt.compose.supersedeSubmit" : "asmt.compose.publish")}
        onPress={() => void submit(true)}
        testID="asmt-publish"
      />
      {supersedeId === "" ? (
        <BlockAction
          busy={busy}
          disabled={summary.trim() === "" && evidenceNote.trim() === ""}
          kind="secondary"
          label={message("asmt.compose.saveDraft")}
          onPress={() => void submit(false)}
          testID="asmt-save-draft"
        />
      ) : null}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  chips: { flexDirection: "row", flexWrap: "wrap", gap: space.s2 },
  groupLabel: {
    color: semantic.textGold,
    ...typeStyles.labelM,
  },
});
