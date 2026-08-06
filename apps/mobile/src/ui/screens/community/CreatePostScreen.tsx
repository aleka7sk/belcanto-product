import { router } from "expo-router";
import { useMemo, useState } from "react";
import { StyleSheet, Text, View } from "react-native";

import { useApiClient } from "@/api";
import { createIntentIdempotency } from "@/controllers";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { InlineNotice, PremiumTextField } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
  ToggleRow,
} from "../../patterns/accountPatterns";
import { AreaChip } from "../../patterns/journalPatterns";
import { semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav, useWorkingRole } from "../account/shared";

/**
 * Composer (Figma COM-POST-02). Text-only while DEC-103 (guardian
 * consent for community media) stays open — the media slot states that
 * honestly instead of pretending. Staff additionally publish official
 * announcements: titled, optionally pinned, replies closed by default
 * (COM-ANN-01).
 */

export function CreatePostScreen() {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const workingRole = useWorkingRole();
  const staff = workingRole !== null && workingRole !== "Student";

  const idempotency = useMemo(() => createIntentIdempotency(), []);
  const [body, setBody] = useState("");
  const [title, setTitle] = useState("");
  const [announcement, setAnnouncement] = useState(false);
  const [staffAudience, setStaffAudience] = useState(false);
  const [pinned, setPinned] = useState(false);
  const [commentsEnabled, setCommentsEnabled] = useState(true);
  const [busy, setBusy] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const submit = async () => {
    const trimmedBody = body.trim();
    const trimmedTitle = title.trim();
    if (trimmedBody === "") return;
    setFormError(null);
    setBusy(true);
    try {
      const created = await runAuthenticated((accessToken) =>
        api.createCommunityPost(
          accessToken,
          {
            body: trimmedBody,
            commentsEnabled,
            ...(announcement
              ? { kind: "announcement" as const, title: trimmedTitle, pinned }
              : {}),
            ...(staffAudience ? { audience: "staff" as const } : {}),
          },
          idempotency.key(),
        ),
      );
      idempotency.complete();
      router.replace({
        pathname: "/(protected)/community/[postId]",
        params: { postId: created.id },
      });
    } catch (cause) {
      setFormError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  const missingTitle = announcement && title.trim() === "";

  return (
    <AccountScreenShell navigation={<AccountNav active="community" />} testID="community-create">
      <ScreenHeading
        eyebrow={message("com.create.eyebrow")}
        subtitle={message("com.create.subtitle")}
        title={message("com.create.title")}
      />
      {staff ? (
        <View style={styles.kindRow}>
          <AreaChip
            accent={semantic.accentViolet}
            active={!announcement}
            label={message("com.create.kindPost")}
            onPress={() => setAnnouncement(false)}
            testID="community-kind-post"
          />
          <AreaChip
            accent={semantic.accentGold}
            active={announcement}
            label={message("com.create.kindAnnouncement")}
            onPress={() => {
              setAnnouncement(true);
              setCommentsEnabled(false);
            }}
            testID="community-kind-announcement"
          />
        </View>
      ) : null}
      {announcement ? (
        <PremiumTextField
          label={message("com.create.titleLabel")}
          onChangeText={setTitle}
          placeholder={message("com.create.titlePlaceholder")}
          testID="community-title-input"
          value={title}
        />
      ) : null}
      <PremiumTextField
        label={message("com.create.bodyLabel")}
        multiline
        onChangeText={setBody}
        placeholder={message("com.create.bodyPlaceholder")}
        testID="community-body-input"
        value={body}
      />
      <Text style={styles.counter}>
        {message("com.create.counter", { count: body.trim().length })}
      </Text>
      <StatusCard
        body={
          staffAudience
            ? message("com.create.audienceStaffBody")
            : message("com.create.audienceSchoolBody")
        }
        status={message(staffAudience ? "com.audience.staff" : "com.audience.school")}
        title={message("com.create.audienceTitle")}
        tone="info"
      />
      {staff ? (
        <ToggleRow
          offLabel={message("com.audience.school")}
          onLabel={message("com.audience.staff")}
          onToggle={(next) => setStaffAudience(next)}
          subtitle={message("com.create.audienceToggleHint")}
          testID="community-audience-toggle"
          title={message("com.create.audienceToggle")}
          value={staffAudience}
        />
      ) : null}
      <ToggleRow
        offLabel={message("com.commentsOff")}
        onLabel={message("com.create.commentsOn")}
        onToggle={(next) => setCommentsEnabled(next)}
        subtitle={message("com.create.commentsToggleHint")}
        testID="community-comments-toggle"
        title={message("com.create.commentsToggle")}
        value={commentsEnabled}
      />
      {announcement ? (
        <ToggleRow
          offLabel={message("com.create.pinnedOff")}
          onLabel={message("com.pinned")}
          onToggle={(next) => setPinned(next)}
          subtitle={message("com.create.pinnedToggleHint")}
          testID="community-pinned-toggle"
          title={message("com.create.pinnedToggle")}
          value={pinned}
        />
      ) : null}
      <StatusCard
        body={message("com.create.mediaBody")}
        status={message("com.create.mediaFooter")}
        title={message("com.create.mediaTitle")}
        tone="muted"
      />
      {formError !== null ? (
        <InlineNotice body={formError} title={message("common.retry")} tone="error" />
      ) : null}
      <BlockAction
        busy={busy}
        disabled={body.trim() === "" || missingTitle}
        label={message("com.create.publish")}
        onPress={() => void submit()}
        testID="community-publish"
      />
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  kindRow: { flexDirection: "row", gap: space.s2 },
  counter: {
    color: semantic.textMuted,
    ...typeStyles.labelM,
  },
});
