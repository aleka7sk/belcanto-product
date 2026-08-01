import type { FormIssueCode } from "@/forms/result";

export const BACKEND_INPUT_LIMITS = {
  identifierBytes: 128,
  idempotencyKeyBytes: 128,
  fullNameRunes: 200,
  enrollmentReferenceRunes: 100,
  reasonRunes: 500,
  firstMinuteRunes: 500,
  localeBytes: 35,
  timezoneBytes: 100,
} as const;

const IDENTIFIER_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]*$/;
const IDEMPOTENCY_KEY_PATTERN = /^[\x21-\x7e]+$/;
const CONTROL_CHARACTER_PATTERN = /[\u0000-\u001f\u007f-\u009f]/;

export function utf8ByteLength(value: string): number | null {
  let bytes = 0;
  for (let index = 0; index < value.length; index += 1) {
    const code = value.charCodeAt(index);
    if (code <= 0x7f) {
      bytes += 1;
    } else if (code <= 0x7ff) {
      bytes += 2;
    } else if (code >= 0xd800 && code <= 0xdbff) {
      const next = value.charCodeAt(index + 1);
      if (!Number.isFinite(next) || next < 0xdc00 || next > 0xdfff) return null;
      bytes += 4;
      index += 1;
    } else if (code >= 0xdc00 && code <= 0xdfff) {
      return null;
    } else {
      bytes += 3;
    }
  }
  return bytes;
}

export function backendTextIssue(
  value: string,
  maximumRunes: number,
): FormIssueCode | null {
  if (
    utf8ByteLength(value) === null ||
    CONTROL_CHARACTER_PATTERN.test(value)
  ) {
    return "invalid_format";
  }
  return Array.from(value).length > maximumRunes ? "too_long" : null;
}

export function backendIdentifierIssue(value: string): FormIssueCode | null {
  return value.length > BACKEND_INPUT_LIMITS.identifierBytes
    ? "too_long"
    : IDENTIFIER_PATTERN.test(value)
      ? null
      : "invalid_format";
}

export function isValidIdempotencyKey(value: string): boolean {
  return (
    value.length >= 1 &&
    value.length <= BACKEND_INPUT_LIMITS.idempotencyKeyBytes &&
    IDEMPOTENCY_KEY_PATTERN.test(value)
  );
}

export function requiredIdempotencyKey(value: string): string {
  const normalized = value.trim();
  if (!isValidIdempotencyKey(normalized)) {
    throw new TypeError(
      "idempotencyKey must contain 1..128 visible ASCII bytes",
    );
  }
  return normalized;
}

export function backendLocaleIssue(value: string): FormIssueCode | null {
  const bytes = utf8ByteLength(value);
  if (bytes === null) return "invalid_format";
  if (bytes > BACKEND_INPUT_LIMITS.localeBytes) return "too_long";
  try {
    Intl.getCanonicalLocales(value);
    return null;
  } catch {
    return "invalid_format";
  }
}

export function backendTimezoneIssue(value: string): FormIssueCode | null {
  const bytes = utf8ByteLength(value);
  if (bytes === null) return "invalid_format";
  if (bytes > BACKEND_INPUT_LIMITS.timezoneBytes) return "too_long";
  try {
    new Intl.DateTimeFormat("en-US", { timeZone: value }).format(0);
    return null;
  } catch {
    return "invalid_format";
  }
}
