import type { ReactNode } from "react";
import { Pressable, StyleSheet, Text, View } from "react-native";

import type { ProgressEvidence } from "@/api/contracts";
import { Chip } from "../chip";
import { radius, semantic, sizes, space, strokes, typeStyles } from "../tokens";

/**
 * Journal & progress pattern kit (Figma Pages 22/26, components
 * «Growth Signal» 302:2, «Evidence Card» 304:2, «Lesson Recap» 309:20119).
 * DEC-006 shapes every element: progress renders as named-area evidence
 * and never as a single score, so the stage bar is opt-in and no pattern
 * accepts a numeric rating.
 */

/** «Growth Signal» — evidence-backed learning signal, never an overall score. */
export function GrowthSignal({
  kind,
  state,
  title,
  body,
  supporting,
  stage,
  testID,
}: {
  kind: string;
  state?: string | undefined;
  title: string;
  body: string;
  supporting?: string | undefined;
  stage?: { filled: number; total: number } | undefined;
  testID?: string | undefined;
}) {
  return (
    <View style={styles.signal} testID={testID}>
      <View style={styles.signalTop}>
        <Text style={styles.signalKind}>{kind.toUpperCase()}</Text>
        {state !== undefined ? (
          <Text style={styles.signalState}>{state.toUpperCase()}</Text>
        ) : null}
      </View>
      <Text style={styles.signalTitle}>{title}</Text>
      <Text style={styles.signalBody}>{body}</Text>
      {stage !== undefined && stage.total > 0 ? (
        <View style={styles.signalStage}>
          {Array.from({ length: stage.total }, (_, index) => (
            <View
              key={index}
              style={[
                styles.signalSegment,
                {
                  backgroundColor:
                    index < stage.filled
                      ? semantic.accentViolet
                      : semantic.borderDefault,
                },
              ]}
            />
          ))}
        </View>
      ) : null}
      {supporting !== undefined ? (
        <Text style={styles.signalSupporting}>{supporting}</Text>
      ) : null}
    </View>
  );
}

/**
 * «Lesson Recap» tones. Draft stays teacher-only, published is what the
 * Student sees, offline marks a locally-held draft that needs sync.
 */
export type RecapTone = "draft" | "published" | "offline";

export const RECAP_TONES: Record<RecapTone, { border: string; label: string }> = {
  draft: { border: semantic.feedbackWarning, label: semantic.feedbackWarning },
  published: { border: semantic.borderAccent, label: semantic.accentCyan },
  offline: { border: semantic.feedbackDanger, label: semantic.feedbackDanger },
};

export function LessonRecapCard({
  tone,
  eyebrow,
  title,
  sections,
  chips,
  privateNote,
  children,
  testID,
}: {
  tone: RecapTone;
  eyebrow: string;
  title: string;
  sections: readonly { label: string; body: string }[];
  chips?: readonly string[] | undefined;
  privateNote?: { label: string; body: string } | undefined;
  children?: ReactNode;
  testID?: string | undefined;
}) {
  const appearance = RECAP_TONES[tone];
  return (
    <View
      style={[styles.recap, { borderColor: appearance.border }]}
      testID={testID}
    >
      <Text style={[styles.recapLabel, { color: appearance.label }]}>
        {eyebrow.toUpperCase()}
      </Text>
      <Text style={styles.recapTitle}>{title}</Text>
      {sections.map((section) => (
        <View key={section.label} style={styles.recapSection}>
          <Text style={[styles.recapLabel, { color: appearance.label }]}>
            {section.label.toUpperCase()}
          </Text>
          <Text style={styles.recapBody}>{section.body}</Text>
        </View>
      ))}
      {chips !== undefined && chips.length > 0 ? (
        <View style={styles.recapChips}>
          {chips.map((chip) => (
            <View key={chip} style={styles.recapChip}>
              <Text style={styles.recapChipLabel}>{chip}</Text>
            </View>
          ))}
        </View>
      ) : null}
      {privateNote !== undefined ? (
        <View style={styles.recapSection}>
          <Text style={styles.recapPrivateLabel}>
            {privateNote.label.toUpperCase()}
          </Text>
          <Text style={styles.recapPrivateBody}>{privateNote.body}</Text>
        </View>
      ) : null}
      {children}
    </View>
  );
}

/**
 * «Evidence Card» media kinds. This slice materializes teacher notes from
 * lesson journals; audio/video arrive with the practice/media slice.
 */
export type EvidenceKind = "audio" | "video" | "teacherNote";

export const EVIDENCE_KINDS: Record<
  EvidenceKind,
  { tag: string; color: string }
> = {
  audio: { tag: "АУ", color: semantic.accentViolet },
  video: { tag: "ВИ", color: semantic.accentMagenta },
  teacherNote: { tag: "НТ", color: semantic.accentCyan },
};

