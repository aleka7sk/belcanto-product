import { router } from "expo-router";
import { useState } from "react";
import { Pressable, StyleSheet, Text, View } from "react-native";

import { navigationRoleFor } from "@/access/activeRole";
import { useApiClient } from "@/api";
import type { CommunityPost, PolicyVersion } from "@/api/contracts";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import { ErrorNotice, InlineNotice } from "../../components";
import {
  AccountScreenShell,
  BlockAction,
  ScreenHeading,
  StatusCard,
} from "../../patterns/accountPatterns";
import { CommunityPostCard } from "../../patterns/communityPatterns";
import { PremiumContextHero } from "../../patterns/premiumContextHero";
import { SegmentedControl } from "../../segmentedControl";
import { radius, semantic, space, typeStyles } from "../../tokens";
import { apiErrorMessage } from "../../viewModels";
import { AccountNav, useAccountResource, useWorkingRole } from "../account/shared";

/**
 * Community home (Figma COM-HOME-01/02). One space for the whole school:
 * feed with pinned announcements, the Events tab linking to the
 * canonical Page 24 catalog, and the Chats tab stating the safety
 * boundary — direct Student↔Student messaging does not exist. First
 * entry passes the community guidelines gate (COM-SAFE-04) backed by
 * the newest effective policy version of kind «community».
 */

type CommunityTab = "feed" | "events" | "chats";

/** The newest effective community guidelines pending acceptance. */
export function pendingCommunityGuidelines(
  policies: readonly PolicyVersion[],
  now: Date,
): PolicyVersion | null {
  let newest: PolicyVersion | null = null;
  for (const policy of policies) {
    if (policy.kind !== "community") continue;
    if (new Date(policy.effectiveFrom).getTime() > now.getTime()) continue;
    if (newest === null || policy.effectiveFrom > newest.effectiveFrom) {
      newest = policy;
    }
  }
  if (newest === null || newest.acceptedAt !== undefined) return null;
  return newest;
}

export function announcementCount(posts: readonly CommunityPost[]): number {
  return posts.filter((post) => post.kind === "announcement").length;
}

function GuidelinesGate({
  policy,
  onAccepted,
}: {
  policy: PolicyVersion;
  onAccepted: () => Promise<void>;
}) {
  const message = useMessage();
  const api = useApiClient();
  const { runAuthenticated } = useSession();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const accept = async () => {
    setError(null);
    setBusy(true);
    try {
      await runAuthenticated((accessToken) =>
        api.acceptPolicy(accessToken, { policyVersionId: policy.id }),
      );
      await onAccepted();
    } catch (cause) {
      setError(apiErrorMessage(cause));
    } finally {
      setBusy(false);
    }
  };

  return (
    <>
      <ScreenHeading
        eyebrow={message("com.gate.eyebrow")}
        subtitle={message("com.gate.subtitle", { version: policy.version })}
        title={policy.title}
      />
      <StatusCard
        body={message("com.gate.support.body")}
        status={message("com.gate.support.footer")}
        title={message("com.gate.support.title")}
        tone="info"
      />
      <StatusCard
        body={message("com.gate.privacy.body")}
        status={message("com.gate.privacy.footer")}
        title={message("com.gate.privacy.title")}
        tone="info"
      />
      <StatusCard
        body={message("com.gate.report.body")}
        status={message("com.gate.report.footer")}
        title={message("com.gate.report.title")}
        tone="info"
      />
      {error !== null ? (
        <InlineNotice body={error} title={message("common.retry")} tone="error" />
      ) : null}
      <BlockAction
        busy={busy}
        label={message("com.gate.accept")}
        onPress={() => void accept()}
        testID="community-guidelines-accept"
      />
    </>
  );
}

