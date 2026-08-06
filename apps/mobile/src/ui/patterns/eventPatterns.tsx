import type { ReactNode } from "react";
import {
  ActivityIndicator,
  Pressable,
  StyleSheet,
  Text,
  View,
} from "react-native";

import { Chip } from "../chip";
import { semantic, space, strokes, typeStyles } from "../tokens";

/**
 * Event patterns (Figma Page 24 + component «RSVP Control · Production»
 * 333:268). The control renders the atomic participation lifecycle:
 * confirmed appears in the personal schedule, cancelled disappears from
 * it but the event stays discoverable; capacity mutations are
 * online-only, so pending states block repeated taps.
 */

export type RsvpControlState =
  | "available"
  | "reserving"
  | "confirmed"
  | "cancelling"
  | "full"
  | "waitlistAvailable"
  | "waitlisted"
  | "spotOffered"
  | "offerExpired"
  | "conflict"
  | "closed"
  | "error";

const CONTROL_APPEARANCE: Record<
  RsvpControlState,
  { background: string; border: string; onAction: boolean; busy?: boolean }
> = {
  available: { background: semantic.bgAction, border: semantic.borderAccent, onAction: true },
  reserving: { background: semantic.bgRaised, border: semantic.borderAccent, onAction: false, busy: true },
  confirmed: { background: semantic.bgRaised, border: semantic.feedbackSuccess, onAction: false },
  cancelling: { background: semantic.bgRaised, border: semantic.feedbackWarning, onAction: false, busy: true },
  full: { background: semantic.bgSunken, border: semantic.borderStrong, onAction: false },
  waitlistAvailable: { background: semantic.bgRaised, border: semantic.accentCyan, onAction: false },
  waitlisted: { background: semantic.bgRaised, border: semantic.accentCyan, onAction: false },
  spotOffered: { background: semantic.bgRaised, border: semantic.accentGold, onAction: false },
  offerExpired: { background: semantic.bgSunken, border: semantic.feedbackWarning, onAction: false },
  conflict: { background: semantic.bgRaised, border: semantic.feedbackWarning, onAction: false },
  closed: { background: semantic.bgSunken, border: semantic.borderDefault, onAction: false },
  error: { background: semantic.bgRaised, border: semantic.feedbackDanger, onAction: false },
};

export function RsvpControl({
  state,
  title,
  subtitle,
  onPress,
  testID,
}: {
  state: RsvpControlState;
  title: string;
  subtitle: string;
  onPress?: (() => void) | undefined;
  testID?: string | undefined;
}) {
  const appearance = CONTROL_APPEARANCE[state];
  const titleColor = appearance.onAction ? semantic.textOnAction : semantic.textPrimary;
  const subtitleColor = appearance.onAction ? semantic.textOnAction : semantic.textSecondary;
  const interactive = onPress !== undefined && appearance.busy !== true;
  const body = (
    <>
      <View style={styles.controlTitleRow}>
        {appearance.busy ? (
          <ActivityIndicator color={titleColor} size="small" />
        ) : null}
        <Text style={[styles.controlTitle, { color: titleColor }]}>{title}</Text>
      </View>
      <Text style={[styles.controlSubtitle, { color: subtitleColor }]}>
        {subtitle}
      </Text>
    </>
  );
  const shellStyle = [
    styles.controlShell,
    { backgroundColor: appearance.background, borderColor: appearance.border },
  ];
  if (!interactive) {
    return (
      <View
        accessibilityLabel={title}
        accessibilityState={{ busy: appearance.busy === true }}
        style={shellStyle}
        testID={testID}
      >
        {body}
      </View>
    );
  }
  return (
    <Pressable
      accessibilityLabel={title}
      accessibilityRole="button"
      onPress={onPress}
      style={({ pressed }) => [...shellStyle, pressed && styles.pressed]}
      testID={testID}
    >
      {body}
    </Pressable>
  );
}

