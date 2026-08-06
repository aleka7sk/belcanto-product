import { useLocalSearchParams } from "expo-router";
import { useMemo, useState } from "react";
import { RefreshControl, StyleSheet, Text, View } from "react-native";

import { ApiTransportError, useApiClient } from "@/api";
import type { SongStage, StudentSong } from "@/api/contracts";
import { SONG_STAGES } from "@/api/contracts";
import { createIntentIdempotency } from "@/controllers";
import { useMessage, type MessageFormatter } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice, PremiumTextField } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  StatusRow,
} from "../../patterns/accountPatterns";
import { AreaChip } from "../../patterns/journalPatterns";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage, formatBelcantoDate } from "../../viewModels";
import { AccountNav, useAccountResource } from "../account/shared";

/**
 * Repertoire (Figma STU-GROWTH-07 + STU-PRACTICE-10/11). Each piece
 * walks the explicit journey «Знакомство → … → готово к сцене»; the
 * path bar shows the journey position — a named step, never a score
 * (DEC-006). The Student reads; the assigned Teacher moves the stage
 * with an append-only history behind it.
 */

export function songStageLabel(stage: SongStage, message: MessageFormatter): string {
  return message(`rep.stage.${stage}`);
}

function stageIndex(stage: SongStage): number {
  return SONG_STAGES.indexOf(stage);
}

function JourneyBar({ stage }: { stage: SongStage }) {
  const position = stageIndex(stage);
  return (
    <View style={styles.journey}>
      {SONG_STAGES.map((candidate, index) => (
        <View
          key={candidate}
          style={[
            styles.journeySegment,
            {
              backgroundColor:
                index <= position ? semantic.accentViolet : semantic.borderDefault,
            },
          ]}
        />
      ))}
    </View>
  );
}

type RepertoireFilter = "all" | "inWork" | "ready";

