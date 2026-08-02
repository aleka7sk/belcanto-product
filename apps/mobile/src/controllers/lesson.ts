import type {
  CreateLessonRequest,
  IsoDateTime,
  LessonTeacherReplacementInput,
  PrimaryTeacherReassignmentInput,
  ReassignPrimaryTeachersRequest,
  ReplaceLessonTeachersRequest,
} from "@/api/contracts";
import {
  invalid,
  normalizedRequired,
  valid,
  type FormIssue,
  type FormResult,
} from "@/forms/result";
import {
  backendIdentifierIssue,
  backendTextIssue,
  requiredIdempotencyKey,
} from "@/validation/backend";
import { parseAlmatyLocalDateTime } from "@/validation/datetime";

const LESSON_TEXT_LIMIT = 200;

function uniqueIdentifiers<Field extends string>(
  values: readonly string[],
  field: Field,
  issues: FormIssue<Field>[],
): string[] | null {
  const normalized = values.map((value) => value.trim());
  if (normalized.length === 0) {
    issues.push({ field, code: "required" });
    return null;
  }
  if (normalized.length > 100) {
    issues.push({ field, code: "invalid_value" });
    return null;
  }
  if (
    normalized.some((value) => backendIdentifierIssue(value) !== null) ||
    new Set(normalized).size !== normalized.length
  ) {
    issues.push({ field, code: "invalid_value" });
    return null;
  }
  return normalized;
}

export interface CreateLessonDraft {
  title: string;
  startsOn: string;
  startsAtTime: string;
  durationMinutes: string;
  location: string;
  teacherAccountId: string;
  studentIds: string[];
}

type CreateLessonField = keyof CreateLessonDraft;

export interface CreateLessonCommand {
  body: CreateLessonRequest;
  idempotencyKey: string;
}

export function prepareCreateLesson(
  draft: CreateLessonDraft,
  suppliedIdempotencyKey: string,
  now = new Date(),
): FormResult<CreateLessonCommand, CreateLessonField> {
  const issues: FormIssue<CreateLessonField>[] = [];
  const title = normalizedRequired(draft.title);
  const location = draft.location.trim();
  const teacherAccountId = normalizedRequired(draft.teacherAccountId);
  const studentIds = uniqueIdentifiers(draft.studentIds, "studentIds", issues);
  const startsAtTimestamp = parseAlmatyLocalDateTime(
    draft.startsOn,
    draft.startsAtTime,
  );
  const durationMinutes = Number(draft.durationMinutes.trim());

  if (title === null) issues.push({ field: "title", code: "required" });
  else {
    const code = backendTextIssue(title, LESSON_TEXT_LIMIT);
    if (code !== null) issues.push({ field: "title", code });
  }
  if (startsAtTimestamp === null) {
    issues.push({ field: "startsOn", code: "invalid_format" });
    issues.push({ field: "startsAtTime", code: "invalid_format" });
  } else if (startsAtTimestamp <= now.getTime()) {
    issues.push({ field: "startsOn", code: "must_be_future" });
  }
  if (
    !Number.isSafeInteger(durationMinutes) ||
    durationMinutes < 1 ||
    durationMinutes > 1440
  ) {
    issues.push({ field: "durationMinutes", code: "invalid_value" });
  }
  if (location.length > 0) {
    const code = backendTextIssue(location, LESSON_TEXT_LIMIT);
    if (code !== null) issues.push({ field: "location", code });
  }
  if (teacherAccountId === null) {
    issues.push({ field: "teacherAccountId", code: "required" });
  } else {
    const code = backendIdentifierIssue(teacherAccountId);
    if (code !== null) issues.push({ field: "teacherAccountId", code });
  }
  if (
    issues.length > 0 ||
    title === null ||
    startsAtTimestamp === null ||
    teacherAccountId === null ||
    studentIds === null
  ) {
    return invalid(issues);
  }
  const body: CreateLessonRequest = {
    title,
    startsAt: new Date(startsAtTimestamp).toISOString() as IsoDateTime,
    durationMinutes,
    teacherAccountId,
    studentIds,
  };
  if (location.length > 0) body.location = location;
  return valid({ body, idempotencyKey: requiredIdempotencyKey(suppliedIdempotencyKey) });
}

export interface ReassignPrimaryTeachersDraft {
  students: PrimaryTeacherReassignmentInput[];
  newTeacherAccountId: string;
  effectiveImmediately: boolean;
  effectiveOn: string;
  effectiveAtTime: string;
}

type ReassignField = keyof ReassignPrimaryTeachersDraft;

export interface ReassignPrimaryTeachersCommand {
  body: ReassignPrimaryTeachersRequest;
  idempotencyKey: string;
}

