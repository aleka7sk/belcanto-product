import { router } from "expo-router";
import { useMemo, useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { ApiTransportError, useApiClient } from "@/api";
import type { AchievementAward, StudentGoal } from "@/api/contracts";
import { createIntentIdempotency } from "@/controllers";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice, PremiumTextField } from "../../components";
import { BlockAction, StatusCard, StatusRow } from "../../patterns/accountPatterns";
import { AreaChip, GrowthSignal } from "../../patterns/journalPatterns";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { useAccountResource } from "../account/shared";

/**
 * Goal and achievement sections of the growth area (Figma
 * STU-GROWTH-04/08). A goal is reframed, never «failed»; achievements
 * are evidence-backed awards from the school catalog — no ratings, XP
 * or streak penalties anywhere (DEC-006).
 */

export function GoalSection({
  studentId,
  canLead,
}: {
  studentId: string;
  canLead: boolean;
}) {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const goals = useAccountResource((accessToken) =>
    api.listStudentGoals(accessToken, studentId),
  );
  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [mode, setMode] = useState<"idle" | "create" | "complete" | "reframe">("idle");
  const [criterion, setCriterion] = useState("");
  const [description, setDescription] = useState("");
  const [note, setNote] = useState("");
  const [reason, setReason] = useState("");
  const [newCriterion, setNewCriterion] = useState("");
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const list = goals.value ?? [];
  const active = list.find((goal) => goal.status === "active") ?? null;
  const history = list.filter((goal) => goal.status !== "active");

  const run = async (operation: (accessToken: string) => Promise<unknown>) => {
    setActionError(null);
    setBusy(true);
    try {
      await runAuthenticated(operation);
      idempotency.complete();
      setMode("idle");
      setCriterion("");
      setDescription("");
      setNote("");
      setReason("");
      setNewCriterion("");
      await goals.reload();
    } catch (cause) {
      if (!(cause instanceof ApiTransportError)) idempotency.abandon();
      setActionError(apiErrorMessage(cause));
      await goals.reload();
    } finally {
      setBusy(false);
    }
  };

  return (
    <View style={styles.section}>
      <Text style={styles.sectionTitle}>{message("goal.section.eyebrow")}</Text>
      <Text style={styles.sectionHint}>{message("goal.section.hint")}</Text>
      {active !== null ? (
        <GrowthSignal
          body={active.description ?? message("goal.section.hint")}
          kind={message("goal.signal.kind")}
          state={message("goal.signal.state")}
          supporting={message("goal.setLine", {
            date: formatBelcantoDate(active.createdAt),
            name: active.createdBy.fullName,
          })}
          testID="growth-active-goal"
          title={active.criterion}
        />
      ) : (
        <Text style={styles.sectionHint}>{message("goal.none")}</Text>
      )}
      {actionError !== null ? (
        <InlineNotice body={actionError} title={message("common.retry")} tone="error" />
      ) : null}
      {canLead && active === null && mode !== "create" ? (
        <BlockAction
          kind="secondary"
          label={message("goal.create.title")}
          onPress={() => setMode("create")}
          testID="goal-create-open"
        />
      ) : null}
      {canLead && mode === "create" ? (
        <>
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("goal.create.criterion")}
            multiline
            onChangeText={setCriterion}
            placeholder={message("goal.create.criterionHint")}
            testID="goal-criterion"
            value={criterion}
          />
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("goal.create.description")}
            multiline
            onChangeText={setDescription}
            testID="goal-description"
            value={description}
          />
          <BlockAction
            busy={busy}
            disabled={criterion.trim() === ""}
            label={message("goal.create.save")}
            onPress={() =>
              void run((accessToken) =>
                api.createGoal(
                  accessToken,
                  studentId,
                  {
                    criterion: criterion.trim(),
                    ...(description.trim() !== ""
                      ? { description: description.trim() }
                      : {}),
                  },
                  idempotency.key(),
                ),
              )
            }
            testID="goal-create-save"
          />
          <BlockAction
            kind="secondary"
            label={message("common.cancel")}
            onPress={() => setMode("idle")}
          />
        </>
      ) : null}
      {canLead && active !== null && mode === "idle" ? (
        <>
          <BlockAction
            kind="secondary"
            label={message("goal.complete.action")}
            onPress={() => setMode("complete")}
            testID="goal-complete-open"
          />
          <BlockAction
            kind="secondary"
            label={message("goal.reframe.action")}
            onPress={() => setMode("reframe")}
            testID="goal-reframe-open"
          />
        </>
      ) : null}
      {canLead && active !== null && mode === "complete" ? (
        <>
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("goal.complete.note")}
            multiline
            onChangeText={setNote}
            testID="goal-complete-note"
            value={note}
          />
          <BlockAction
            busy={busy}
            disabled={note.trim() === ""}
            label={message("goal.complete.save")}
            onPress={() =>
              void run((accessToken) =>
                api.completeGoal(
                  accessToken,
                  active.id,
                  { completionNote: note.trim(), expectedVersion: active.version },
                  idempotency.key(),
                ),
              )
            }
            testID="goal-complete-save"
          />
          <BlockAction
            kind="secondary"
            label={message("common.cancel")}
            onPress={() => setMode("idle")}
          />
        </>
      ) : null}
      {canLead && active !== null && mode === "reframe" ? (
        <>
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("goal.reframe.reason")}
            multiline
            onChangeText={setReason}
            testID="goal-reframe-reason"
            value={reason}
          />
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("goal.reframe.newCriterion")}
            multiline
            onChangeText={setNewCriterion}
            testID="goal-reframe-criterion"
            value={newCriterion}
          />
          <BlockAction
            busy={busy}
            disabled={reason.trim() === ""}
            label={message("goal.reframe.save")}
            onPress={() =>
              void run((accessToken) =>
                api.reframeGoal(
                  accessToken,
                  active.id,
                  {
                    reason: reason.trim(),
                    ...(newCriterion.trim() !== ""
                      ? { newCriterion: newCriterion.trim() }
                      : {}),
                    expectedVersion: active.version,
                  },
                  idempotency.key(),
                ),
              )
            }
            testID="goal-reframe-save"
          />
          <BlockAction
            kind="secondary"
            label={message("common.cancel")}
            onPress={() => setMode("idle")}
          />
        </>
      ) : null}
      {history.length > 0 ? (
        <>
          <Text style={styles.sectionTitle}>{message("goal.history.title")}</Text>
          {history.map((goal: StudentGoal) => (
            <StatusRow
              key={goal.id}
              status={
                goal.status === "completed"
                  ? message("goal.status.completed", {
                      note: goal.completionNote ?? "",
                    })
                  : message("goal.status.cancelled", {
                      reason: goal.cancelReason ?? "",
                    })
              }
              subtitle={formatBelcantoDate(goal.updatedAt)}
              title={goal.criterion}
              tone={goal.status === "completed" ? "success" : "muted"}
            />
          ))}
        </>
      ) : null}
    </View>
  );
}

