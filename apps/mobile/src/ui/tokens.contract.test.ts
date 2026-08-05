import {
  elevation,
  motion,
  palette,
  semanticByMode,
  sizes,
  space,
  typeStyles,
} from "./tokens";

describe("design token projection (Figma variables, 2026-08-05)", () => {
  it("defines every semantic role in both light and dark modes", () => {
    const darkKeys = Object.keys(semanticByMode.dark).sort();
    const lightKeys = Object.keys(semanticByMode.light).sort();
    expect(lightKeys).toEqual(darkKeys);
    expect(darkKeys).toHaveLength(30);
  });

  it("keeps the dark canvas and action colors on the canonical palette", () => {
    expect(semanticByMode.dark.bgCanvas).toBe("#070611");
    expect(semanticByMode.dark.bgAction).toBe(palette.violet500);
    expect(semanticByMode.dark.textOnAction).toBe(palette.ink950);
    expect(semanticByMode.light.bgAction).toBe(palette.violet600);
    expect(semanticByMode.dark.feedbackSuccess).toBe("#42C297");
  });

  it("keeps standard motion inside the 100-320ms band with reduced motion at zero", () => {
    for (const duration of [
      motion.durationQuick,
      motion.durationStandard,
      motion.durationExpressive,
    ]) {
      expect(duration).toBeGreaterThanOrEqual(100);
      expect(duration).toBeLessThanOrEqual(320);
    }
    expect(motion.durationCelebration).toBe(480);
    expect(motion.durationReduced).toBe(0);
  });

  it("keeps the minimum touch target at 48pt and spacing on the 4pt scale", () => {
    expect(sizes.touchMin).toBe(48);
    for (const value of Object.values(space)) {
      expect(value % 4).toBe(0);
    }
  });

  it("uses only loaded Onest families in the eleven text styles", () => {
    const loaded = new Set([
      "Onest_400Regular",
      "Onest_500Medium",
      "Onest_600SemiBold",
      "Onest_700Bold",
      "Onest_800ExtraBold",
    ]);
    expect(Object.keys(typeStyles)).toHaveLength(11);
    for (const style of Object.values(typeStyles)) {
      expect(loaded.has(style.fontFamily)).toBe(true);
      expect(style.lineHeight).toBeGreaterThanOrEqual(style.fontSize);
    }
  });

  it("keeps elevation projections aligned with the Figma effect definitions", () => {
    expect(elevation.subtle.figma.blur).toBe(12);
    expect(elevation.raised.figma.blur).toBe(28);
    expect(elevation.overlay.figma.blur).toBe(44);
    expect(elevation.glowViolet.shadowColor).toBe(palette.violet500);
  });
});
