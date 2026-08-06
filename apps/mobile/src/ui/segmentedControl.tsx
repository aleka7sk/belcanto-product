import { Pressable, StyleSheet, Text, View } from "react-native";

import { chipHitSlop } from "./chip";
import { radius, semantic, typeStyles, space } from "./tokens";

/**
 * SegmentedControl — the one segment/tab row of the application
 * (owner overview, teacher students, community home). Screen-reader
 * contract: the container is a tablist, each segment a tab with a
 * selected state — selection is never conveyed by color alone. The
 * visual height matches the design's 38-pt pills (COM-HOME-01 347:10);
 * the interactive target is extended to 48 through hitSlop.
 */
export const SEGMENT_VISUAL_MIN_HEIGHT = 38;

const HIT_SLOP = chipHitSlop(SEGMENT_VISUAL_MIN_HEIGHT);

export interface SegmentItem<Key extends string> {
  key: Key;
  label: string;
  /**
   * «link» marks an item that navigates away instead of switching the
   * segment (the События entry of COM-HOME-01) so assistive tech does
   * not announce a tab that never becomes selected.
   */
  role?: "tab" | "link" | undefined;
}

export interface SegmentedControlProps<Key extends string> {
  items: readonly SegmentItem<Key>[];
  active: Key;
  onSelect: (key: Key) => void;
  /** Active-pill fill; community surfaces use the magenta accent. */
  activeColor?: string | undefined;
  accessibilityLabel?: string | undefined;
  testIDPrefix?: string | undefined;
}

export function SegmentedControl<Key extends string>({
  items,
  active,
  onSelect,
  activeColor = semantic.bgAction,
  accessibilityLabel,
  testIDPrefix,
}: SegmentedControlProps<Key>) {
  return (
    <View
      accessibilityLabel={accessibilityLabel}
      accessibilityRole="tablist"
      style={styles.row}
    >
      {items.map((item) => {
        const selected = item.role !== "link" && item.key === active;
        return (
          <Pressable
            accessibilityLabel={item.label}
            accessibilityRole={item.role === "link" ? "link" : "tab"}
            accessibilityState={item.role === "link" ? undefined : { selected }}
            hitSlop={HIT_SLOP}
            key={item.key}
            onPress={() => onSelect(item.key)}
            style={({ pressed }) => [
              styles.segment,
              selected && { backgroundColor: activeColor },
              pressed && !selected && styles.pressed,
            ]}
            testID={testIDPrefix !== undefined ? `${testIDPrefix}-${item.key}` : undefined}
          >
            <Text style={[styles.label, selected && styles.labelActive]}>
              {item.label}
            </Text>
          </Pressable>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  label: { color: semantic.textSecondary, textAlign: "center", ...typeStyles.labelM },
  labelActive: { color: semantic.textOnAction },
  pressed: { opacity: 0.85 },
  row: { flexDirection: "row", gap: space.s2 },
  segment: {
    alignItems: "center",
    backgroundColor: semantic.bgRaised,
    borderRadius: radius.md,
    flex: 1,
    justifyContent: "center",
    minHeight: SEGMENT_VISUAL_MIN_HEIGHT,
    paddingHorizontal: space.s2,
  },
});