export function AchievementsSection({
  studentId,
  canLead,
  canManageCatalog,
}: {
  studentId: string;
  canLead: boolean;
  canManageCatalog: boolean;
}) {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const awards = useAccountResource((accessToken) =>
    api.listStudentAwards(accessToken, studentId),
  );
  const definitions = useAccountResource((accessToken) =>
    canLead ? api.listAchievementDefinitions(accessToken) : Promise.resolve([]),
  );
  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [mode, setMode] = useState<"idle" | "award" | "definition">("idle");
  const [definitionId, setDefinitionId] = useState<string | null>(null);
  const [evidenceNote, setEvidenceNote] = useState("");
  const [revokingAwardId, setRevokingAwardId] = useState<string | null>(null);
  const [revokeReason, setRevokeReason] = useState("");
  const [defName, setDefName] = useState("");
  const [defDescription, setDefDescription] = useState("");
  const [defCategory, setDefCategory] = useState("");
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const list = awards.value ?? [];
  const published = (definitions.value ?? []).filter(
    (definition) => definition.status === "published",
  );

  const run = async (operation: (accessToken: string) => Promise<unknown>) => {
    setActionError(null);
    setBusy(true);
    try {
      await runAuthenticated(operation);
      idempotency.complete();
      setMode("idle");
      setDefinitionId(null);
      setEvidenceNote("");
      setRevokingAwardId(null);
      setRevokeReason("");
      setDefName("");
      setDefDescription("");
      setDefCategory("");
      await Promise.all([awards.reload(), definitions.reload()]);
    } catch (cause) {
      if (!(cause instanceof ApiTransportError)) idempotency.abandon();
      setActionError(apiErrorMessage(cause));
      await awards.reload();
    } finally {
      setBusy(false);
    }
  };

  return (
    <View style={styles.section}>
      <Text style={styles.sectionTitle}>{message("ach.section.title")}</Text>
      <Text style={styles.sectionHint}>{message("ach.section.hint")}</Text>
      {list.length === 0 ? (
        <Text style={styles.sectionHint}>{message("ach.none")}</Text>
      ) : null}
      {list.map((award: AchievementAward) => (
        <View key={award.id} style={styles.awardBlock}>
          <StatusCard
            body={award.evidenceNote}
            status={
              award.status === "revoked"
                ? message("ach.award.revoked", { reason: award.revokeReason ?? "" })
                : message("ach.award.line", {
                    category: award.category,
                    date: formatBelcantoDate(award.awardedAt),
                    name: award.awardedBy.fullName,
                  })
            }
            title={award.definitionName}
            tone={award.status === "revoked" ? "danger" : "success"}
          />
          {canLead && award.status === "awarded" ? (
            revokingAwardId === award.id ? (
              <>
                <PremiumTextField
                  autoCapitalize="sentences"
                  label={message("ach.revoke.reason")}
                  onChangeText={setRevokeReason}
                  testID={`award-revoke-reason-${award.id}`}
                  value={revokeReason}
                />
                <BlockAction
                  busy={busy}
                  disabled={revokeReason.trim() === ""}
                  kind="secondary"
                  label={message("ach.revoke.action")}
                  onPress={() =>
                    void run((accessToken) =>
                      api.revokeAchievement(
                        accessToken,
                        award.id,
                        { reason: revokeReason.trim() },
                        idempotency.key(),
                      ),
                    )
                  }
                  testID={`award-revoke-save-${award.id}`}
                />
              </>
            ) : (
              <BlockAction
                kind="secondary"
                label={message("ach.revoke.action")}
                onPress={() => setRevokingAwardId(award.id)}
                testID={`award-revoke-open-${award.id}`}
              />
            )
          ) : null}
        </View>
      ))}
      {actionError !== null ? (
        <InlineNotice body={actionError} title={message("common.retry")} tone="error" />
      ) : null}
      {canLead && mode === "idle" ? (
        <BlockAction
          kind="secondary"
          label={message("ach.award.title")}
          onPress={() => setMode("award")}
          testID="award-open"
        />
      ) : null}
      {canLead && mode === "award" ? (
        <>
          <Text style={styles.sectionHint}>{message("ach.award.pick")}</Text>
          <View style={styles.chips}>
            {published.map((definition) => (
              <AreaChip
                accent={semantic.accentGold}
                active={definitionId === definition.id}
                key={definition.id}
                label={definition.name}
                onPress={() => setDefinitionId(definition.id)}
                testID={`award-definition-${definition.id}`}
              />
            ))}
          </View>
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("ach.award.evidence")}
            multiline
            onChangeText={setEvidenceNote}
            testID="award-evidence"
            value={evidenceNote}
          />
          <BlockAction
            busy={busy}
            disabled={definitionId === null || evidenceNote.trim() === ""}
            label={message("ach.award.save")}
            onPress={() =>
              void run((accessToken) =>
                api.awardAchievement(
                  accessToken,
                  studentId,
                  { definitionId: definitionId!, evidenceNote: evidenceNote.trim() },
                  idempotency.key(),
                ),
              )
            }
            testID="award-save"
          />
          <BlockAction
            kind="secondary"
            label={message("common.cancel")}
            onPress={() => setMode("idle")}
          />
        </>
      ) : null}
      {canManageCatalog && mode === "idle" ? (
        <BlockAction
          kind="secondary"
          label={message("ach.def.title")}
          onPress={() => setMode("definition")}
          testID="definition-open"
        />
      ) : null}
      {canManageCatalog && mode === "definition" ? (
        <>
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("ach.def.name")}
            onChangeText={setDefName}
            testID="definition-name"
            value={defName}
          />
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("ach.def.description")}
            multiline
            onChangeText={setDefDescription}
            testID="definition-description"
            value={defDescription}
          />
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("ach.def.category")}
            onChangeText={setDefCategory}
            testID="definition-category"
            value={defCategory}
          />
          <BlockAction
            busy={busy}
            disabled={
              defName.trim() === "" ||
              defDescription.trim() === "" ||
              defCategory.trim() === ""
            }
            label={message("ach.def.save")}
            onPress={() =>
              void run((accessToken) =>
                api.createAchievementDefinition(
                  accessToken,
                  {
                    name: defName.trim(),
                    description: defDescription.trim(),
                    category: defCategory.trim(),
                  },
                  idempotency.key(),
                ),
              )
            }
            testID="definition-save"
          />
          <BlockAction
            kind="secondary"
            label={message("common.cancel")}
            onPress={() => setMode("idle")}
          />
        </>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  section: { gap: space.s3 },
  sectionTitle: {
    color: semantic.textPrimary,
    marginTop: space.s2,
    ...typeStyles.headingM,
  },
  sectionHint: { color: semantic.textSecondary, ...typeStyles.caption },
  awardBlock: { gap: space.s3 },
  chips: { flexDirection: "row", flexWrap: "wrap", gap: space.s2 },
});

