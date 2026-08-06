import { MINIMUM_INTERACTIVE_TARGET } from "../accessibility/policy";
import { BUTTON_MIN_HEIGHT } from "./button";
import { CHIP_VISUAL_MIN_HEIGHT, chipHitSlop } from "./chip";
import { SEGMENT_VISUAL_MIN_HEIGHT } from "./segmentedControl";
import { navigation, sizes } from "./tokens";

describe("interactive target contract", () => {
  it("declares the minimum target once, from the token layer", () => {
    expect(MINIMUM_INTERACTIVE_TARGET).toBe(sizes.touchMin);
    expect(sizes.touchMin).toBe(48);
  });

  it("keeps every button shape at or above the minimum", () => {
    for (const minHeight of Object.values(BUTTON_MIN_HEIGHT)) {
      expect(minHeight).toBeGreaterThanOrEqual(sizes.touchMin);
    }
  });

  it("extends compact chips and segments to the minimum through hitSlop", () => {
    for (const visual of [CHIP_VISUAL_MIN_HEIGHT, SEGMENT_VISUAL_MIN_HEIGHT]) {
      const slop = chipHitSlop(visual);
      expect(visual + slop.top + slop.bottom).toBeGreaterThanOrEqual(sizes.touchMin);
    }
  });

  it("does not add hitSlop to targets already at the minimum", () => {
    expect(chipHitSlop(sizes.touchMin)).toEqual({ top: 0, bottom: 0 });
    expect(chipHitSlop(64)).toEqual({ top: 0, bottom: 0 });
  });
});

describe("bottom navigation host geometry (Figma 310:20542)", () => {
  it("mirrors the fixed host frame of the production screens", () => {
    expect(navigation.height).toBe(68);
    expect(navigation.maxWidth).toBe(366);
    expect(navigation.sideInset).toBe(12);
    expect(navigation.bottomGap).toBe(8);
    expect(navigation.itemMinHeight).toBeGreaterThanOrEqual(sizes.touchMin);
  });
});
