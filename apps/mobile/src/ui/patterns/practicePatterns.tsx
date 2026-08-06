import { Pressable, StyleSheet, Text, View } from "react-native";

import type { HomeworkStatus, HomeworkTask } from "@/api/contracts";
import type { StatusTone } from "./accountPatterns";
import { semantic, space, strokes, typeStyles } from "../tokens";

/**
 * Practice pattern kit (Figma Page 23, STU-PRACTICE-01..16). Upload
 * progress mirrors the resumable media lifecycle — the bar renders real
 * uploadedBytes, never a fake percent — and homework states carry the
 * approved lifecycle of domain/homework.md.
 */

/** Status → tone for status lines and chips across practice screens. */
export const HOMEWORK_STATUS_TONE: Record<HomeworkStatus, StatusTone> = {
  draft: "muted",
  assigned: "info",
  in_progress: "info",
  submitted: "warning",
  reviewed: "warning",
  completed: "success",
  cancelled: "danger",
  expired: "muted",
};

/** Accent cycle for parallel upload rows (violet, then cyan — 327:463/469). */
const UPLOAD_ACCENTS = [semantic.accentViolet, semantic.accentCyan] as const;

export function uploadAccent(index: number): string {
  return UPLOAD_ACCENTS[index % UPLOAD_ACCENTS.length]!;
}

export function uploadPercent(uploadedBytes: number, byteSize: number): number {
  if (byteSize <= 0) return 0;
  const percent = Math.floor((uploadedBytes / byteSize) * 100);
  return Math.min(100, Math.max(0, percent));
}

/** «Progress · Попытка N» — resumable upload row with a real byte bar. */
export function UploadProgressCard({
  title,
  uploadedBytes,
  byteSize,
  accent,
  testID,
}: {
  title: string;
  uploadedBytes: number;
  byteSize: number;
  accent: string;
  testID?: string | undefined;
}) {
  const percent = uploadPercent(uploadedBytes, byteSize);
  return (
    <View
      accessibilityLabel={`${title} · ${percent}%`}
      style={styles.uploadCard}
      testID={testID}
    >
      <View style={styles.uploadHeader}>
        <Text style={styles.uploadTitle}>{title}</Text>
        <Text style={styles.uploadPercent}>{percent}%</Text>
      </View>
      <View style={styles.uploadTrack}>
        <View
          style={[
            styles.uploadValue,
            { backgroundColor: accent, width: `${percent}%` },
          ]}
        />
      </View>
    </View>
  );
}

/**
 * Plan row («План на 15 минут», 327:127): a numbered task line the
 * Student can toggle while the homework is in progress.
 */
export function TaskRow({
  task,
  doneLabel,
  pendingLabel,
  onToggle,
  testID,
}: {
  task: HomeworkTask;
  doneLabel: string;
  pendingLabel: string;
  onToggle?: (() => void) | undefined;
  testID?: string | undefined;
}) {
  const done = task.status === "done";
  const line = [
    `${task.position}. ${task.title}`,
    task.recommendedMinutes !== undefined ? `${task.recommendedMinutes} мин` : null,
  ]
    .filter((part): part is string => part !== null)
    .join(" · ");
  const body = (
    <>
      <Text style={[styles.taskLine, done && styles.taskLineDone]}>{line}</Text>
      <Text
        style={[
          styles.taskStatus,
          { color: done ? semantic.feedbackSuccess : semantic.textMuted },
        ]}
      >
        {done ? doneLabel : pendingLabel}
      </Text>
    </>
  );
  if (onToggle === undefined) {
    return (
      <View style={styles.taskRow} testID={testID}>
        {body}
      </View>
    );
  }
  return (
    <Pressable
      accessibilityLabel={task.title}
      accessibilityRole="checkbox"
      accessibilityState={{ checked: done }}
      onPress={onToggle}
      style={({ pressed }) => [styles.taskRow, pressed && styles.pressed]}
      testID={testID}
    >
      {body}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  uploadCard: {
    backgroundColor: semantic.bgSurface,
    borderColor: semantic.borderDefault,
    borderRadius: 20,
    borderWidth: strokes.hairline,
    gap: 10,
    padding: space.s4,
  },
  uploadHeader: {
    alignItems: "center",
    flexDirection: "row",
    justifyContent: "space-between",
  },
  uploadTitle: { color: semantic.textPrimary, ...typeStyles.labelL },
  uploadPercent: { color: semantic.textSecondary, ...typeStyles.labelM },
  uploadTrack: {
    backgroundColor: semantic.bgRaised,
    borderRadius: 4,
    height: 8,
    overflow: "hidden",
  },
  uploadValue: { borderRadius: 4, height: 8 },
  taskRow: {
    alignItems: "center",
    flexDirection: "row",
    gap: space.s3,
    justifyContent: "space-between",
    minHeight: 32,
  },
  taskLine: { color: semantic.textSecondary, flex: 1, ...typeStyles.bodyS },
  taskLineDone: { color: semantic.textMuted, textDecorationLine: "line-through" },
  taskStatus: { ...typeStyles.labelM },
  pressed: { opacity: 0.85 },
});