export function EvidenceTile({
  kind,
  title,
  sourceLine,
  note,
  statusLine,
  visibility,
  onPress,
  testID,
}: {
  kind: EvidenceKind;
  title: string;
  sourceLine: string;
  note?: string | undefined;
  statusLine?: string | undefined;
  visibility: string;
  onPress?: (() => void) | undefined;
  testID?: string | undefined;
}) {
  const appearance = EVIDENCE_KINDS[kind];
  const body = (
    <>
      <View style={styles.evidenceRow}>
        <View style={[styles.evidenceMedia, { backgroundColor: appearance.color }]}>
          <Text style={styles.evidenceTag}>{appearance.tag}</Text>
        </View>
        <View style={styles.evidenceCopy}>
          <Text style={styles.evidenceTitle}>{title}</Text>
          <Text style={styles.evidenceSource}>{sourceLine}</Text>
        </View>
      </View>
      {note !== undefined ? <Text style={styles.evidenceNote}>{note}</Text> : null}
      {statusLine !== undefined ? (
        <Text style={[styles.evidenceStatus, { color: appearance.color }]}>
          {statusLine.toUpperCase()}
        </Text>
      ) : null}
      <Text style={styles.evidenceVisibility}>{visibility}</Text>
    </>
  );
  if (onPress === undefined) {
    return (
      <View style={styles.evidenceShell} testID={testID}>
        {body}
      </View>
    );
  }
  return (
    <Pressable
      accessibilityLabel={title}
      accessibilityRole="button"
      onPress={onPress}
      style={({ pressed }) => [styles.evidenceShell, pressed && styles.pressed]}
      testID={testID}
    >
      {body}
    </Pressable>
  );
}

/**
 * Skill-constellation pill adapted as an area filter (STU-GROWTH-01).
 * @deprecated Thin delegate over the unified Chip.
 */
export function AreaChip({
  label,
  accent,
  active,
  onPress,
  testID,
}: {
  label: string;
  accent: string;
  active: boolean;
  onPress: () => void;
  testID?: string | undefined;
}) {
  return (
    <Chip accent={accent} active={active} label={label} onPress={onPress} testID={testID} />
  );
}

/**
 * Area accent cycle mirroring the five constellation pills on
 * STU-GROWTH-01 (violet, cyan, gold, magenta, lavender).
 */
const AREA_ACCENTS = [
  semantic.accentViolet,
  semantic.accentCyan,
  semantic.accentGold,
  semantic.accentMagenta,
  semantic.textAccent,
] as const;

export function journalAreaAccent(area: string): string {
  let hash = 0;
  for (const char of area) {
    hash = (hash * 31 + char.charCodeAt(0)) >>> 0;
  }
  return AREA_ACCENTS[hash % AREA_ACCENTS.length]!;
}

export interface AreaGroup {
  area: string;
  accent: string;
  entries: ProgressEvidence[];
}

/** Group evidence by named area, preserving the list's recency order. */
export function groupEvidenceByArea(
  entries: readonly ProgressEvidence[],
): AreaGroup[] {
  const groups = new Map<string, AreaGroup>();
  for (const entry of entries) {
    let group = groups.get(entry.area);
    if (group === undefined) {
      group = { area: entry.area, accent: journalAreaAccent(entry.area), entries: [] };
      groups.set(entry.area, group);
    }
    group.entries.push(entry);
  }
  return [...groups.values()];
}

const styles = StyleSheet.create({
  signal: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderAccent,
    borderRadius: radius.xl,
    borderWidth: strokes.default,
    gap: space.s3,
    padding: space.s5,
  },
  signalTop: {
    flexDirection: "row",
    justifyContent: "space-between",
  },
  signalKind: { color: semantic.accentViolet, ...typeStyles.labelM },
  signalState: { color: semantic.textAccent, ...typeStyles.labelM },
  signalTitle: { color: semantic.textPrimary, ...typeStyles.headingM },
  signalBody: { color: semantic.textSecondary, ...typeStyles.bodyS },
  signalStage: { flexDirection: "row", gap: space.s1, height: 6 },
  signalSegment: { borderRadius: radius.pill, flex: 1, height: 6 },
  signalSupporting: { color: semantic.textAccent, ...typeStyles.labelL },
  recap: {
    backgroundColor: semantic.bgSurface,
    borderRadius: radius.xl,
    borderWidth: strokes.default,
    gap: space.s3,
    padding: space.s5,
  },
  recapLabel: { ...typeStyles.labelM },
  recapTitle: { color: semantic.textPrimary, ...typeStyles.headingM },
  recapSection: { gap: space.s1 },
  recapBody: { color: semantic.textSecondary, ...typeStyles.bodyS },
  recapChips: { flexDirection: "row", flexWrap: "wrap", gap: space.s2 },
  recapChip: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderDefault,
    borderRadius: radius.pill,
    borderWidth: strokes.hairline,
    paddingHorizontal: space.s2,
    paddingVertical: space.s1,
  },
  recapChipLabel: { color: semantic.textAccent, ...typeStyles.caption },
  recapPrivateLabel: { color: semantic.textGold, ...typeStyles.labelM },
  recapPrivateBody: { color: semantic.textMuted, ...typeStyles.caption },
  evidenceShell: {
    backgroundColor: semantic.bgSurface,
    borderColor: semantic.borderDefault,
    borderRadius: radius.xl,
    borderWidth: strokes.default,
    gap: space.s3,
    padding: space.s4,
  },
  evidenceRow: { alignItems: "center", flexDirection: "row", gap: space.s3 },
  evidenceMedia: {
    alignItems: "center",
    borderRadius: radius.lg,
    height: sizes.avatarMd,
    justifyContent: "center",
    width: sizes.avatarMd,
  },
  evidenceTag: { color: semantic.textInverse, ...typeStyles.labelL },
  evidenceCopy: { flex: 1, gap: space.s1 },
  evidenceTitle: { color: semantic.textPrimary, ...typeStyles.headingM },
  evidenceSource: { color: semantic.textSecondary, ...typeStyles.caption },
  evidenceNote: { color: semantic.textSecondary, ...typeStyles.bodyS },
  evidenceStatus: { ...typeStyles.labelM },
  evidenceVisibility: { color: semantic.textMuted, ...typeStyles.caption },
  pressed: { opacity: 0.85 },
});
