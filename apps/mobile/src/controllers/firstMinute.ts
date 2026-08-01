import type { PublishFirstMinuteRequest } from "@/api/contracts";
import {
  invalid,
  normalizedRequired,
  valid,
  type FormIssue,
  type FormResult,
} from "@/forms/result";
import {
  BACKEND_INPUT_LIMITS,
  backendIdentifierIssue,
  backendTextIssue,
  requiredIdempotencyKey,
} from "@/validation/backend";


export interface FirstMinuteDraft {
  studentId: string;
  whatWorked: string;
  currentFocus: string;
  nextStep: string;
  expectedVersion: number;
}

type FirstMinuteField = keyof FirstMinuteDraft;

export interface FirstMinuteCommand {
  studentId: string;
  body: PublishFirstMinuteRequest;
  idempotencyKey: string;
}

export function prepareFirstMinute(
  draft: FirstMinuteDraft,
  suppliedIdempotencyKey: string,
): FormResult<FirstMinuteCommand, FirstMinuteField> {
  const issues: FormIssue<FirstMinuteField>[] = [];
  const studentId = normalizedRequired(draft.studentId);
  const whatWorked = normalizedRequired(draft.whatWorked);
  const currentFocus = normalizedRequired(draft.currentFocus);
  const nextStep = normalizedRequired(draft.nextStep);
  const idempotencyKey = requiredIdempotencyKey(suppliedIdempotencyKey);
  if (studentId === null) issues.push({ field: "studentId", code: "required" });
  else {
    const code = backendIdentifierIssue(studentId);
    if (code !== null) issues.push({ field: "studentId", code });
  }
  for (const [field, value] of [
    ["whatWorked", whatWorked],
    ["currentFocus", currentFocus],
    ["nextStep", nextStep],
  ] as const) {
    if (value === null) {
      issues.push({ field, code: "required" });
    } else {
      const code = backendTextIssue(
        value,
        BACKEND_INPUT_LIMITS.firstMinuteRunes,
      );
      if (code !== null) issues.push({ field, code });
    }
  }
  if (!Number.isSafeInteger(draft.expectedVersion) || draft.expectedVersion < 0) {
    issues.push({ field: "expectedVersion", code: "invalid_value" });
  }
  if (
    issues.length > 0 ||
    studentId === null ||
    whatWorked === null ||
    currentFocus === null ||
    nextStep === null
  ) {
    return invalid(issues);
  }
  return valid({
    studentId,
    body: {
      whatWorked,
      currentFocus,
      nextStep,
      expectedVersion: draft.expectedVersion,
    },
    idempotencyKey,
  });
}