/**
 * Published assessment history (domain/assessment.md; TCH-REVIEW-06
 * «Что увидит ученица»). The server already scopes visibility — the
 * Student receives student_visible publications only.
 */
export function AssessmentsSection({ studentId }: { studentId: string }) {
  const message = useMessage();
  const api = useApiClient();
  const assessments = useAccountResource((accessToken) =>
    api.listStudentAssessments(accessToken, studentId),
  );
  const list = (assessments.value ?? []).filter(
    (assessment) => assessment.status === "published" || assessment.status === "superseded",
  );
  if (assessments.value !== null && list.length === 0) return null;
  return (
    <View style={assessmentStyles.block}>
      <Text style={assessmentStyles.title}>
        {message("asmt.section.title").toUpperCase()}
      </Text>
      {assessments.error !== null ? (
        <InlineNotice
          body={apiErrorMessage(assessments.error)}
          title={message("common.retry")}
          tone="error"
        />
      ) : null}
      {list.map((assessment) => (
        <StatusRow
          key={assessment.id}
          onPress={() =>
            router.push({
              pathname: "/(protected)/assessment/[assessmentId]",
              params: { assessmentId: assessment.id },
            })
          }
          status={
            assessment.status === "superseded"
              ? message("asmt.status.superseded")
              : message("asmt.section.published", {
                  date: assessment.assessmentDate,
                })
          }
          subtitle={assessment.summary ?? ""}
          testID={`assessment-row-${assessment.id}`}
          title={`${message(`asmt.context.${assessment.contextType}`)} · ${assessment.author.fullName}`}
          tone={assessment.status === "superseded" ? "muted" : "success"}
        />
      ))}
    </View>
  );
}

const assessmentStyles = StyleSheet.create({
  block: { gap: space.s3 },
  title: { color: semantic.textGold, ...typeStyles.labelM },
});
