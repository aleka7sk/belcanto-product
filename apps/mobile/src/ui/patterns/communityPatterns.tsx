import { Pressable, StyleSheet, Text, View } from "react-native";

import type { CommunityComment, CommunityPost } from "@/api/contracts";
import type { MessageFormatter } from "@/i18n";
import { personAccent } from "../domainAccent";
import { radius, semantic, sizes, space, strokes, typeStyles } from "../tokens";

import { formatBelcantoDate } from "../viewModels";

/**
 * Community patterns from Figma «Community Post Card» (75:1308) and
 * «Message Bubble» (74:1294), Page 28. Curated growth stories and
 * school announcements — never an endless entertainment feed, never
 * popularity or ranking signals (DEC-006). Hidden or removed content is
 * a tombstone: the mark and the moment stay, the words and the author
 * leave (COM-SAFE-05).
 */

export function communityInitials(fullName: string): string {
  const parts = fullName.trim().split(/\s+/u).filter(Boolean);
  const first = parts[0]?.[0] ?? "";
  const second = parts[1]?.[0] ?? "";
  return `${first}${second}`.toUpperCase() || "•";
}

/** Accent per post nature: announcements gold, staff room cyan. */
export function postAccent(post: Pick<CommunityPost, "kind" | "audience">): string {
  if (post.kind === "announcement") return semantic.accentGold;
  if (post.audience === "staff") return semantic.accentCyan;
  return semantic.accentViolet;
}

export function postKickerKey(
  post: Pick<CommunityPost, "kind" | "audience">,
): "com.kicker.announcement" | "com.kicker.staff" | "com.kicker.feed" {
  if (post.kind === "announcement") return "com.kicker.announcement";
  if (post.audience === "staff") return "com.kicker.staff";
  return "com.kicker.feed";
}

export function communityRoleKey(
  role: string,
):
  | "com.role.Owner"
  | "com.role.Administrator"
  | "com.role.Teacher"
  | "com.role.Student"
  | "com.role.member" {
  switch (role) {
    case "Owner":
      return "com.role.Owner";
    case "Administrator":
      return "com.role.Administrator";
    case "Teacher":
      return "com.role.Teacher";
    case "Student":
      return "com.role.Student";
    default:
      return "com.role.member";
  }
}

function AuthorAvatar({ name }: { name: string }) {
  return (
    <View style={[avatarStyles.circle, { backgroundColor: personAccent(name) }]}>
      <Text style={avatarStyles.initials}>{communityInitials(name)}</Text>
    </View>
  );
}

const avatarStyles = StyleSheet.create({
  circle: {
    alignItems: "center",
    backgroundColor: semantic.accentViolet,
    borderColor: semantic.borderAccent,
    borderRadius: radius.pill,
    borderWidth: 2,
    height: sizes.avatarSm,
    justifyContent: "center",
    width: sizes.avatarSm,
  },
  initials: {
    color: semantic.textOnAction,
    ...typeStyles.labelM,
  },
});

export type CommunityPostCardProps = {
  post: CommunityPost;
  message: MessageFormatter;
  onOpen?: (() => void) | undefined;
  testID?: string | undefined;
};

/** Feed and thread card. Tombstoned posts keep the mark and moment. */
export function CommunityPostCard({ post, message, onOpen, testID }: CommunityPostCardProps) {
  const accent = postAccent(post);
  const tombstone = post.status !== "published" && post.body === undefined;
  const body = (
    <View style={cardStyles.card} testID={testID}>
      <View style={[cardStyles.rail, { backgroundColor: accent }]} />
      {tombstone ? (
        <>
          <Text style={cardStyles.kicker}>{message("com.tombstone.kicker")}</Text>
          <Text style={cardStyles.title}>{message("com.tombstone.title")}</Text>
          <Text style={cardStyles.body}>{message("com.tombstone.note")}</Text>
          <Text style={cardStyles.meta}>{formatBelcantoDate(post.createdAt)}</Text>
        </>
      ) : (
        <>
          <View style={cardStyles.header}>
            <AuthorAvatar name={post.author.fullName || "•"} />
            <View style={cardStyles.identity}>
              <Text style={cardStyles.name}>{post.author.fullName}</Text>
              <Text style={cardStyles.meta}>
                {message(communityRoleKey(post.author.role))} ·{" "}
                {formatBelcantoDate(post.createdAt)}
              </Text>
            </View>
            {post.pinned ? (
              <Text style={[cardStyles.pinned, { color: accent }]}>
                {message("com.pinned")}
              </Text>
            ) : null}
          </View>
          <Text style={[cardStyles.kicker, { color: accent }]}>
            {message(postKickerKey(post))}
          </Text>
          {post.title !== undefined ? (
            <Text style={cardStyles.title}>{post.title}</Text>
          ) : null}
          <Text style={cardStyles.body}>{post.body}</Text>
          <Text style={cardStyles.meta}>
            {post.commentsEnabled
              ? message("com.commentCount", { count: post.commentCount })
              : message("com.commentsOff")}
            {post.status !== "published"
              ? ` · ${message(`com.status.${post.status}`)}`
              : ""}
          </Text>
        </>
      )}
    </View>
  );
  if (onOpen === undefined) return body;
  return (
    <Pressable
      accessibilityLabel={
        tombstone ? message("com.tombstone.title") : (post.title ?? post.body ?? "")
      }
      accessibilityRole="button"
      onPress={onOpen}
    >
      {body}
    </Pressable>
  );
}

