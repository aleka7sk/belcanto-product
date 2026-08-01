import type {
  GrantDelegationRequest,
  IsoDateTime,
  RevokeDelegationRequest,
} from "@/api/contracts";
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
import { parseStrictRfc3339 } from "@/validation/datetime";

export interface GrantDelegationDraft {
  administratorAccountId: string;
  reason: string;
  expiresAt: string;
  currentPassword: string;
}

type GrantField = keyof GrantDelegationDraft;

export interface GrantDelegationCommand {
  body: GrantDelegationRequest;
  idempotencyKey: string;
}

export function prepareGrantDelegation(
  draft: GrantDelegationDraft,
  suppliedIdempotencyKey: string,
  now = new Date(),
): FormResult<GrantDelegationCommand, GrantField> {
  const issues: FormIssue<GrantField>[] = [];
  const administratorAccountId = normalizedRequired(draft.administratorAccountId);
  const reason = normalizedRequired(draft.reason);
  const currentPassword = normalizedRequired(draft.currentPassword, (value) => value);
  const idempotencyKey = requiredIdempotencyKey(suppliedIdempotencyKey);
  if (administratorAccountId === null)
    issues.push({ field: "administratorAccountId", code: "required" });
  else {
    const code = backendIdentifierIssue(administratorAccountId);
    if (code !== null)
      issues.push({ field: "administratorAccountId", code });
  }
  if (reason === null) issues.push({ field: "reason", code: "required" });
  else {
    const code = backendTextIssue(reason, BACKEND_INPUT_LIMITS.reasonRunes);
    if (code !== null) issues.push({ field: "reason", code });
  }
  if (currentPassword === null)
    issues.push({ field: "currentPassword", code: "required" });

  let expiresAt: IsoDateTime | undefined;
  if (draft.expiresAt.trim().length > 0) {
    const timestamp = parseStrictRfc3339(draft.expiresAt);
    if (timestamp === null) {
      issues.push({ field: "expiresAt", code: "invalid_format" });
    } else if (timestamp <= now.getTime()) {
      issues.push({ field: "expiresAt", code: "must_be_future" });
    } else {
      expiresAt = new Date(timestamp).toISOString() as IsoDateTime;
    }
  }
  if (
    issues.length > 0 ||
    administratorAccountId === null ||
    reason === null ||
    currentPassword === null
  ) {
    return invalid(issues);
  }

  const body: GrantDelegationRequest = {
    administratorAccountId,
    reason,
    currentPassword,
  };
  if (expiresAt !== undefined) body.expiresAt = expiresAt;
  return valid({ body, idempotencyKey });
}

export interface RevokeDelegationDraft {
  delegationId: string;
  reason: string;
  currentPassword: string;
}

type RevokeField = keyof RevokeDelegationDraft;

export interface RevokeDelegationCommand {
  delegationId: string;
  body: RevokeDelegationRequest;
  idempotencyKey: string;
}

export function prepareRevokeDelegation(
  draft: RevokeDelegationDraft,
  suppliedIdempotencyKey: string,
): FormResult<RevokeDelegationCommand, RevokeField> {
  const issues: FormIssue<RevokeField>[] = [];
  const delegationId = normalizedRequired(draft.delegationId);
  const reason = normalizedRequired(draft.reason);
  const currentPassword = normalizedRequired(draft.currentPassword, (value) => value);
  const idempotencyKey = requiredIdempotencyKey(suppliedIdempotencyKey);
  if (delegationId === null) issues.push({ field: "delegationId", code: "required" });
  else {
    const code = backendIdentifierIssue(delegationId);
    if (code !== null) issues.push({ field: "delegationId", code });
  }
  if (reason === null) issues.push({ field: "reason", code: "required" });
  else {
    const code = backendTextIssue(reason, BACKEND_INPUT_LIMITS.reasonRunes);
    if (code !== null) issues.push({ field: "reason", code });
  }
  if (currentPassword === null)
    issues.push({ field: "currentPassword", code: "required" });
  if (
    issues.length > 0 ||
    delegationId === null ||
    reason === null ||
    currentPassword === null
  ) {
    return invalid(issues);
  }
  return valid({
    delegationId,
    body: { reason, currentPassword },
    idempotencyKey,
  });
}
