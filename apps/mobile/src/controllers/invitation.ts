import {
  invalid,
  normalizedRequired,
  valid,
  type FormIssue,
  type FormResult,
} from "@/forms/result";
import {
  backendIdentifierIssue,
  requiredIdempotencyKey,
} from "@/validation/backend";

export type InvitationMode = "issue" | "reissue";

export interface InvitationDraft {
  studentId: string;
}

export interface InvitationCommand {
  studentId: string;
  idempotencyKey: string;
  mode: InvitationMode;
}

export function prepareInvitation(
  draft: InvitationDraft,
  suppliedIdempotencyKey: string,
  mode: InvitationMode,
): FormResult<InvitationCommand, keyof InvitationDraft> {
  const issues: FormIssue<keyof InvitationDraft>[] = [];
  const studentId = normalizedRequired(draft.studentId);
  const idempotencyKey = requiredIdempotencyKey(suppliedIdempotencyKey);
  if (studentId === null) issues.push({ field: "studentId", code: "required" });
  else {
    const code = backendIdentifierIssue(studentId);
    if (code !== null) issues.push({ field: "studentId", code });
  }
  if (issues.length > 0 || studentId === null) {
    return invalid(issues);
  }
  return valid({ studentId, idempotencyKey, mode });
}

export interface RevokeInvitationDraft {
  invitationId: string;
}

export interface RevokeInvitationCommand {
  invitationId: string;
  idempotencyKey: string;
}

export function prepareRevokeInvitation(
  draft: RevokeInvitationDraft,
  suppliedIdempotencyKey: string,
): FormResult<RevokeInvitationCommand, keyof RevokeInvitationDraft> {
  const issues: FormIssue<keyof RevokeInvitationDraft>[] = [];
  const invitationId = normalizedRequired(draft.invitationId);
  const idempotencyKey = requiredIdempotencyKey(suppliedIdempotencyKey);
  if (invitationId === null) issues.push({ field: "invitationId", code: "required" });
  else {
    const code = backendIdentifierIssue(invitationId);
    if (code !== null) issues.push({ field: "invitationId", code });
  }
  if (issues.length > 0 || invitationId === null) {
    return invalid(issues);
  }
  return valid({ invitationId, idempotencyKey });
}
