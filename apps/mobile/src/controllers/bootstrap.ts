import type { BootstrapView } from "@/api/contracts";

export type BootstrapInterpretation =
  | { ready: true; view: BootstrapView }
  | {
      ready: false;
      reason:
        | "student_identity_missing"
        | "student_full_name_missing"
        | "first_minute_missing"
        | "first_minute_student_mismatch";
    };

export function interpretBootstrap(view: BootstrapView): BootstrapInterpretation {
  if (view.roles.includes("Student") && view.studentId === undefined) {
    return { ready: false, reason: "student_identity_missing" };
  }
  if (view.roles.includes("Student") && view.fullName === undefined) {
    return { ready: false, reason: "student_full_name_missing" };
  }
  if (view.roles.includes("Student") && view.firstMinute === undefined) {
    return { ready: false, reason: "first_minute_missing" };
  }
  if (
    view.firstMinute !== undefined &&
    view.studentId !== view.firstMinute.studentId
  ) {
    return { ready: false, reason: "first_minute_student_mismatch" };
  }
  return { ready: true, view };
}