export function CommunityScreen() {
  const message = useMessage();
  const api = useApiClient();
  const workingRole = useWorkingRole();
  const [tab, setTab] = useState<CommunityTab>("feed");
  const [announcementsOnly, setAnnouncementsOnly] = useState(false);

  const policies = useAccountResource((accessToken) => api.listPolicies(accessToken));
  const feed = useAccountResource((accessToken) => api.communityFeed(accessToken));

  const pendingGuidelines =
    policies.value === null
      ? null
      : pendingCommunityGuidelines(policies.value, new Date());
  const posts = feed.value ?? [];
  const visible = announcementsOnly
    ? posts.filter((post) => post.kind === "announcement")
    : posts;
  const moderator = workingRole === "Owner" || workingRole === "Administrator";

  const reloadAll = async () => {
    await Promise.all([policies.reload(), feed.reload()]);
  };

  return (
    <AccountScreenShell
      navigation={<AccountNav active="community" />}
      testID="community-home"
    >
      {policies.error !== null ? (
        <ErrorNotice
          actionLabel={message("common.retry")}
          body={apiErrorMessage(policies.error)}
          onAction={() => void policies.reload()}
          title={message("com.gate.eyebrow")}
        />
      ) : null}
      {pendingGuidelines !== null ? (
        <GuidelinesGate onAccepted={reloadAll} policy={pendingGuidelines} />
      ) : (
        <>
          <PremiumContextHero
            body={message("com.hero.body")}
            eyebrow={message("com.hero.eyebrow")}
            {...(feed.value === null
              ? {}
              : { metric: message("com.hero.metric", { count: posts.length }) })}
            role={workingRole === null ? "Student" : navigationRoleFor(workingRole)}
            title={message("com.hero.title")}
          />
          <SegmentedControl<CommunityTab>
            accessibilityLabel={message("com.hero.title")}
            active={tab}
            activeColor={semantic.bgCommunity}
            items={[
              { key: "feed", label: message("com.tab.feed") },
              { key: "events", label: message("com.tab.events"), role: "link" },
              { key: "chats", label: message("com.tab.chats") },
            ]}
            onSelect={(key) => {
              if (key === "events") {
                router.push("/(protected)/events");
                return;
              }
              setTab(key);
            }}
            testIDPrefix="community-tab"
          />
          {feed.error !== null ? (
            <ErrorNotice
              actionLabel={message("common.retry")}
              body={apiErrorMessage(feed.error)}
              onAction={() => void feed.reload()}
              title={message("com.hero.title")}
            />
          ) : null}
          {tab === "chats" ? (
            <>
              <StatusCard
                body={message("com.chats.rule.body")}
                status={message("com.chats.rule.footer")}
                title={message("com.chats.rule.title")}
                tone="info"
              />
              <Text style={styles.deferred}>{message("com.chats.deferred")}</Text>
            </>
          ) : (
            <>
              <Pressable
                accessibilityLabel={message("com.space.announcements")}
                accessibilityRole="switch"
                accessibilityState={{ checked: announcementsOnly }}
                onPress={() => setAnnouncementsOnly((current) => !current)}
                style={[styles.spaceCard, announcementsOnly && styles.spaceCardActive]}
                testID="community-space-announcements"
              >
                <View style={styles.spaceText}>
                  <Text style={styles.spaceTitle}>{message("com.space.announcements")}</Text>
                  <Text style={styles.spaceBody}>{message("com.space.announcementsBody")}</Text>
                  <Text style={styles.spaceCount}>
                    {message("com.space.count", { count: announcementCount(posts) })}
                  </Text>
                </View>
              </Pressable>
              {feed.value !== null && visible.length === 0 ? (
                <Text style={styles.deferred}>{message("com.feed.empty")}</Text>
              ) : null}
              {visible.map((post) => (
                <CommunityPostCard
                  key={post.id}
                  message={message}
                  onOpen={() =>
                    router.push({
                      pathname: "/(protected)/community/[postId]",
                      params: { postId: post.id },
                    })
                  }
                  post={post}
                  testID={`community-post-${post.id}`}
                />
              ))}
              <BlockAction
                label={message("com.compose")}
                onPress={() => router.push("/(protected)/community/create")}
                testID="community-compose"
              />
              {moderator ? (
                <BlockAction
                  kind="secondary"
                  label={message("com.moderation.open")}
                  onPress={() => router.push("/(protected)/community/moderation")}
                  testID="community-moderation-link"
                />
              ) : null}
            </>
          )}
        </>
      )}
    </AccountScreenShell>
  );
}

const styles = StyleSheet.create({
  spaceCard: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderGlass,
    borderRadius: radius.xl,
    borderWidth: 1,
    padding: space.s4,
  },
  spaceCardActive: { borderColor: semantic.accentGold },
  spaceText: { gap: space.s2 },
  spaceTitle: {
    color: semantic.textPrimary,
    ...typeStyles.labelL,
  },
  spaceBody: {
    color: semantic.textSecondary,
    ...typeStyles.caption,
  },
  spaceCount: {
    color: semantic.accentGold,
    ...typeStyles.labelM,
  },
  deferred: {
    color: semantic.textSecondary,
    ...typeStyles.bodyS,
  },
});
