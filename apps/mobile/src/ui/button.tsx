import { ActivityIndicator, Pressable, StyleSheet, Text } from "react-native";

import { opacities, radius, semantic, sizes, space, strokes, typeStyles } from "./tokens";

/**
 * Button — the one interactive-action implementation of the application.
 * Two shapes cover the design language: «block» is the 52-pt full-width
 * «Action · …» rectangle of the production frames, «pill» is the rounded
 * capsule of the auth flows. Every variant grows from a minimum height —
 * never a fixed one — so large font scales expand the target instead of
 * clipping the label (§3 accessibility contract, sizes.touchMin).
 */
export type ButtonKind = "primary" | "secondary" | "text";
export type ButtonShape = "block" | "pill";
export type ButtonTone = "default" | "danger";

/** Block actions are 52pt per the Figma «Action · …» frames; pills 48pt. */
export const BUTTON_MIN_HEIGHT: Record<ButtonShape, number> = {
  block: 52,
  pill: sizes.touchMin,
};

export interface ButtonProps {
  label: string;
  onPress: () => void;
  kind?: ButtonKind | undefined;
  shape?: ButtonShape | undefined;
  tone?: ButtonTone | undefined;
  busy?: boolean | undefined;
  disabled?: boolean | undefined;
  align?: "stretch" | "center" | "right" | undefined;
  accessibilityHint?: string | undefined;
  testID?: string | undefined;
}

export function Button({
  label,
  onPress,
  kind = "primary",
  shape = "block",
  tone = "default",
  busy = false,
  disabled = false,
  align = "stretch",
  accessibilityHint,
  testID,
}: ButtonProps) {
  const unavailable = busy || disabled;
  const labelColor =
    kind === "primary"
      ? semantic.textOnAction
      : tone === "danger"
        ? semantic.feedbackDanger
        : kind === "text"
          ? semantic.textAccent
          : semantic.textPrimary;
  return (
    <Pressable
      accessibilityHint={accessibilityHint}
      accessibilityLabel={label}
      accessibilityRole="button"
      accessibilityState={{ busy, disabled: unavailable }}
      disabled={unavailable}
      onPress={onPress}
      style={({ pressed }) => [
        styles.base,
        { minHeight: BUTTON_MIN_HEIGHT[shape] },
        shape === "pill" ? styles.pill : styles.block,
        kind === "primary" && styles.primary,
        kind === "secondary" && styles.secondary,
        kind === "primary" && pressed && !unavailable && styles.primaryPressed,
        kind !== "primary" && pressed && !unavailable && styles.pressed,
        align === "center" && styles.alignCenter,
        align === "right" && styles.alignRight,
        unavailable && styles.disabled,
      ]}
      testID={testID}
    >
      {busy ? (
        <ActivityIndicator
          color={kind === "primary" ? semantic.textOnAction : semantic.textPrimary}
        />
      ) : (
        <Text style={[kind === "text" ? styles.textLabel : styles.label, { color: labelColor }]}>
          {label}
        </Text>
      )}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  alignCenter: { alignSelf: "center", minWidth: sizes.touchMin },
  alignRight: { alignSelf: "flex-end", minWidth: sizes.touchMin },
  base: {
    alignItems: "center",
    justifyContent: "center",
    paddingHorizontal: space.s5,
  },
  block: { borderRadius: radius.lg },
  disabled: { opacity: opacities.disabled },
  label: { ...typeStyles.labelL, textAlign: "center" },
  pill: { borderRadius: radius.pill },
  pressed: { opacity: 0.85 },
  primary: { backgroundColor: semantic.bgAction },
  primaryPressed: { backgroundColor: semantic.bgActionPressed },
  secondary: {
    backgroundColor: semantic.bgRaised,
    borderColor: semantic.borderDefault,
    borderWidth: strokes.hairline,
  },
  textLabel: { ...typeStyles.labelM, textAlign: "center" },
});
