export type FormIssueCode =
  | "required"
  | "invalid_format"
  | "must_be_future"
  | "must_confirm"
  | "too_short"
  | "too_long"
  | "mismatch"
  | "invalid_value";

export interface FormIssue<Field extends string> {
  field: Field;
  code: FormIssueCode;
}

export type FormResult<Value, Field extends string> =
  | { ok: true; value: Value }
  | { ok: false; issues: FormIssue<Field>[] };

export function invalid<Value, Field extends string>(
  issues: FormIssue<Field>[],
): FormResult<Value, Field> {
  return { ok: false, issues };
}

export function valid<Value, Field extends string>(
  value: Value,
): FormResult<Value, Field> {
  return { ok: true, value };
}

export function normalizedRequired(
  value: string,
  normalize: (value: string) => string = (candidate) => candidate.trim(),
): string | null {
  const normalized = normalize(value);
  return normalized.length === 0 ? null : normalized;
}

export function normalizePhone(value: string): string | null {
  const normalized = value.trim().replace(/[\s\-()]/g, "");
  return /^\+[1-9][0-9]{7,14}$/.test(normalized) ? normalized : null;
}

export function unicodeLength(value: string): number {
  return Array.from(value).length;
}
