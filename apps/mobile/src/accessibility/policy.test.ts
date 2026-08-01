import {
  feedbackPolicy,
  meetsMinimumInteractiveTarget,
  motionPolicy,
} from "./policy";

describe("accessibility policy", () => {
  it("enforces the 48 by 48 interaction target", () => {
    expect(meetsMinimumInteractiveTarget(48, 48)).toBe(true);
    expect(meetsMinimumInteractiveTarget(47, 48)).toBe(false);
  });

  it("removes spatial and decorative motion without hiding state", () => {
    expect(motionPolicy(true)).toEqual({
      allowSpatialMotion: false,
      allowDecorativeMotion: false,
      preserveImmediateStateOutcome: true,
    });
  });

  it("never uses haptics for generic presses or as sole feedback", () => {
    expect(
      feedbackPolicy("generic_press", {
        hapticsAvailable: true,
        hapticsEnabled: true,
      }),
    ).toEqual({
      visibleOutcomeRequired: true,
      accessibleAnnouncementRequired: true,
      hapticAllowed: false,
    });
  });
});
