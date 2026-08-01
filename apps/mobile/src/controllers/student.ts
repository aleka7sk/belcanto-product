import type { CreateStudentRequest } from "@/api/contracts";
import {
  invalid,
  normalizePhone,
  normalizedRequired,
  valid,
  type FormIssue,
  type FormResult,
} from "@/forms/result";
import {
  BACKEND_INPUT_LIMITS,
  backendIdentifierIssue,
  backendLocaleIssue,
  backendTextIssue,
  backendTimezoneIssue,
  requiredIdempotencyKey,
} from "@/validation/backend";

export interface CreateStudentDraft {
  fullName: string;
  phone: string;
  enrollmentReference: string;
  teacherAccountId: string;
  locale: string;
  timezone: string;
  adultConfirmed: boolean;
}

type StudentField = keyof CreateStudentDraft;

export interface CreateStudentCommand {
  body: CreateStudentRequest;
  idempotencyKey: string;
}

export function prepareCreateStudent(
  draft: CreateStudentDraft,
  suppliedIdempotencyKey: string,
): FormResult<CreateStudentCommand, StudentField> {
  const issues: FormIssue<StudentField>[] = [];
  const fullName = normalizedRequired(draft.fullName);
  const phone = normalizePhone(draft.phone);
  const enrollmentReference = normalizedRequired(draft.enrollmentReference);
  const teacherAccountId = normalizedRequired(draft.teacherAccountId);
  const locale = draft.locale === "" ? "ru-KZ" : draft.locale.trim();
  const timezone =
    draft.timezone === "" ? "Asia/Almaty" : draft.timezone.trim();
  const idempotencyKey = requiredIdempotencyKey(suppliedIdempotencyKey);
  if (fullName === null) issues.push({ field: "fullName", code: "required" });
  else {
    const code = backendTextIssue(fullName, BACKEND_INPUT_LIMITS.fullNameRunes);
    if (code !== null) issues.push({ field: "fullName", code });
  }
  if (phone === null) issues.push({ field: "phone", code: "invalid_format" });
  if (enrollmentReference === null)
    issues.push({ field: "enrollmentReference", code: "required" });
  else {
    const code = backendTextIssue(
      enrollmentReference,
      BACKEND_INPUT_LIMITS.enrollmentReferenceRunes,
    );
    if (code !== null) issues.push({ field: "enrollmentReference", code });
  }
  if (teacherAccountId === null)
    issues.push({ field: "teacherAccountId", code: "required" });
  else {
    const code = backendIdentifierIssue(teacherAccountId);
    if (code !== null) issues.push({ field: "teacherAccountId", code });
  }
  const localeCode = backendLocaleIssue(locale);
  if (localeCode !== null) issues.push({ field: "locale", code: localeCode });
  const timezoneCode = backendTimezoneIssue(timezone);
  if (timezoneCode !== null)
    issues.push({ field: "timezone", code: timezoneCode });
  if (!draft.adultConfirmed)
    issues.push({ field: "adultConfirmed", code: "must_confirm" });
  if (
    issues.length > 0 ||
    fullName === null ||
    phone === null ||
    enrollmentReference === null ||
    teacherAccountId === null
  ) {
    return invalid(issues);
  }
  return valid({
    body: {
      fullName,
      phone,
      enrollmentReference,
      teacherAccountId,
      locale,
      timezone,
      adultConfirmed: true,
    },
    idempotencyKey,
  });
}
