import { useMemo, useState } from "react";
import { RefreshControl, StyleSheet, Text } from "react-native";

import { useApiClient } from "@/api";
import type { CoreLessonSeries } from "@/api/contracts";
import { createIntentIdempotency } from "@/controllers";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice, PremiumTextField } from "../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../patterns/accountPatterns";
import { domainAccent } from "../domainAccent";
import { semantic, typeStyles } from "../tokens";
import { apiErrorMessage } from "../viewModels";
import { AccountNav, useAccountResource, useWorkingRole } from "./account/shared";

/**
 * Weekly series and rooms management for Owner/Administrator. The
 * lifecycle follows the schema: active ⇄ paused and the terminal
 * ended — pausing or ending stops future occurrence generation only,
 * already-scheduled Lessons stay and are changed through the explicit
 * Lesson operations. Series creation ships with the Page 29
 * administrator workspace.
 */

const WEEKDAYS = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"] as const;

export function seriesTimeLabel(series: Pick<CoreLessonSeries, "weekday" | "startMinutes" | "durationMinutes">): string {
  const day = WEEKDAYS[series.weekday] ?? "?";
  const hours = Math.floor(series.startMinutes / 60);
  const minutes = series.startMinutes % 60;
  const start = `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
  return `${day} · ${start} · ${series.durationMinutes} мин`;
}

export function SeriesManagementScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const workingRole = useWorkingRole();
  const manager = workingRole === "Owner" || workingRole === "Administrator";

  const series = useAccountResource((accessToken) => api.listLessonSeries(accessToken));
  const rooms = useAccountResource((accessToken) => api.listRooms(accessToken));
  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [roomName, setRoomName] = useState("");
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const [actionNote, setActionNote] = useState<string | null>(null);

  if (!manager) {
    return (
      <AccountScreenShell navigation={<AccountNav active="schedule" />} testID="series-guard">
        <InlineNotice
          body={message("ser.guardBody")}
          title={message("ser.guardTitle")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const items = series.value ?? [];
  const selected = items.find((item) => item.id === selectedId) ?? null;

  const run = async (
    action: (accessToken: string) => Promise<unknown>,
    note: string | null,
  ) => {
    setActionError(null);
    setActionNote(null);
    setBusy(true);
    try {
      await runAuthenticated(action);
      await Promise.all([series.reload(), rooms.reload()]);
      if (note !== null) setActionNote(note);
      return true;
    } catch (cause) {
      setActionError(apiErrorMessage(cause));
      return false;
    } finally {
      setBusy(false);
    }
  };

  const changeStatus = async (status: "active" | "paused" | "ended") => {
    if (selected === null) return;
    const done = await run(
      (accessToken) =>
        api.changeLessonSeriesStatus(
          accessToken,
          selected.id,
          { status, expectedVersion: selected.version },
          idempotency.key(),
        ),
      message(`ser.changed.${status}`),
    );
    if (done) idempotency.complete();
  };

  const generate = async () => {
    if (selected === null) return;
    const done = await run(
      (accessToken) =>
        api.generateSeriesOccurrences(
          accessToken,
          selected.id,
          { weeks: 4 },
          idempotency.key(),
        ),
      message("ser.generated"),
    );
    if (done) idempotency.complete();
  };

  const addRoom = async () => {
    const name = roomName.trim();
    if (name === "") return;
    const done = await run(
      (accessToken) => api.createRoom(accessToken, { name }),
      message("ser.roomCreated"),
    );
    if (done) setRoomName("");
  };

  return (
    <AccountScreenShell
      navigation={<AccountNav active="schedule" />}
      refreshControl={
        <RefreshControl
          onRefresh={() => {
            void series.reload();
            void rooms.reload();
          }}
          refreshing={series.refreshing || rooms.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="series-management"
    >
      <ScreenHeading
        accent={domainAccent("schedule")}
        eyebrow={message("ser.eyebrow")}
        subtitle={message("ser.subtitle")}
        title={message("ser.title")}
      />
      {series.error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={apiErrorMessage(series.error)}
          onAction={() => void series.reload()}
          title={message("ser.title")}
        />
      ) : null}
      {actionError !== null ? (
        <InlineNotice body={actionError} title={message("common.retry")} tone="error" />
      ) : null}
      {actionNote !== null ? (
        <InlineNotice body={actionNote} title={message("ser.done")} tone="success" />
      ) : null}
      {selected === null ? (
        <>
          {series.value !== null && items.length === 0 ? (
            <Text style={styles.muted}>{message("ser.empty")}</Text>
          ) : null}
          {items.map((item) => (
            <StatusRow
              key={item.id}
              onPress={() => {
                setActionNote(null);
                setSelectedId(item.id);
              }}
              status={message(`ser.status.${item.status === "active" ? "active" : item.status === "paused" ? "paused" : "ended"}`)}
              subtitle={`${seriesTimeLabel(item)} · ${item.teacher.fullName}`}
              testID={`series-${item.id}`}
              title={`${item.title} · ${message(
                item.format === "group" ? "ser.format.group" : "ser.format.individual",
              )}`}
              tone={
                item.status === "active"
                  ? "success"
                  : item.status === "paused"
                    ? "warning"
                    : "muted"
              }
            />
          ))}
          <ScreenHeading
            accent={domainAccent("schedule")}
            eyebrow={message("ser.rooms.eyebrow")}
            subtitle={message("ser.rooms.subtitle")}
            title={message("ser.rooms.title")}
          />
          {(rooms.value ?? []).map((room) => (
            <StatusRow
              key={room.id}
              status={message("ser.rooms.active")}
              subtitle={
                room.capacity !== undefined
                  ? message("ser.rooms.capacity", { count: room.capacity })
                  : message("ser.rooms.noCapacity")
              }
              testID={`room-${room.id}`}
              title={room.name}
              tone="info"
            />
          ))}
          <PremiumTextField
            label={message("ser.rooms.nameLabel")}
            onChangeText={setRoomName}
            placeholder={message("ser.rooms.namePlaceholder")}
            testID="room-name-input"
            value={roomName}
          />
          <BlockAction
            busy={busy}
            disabled={roomName.trim() === ""}
            kind="secondary"
            label={message("ser.rooms.add")}
            onPress={() => void addRoom()}
            testID="room-add"
          />
        </>
      ) : (
        <>
          <StatusCard
            body={`${seriesTimeLabel(selected)} · ${selected.teacher.fullName}`}
            status={message(`ser.status.${selected.status === "active" ? "active" : selected.status === "paused" ? "paused" : "ended"}`)}
            title={selected.title}
            tone={selected.status === "active" ? "success" : "warning"}
          />
          <StatusCard
            body={message("ser.lifecycleBody")}
            status={message("ser.lifecycleFooter")}
            title={message("ser.lifecycleTitle")}
            tone="info"
          />
          {selected.status === "active" ? (
            <>
              <BlockAction
                busy={busy}
                label={message("ser.generate")}
                onPress={() => void generate()}
                testID="series-generate"
              />
              <BlockAction
                busy={busy}
                kind="secondary"
                label={message("ser.pause")}
                onPress={() => void changeStatus("paused")}
                testID="series-pause"
              />
            </>
          ) : null}
          {selected.status === "paused" ? (
            <BlockAction
              busy={busy}
              label={message("ser.resume")}
              onPress={() => void changeStatus("active")}
              testID="series-resume"
            />
          ) : null}
          {selected.status !== "ended" ? (
            <BlockAction
              busy={busy}
              kind="secondary"
              label={message("ser.end")}
              onPress={() => void changeStatus("ended")}
              testID="series-end"
            />
          ) : null}
          <BlockAction
            kind="secondary"
            label={message("common.back")}
            onPress={() => setSelectedId(null)}
            testID="series-back"
          />
        </>
      )}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  muted: { color: semantic.textSecondary, ...typeStyles.bodyS },
});
