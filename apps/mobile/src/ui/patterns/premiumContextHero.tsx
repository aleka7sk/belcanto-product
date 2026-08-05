import { Pressable, StyleSheet, Text, View } from "react-native";

import { radius, semantic, sizes, space, strokes, typeStyles } from "../tokens";

/**
 * PremiumContextHero — implementation of «Premium Context Hero»
 * (Figma 300:2, Page 35): role-aware hero for the primary context of a
 * screen. One hero per screen; heroes are never stacked.
 */
export type HeroRole = "Student" | "Teacher" | "Administrator" | "Owner" | "Profile";

export function heroAccentColor(role: HeroRole): string {
  switch (role) {
    case "Teacher":
      return semantic.accentCyan;
    case "Administrator":
      return semantic.accentGold;
    case "Owner":
      return semantic.accentMagenta;
    case "Student":
    case "Profile":
      return semantic.accentViolet;
  }
}

export type PremiumContextHeroProps = {
  role: HeroRole;
  eyebrow: string;
  title: string;
  body: string;
  metric?: string;
  onPressMetric?: () => void;
  metricAccessibilityHint?: string;
  testID?: string;
};

export function PremiumContextHero({
  role,
  eyebrow,
  title,
  body,
  metric,
  onPressMetric,
  metricAccessibilityHint,
  testID,
}: PremiumContextHeroProps) {
  const accent = heroAccentColor(role);
  return (
    <View style={styles.card} testID={testID}>
      <View style={[styles.accentBar, { backgroundColor: accent }]} />
      <Text style={[styles.eyebrow, { color: accent }]}>{eyebrow}</Text>
      <Text accessibilityRole="header" style={styles.title}>
        {title}
      </Text>
      <Text style={styles.body}>{body}</Text>
      {metric === undefined ? null : onPressMetric ? (
        <Pressable
          accessibilityHint={metricAccessibilityHint}
          accessibilityLabel={metric}
          accessibilityRole="button"
          onPress={onPressMetric}
          style={styles.metricAction}
        >
          <Text style={styles.metric}>{metric}</Text>
        </Pressable>
      ) : (
        <Text style={styles.metric}>{metric}</Text>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderDefault,
    borderRadius: radius.xxl,
    borderWidth: strokes.default,
    gap: space.s3,
    padding: space.s6,
    width: "100%",
  },
  accentBar: {
    borderRadius: radius.pill,
    height: 4,
    width: 48,
  },
  eyebrow: {
    ...typeStyles.labelM,
  },
  title: {
    ...typeStyles.headingL,
    color: semantic.textPrimary,
  },
  body: {
    ...typeStyles.bodyS,
    color: semantic.textSecondary,
  },
  metricAction: {
    justifyContent: "center",
    minHeight: sizes.touchMin,
  },
  metric: {
    ...typeStyles.labelL,
    color: semantic.textAccent,
  },
});
