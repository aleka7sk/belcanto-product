import { useMemo, useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { ApiTransportError, useApiClient } from "@/api";
import type { AttendanceRecord, AttendanceStatus, LessonStudent } from "@/api/contracts";
import { createIntentIdempotency } from "@/controllers";
import { useMessage, type MessageFormatter } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice, PremiumTextField } from "../../components";
import { BlockAction, StatusRow } from "../../patterns/accountPatterns";
import { AreaChip } from "../../patterns/journalPatterns";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage } from "../../viewModels";
import { useAccountResource } from "../account/shared";

/**
 * Attendance block (Figma TCH-JOURNAL-01/02) inside the teacher lesson
 * context. Present / late (with minutes) / absent (with a mandatory
 * note); changing a recorded mark demands a reason; an empty group seat
 * simply has no row and is never an absence.
 */

const STATUS_ORDER: AttendanceStatus[] = ["present", "late", "absent"];

function statusLabel(status: AttendanceStatus, message: MessageFormatter): string {
  return message(`att.status.${status}`);
}

function recordLine(record: AttendanceRecord, message: MessageFormatter): string {
  if (record.status === "late") {
    return message("att.line.late", { minutes: record.lateMinutes ?? 0 });
  }
  return statusLabel(record.status, message);
}

export function AttendanceBlock({
  lessonId,
  students,
}: {
  lessonId: string;
  students: readonly LessonStudent[];
}) {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const attendance = useAccountResource((accessToken) =>
    api.listLessonAttendance(accessToken, lessonId),
  );
  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [editing, setEditing] = useState<string | null>(null);
  const [status, setStatus] = useState<AttendanceStatus>("present");
  const [minutes, setMinutes] = useState("");
  const [note, setNote] = useState("");
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const records = attendance.value ?? [];
  const recordOf = (studentId: string) =>
    records.find((record) => record.studentId === studentId);

  const openEditor = (studentId: string) => {
    const existing = recordOf(studentId);
    setEditing(studentId);
    setStatus(existing?.status ?? "present");
    setMinutes(existing?.lateMinutes !== undefined ? String(existing.lateMinutes) : "");
    setNote(existing?.note ?? "");
    setReason("");
    setActionError(null);
  };

  const save = async (studentId: string) => {
    const existing = recordOf(studentId);
    const parsedMinutes = status === "late" ? Number(minutes) : 0;
    if (status === "late" && (!Number.isInteger(parsedMinutes) || parsedMinutes < 1)) {
      setActionError(message("att.minutesInvalid"));
      return;
    }
    setActionError(null);
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.markAttendance(
          accessToken,
          lessonId,
          studentId,
          {
            status,
            ...(status === "late" ? { lateMinutes: parsedMinutes } : {}),
            ...(note.trim() !== "" ? { note: note.trim() } : {}),
            ...(existing !== undefined ? { changeReason: reason.trim() } : {}),
          },
          idempotency.key(),
        ),
      );
      idempotency.complete();
      setEditing(null);
      await attendance.reload();
    } catch (cause) {
      if (!(cause instanceof ApiTransportError)) idempotency.abandon();
      setActionError(apiErrorMessage(cause));
      await attendance.reload();
    } finally {
      setBusy(false);
    }
  };

  return (
    <View style={styles.block}>
      <Text style={styles.sectionTitle}>{message("att.title")}</Text>
      {students.map((student) => {
        const record = recordOf(student.studentId);
        const isEditing = editing === student.studentId;
        return (
          <View key={student.studentId} style={styles.entry}>
            <StatusRow
              onPress={() => openEditor(student.studentId)}
              status={
                record !== undefined
                  ? recordLine(record, message)
                  : message("att.unmarked")
              }
              subtitle={
                record !== undefined ? message("att.edit") : message("att.mark")
              }
              testID={`attendance-row-${student.studentId}`}
              title={student.fullName}
              tone={
                record === undefined
                  ? "muted"
                  : record.status === "present"
                    ? "success"
                    : record.status === "late"
                      ? "warning"
                      : "danger"
              }
            />
            {isEditing ? (
              <View style={styles.editor}>
                <View style={styles.chips}>
                  {STATUS_ORDER.map((candidate) => (
                    <AreaChip
                      accent={
                        candidate === "present"
                          ? semantic.feedbackSuccess
                          : candidate === "late"
                            ? semantic.feedbackWarning
                            : semantic.feedbackDanger
                      }
                      active={status === candidate}
                      key={candidate}
                      label={statusLabel(candidate, message)}
                      onPress={() => setStatus(candidate)}
                      testID={`attendance-status-${candidate}`}
                    />
                  ))}
                </View>
                {status === "late" ? (
                  <PremiumTextField
                    keyboardType="number-pad"
                    label={message("att.minutes")}
                    onChangeText={setMinutes}
                    testID="attendance-minutes"
                    value={minutes}
                  />
                ) : null}
                <PremiumTextField
                  autoCapitalize="sentences"
                  helper={
                    status === "absent"
                      ? message("att.noteRequired")
                      : message("att.notePrivate")
                  }
                  label={message("att.note")}
                  multiline
                  onChangeText={setNote}
                  testID="attendance-note"
                  value={note}
                />
                {record !== undefined ? (
                  <PremiumTextField
                    autoCapitalize="sentences"
                    helper={message("att.reasonHelper")}
                    label={message("att.reason")}
                    onChangeText={setReason}
                    testID="attendance-reason"
                    value={reason}
                  />
                ) : null}
                {actionError !== null ? (
                  <InlineNotice
                    body={actionError}
                    title={message("common.retry")}
                    tone="error"
                  />
                ) : null}
                <BlockAction
                  busy={busy}
                  disabled={
                    (status === "absent" && note.trim() === "") ||
                    (record !== undefined && reason.trim() === "")
                  }
                  label={message("att.save")}
                  onPress={() => void save(student.studentId)}
                  testID="attendance-save"
                />
                <BlockAction
                  kind="secondary"
                  label={message("common.cancel")}
                  onPress={() => setEditing(null)}
                />
              </View>
            ) : null}
          </View>
        );
      })}
      {students.length > 1 ? (
        <Text style={styles.emptySeat}>{message("att.emptySeat")}</Text>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  block: { gap: space.s3 },
  sectionTitle: {
    color: semantic.textPrimary,
    marginTop: space.s2,
    ...typeStyles.headingM,
  },
  entry: { gap: space.s3 },
  editor: { gap: space.s3 },
  chips: { flexDirection: "row", flexWrap: "wrap", gap: space.s2 },
  emptySeat: { color: semantic.textMuted, ...typeStyles.caption },
});
