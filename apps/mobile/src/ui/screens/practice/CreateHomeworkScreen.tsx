import { router, useLocalSearchParams } from "expo-router";
import { useMemo, useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { ApiTransportError, useApiClient } from "@/api";
import type { HomeworkTaskInput, IsoDateTime } from "@/api/contracts";
import { createIntentIdempotency } from "@/controllers";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { parseAlmatyLocalDateTime } from "@/validation/datetime";
import { InlineNotice, PremiumTextField, TextAction } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
} from "../../patterns/accountPatterns";
import { EventDetailCard } from "../../patterns/eventPatterns";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav } from "../account/shared";

/**
 * New homework (Figma TCH-JOURNAL-06 «Новое домашнее», flow E start).
 * The Lesson's Teacher assembles goal, optional deadline in Asia/Almaty
 * civil time, a short task plan and readiness criteria; saving assigns
 * the homework to the Student immediately. Media materials join once
 * the recorder/picker modules land (recorded phase-7 deferral).
 */

interface TaskDraft {
  title: string;
  minutes: string;
}

export function CreateHomeworkScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ occurrenceId?: string; studentId?: string }>();
  const occurrenceId = typeof params.occurrenceId === "string" ? params.occurrenceId : "";
  const studentId = typeof params.studentId === "string" ? params.studentId : "";

  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [goal, setGoal] = useState("");
  const [readiness, setReadiness] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [dueTime, setDueTime] = useState("");
  const [tasks, setTasks] = useState<TaskDraft[]>([]);
  const [busy, setBusy] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  if (occurrenceId === "" || studentId === "") {
    return (
      <AccountScreenShell navigation={<AccountNav active="schedule" />} testID="homework-create-guard">
        <InlineNotice
          body={message("prac.guard.body")}
          title={message("prac.guard.title")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const dueProvided = dueDate.trim() !== "" || dueTime.trim() !== "";
  const submit = async () => {
    setFormError(null);
    let dueAt: IsoDateTime | undefined;
    if (dueProvided) {
      const timestamp = parseAlmatyLocalDateTime(dueDate, dueTime);
      if (timestamp === null || timestamp <= Date.now()) {
        setFormError(message("prac.create.dueInvalid"));
        return;
      }
      dueAt = new Date(timestamp).toISOString() as IsoDateTime;
    }
    const taskInputs: HomeworkTaskInput[] = [];
    for (const task of tasks) {
      if (task.title.trim() === "") continue;
      const minutes = task.minutes.trim() === "" ? 0 : Number(task.minutes);
      if (!Number.isInteger(minutes) || minutes < 0 || minutes > 600) {
        setFormError(message("prac.create.minutesInvalid"));
        return;
      }
      taskInputs.push({
        title: task.title.trim(),
        ...(minutes > 0 ? { recommendedMinutes: minutes } : {}),
      });
    }
    setBusy(true);
    try {
      const created = await runAuthenticated((accessToken) =>
        api.createHomework(
          accessToken,
          {
            occurrenceId,
            studentId,
            goal: goal.trim(),
            ...(readiness.trim() !== "" ? { readinessCriteria: readiness.trim() } : {}),
            ...(dueAt !== undefined ? { dueAt } : {}),
            ...(taskInputs.length > 0 ? { tasks: taskInputs } : {}),
            assign: true,
          },
          idempotency.key(),
        ),
      );
      idempotency.complete();
      router.replace({
        pathname: "/(protected)/practice/[homeworkId]",
        params: { homeworkId: created.id },
      });
    } catch (cause) {
      if (!(cause instanceof ApiTransportError)) idempotency.abandon();
      setFormError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AccountScreenShell navigation={<AccountNav active="schedule" />} testID="homework-create">
      <ScreenHeading
        eyebrow={message("prac.create.eyebrow")}
        subtitle={message("prac.create.subtitle")}
        title={message("prac.create.title")}
      />
      <PremiumTextField
        autoCapitalize="sentences"
        helper={message("jrnl.field.visible")}
        label={message("prac.goal.title")}
        multiline
        onChangeText={setGoal}
        placeholder={message("prac.create.goalHint")}
        testID="homework-create-goal"
        value={goal}
      />
      <PremiumTextField
        autoCapitalize="sentences"
        label={message("prac.create.readiness")}
        multiline
        onChangeText={setReadiness}
        placeholder={message("prac.create.readinessHint")}
        testID="homework-create-readiness"
        value={readiness}
      />
      <View style={styles.dueRow}>
        <View style={styles.dueField}>
          <PremiumTextField
            label={message("prac.create.dueDate")}
            onChangeText={setDueDate}
            placeholder="10.08.2026"
            testID="homework-create-due-date"
            value={dueDate}
          />
        </View>
        <View style={styles.dueField}>
          <PremiumTextField
            label={message("prac.create.dueTime")}
            onChangeText={setDueTime}
            placeholder="18:00"
            testID="homework-create-due-time"
            value={dueTime}
          />
        </View>
      </View>
      <Text style={styles.sectionTitle}>{message("prac.plan.title")}</Text>
      {tasks.map((task, index) => (
        <View key={index} style={styles.taskRow}>
          <PremiumTextField
            autoCapitalize="sentences"
            label={message("prac.create.taskTitle", { position: index + 1 })}
            onChangeText={(value) =>
              setTasks((current) =>
                current.map((entry, at) =>
                  at === index ? { ...entry, title: value } : entry,
                ),
              )
            }
            testID={`homework-create-task-${index}`}
            value={task.title}
          />
          <PremiumTextField
            keyboardType="number-pad"
            label={message("prac.create.taskMinutes")}
            onChangeText={(value) =>
              setTasks((current) =>
                current.map((entry, at) =>
                  at === index ? { ...entry, minutes: value } : entry,
                ),
              )
            }
            value={task.minutes}
          />
          <TextAction
            align="right"
            label={message("jrnl.evidence.remove")}
            onPress={() =>
              setTasks((current) => current.filter((_, at) => at !== index))
            }
          />
        </View>
      ))}
      {tasks.length < 10 ? (
        <BlockAction
          kind="secondary"
          label={message("prac.create.addTask")}
          onPress={() => setTasks((current) => [...current, { title: "", minutes: "" }])}
          testID="homework-create-add-task"
        />
      ) : null}
      <EventDetailCard
        accent={semantic.accentCyan}
        body={message("prac.create.materialsBody")}
        status={message("prac.create.materialsStatus")}
        statusColor={semantic.accentCyan}
        title={message("prac.materials.title")}
      />
      {formError !== null ? (
        <InlineNotice body={formError} title={message("common.retry")} tone="error" />
      ) : null}
      <BlockAction
        busy={busy}
        disabled={goal.trim() === ""}
        label={message("prac.create.save")}
        onPress={() => void submit()}
        testID="homework-create-save"
      />
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  dueRow: { flexDirection: "row", gap: space.s3 },
  dueField: { flex: 1 },
  sectionTitle: {
    color: semantic.textPrimary,
    marginTop: space.s2,
    ...typeStyles.headingM,
  },
  taskRow: { gap: space.s3 },
});