/** Category accent cycle for catalog cards and detail cards. */
const CATEGORY_ACCENTS = [
  semantic.accentViolet,
  semantic.accentGold,
  semantic.accentCyan,
  semantic.accentMagenta,
] as const;

export function categoryAccent(categoryId: string): string {
  let hash = 0;
  for (const char of categoryId) {
    hash = (hash * 31 + char.charCodeAt(0)) >>> 0;
  }
  return CATEGORY_ACCENTS[hash % CATEGORY_ACCENTS.length]!;
}

export function EventCard({
  category,
  title,
  meta,
  seats,
  accent,
  onPress,
  testID,
}: {
  category: string;
  title: string;
  meta: string;
  seats: string;
  accent: string;
  onPress: () => void;
  testID?: string | undefined;
}) {
  return (
    <Pressable
      accessibilityLabel={title}
      accessibilityRole="button"
      onPress={onPress}
      style={({ pressed }) => [
        styles.eventCard,
        { borderColor: accent },
        pressed && styles.pressed,
      ]}
      testID={testID}
    >
      <Text style={[styles.eventCategory, { color: accent }]}>
        {category.toUpperCase()}
      </Text>
      <Text style={styles.eventTitle}>{title}</Text>
      <Text style={styles.eventMeta}>{meta}</Text>
      <Text style={styles.eventSeats}>{seats}</Text>
    </Pressable>
  );
}

/** @deprecated Thin delegate over the unified Chip. */
export function FilterChip({
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

/** Accent-bordered detail card («Card · …» on Page 24 detail screens). */
export function EventDetailCard({
  title,
  body,
  status,
  accent = semantic.borderAccent,
  statusColor,
  children,
}: {
  title: string;
  body: string;
  status?: string | undefined;
  accent?: string | undefined;
  statusColor?: string | undefined;
  children?: ReactNode;
}) {
  return (
    <View style={[styles.detailCard, { borderColor: accent }]}>
      <Text style={styles.detailTitle}>{title}</Text>
      <Text style={styles.detailBody}>{body}</Text>
      {status ? (
        <Text
          style={[
            styles.detailStatus,
            { color: statusColor ?? semantic.textAccent },
          ]}
        >
          {status}
        </Text>
      ) : null}
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  controlShell: {
    borderRadius: 18,
    borderWidth: strokes.hairline,
    gap: space.s1,
    minHeight: 84,
    paddingHorizontal: space.s4,
    paddingVertical: 14,
  },
  controlTitleRow: { alignItems: "center", flexDirection: "row", gap: space.s2 },
  controlTitle: { ...typeStyles.labelL },
  controlSubtitle: { ...typeStyles.caption },
  eventCard: {
    backgroundColor: semantic.bgSurface,
    borderRadius: 20,
    borderWidth: strokes.hairline,
    gap: 10,
    padding: space.s4,
  },
  eventCategory: { ...typeStyles.labelM },
  eventTitle: {
    color: semantic.textPrimary,
    fontFamily: "Onest_600SemiBold",
    fontSize: 20,
    letterSpacing: -0.1,
    lineHeight: 28,
  },
  eventMeta: { color: semantic.textSecondary, ...typeStyles.bodyS },
  eventSeats: { color: semantic.textSecondary, ...typeStyles.labelM },
  detailCard: {
    backgroundColor: semantic.bgSurface,
    borderRadius: 20,
    borderWidth: strokes.hairline,
    gap: 10,
    padding: space.s4,
  },
  detailTitle: {
    color: semantic.textPrimary,
    fontFamily: "Onest_600SemiBold",
    fontSize: 20,
    letterSpacing: -0.1,
    lineHeight: 28,
  },
  detailBody: { color: semantic.textSecondary, ...typeStyles.bodyS },
  detailStatus: { ...typeStyles.labelM },
  pressed: { opacity: 0.85 },
});