export function prepareReassignPrimaryTeachers(
  draft: ReassignPrimaryTeachersDraft,
  suppliedIdempotencyKey: string,
  now = new Date(),
): FormResult<ReassignPrimaryTeachersCommand, ReassignField> {
  const issues: FormIssue<ReassignField>[] = [];
  const studentIds = draft.students.map((student) => student.studentId.trim());
  const newTeacherAccountId = normalizedRequired(draft.newTeacherAccountId);
  const effectiveFromTimestamp = draft.effectiveImmediately
    ? null
    : parseAlmatyLocalDateTime(draft.effectiveOn, draft.effectiveAtTime);

  if (draft.students.length === 0) {
    issues.push({ field: "students", code: "required" });
  } else if (
    studentIds.some((studentId) => backendIdentifierIssue(studentId) !== null) ||
    new Set(studentIds).size !== studentIds.length ||
    draft.students.some(
      (student) =>
        !Number.isSafeInteger(student.expectedAssignmentVersion) ||
        student.expectedAssignmentVersion < 0,
    )
  ) {
    issues.push({ field: "students", code: "invalid_value" });
  }

  if (newTeacherAccountId === null) {
    issues.push({ field: "newTeacherAccountId", code: "required" });
  } else {
    const code = backendIdentifierIssue(newTeacherAccountId);
    if (code !== null) issues.push({ field: "newTeacherAccountId", code });
  }
  if (!draft.effectiveImmediately && effectiveFromTimestamp === null) {
    issues.push({ field: "effectiveOn", code: "invalid_format" });
    issues.push({ field: "effectiveAtTime", code: "invalid_format" });
  } else if (
    !draft.effectiveImmediately &&
    effectiveFromTimestamp !== null &&
    effectiveFromTimestamp < now.getTime()
  ) {
    issues.push({ field: "effectiveOn", code: "must_be_future" });
  }
  if (draft.students.length > 100) {
    issues.push({ field: "students", code: "invalid_value" });
  }
  if (
    issues.length > 0 ||
    draft.students.length === 0 ||
    newTeacherAccountId === null ||
    (!draft.effectiveImmediately && effectiveFromTimestamp === null)
  ) {
    return invalid(issues);
  }
  const students = draft.students.map((student, index) => ({
    studentId: studentIds[index]!,
    expectedAssignmentVersion: student.expectedAssignmentVersion,
  }));
  const body: ReassignPrimaryTeachersRequest = draft.effectiveImmediately
    ? { students, newTeacherAccountId, effectiveMode: "immediate" }
    : {
        students,
        newTeacherAccountId,
        effectiveMode: "scheduled",
        effectiveFrom: new Date(effectiveFromTimestamp!).toISOString() as IsoDateTime,
      };
  return valid({
    body,
    idempotencyKey: requiredIdempotencyKey(suppliedIdempotencyKey),
  });
}

export interface ReplaceLessonTeachersDraft {
  lessons: LessonTeacherReplacementInput[];
  newTeacherAccountId: string;
}

type ReplaceField = keyof ReplaceLessonTeachersDraft;

export interface ReplaceLessonTeachersCommand {
  body: ReplaceLessonTeachersRequest;
  idempotencyKey: string;
}

export function prepareReplaceLessonTeachers(
  draft: ReplaceLessonTeachersDraft,
  suppliedIdempotencyKey: string,
): FormResult<ReplaceLessonTeachersCommand, ReplaceField> {
  const issues: FormIssue<ReplaceField>[] = [];
  const newTeacherAccountId = normalizedRequired(draft.newTeacherAccountId);
  const lessonIds = draft.lessons.map((lesson) => lesson.lessonId.trim());

  if (draft.lessons.length === 0) {
    issues.push({ field: "lessons", code: "required" });
  } else if (draft.lessons.length > 100) {
    issues.push({ field: "lessons", code: "invalid_value" });
  } else if (
    lessonIds.some((lessonId) => backendIdentifierIssue(lessonId) !== null) ||
    draft.lessons.some(
      (lesson) =>
        backendIdentifierIssue(lesson.expectedPreviousTeacherAccountId.trim()) !== null,
    ) ||
    new Set(lessonIds).size !== lessonIds.length ||
    draft.lessons.some(
      (lesson) =>
        !Number.isSafeInteger(lesson.expectedVersion) || lesson.expectedVersion < 0,
    )
  ) {
    issues.push({ field: "lessons", code: "invalid_value" });
  }
  if (newTeacherAccountId === null) {
    issues.push({ field: "newTeacherAccountId", code: "required" });
  } else {
    const code = backendIdentifierIssue(newTeacherAccountId);
    if (code !== null) issues.push({ field: "newTeacherAccountId", code });
  }
  if (issues.length > 0 || newTeacherAccountId === null) {
    return invalid(issues);
  }
  return valid({
    body: {
      lessons: draft.lessons.map((lesson, index) => ({
        lessonId: lessonIds[index]!,
        expectedVersion: lesson.expectedVersion,
        expectedPreviousTeacherAccountId:
          lesson.expectedPreviousTeacherAccountId.trim(),
      })),
      newTeacherAccountId,
    },
    idempotencyKey: requiredIdempotencyKey(suppliedIdempotencyKey),
  });
}
