export const MINIMUM_INTERACTIVE_TARGET = 48;

export function meetsMinimumInteractiveTarget(width: number, height: number): boolean {
  return width >= MINIMUM_INTERACTIVE_TARGET && height >= MINIMUM_INTERACTIVE_TARGET;
}

export interface MotionPolicy {
  allowSpatialMotion: boolean;
  allowDecorativeMotion: boolean;
  preserveImmediateStateOutcome: true;
}

export function motionPolicy(reducedMotionEnabled: boolean): MotionPolicy {
  return {
    allowSpatialMotion: !reducedMotionEnabled,
    allowDecorativeMotion: !reducedMotionEnabled,
    preserveImmediateStateOutcome: true,
  };
}

export type FeedbackIntent =
  | "generic_press"
  | "selection"
  | "confirmation"
  | "warning"
  | "rejection";

export interface FeedbackPolicy {
  visibleOutcomeRequired: true;
  accessibleAnnouncementRequired: true;
  hapticAllowed: boolean;
}

export function feedbackPolicy(
  intent: FeedbackIntent,
  options: { hapticsAvailable: boolean; hapticsEnabled: boolean },
): FeedbackPolicy {
  return {
    visibleOutcomeRequired: true,
    accessibleAnnouncementRequired: true,
    hapticAllowed:
      intent !== "generic_press" &&
      options.hapticsAvailable &&
      options.hapticsEnabled,
  };
}