const cardStyles = StyleSheet.create({
  card: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderGlass,
    borderRadius: radius.xl,
    borderWidth: strokes.hairline,
    gap: space.s3,
    padding: space.s4,
    width: "100%",
  },
  rail: {
    borderRadius: radius.pill,
    height: 4,
    width: "100%",
  },
  header: {
    alignItems: "center",
    flexDirection: "row",
    gap: space.s2,
  },
  identity: { flex: 1, gap: space.s1 },
  name: {
    color: semantic.textPrimary,
    ...typeStyles.labelM,
  },
  pinned: {
    ...typeStyles.labelM,
  },
  kicker: {
    color: semantic.accentGold,
    ...typeStyles.labelM,
  },
  title: {
    color: semantic.textPrimary,
    ...typeStyles.labelL,
  },
  body: {
    color: semantic.textSecondary,
    ...typeStyles.bodyS,
  },
  meta: {
    color: semantic.textMuted,
    ...typeStyles.labelM,
  },
});

export type CommentBubbleProps = {
  comment: CommunityComment;
  own: boolean;
  message: MessageFormatter;
  onReport?: (() => void) | undefined;
  testID?: string | undefined;
};

/**
 * Thread reply bubble. The member's own reply uses the action surface
 * («Вы»); a tombstoned reply keeps its place and moment only.
 */
export function CommentBubble({ comment, own, message, onReport, testID }: CommentBubbleProps) {
  const tombstone = comment.body === undefined;
  if (tombstone) {
    return (
      <View style={[bubbleStyles.bubble, bubbleStyles.tombstone]} testID={testID}>
        <View style={bubbleStyles.header}>
          <Text style={bubbleStyles.tombstoneText}>
            {message(
              comment.status === "removed"
                ? "com.comment.removed"
                : "com.comment.hidden",
            )}
          </Text>
          <Text style={bubbleStyles.timeMuted}>{formatBelcantoDate(comment.createdAt)}</Text>
        </View>
      </View>
    );
  }
  return (
    <View
      style={[bubbleStyles.bubble, own ? bubbleStyles.own : bubbleStyles.other]}
      testID={testID}
    >
      <View style={bubbleStyles.header}>
        {own ? null : <AuthorAvatar name={comment.author.fullName || "•"} />}
        <Text style={[bubbleStyles.name, own && bubbleStyles.onAction]}>
          {own ? message("com.you") : comment.author.fullName}
        </Text>
        <Text style={[bubbleStyles.timeMuted, own && bubbleStyles.onAction]}>
          {formatBelcantoDate(comment.createdAt)}
        </Text>
      </View>
      <Text style={[bubbleStyles.body, own && bubbleStyles.onAction]}>{comment.body}</Text>
      {comment.status !== "published" ? (
        <Text style={bubbleStyles.timeMuted}>{message(`com.status.${comment.status}`)}</Text>
      ) : null}
      {onReport === undefined ? null : (
        <Pressable
          accessibilityLabel={message("com.report.action")}
          accessibilityRole="button"
          onPress={onReport}
          style={bubbleStyles.reportAction}
        >
          <Text style={bubbleStyles.reportLabel}>{message("com.report.action")}</Text>
        </Pressable>
      )}
    </View>
  );
}

const bubbleStyles = StyleSheet.create({
  bubble: {
    borderRadius: radius.lg,
    gap: space.s2,
    padding: space.s3,
    width: "100%",
  },
  other: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderGlass,
    borderWidth: strokes.hairline,
  },
  own: {
    backgroundColor: semantic.bgAction,
    borderColor: semantic.borderAccent,
    borderWidth: strokes.hairline,
  },
  tombstone: {
    backgroundColor: semantic.bgSurface,
    borderColor: semantic.borderDefault,
    borderWidth: strokes.hairline,
  },
  header: {
    alignItems: "center",
    flexDirection: "row",
    gap: space.s2,
  },
  name: {
    color: semantic.textPrimary,
    flex: 1,
    ...typeStyles.labelM,
  },
  body: {
    color: semantic.textSecondary,
    ...typeStyles.bodyS,
  },
  onAction: { color: semantic.textOnAction },
  timeMuted: {
    color: semantic.textMuted,
    ...typeStyles.labelM,
  },
  tombstoneText: {
    color: semantic.textMuted,
    flex: 1,
    fontStyle: "italic",
    ...typeStyles.bodyS,
  },
  reportAction: {
    justifyContent: "center",
    minHeight: sizes.touchMin,
  },
  reportLabel: {
    color: semantic.textMuted,
    ...typeStyles.labelM,
  },
});
