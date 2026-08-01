import type { SignInRequest } from "@/api/contracts";
import {
  invalid,
  normalizePhone,
  normalizedRequired,
  valid,
  type FormIssue,
  type FormResult,
} from "@/forms/result";

export interface SignInDraft {
  phone: string;
  password: string;
}

export function prepareSignIn(
  draft: SignInDraft,
): FormResult<SignInRequest, keyof SignInDraft> {
  const issues: FormIssue<keyof SignInDraft>[] = [];
  const phone = normalizePhone(draft.phone);
  const password = normalizedRequired(draft.password, (value) => value);
  if (phone === null) issues.push({ field: "phone", code: "invalid_format" });
  if (password === null) issues.push({ field: "password", code: "required" });
  if (issues.length > 0 || phone === null || password === null) {
    return invalid(issues);
  }
  return valid({ phone, password });
}