export function RepertoireScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { state, runAuthenticated } = useSession();
  const params = useLocalSearchParams<{ studentId?: string }>();
  const paramStudentId = typeof params.studentId === "string" ? params.studentId : null;
  const studentId = paramStudentId ?? state.bootstrap?.studentId ?? null;

  const songs = useAccountResource((accessToken) =>
    studentId === null
      ? Promise.resolve<StudentSong[]>([])
      : api.listStudentSongs(accessToken, studentId),
  );
  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [filter, setFilter] = useState<RepertoireFilter>("all");
  const [openSongId, setOpenSongId] = useState<string | null>(null);
  const [editStage, setEditStage] = useState<SongStage | null>(null);
  const [editNote, setEditNote] = useState("");
  const [adding, setAdding] = useState(false);
  const [newTitle, setNewTitle] = useState("");
  const [newArtist, setNewArtist] = useState("");
  const [busy, setBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  if (studentId === null) {
    return (
      <AccountScreenShell navigation={<AccountNav />} testID="repertoire-guard">
        <InlineNotice
          body={message("prac.guard.body")}
          title={message("prac.guard.title")}
          tone="error"
        />
      </AccountScreenShell>
    );
  }

  const list = songs.value ?? [];
  const visible = list.filter((song) => {
    if (filter === "inWork") return song.stage !== "stage_ready";
    if (filter === "ready") return song.stage === "stage_ready";
    return true;
  });
  const empty = songs.value !== null && list.length === 0;
  const canLead =
    state.bootstrap !== null &&
    (state.bootstrap.roles.includes("Teacher") ||
      state.bootstrap.roles.includes("Administrator"));

  const openEditor = (song: StudentSong) => {
    setOpenSongId(song.id);
    setEditStage(song.stage);
    setEditNote(song.stageNote ?? "");
    setActionError(null);
  };

  const addSong = async () => {
    setActionError(null);
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.addStudentSong(
          accessToken,
          studentId,
          {
            title: newTitle.trim(),
            ...(newArtist.trim() !== "" ? { artist: newArtist.trim() } : {}),
          },
          idempotency.key(),
        ),
      );
      idempotency.complete();
      setAdding(false);
      setNewTitle("");
      setNewArtist("");
      await songs.reload();
    } catch (cause) {
      if (!(cause instanceof ApiTransportError)) idempotency.abandon();
      setActionError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  const saveStage = async (song: StudentSong) => {
    if (editStage === null) return;
    setActionError(null);
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.changeSongStage(
          accessToken,
          song.id,
          {
            stage: editStage,
            ...(editNote.trim() !== "" ? { stageNote: editNote.trim() } : {}),
            expectedVersion: song.version,
          },
          idempotency.key(),
        ),
      );
      idempotency.complete();
      setOpenSongId(null);
      await songs.reload();
    } catch (cause) {
      if (!(cause instanceof ApiTransportError)) idempotency.abandon();
      setActionError(apiErrorMessage(cause));
      await songs.reload();
    } finally {
      setBusy(false);
    }
  };

  return (
    <AccountScreenShell
      navigation={<AccountNav active="practice" />}
      refreshControl={
        <RefreshControl
          onRefresh={() => void songs.reload()}
          refreshing={songs.refreshing}
          tintColor={semantic.accentViolet}
        />
      }
      testID="repertoire"
    >
      <ScreenHeading
        eyebrow={message("rep.eyebrow")}
        subtitle={
          empty
            ? message("rep.empty.body")
            : message("rep.subtitle", { count: list.length })
        }
        title={empty ? message("rep.empty.title") : message("rep.title")}
      />
      {songs.error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={apiErrorMessage(songs.error)}
          onAction={() => void songs.reload()}
          title={message("rep.title")}
        />
      ) : null}
      {!empty ? (
        <View style={styles.filters}>
          <AreaChip
            accent={semantic.accentViolet}
            active={filter === "all"}
            label={message("rep.filter.all")}
            onPress={() => setFilter("all")}
            testID="repertoire-filter-all"
          />
          <AreaChip
            accent={semantic.accentCyan}
            active={filter === "inWork"}
            label={message("rep.filter.inWork")}
            onPress={() => setFilter("inWork")}
          />
          <AreaChip
            accent={semantic.accentGold}
            active={filter === "ready"}
            label={message("rep.filter.ready")}
            onPress={() => setFilter("ready")}
          />
        </View>
      ) : null}
      {empty ? (
        <StatusCard
          body={message("rep.empty.card.body")}
          status={message("rep.empty.card.status")}
          title={message("rep.empty.card.title")}
          tone="info"
        />
      ) : null}
      {canLead ? (
        adding ? (
          <View style={styles.detail}>
            <Text style={styles.detailTitle}>{message("rep.add.title")}</Text>
            <PremiumTextField
              autoCapitalize="sentences"
              label={message("rep.add.songTitle")}
              onChangeText={setNewTitle}
              testID="repertoire-add-title"
              value={newTitle}
            />
            <PremiumTextField
              autoCapitalize="words"
              label={message("rep.add.artist")}
              onChangeText={setNewArtist}
              testID="repertoire-add-artist"
              value={newArtist}
            />
            {actionError !== null && openSongId === null ? (
              <InlineNotice
                body={actionError}
                title={message("common.retry")}
                tone="error"
              />
            ) : null}
            <BlockAction
              busy={busy}
              disabled={newTitle.trim() === ""}
              label={message("rep.add.save")}
              onPress={() => void addSong()}
              testID="repertoire-add-save"
            />
            <BlockAction
              kind="secondary"
              label={message("common.cancel")}
              onPress={() => setAdding(false)}
            />
          </View>
        ) : (
          <BlockAction
            kind="secondary"
            label={message("rep.add.title")}
            onPress={() => {
              setAdding(true);
              setActionError(null);
            }}
            testID="repertoire-add-open"
          />
        )
      ) : null}
      {visible.map((song) => {
        const open = openSongId === song.id;
        return (
          <View key={song.id} style={styles.songBlock}>
            <StatusRow
              onPress={() => (open ? setOpenSongId(null) : openEditor(song))}
              status={songStageLabel(song.stage, message)}
              subtitle={[song.artist, song.stageNote]
                .filter((part): part is string => part !== undefined && part !== "")
                .join(" · ")}
              testID={`repertoire-song-${song.id}`}
              title={song.title}
              tone={song.stage === "stage_ready" ? "success" : "info"}
            />
            {open ? (
              <View style={styles.detail}>
                <Text style={styles.detailTitle}>{message("rep.journey.title")}</Text>
                <JourneyBar stage={song.stage} />
                <Text style={styles.journeyLine}>
                  {SONG_STAGES.map((candidate) =>
                    songStageLabel(candidate, message),
                  ).join(" → ")}
                </Text>
                <Text style={styles.metaLine}>
                  {message("rep.assignedBy", {
                    name: song.assignedBy.fullName,
                    date: formatBelcantoDate(song.createdAt),
                  })}
                </Text>
                {song.history.map((change) => (
                  <StatusRow
                    key={change.changedAt}
                    status={
                      change.fromStage !== undefined
                        ? `${songStageLabel(change.fromStage, message)} → ${songStageLabel(change.toStage, message)}`
                        : songStageLabel(change.toStage, message)
                    }
                    subtitle={change.note ?? ""}
                    title={formatBelcantoDate(change.changedAt)}
                    tone="muted"
                  />
                ))}
                {canLead ? (
                  <>
                    <Text style={styles.detailTitle}>
                      {message("rep.edit.title")}
                    </Text>
                    <View style={styles.filters}>
                      {SONG_STAGES.map((candidate) => (
                        <AreaChip
                          accent={semantic.accentViolet}
                          active={editStage === candidate}
                          key={candidate}
                          label={songStageLabel(candidate, message)}
                          onPress={() => setEditStage(candidate)}
                          testID={`repertoire-stage-${candidate}`}
                        />
                      ))}
                    </View>
                    <PremiumTextField
                      autoCapitalize="sentences"
                      label={message("rep.edit.note")}
                      multiline
                      onChangeText={setEditNote}
                      testID="repertoire-stage-note"
                      value={editNote}
                    />
                    {actionError !== null ? (
                      <InlineNotice
                        body={actionError}
                        title={message("common.retry")}
                        tone="error"
                      />
                    ) : null}
                    <BlockAction
                      busy={busy}
                      disabled={editStage === null}
                      label={message("rep.edit.save")}
                      onPress={() => void saveStage(song)}
                      testID="repertoire-stage-save"
                    />
                  </>
                ) : null}
              </View>
            ) : null}
          </View>
        );
      })}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  filters: { flexDirection: "row", flexWrap: "wrap", gap: space.s2 },
  songBlock: { gap: space.s3 },
  detail: {
    backgroundColor: semantic.bgSurface,
    borderColor: semantic.borderAccent,
    borderRadius: 20,
    borderWidth: 1,
    gap: space.s3,
    padding: space.s4,
  },
  detailTitle: { color: semantic.textPrimary, ...typeStyles.headingM },
  journey: { flexDirection: "row", gap: space.s1, height: 6 },
  journeySegment: { borderRadius: 999, flex: 1, height: 6 },
  journeyLine: { color: semantic.textSecondary, ...typeStyles.caption },
  metaLine: { color: semantic.textMuted, ...typeStyles.caption },
});
