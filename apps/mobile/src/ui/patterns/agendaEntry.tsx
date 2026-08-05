import { Pressable, StyleSheet, Text, View } from "react-native";

import { radius, semantic, space, strokes, typeStyles } from "../tokens";

/**
 * AgendaEntry — implementation of «Agenda Entry · Production»
 * (Figma 306:2, Page 35): the canonical personal agenda entry. Core
 * individual, core group and registered event share one premium
 * composition but remain distinct domain types; Student group cards never
 * expose groupmates or capacity (DEC-010).
 */
export type AgendaEntryType = "individual" | "group" | "event";

export type AgendaEntryState =
  | "upcoming"
  | "now"
  | "changed"
  | "cancelled"
  | "completed";

/** Rail and eyebrow accent per Figma variant matrix (state wins over type). */
export function agendaAccentColor(
  type: AgendaEntryType,
  state: AgendaEntryState,
): string {
  if (state === "completed") {
    return semantic.textMuted;
  }
  if (state === "cancelled") {
    return semantic.feedbackDanger;
  }
  if (state === "changed") {
    return semantic.feedbackWarning;
  }
  if (state === "upcoming" && type === "event") {
    return semantic.accentMagenta;
  }
  if (state === "upcoming" && type === "group") {
    return semantic.accentCyan;
  }
  return semantic.accentViolet;
}

export function agendaCardColors(state: AgendaEntryState): {
  backgroundColor: string;
  borderColor: string;
} {
  switch (state) {
    case "cancelled":
      return {
        backgroundColor: semantic.bgSurface,
        borderColor: semantic.feedbackDanger,
      };
    case "changed":
      return {
        backgroundColor: semantic.bgSurface,
        borderColor: semantic.feedbackWarning,
      };
    case "now":
      return {
        backgroundColor: semantic.bgRaised,
        borderColor: semantic.borderAccent,
      };
    case "upcoming":
    case "completed":
      return {
        backgroundColor: semantic.bgSurface,
        borderColor: semantic.borderDefault,
      };
  }
}

export type AgendaEntryProps = {
  type: AgendaEntryType;
  state: AgendaEntryState;
  eyebrow: string;
  title: string;
  timePlace: string;
  action: string;
  supporting: string;
  onPress?: () => void;
  accessibilityHint?: string;
  testID?: string;
};

export function AgendaEntry({
  type,
  state,
  eyebrow,
  title,
  timePlace,
  action,
  supporting,
  onPress,
  accessibilityHint,
  testID,
}: AgendaEntryProps) {
  const accent = agendaAccentColor(type, state);
  const card = agendaCardColors(state);
  const cancelled = state === "cancelled";
  const body = (
    <>
      <View style={[styles.rail, { backgroundColor: accent }]} />
      <View style={styles.content}>
        <Text style={[styles.eyebrow, { color: accent }]}>{eyebrow}</Text>
        <Text
          style={[styles.title, cancelled ? styles.titleCancelled : null]}
        >
          {title}
        </Text>
        <Text style={styles.timePlace}>{timePlace}</Text>
        <Text style={[styles.action, { color: accent }]}>{action}</Text>
        <Text style={styles.supporting}>{supporting}</Text>
      </View>
    </>
  );
  if (!onPress) {
    return (
      <View
        style={[styles.card, card]}
        testID={testID}
      >
        {body}
      </View>
    );
  }
  return (
    <Pressable
      accessibilityHint={accessibilityHint}
      accessibilityLabel={`${eyebrow}. ${title}. ${timePlace}. ${action}`}
      accessibilityRole="button"
      onPress={onPress}
      style={[styles.card, card]}
      testID={testID}
    >
      {body}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: {
    borderRadius: radius.xl,
    borderWidth: strokes.default,
    flexDirection: "row",
    overflow: "hidden",
    width: "100%",
  },
  rail: {
    alignSelf: "stretch",
    width: 5,
  },
  content: {
    flex: 1,
    gap: space.s2,
    padding: space.s4,
  },
  eyebrow: {
    ...typeStyles.labelM,
  },
  title: {
    ...typeStyles.headingM,
    color: semantic.textPrimary,
  },
  titleCancelled: {
    color: semantic.textMuted,
  },
  timePlace: {
    ...typeStyles.bodyS,
    color: semantic.textSecondary,
  },
  action: {
    ...typeStyles.labelL,
  },
  supporting: {
    ...typeStyles.caption,
    color: semantic.textMuted,
  },
});
