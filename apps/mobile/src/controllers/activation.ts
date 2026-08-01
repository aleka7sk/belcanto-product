import type {
  ActivationPreviewRequest,
  CompleteActivationRequest,
} from "@/api/contracts";
import { parseOpaqueActivationToken } from "@/activation/links";
import {
  invalid,
  normalizePhone,
  unicodeLength,
  valid,
  type FormIssue,
  type FormResult,
} from "@/forms/result";
import {
  requiredIdempotencyKey,
  utf8ByteLength,
} from "@/validation/backend";
import { COMMON_PASSWORDS } from "@/validation/commonPasswords";

export interface ActivationPreviewDraft {
  token: string;
}

export function prepareActivationPreview(
  draft: ActivationPreviewDraft,
): FormResult<ActivationPreviewRequest, "token"> {
  const token = parseOpaqueActivationToken(draft.token);
  return token === null
    ? invalid([{ field: "token", code: "invalid_format" }])
    : valid({ token });
}

export interface CompleteActivationDraft {
  token: string;
  phone: string;
  password: string;
  passwordConfirmation: string;
}

type ActivationField = keyof CompleteActivationDraft;

export interface CompleteActivationCommand {
  body: CompleteActivationRequest;
  idempotencyKey: string;
}

export function prepareCompleteActivation(
  draft: CompleteActivationDraft,
  suppliedIdempotencyKey: string,
): FormResult<CompleteActivationCommand, ActivationField> {
  const issues: FormIssue<ActivationField>[] = [];
  const token = parseOpaqueActivationToken(draft.token);
  const phone = normalizePhone(draft.phone);
  const password = draft.password.normalize("NFC");
  const confirmation = draft.passwordConfirmation.normalize("NFC");
  const idempotencyKey = requiredIdempotencyKey(suppliedIdempotencyKey);
  if (token === null) issues.push({ field: "token", code: "invalid_format" });
  if (phone === null) issues.push({ field: "phone", code: "invalid_format" });
  const passwordLength = unicodeLength(password);
  if (utf8ByteLength(password) === null) {
    issues.push({ field: "password", code: "invalid_format" });
  } else if (passwordLength < 15) {
    issues.push({ field: "password", code: "too_short" });
  } else if (passwordLength > 128) {
    issues.push({ field: "password", code: "too_long" });
  } else if (COMMON_PASSWORDS.has(password.toLowerCase())) {
    issues.push({ field: "password", code: "invalid_value" });
  }
  if (password !== confirmation) {
    issues.push({ field: "passwordConfirmation", code: "mismatch" });
  }
  if (issues.length > 0 || token === null || phone === null) {
    return invalid(issues);
  }
  return valid({ body: { token, phone, password }, idempotencyKey });
}
