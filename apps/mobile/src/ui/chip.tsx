import { Pressable, StyleSheet, Text } from "react-native";

import { radius, semantic, sizes, strokes, space, typeStyles } from "./tokens";

/**
 * Chip — the one filter/selection pill (area filters of STU-GROWTH-01,
 * category filters of Page 24, composer choices). The visual height is
 * the design's 34pt; the interactive target is extended to
 * sizes.touchMin through hitSlop so the 48-pt minimum holds without
 * inflating the visual row (§3 accessibility contract).
 */
export const CHIP_VISUAL_MIN_HEIGHT = 34;

/** Symmetric hitSlop growing the visual height to the minimum target. */
export function chipHitSlop(visualMinHeight: number): { top: number; bottom: number } {
  const missing = Math.max(0, sizes.touchMin - visualMinHeight);
  const half = Math.ceil(missing / 2);
  return { top: half, bottom: half };
}

const HIT_SLOP = chipHitSlop(CHIP_VISUAL_MIN_HEIGHT);

export interface ChipProps {
  label: string;
  active: boolean;
  onPress: () => void;
  /** Accent color for the active border and label. */
  accent?: string | undefined;
  testID?: string | undefined;
}

export function Chip({
  label,
  active,
  onPress,
  accent = semantic.accentViolet,
  testID,
}: ChipProps) {
  return (
    <Pressable
      accessibilityLabel={label}
      accessibilityRole="button"
      accessibilityState={{ selected: active }}
      hitSlop={HIT_SLOP}
      onPress={onPress}
      style={({ pressed }) => [
        styles.chip,
        { borderColor: active ? accent : semantic.borderDefault },
        pressed && styles.pressed,
      ]}
      testID={testID}
    >
      <Text style={[styles.label, { color: active ? accent : semantic.textMuted }]}>
        {label}
      </Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  chip: {
    alignItems: "center",
    backgroundColor: semantic.bgSurface,
    borderRadius: radius.pill,
    borderWidth: strokes.default,
    justifyContent: "center",
    minHeight: CHIP_VISUAL_MIN_HEIGHT,
    paddingHorizontal: space.s3,
  },
  label: { ...typeStyles.caption },
  pressed: { opacity: 0.85 },
});
