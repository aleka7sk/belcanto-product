import {
  parseCanonicalActivationLink,
  type ActivationLinkPolicy,
} from "@/activation/links";
import { parseStrictRfc3339 } from "@/validation/datetime";

export const ROLES = ["Owner", "Administrator", "Teacher", "Student"] as const;

export type Role = (typeof ROLES)[number];

export const STUDENT_ONBOARDING_MANAGER_BUNDLE =
  "StudentOnboardingManager.v1" as const;

export type CapabilityBundle = typeof STUDENT_ONBOARDING_MANAGER_BUNDLE;

export const PERMISSIONS = [
  "students.create",
  "student_onboarding.read",
  "student_invitations.issue",
  "student_invitations.reissue",
  "student_invitations.revoke",
  "student_onboarding.delegate",
  "lessons.read",
  "lessons.create",
  "lesson_teachers.replace",
  "student_primary_teachers.reassign",
] as const;

export type Permission = (typeof PERMISSIONS)[number];

const DELEGATED_STUDENT_ONBOARDING_PERMISSIONS = [
  "students.create",
  "student_onboarding.read",
] as const satisfies readonly Permission[];

const OWNER_STUDENT_ONBOARDING_PERMISSIONS = [
  "students.create",
  "student_onboarding.read",
  "student_invitations.issue",
  "student_invitations.reissue",
  "student_invitations.revoke",
  "student_onboarding.delegate",
] as const satisfies readonly Permission[];

const LESSON_READER_PERMISSIONS = ["lessons.read"] as const satisfies readonly Permission[];
const LESSON_CREATOR_PERMISSIONS = [
  "lessons.read",
  "lessons.create",
] as const satisfies readonly Permission[];
const LESSON_MANAGER_PERMISSIONS = [
  "lessons.read",
  "lessons.create",
  "lesson_teachers.replace",
  "student_primary_teachers.reassign",
] as const satisfies readonly Permission[];

export type IsoDateTime = string & { readonly __isoDateTime: unique symbol };

export const API_ERROR_CODES = [
  "INVALID_INPUT",
  "UNAUTHENTICATED",
  "FORBIDDEN",
  "NOT_FOUND",
  "CONFLICT",
  "INVALID_STATE",
  "ACTIVATION_INVALID",
  "RATE_LIMITED",
  "UNAVAILABLE",
  "INTERNAL",
] as const;

export type ApiErrorCode = (typeof API_ERROR_CODES)[number];

export interface ApiErrorEnvelope {
  error: {
    code: ApiErrorCode;
    message: string;
    requestId: string;
  };
}

export interface ActivationPreviewRequest {
  token: string;
}

export interface ActivationPreview {
  invitationId: string;
  kind: "owner_bootstrap" | "staff_activation" | "student_activation";
  displayName: string;
  maskedPhone: string;
  expiresAt: IsoDateTime;
}

export interface CompleteActivationRequest {
  token: string;
  phone: string;
  password: string;
}

export type SessionPlatform = "ios" | "android" | "web";

export const SESSION_PLATFORMS = ["ios", "android", "web"] as const;

export interface SignInRequest {
  phone: string;
  password: string;
  deviceLabel?: string;
  platform?: SessionPlatform;
}

export interface RefreshSessionRequest {
  refreshToken: string;
}

export interface SessionTokens {
  accessToken: string;
  refreshToken: string;
  accessExpiresAt: IsoDateTime;
  refreshExpiresAt: IsoDateTime;
}

export interface BootstrapAccessView {
  accountId: string;
  roles: Role[];
  accessProfiles: CapabilityBundle[];
  permissions: Permission[];
}

export type BootstrapView = BootstrapAccessView &
  (
    | {
        studentId: string;
        fullName: string;
        firstMinute: FirstMinute;
      }
    | {
        studentId?: never;
        fullName?: never;
        firstMinute?: never;
      }
  );

export type StaffRole = Extract<Role, "Administrator" | "Teacher">;

export interface StaffMember {
  accountId: string;
  fullName: string;
  roles: Role[];
  accessProfiles: CapabilityBundle[];
  onboardingDelegationId?: string;
  onboardingDelegationExpiresAt?: IsoDateTime;
}

export const STUDENT_ONBOARDING_STATES = [
  "awaiting_first_minute",
  "ready_to_invite",
  "invited",
  "activated",
] as const;

export type StudentOnboardingState =
  (typeof STUDENT_ONBOARDING_STATES)[number];

export interface StudentOnboardingItem {
  studentId: string;
  fullName: string;
  enrollmentReference: string;
  teacherAccountId: string;
  studentVersion: number;
  onboardingState: StudentOnboardingState;
  invitationId?: string;
  invitationExpiresAt?: IsoDateTime;
}

export interface GrantDelegationRequest {
  administratorAccountId: string;
  reason: string;
  expiresAt?: IsoDateTime | null;
  currentPassword: string;
}

export interface DelegationResult {
  id: string;
  administratorAccountId: string;
  bundle: CapabilityBundle;
  status: "active";
  grantedAt: IsoDateTime;
  expiresAt?: IsoDateTime;
}

export interface RevokeDelegationRequest {
  reason: string;
  currentPassword: string;
}

export interface CreateStudentRequest {
  fullName: string;
  phone: string;
  enrollmentReference: string;
  teacherAccountId: string;
  locale?: string;
  timezone?: string;
  adultConfirmed: true;
}

export interface StudentResult {
  studentId: string;
  accountId: string;
  onboardingState: "awaiting_first_minute";
}

export interface PublishFirstMinuteRequest {
  whatWorked: string;
  currentFocus: string;
  nextStep: string;
  expectedVersion: number;
}

export interface FirstMinute {
  studentId: string;
  revision: number;
  whatWorked: string;
  currentFocus: string;
  nextStep: string;
  publishedAt: IsoDateTime;
}

export interface InvitationResult {
  invitationId: string;
  studentId: string;
  status: "issued";
  expiresAt: IsoDateTime;
  activationLink: string;
}

export interface LessonTeacher {
  accountId: string;
  fullName: string;
}

export interface AssignedTeacherSummary extends LessonTeacher {
  status: "active" | "inactive";
}

export interface LessonStudent {
  studentId: string;
  fullName: string;
}

export const LESSON_OCCURRENCE_STATUSES = [
  "scheduled",
  "completed",
  "cancelled_school",
  "cancelled_student",
  "rescheduled",
  "no_show",
] as const;

export type LessonOccurrenceStatus = (typeof LESSON_OCCURRENCE_STATUSES)[number];

export interface Lesson {
  id: string;
  title: string;
  /** Core Lesson format (DEC-002); absent on legacy pre-core Lessons. */
  format?: "individual" | "group";
  startsAt: IsoDateTime;
  durationMinutes: number;
  location?: string;
  teacher: LessonTeacher;
  students: LessonStudent[];
  status: LessonOccurrenceStatus;
  version: number;
}

export type LessonSummary = Lesson;
export type LessonDetail = Lesson;

export interface LessonListQuery {
  from: IsoDateTime;
  to: IsoDateTime;
  studentId?: string;
  teacherAccountId?: string;
}

export interface StudentDirectoryQuery {
  asOf?: IsoDateTime;
}

export interface CreateLessonRequest {
  title: string;
  startsAt: IsoDateTime;
  durationMinutes: number;
  location?: string;
  teacherAccountId: string;
  studentIds: string[];
}

export interface StudentDirectoryItem {
  studentId: string;
  fullName: string;
  primaryTeacher: AssignedTeacherSummary;
  primaryTeacherAssignmentVersion: number;
}

export interface PrimaryTeacherReassignmentInput {
  studentId: string;
  expectedAssignmentVersion: number;
}

export type ReassignPrimaryTeachersRequest =
  | {
      students: PrimaryTeacherReassignmentInput[];
      newTeacherAccountId: string;
      effectiveMode: "immediate";
      effectiveFrom?: never;
    }
  | {
      students: PrimaryTeacherReassignmentInput[];
      newTeacherAccountId: string;
      effectiveMode: "scheduled";
      effectiveFrom: IsoDateTime;
    };

export interface PrimaryTeacherAssignmentResult {
  studentId: string;
  previousTeacherAccountId: string;
  newTeacherAccountId: string;
  effectiveFrom: IsoDateTime;
  version: number;
}

export interface ReassignPrimaryTeachersResult {
  reassignedCount: number;
  assignments: PrimaryTeacherAssignmentResult[];
}

export interface LessonTeacherReplacementInput {
  lessonId: string;
  expectedVersion: number;
  expectedPreviousTeacherAccountId: string;
}

export interface ReplaceLessonTeachersRequest {
  lessons: LessonTeacherReplacementInput[];
  newTeacherAccountId: string;
}

export interface ReplaceLessonTeachersResult {
  updatedCount: number;
  lessons: Lesson[];
}

export interface EmptyRequest {
  readonly __empty?: never;
}

type JsonRecord = Record<string, unknown>;

const OPAQUE_TOKEN_PATTERN = /^[A-Za-z0-9_-]{43}$/;
const IDENTIFIER_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._:-]*$/;

export type Decoder<T> = (value: unknown) => T;

export class ContractDecodeError extends Error {
  constructor(readonly contract: string, readonly path: string) {
    super(`Invalid ${contract} response at ${path}`);
    this.name = "ContractDecodeError";
  }
}

function record(value: unknown, contract: string, path = "$"): JsonRecord {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new ContractDecodeError(contract, path);
  }
  return value as JsonRecord;
}

function exactKeys(
  value: JsonRecord,
  keys: readonly string[],
  contract: string,
  path = "$",
): void {
  const allowed = new Set(keys);
  const unexpected = Object.keys(value).find((key) => !allowed.has(key));
  if (unexpected !== undefined) {
    throw new ContractDecodeError(contract, `${path}.${unexpected}`);
  }
}

function stringField(
  value: JsonRecord,
  key: string,
  contract: string,
  optional = false,
): string | undefined {
  const field = value[key];
  if (field === undefined && optional) {
    return undefined;
  }
  if (typeof field !== "string" || field.length === 0) {
    throw new ContractDecodeError(contract, `$.${key}`);
  }
  return field;
}

function opaqueTokenField(
  value: JsonRecord,
  key: string,
  contract: string,
): string {
  const field = stringField(value, key, contract)!;
  if (!OPAQUE_TOKEN_PATTERN.test(field)) {
    throw new ContractDecodeError(contract, `$.${key}`);
  }
  return field;
}

function identifierField(
  value: JsonRecord,
  key: string,
  contract: string,
  optional = false,
): string | undefined {
  const field = stringField(value, key, contract, optional);
  if (
    field !== undefined &&
    (field.length > 128 || !IDENTIFIER_PATTERN.test(field))
  ) {
    throw new ContractDecodeError(contract, `$.${key}`);
  }
  return field;
}

function unique<const T extends string>(
  values: T[],
  contract: string,
  path: string,
): T[] {
  if (new Set(values).size !== values.length) {
    throw new ContractDecodeError(contract, path);
  }
  return values;
}

function numberField(
  value: JsonRecord,
  key: string,
  contract: string,
  minimum = 0,
): number {
  const field = value[key];
  if (
    typeof field !== "number" ||
    !Number.isSafeInteger(field) ||
    field < minimum
  ) {
    throw new ContractDecodeError(contract, `$.${key}`);
  }
  return field;
}

function isoDateField(
  value: JsonRecord,
  key: string,
  contract: string,
  optional = false,
): IsoDateTime | undefined {
  const field = stringField(value, key, contract, optional);
  if (field === undefined) {
    return undefined;
  }
  if (parseStrictRfc3339(field) === null) {
    throw new ContractDecodeError(contract, `$.${key}`);
  }
  return field as IsoDateTime;
}

function oneOf<const T extends readonly string[]>(
  value: unknown,
  choices: T,
  contract: string,
  path: string,
): T[number] {
  if (typeof value !== "string" || !choices.includes(value)) {
    throw new ContractDecodeError(contract, path);
  }
  return value as T[number];
}

function rolesField(
  source: JsonRecord,
  contract: string,
): Role[] {
  if (!Array.isArray(source.roles)) {
    throw new ContractDecodeError(contract, "$.roles");
  }
  return unique(
    source.roles.map((role, index) =>
      oneOf(role, ROLES, contract, `$.roles[${index}]`),
    ),
    contract,
    "$.roles",
  );
}

function accessProfilesField(
  source: JsonRecord,
  contract: string,
): CapabilityBundle[] {
  if (!Array.isArray(source.accessProfiles)) {
    throw new ContractDecodeError(contract, "$.accessProfiles");
  }
  return unique(
    source.accessProfiles.map((profile, index) =>
      oneOf(
        profile,
        [STUDENT_ONBOARDING_MANAGER_BUNDLE] as const,
        contract,
        `$.accessProfiles[${index}]`,
      ),
    ),
    contract,
    "$.accessProfiles",
  );
}

export const decodeActivationPreview: Decoder<ActivationPreview> = (value) => {
  const contract = "ActivationPreview";
  const source = record(value, contract);
  exactKeys(
    source,
    ["invitationId", "kind", "displayName", "maskedPhone", "expiresAt"],
    contract,
  );
  return {
    invitationId: stringField(source, "invitationId", contract)!,
    kind: oneOf(
      source.kind,
      ["owner_bootstrap", "staff_activation", "student_activation"] as const,
      contract,
      "$.kind",
    ),
    displayName: stringField(source, "displayName", contract)!,
    maskedPhone: stringField(source, "maskedPhone", contract)!,
    expiresAt: isoDateField(source, "expiresAt", contract)!,
  };
};

export const decodeSessionTokens: Decoder<SessionTokens> = (value) => {
  const contract = "SessionTokens";
  const source = record(value, contract);
  exactKeys(
    source,
    ["accessToken", "refreshToken", "accessExpiresAt", "refreshExpiresAt"],
    contract,
  );
  return {
    accessToken: opaqueTokenField(source, "accessToken", contract),
    refreshToken: opaqueTokenField(source, "refreshToken", contract),
    accessExpiresAt: isoDateField(source, "accessExpiresAt", contract)!,
    refreshExpiresAt: isoDateField(source, "refreshExpiresAt", contract)!,
  };
};

export const decodeFirstMinute: Decoder<FirstMinute> = (value) => {
  const contract = "FirstMinute";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "studentId",
      "revision",
      "whatWorked",
      "currentFocus",
      "nextStep",
      "publishedAt",
    ],
    contract,
  );
  return {
    studentId: stringField(source, "studentId", contract)!,
    revision: numberField(source, "revision", contract, 1),
    whatWorked: stringField(source, "whatWorked", contract)!,
    currentFocus: stringField(source, "currentFocus", contract)!,
    nextStep: stringField(source, "nextStep", contract)!,
    publishedAt: isoDateField(source, "publishedAt", contract)!,
  };
};

function decodeLessonTeacher(value: unknown, contract: string, path: string): LessonTeacher {
  const source = record(value, contract, path);
  exactKeys(source, ["accountId", "fullName"], contract, path);
  return {
    accountId: identifierField(source, "accountId", contract)!,
    fullName: stringField(source, "fullName", contract)!,
  };
}

function decodeAssignedTeacher(
  value: unknown,
  contract: string,
  path: string,
): AssignedTeacherSummary {
  const source = record(value, contract, path);
  exactKeys(source, ["accountId", "fullName", "status"], contract, path);
  return {
    accountId: identifierField(source, "accountId", contract)!,
    fullName: stringField(source, "fullName", contract)!,
    status: oneOf(
      source.status,
      ["active", "inactive"] as const,
      contract,
      `${path}.status`,
    ),
  };
}

function decodeLessonStudent(value: unknown, contract: string, path: string): LessonStudent {
  const source = record(value, contract, path);
  exactKeys(source, ["studentId", "fullName"], contract, path);
  return {
    studentId: identifierField(source, "studentId", contract)!,
    fullName: stringField(source, "fullName", contract)!,
  };
}

export const decodeLesson: Decoder<Lesson> = (value) => {
  const contract = "Lesson";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "id",
      "title",
      "format",
      "startsAt",
      "durationMinutes",
      "location",
      "teacher",
      "students",
      "status",
      "version",
    ],
    contract,
  );
  if (!Array.isArray(source.students) || source.students.length === 0) {
    throw new ContractDecodeError(contract, "$.students");
  }
  const students = source.students.map((student, index) =>
    decodeLessonStudent(student, contract, `$.students[${index}]`),
  );
  if (new Set(students.map((student) => student.studentId)).size !== students.length) {
    throw new ContractDecodeError(contract, "$.students");
  }
  const lesson: Lesson = {
    id: identifierField(source, "id", contract)!,
    title: stringField(source, "title", contract)!,
    startsAt: isoDateField(source, "startsAt", contract)!,
    durationMinutes: numberField(source, "durationMinutes", contract, 1),
    teacher: decodeLessonTeacher(source.teacher, contract, "$.teacher"),
    students,
    status: oneOf(source.status, LESSON_OCCURRENCE_STATUSES, contract, "$.status"),
    version: numberField(source, "version", contract),
  };
  const location = stringField(source, "location", contract, true);
  if (location !== undefined) lesson.location = location;
  if (source.format !== undefined) {
    lesson.format = oneOf(
      source.format,
      ["individual", "group"] as const,
      contract,
      "$.format",
    );
    // DEC-002: a group Lesson never exceeds three participants, and the
    // format never contradicts the visible participant count.
    if (lesson.format === "individual" && students.length > 1) {
      throw new ContractDecodeError(contract, "$.format");
    }
  }
  return lesson;
};

export const decodeLessons: Decoder<Lesson[]> = (value) => {
  if (!Array.isArray(value)) throw new ContractDecodeError("Lesson[]", "$");
  const lessons = value.map((lesson, index) => {
    try {
      return decodeLesson(lesson);
    } catch (error) {
      if (error instanceof ContractDecodeError) {
        throw new ContractDecodeError("Lesson[]", `$[${index}]${error.path.slice(1)}`);
      }
      throw error;
    }
  });
  if (new Set(lessons.map((lesson) => lesson.id)).size !== lessons.length) {
    throw new ContractDecodeError("Lesson[]", "$");
  }
  return lessons;
};

export const decodeStudentDirectory: Decoder<StudentDirectoryItem[]> = (value) => {
  const contract = "StudentDirectoryItem[]";
  if (!Array.isArray(value)) throw new ContractDecodeError(contract, "$");
  const students = value.map((item, index) => {
    const source = record(item, contract, `$[${index}]`);
    exactKeys(
      source,
      ["studentId", "fullName", "primaryTeacher", "primaryTeacherAssignmentVersion"],
      contract,
      `$[${index}]`,
    );
    return {
      studentId: identifierField(source, "studentId", contract)!,
      fullName: stringField(source, "fullName", contract)!,
      primaryTeacher: decodeAssignedTeacher(
        source.primaryTeacher,
        contract,
        `$[${index}].primaryTeacher`,
      ),
      primaryTeacherAssignmentVersion: numberField(
        source,
        "primaryTeacherAssignmentVersion",
        contract,
      ),
    };
  });
  if (new Set(students.map((student) => student.studentId)).size !== students.length) {
    throw new ContractDecodeError(contract, "$");
  }
  return students;
};

export const decodeReassignPrimaryTeachersResult: Decoder<
  ReassignPrimaryTeachersResult
> = (value) => {
  const contract = "ReassignPrimaryTeachersResult";
  const source = record(value, contract);
  exactKeys(source, ["reassignedCount", "assignments"], contract);
  if (!Array.isArray(source.assignments)) {
    throw new ContractDecodeError(contract, "$.assignments");
  }
  if (source.assignments.length === 0) {
    throw new ContractDecodeError(contract, "$.assignments");
  }
  const assignments = source.assignments.map((item, index) => {
    const assignment = record(item, contract, `$.assignments[${index}]`);
    exactKeys(
      assignment,
      [
        "studentId",
        "previousTeacherAccountId",
        "newTeacherAccountId",
        "effectiveFrom",
        "version",
      ],
      contract,
      `$.assignments[${index}]`,
    );
    return {
      studentId: identifierField(assignment, "studentId", contract)!,
      previousTeacherAccountId: identifierField(
        assignment,
        "previousTeacherAccountId",
        contract,
      )!,
      newTeacherAccountId: identifierField(
        assignment,
        "newTeacherAccountId",
        contract,
      )!,
      effectiveFrom: isoDateField(assignment, "effectiveFrom", contract)!,
      version: numberField(assignment, "version", contract, 1),
    };
  });
  const reassignedCount = numberField(source, "reassignedCount", contract, 1);
  if (reassignedCount !== assignments.length) {
    throw new ContractDecodeError(contract, "$.reassignedCount");
  }
  return { reassignedCount, assignments };
};

export const decodeReplaceLessonTeachersResult: Decoder<
  ReplaceLessonTeachersResult
> = (value) => {
  const contract = "ReplaceLessonTeachersResult";
  const source = record(value, contract);
  exactKeys(source, ["updatedCount", "lessons"], contract);
  const lessons = decodeLessons(source.lessons);
  if (lessons.length === 0) {
    throw new ContractDecodeError(contract, "$.lessons");
  }
  const updatedCount = numberField(source, "updatedCount", contract, 1);
  if (updatedCount !== lessons.length) {
    throw new ContractDecodeError(contract, "$.updatedCount");
  }
  return { updatedCount, lessons };
};

export const decodeBootstrapView: Decoder<BootstrapView> = (value) => {
  const contract = "BootstrapView";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "accountId",
      "roles",
      "accessProfiles",
      "permissions",
      "studentId",
      "fullName",
      "firstMinute",
    ],
    contract,
  );
  const roles = rolesField(source, contract);
  const accessProfiles = accessProfilesField(source, contract);
  if (!Array.isArray(source.permissions)) {
    throw new ContractDecodeError(contract, "$.permissions");
  }
  const permissions = unique(
    source.permissions.map((permission, index) =>
      oneOf(permission, PERMISSIONS, contract, `$.permissions[${index}]`),
    ),
    contract,
    "$.permissions",
  );
  const isOwner = roles.includes("Owner");
  const hasOnboardingProfile = accessProfiles.includes(
    STUDENT_ONBOARDING_MANAGER_BUNDLE,
  );
  const samePermissionSet = (
    expected: readonly Permission[],
  ): boolean =>
    permissions.length === expected.length &&
    expected.every((permission) => permissions.includes(permission));
  const expectedPermissions = new Set<Permission>();
  const addExpected = (values: readonly Permission[]) => {
    for (const permission of values) expectedPermissions.add(permission);
  };
  if (isOwner) {
    addExpected(OWNER_STUDENT_ONBOARDING_PERMISSIONS);
    addExpected(LESSON_MANAGER_PERMISSIONS);
  } else {
    if (roles.includes("Administrator")) addExpected(LESSON_MANAGER_PERMISSIONS);
    if (roles.includes("Teacher")) addExpected(LESSON_CREATOR_PERMISSIONS);
    if (roles.includes("Student")) addExpected(LESSON_READER_PERMISSIONS);
    if (hasOnboardingProfile) addExpected(DELEGATED_STUDENT_ONBOARDING_PERMISSIONS);
  }
  if (
    (isOwner && hasOnboardingProfile) ||
    (hasOnboardingProfile && !roles.includes("Administrator")) ||
    !samePermissionSet([...expectedPermissions])
  ) {
    throw new ContractDecodeError(contract, "$.permissions");
  }
  const studentId = stringField(source, "studentId", contract, true);
  const fullName = stringField(source, "fullName", contract, true);
  const firstMinute =
    source.firstMinute === undefined
      ? undefined
      : decodeFirstMinute(source.firstMinute);
  const result: BootstrapAccessView = {
    accountId: stringField(source, "accountId", contract)!,
    roles,
    accessProfiles,
    permissions,
  };
  const hasStudentRole = roles.includes("Student");
  const hasAnyStudentIdentity =
    studentId !== undefined ||
    fullName !== undefined ||
    firstMinute !== undefined;
  if (!hasStudentRole && hasAnyStudentIdentity) {
    throw new ContractDecodeError(contract, "$.studentId");
  }
  if (hasStudentRole) {
    if (studentId === undefined) {
      throw new ContractDecodeError(contract, "$.studentId");
    }
    if (fullName === undefined) {
      throw new ContractDecodeError(contract, "$.fullName");
    }
    if (firstMinute === undefined) {
      throw new ContractDecodeError(contract, "$.firstMinute");
    }
    if (firstMinute.studentId !== studentId) {
      throw new ContractDecodeError(contract, "$.firstMinute.studentId");
    }
    return { ...result, studentId, fullName, firstMinute };
  }
  return result;
};

export const decodeStaffMembers: Decoder<StaffMember[]> = (value) => {
  const contract = "StaffMember[]";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  return value.map((item, index) => {
    const source = record(item, contract, `$[${index}]`);
    exactKeys(
      source,
      [
        "accountId",
        "fullName",
        "roles",
        "accessProfiles",
        "onboardingDelegationId",
        "onboardingDelegationExpiresAt",
      ],
      contract,
      `$[${index}]`,
    );
    const onboardingDelegationId = identifierField(
      source,
      "onboardingDelegationId",
      contract,
      true,
    );
    const onboardingDelegationExpiresAt = isoDateField(
      source,
      "onboardingDelegationExpiresAt",
      contract,
      true,
    );
    const result: StaffMember = {
      accountId: stringField(source, "accountId", contract)!,
      fullName: stringField(source, "fullName", contract)!,
      roles: rolesField(source, contract),
      accessProfiles: accessProfilesField(source, contract),
    };
    const hasOnboardingProfile = result.accessProfiles.includes(
      STUDENT_ONBOARDING_MANAGER_BUNDLE,
    );
    if (
      hasOnboardingProfile !== (onboardingDelegationId !== undefined) ||
      (hasOnboardingProfile && !result.roles.includes("Administrator")) ||
      (onboardingDelegationExpiresAt !== undefined &&
        onboardingDelegationId === undefined)
    ) {
      throw new ContractDecodeError(
        contract,
        "$.onboardingDelegationExpiresAt",
      );
    }
    if (onboardingDelegationId !== undefined) {
      result.onboardingDelegationId = onboardingDelegationId;
    }
    if (onboardingDelegationExpiresAt !== undefined) {
      result.onboardingDelegationExpiresAt = onboardingDelegationExpiresAt;
    }
    return result;
  });
};

export const decodeStudentOnboardingItems: Decoder<
  StudentOnboardingItem[]
> = (value) => {
  const contract = "StudentOnboardingItem[]";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  return value.map((item, index) => {
    const source = record(item, contract, `$[${index}]`);
    exactKeys(
      source,
      [
        "studentId",
        "fullName",
        "enrollmentReference",
        "teacherAccountId",
        "studentVersion",
        "onboardingState",
        "invitationId",
        "invitationExpiresAt",
      ],
      contract,
      `$[${index}]`,
    );
    const invitationId = stringField(
      source,
      "invitationId",
      contract,
      true,
    );
    const invitationExpiresAt = isoDateField(
      source,
      "invitationExpiresAt",
      contract,
      true,
    );
    const result: StudentOnboardingItem = {
      studentId: stringField(source, "studentId", contract)!,
      fullName: stringField(source, "fullName", contract)!,
      enrollmentReference: stringField(
        source,
        "enrollmentReference",
        contract,
      )!,
      teacherAccountId: stringField(source, "teacherAccountId", contract)!,
      studentVersion: numberField(source, "studentVersion", contract),
      onboardingState: oneOf(
        source.onboardingState,
        STUDENT_ONBOARDING_STATES,
        contract,
        "$.onboardingState",
      ),
    };
    const isInvited = result.onboardingState === "invited";
    if (
      (invitationId !== undefined) !==
        (invitationExpiresAt !== undefined) ||
      isInvited !== (invitationId !== undefined)
    ) {
      throw new ContractDecodeError(contract, "$.invitationId");
    }
    if (invitationId !== undefined) {
      result.invitationId = invitationId;
    }
    if (invitationExpiresAt !== undefined) {
      result.invitationExpiresAt = invitationExpiresAt;
    }
    return result;
  });
};

export const decodeDelegationResult: Decoder<DelegationResult> = (value) => {
  const contract = "DelegationResult";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "id",
      "administratorAccountId",
      "bundle",
      "status",
      "grantedAt",
      "expiresAt",
    ],
    contract,
  );
  const result: DelegationResult = {
    id: stringField(source, "id", contract)!,
    administratorAccountId: stringField(
      source,
      "administratorAccountId",
      contract,
    )!,
    bundle: oneOf(
      source.bundle,
      [STUDENT_ONBOARDING_MANAGER_BUNDLE] as const,
      contract,
      "$.bundle",
    ),
    status: oneOf(
      source.status,
      ["active"] as const,
      contract,
      "$.status",
    ),
    grantedAt: isoDateField(source, "grantedAt", contract)!,
  };
  const expiresAt = isoDateField(source, "expiresAt", contract, true);
  if (expiresAt !== undefined) {
    result.expiresAt = expiresAt;
  }
  return result;
};

export const decodeStudentResult: Decoder<StudentResult> = (value) => {
  const contract = "StudentResult";
  const source = record(value, contract);
  exactKeys(
    source,
    ["studentId", "accountId", "onboardingState"],
    contract,
  );
  return {
    studentId: stringField(source, "studentId", contract)!,
    accountId: stringField(source, "accountId", contract)!,
    onboardingState: oneOf(
      source.onboardingState,
      ["awaiting_first_minute"] as const,
      contract,
      "$.onboardingState",
    ),
  };
};

export function decodeInvitationResult(
  value: unknown,
  policy: ActivationLinkPolicy = {},
): InvitationResult {
  const contract = "InvitationResult";
  const source = record(value, contract);
  exactKeys(
    source,
    ["invitationId", "studentId", "status", "expiresAt", "activationLink"],
    contract,
  );
  const activationLink = stringField(source, "activationLink", contract)!;
  if (!parseCanonicalActivationLink(activationLink, policy).ok) {
    throw new ContractDecodeError(contract, "$.activationLink");
  }
  return {
    invitationId: stringField(source, "invitationId", contract)!,
    studentId: stringField(source, "studentId", contract)!,
    status: oneOf(source.status, ["issued"] as const, contract, "$.status"),
    expiresAt: isoDateField(source, "expiresAt", contract)!,
    activationLink,
  };
}

export const decodeApiErrorEnvelope: Decoder<ApiErrorEnvelope> = (value) => {
  const contract = "ApiErrorEnvelope";
  const source = record(value, contract);
  exactKeys(source, ["error"], contract);
  const error = record(source.error, contract, "$.error");
  exactKeys(error, ["code", "message", "requestId"], contract, "$.error");
  return {
    error: {
      code: oneOf(error.code, API_ERROR_CODES, contract, "$.error.code"),
      message: stringField(error, "message", contract)!,
      requestId: stringField(error, "requestId", contract)!,
    },
  };
};

export const decodeVoid: Decoder<void> = (value) => {
  if (value !== undefined) {
    throw new ContractDecodeError("NoContent", "$");
  }
};

// ---- P.1 session security (Page 32: ACC-05/08/09, AUTH-07/08) ----

export interface RequestPasswordResetRequest {
  phone: string;
}

export interface CompletePasswordResetRequest {
  token: string;
  newPassword: string;
}

export interface RevokeSessionRequest {
  currentPassword: string;
}

export interface SessionDevice {
  sessionId: string;
  deviceLabel?: string;
  platform?: SessionPlatform;
  createdAt: IsoDateTime;
  lastSeenAt?: IsoDateTime;
  current: boolean;
}

export interface RevokeOtherSessionsResult {
  revokedCount: number;
}

export const SECURITY_EVENT_ACTIONS = [
  "SessionCreated",
  "SessionRefreshed",
  "SessionRevoked",
  "RefreshTokenReuseDetected",
  "AccountActivated",
  "PasswordResetRequested",
  "PasswordResetCompleted",
  "OtherSessionsRevoked",
  "ContactChangeStarted",
  "ContactVerified",
  "TwofaEnrolled",
  "TwofaDisabled",
  "TwofaChallengeFailed",
  "ProfileUpdated",
  "PolicyAccepted",
  "PrivacySettingsUpdated",
  "DataExportRequested",
  "DeletionRequested",
  "DeletionRequestCancelled",
] as const;

export type SecurityEventAction = (typeof SECURITY_EVENT_ACTIONS)[number];

export interface SecurityEvent {
  id: number;
  action: SecurityEventAction;
  decision: "allow" | "deny";
  reasonCode?: string;
  targetType?: string;
  targetId?: string;
  recordedAt: IsoDateTime;
}

export interface SecurityEventsPage {
  events: SecurityEvent[];
  nextCursor?: string;
}

export interface SecurityEventsQuery {
  cursor?: string;
  limit?: number;
}

function booleanField(
  value: JsonRecord,
  key: string,
  contract: string,
): boolean {
  const field = value[key];
  if (typeof field !== "boolean") {
    throw new ContractDecodeError(contract, `$.${key}`);
  }
  return field;
}

function decodeSessionDevice(
  value: unknown,
  contract: string,
  path: string,
): SessionDevice {
  const source = record(value, contract, path);
  exactKeys(
    source,
    ["sessionId", "deviceLabel", "platform", "createdAt", "lastSeenAt", "current"],
    contract,
    path,
  );
  const deviceLabel = stringField(source, "deviceLabel", contract, true);
  if (deviceLabel !== undefined && deviceLabel.length > 120) {
    throw new ContractDecodeError(contract, `${path}.deviceLabel`);
  }
  const platform =
    source.platform === undefined
      ? undefined
      : oneOf(source.platform, SESSION_PLATFORMS, contract, `${path}.platform`);
  const device: SessionDevice = {
    sessionId: identifierField(source, "sessionId", contract)!,
    createdAt: isoDateField(source, "createdAt", contract)!,
    current: booleanField(source, "current", contract),
  };
  if (deviceLabel !== undefined) {
    device.deviceLabel = deviceLabel;
  }
  if (platform !== undefined) {
    device.platform = platform;
  }
  const lastSeenAt = isoDateField(source, "lastSeenAt", contract, true);
  if (lastSeenAt !== undefined) {
    device.lastSeenAt = lastSeenAt;
  }
  return device;
}

export const decodeSessionDevices: Decoder<SessionDevice[]> = (value) => {
  const contract = "SessionDeviceList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const devices = value.map((entry, index) =>
    decodeSessionDevice(entry, contract, `$[${index}]`),
  );
  unique(
    devices.map((device) => device.sessionId),
    contract,
    "$[].sessionId",
  );
  if (devices.filter((device) => device.current).length > 1) {
    throw new ContractDecodeError(contract, "$[].current");
  }
  return devices;
};

export const decodeRevokeOtherSessionsResult: Decoder<
  RevokeOtherSessionsResult
> = (value) => {
  const contract = "RevokeOtherSessionsResult";
  const source = record(value, contract);
  exactKeys(source, ["revokedCount"], contract);
  return { revokedCount: numberField(source, "revokedCount", contract) };
};

export const decodeSecurityEventsPage: Decoder<SecurityEventsPage> = (
  value,
) => {
  const contract = "SecurityEventsPage";
  const source = record(value, contract);
  exactKeys(source, ["events", "nextCursor"], contract);
  if (!Array.isArray(source.events)) {
    throw new ContractDecodeError(contract, "$.events");
  }
  const events = source.events.map((entry, index) => {
    const path = `$.events[${index}]`;
    const eventSource = record(entry, contract, path);
    exactKeys(
      eventSource,
      ["id", "action", "decision", "reasonCode", "targetType", "targetId", "recordedAt"],
      contract,
      path,
    );
    const event: SecurityEvent = {
      id: numberField(eventSource, "id", contract, 1),
      action: oneOf(
        eventSource.action,
        SECURITY_EVENT_ACTIONS,
        contract,
        `${path}.action`,
      ),
      decision: oneOf(
        eventSource.decision,
        ["allow", "deny"] as const,
        contract,
        `${path}.decision`,
      ),
      recordedAt: isoDateField(eventSource, "recordedAt", contract)!,
    };
    const reasonCode = stringField(eventSource, "reasonCode", contract, true);
    if (reasonCode !== undefined) {
      event.reasonCode = reasonCode;
    }
    const targetType = stringField(eventSource, "targetType", contract, true);
    if (targetType !== undefined) {
      event.targetType = targetType;
    }
    const targetId = identifierField(eventSource, "targetId", contract, true);
    if (targetId !== undefined) {
      event.targetId = targetId;
    }
    return event;
  });
  for (let index = 1; index < events.length; index += 1) {
    if (events[index]!.id >= events[index - 1]!.id) {
      throw new ContractDecodeError(contract, `$.events[${index}].id`);
    }
  }
  const page: SecurityEventsPage = { events };
  const nextCursor = stringField(source, "nextCursor", contract, true);
  if (nextCursor !== undefined) {
    if (nextCursor.length > 64) {
      throw new ContractDecodeError(contract, "$.nextCursor");
    }
    page.nextCursor = nextCursor;
  }
  return page;
};


// ---- P.1 contacts, 2FA and multi-step activation (AUTH-01..10, ACC-03/06) ----

export type SignInOutcome =
  | { tokens: SessionTokens; twofaChallenge?: never; twofaExpiresAt?: never }
  | { tokens?: never; twofaChallenge: string; twofaExpiresAt: IsoDateTime };

export interface TwofaSignInRequest {
  challenge: string;
  code: string;
}

export type ContactKind = "email" | "phone";

export const CONTACT_KINDS = ["email", "phone"] as const;

export interface ActivationTokenRequest {
  token: string;
}

export interface ActivationPasswordRequest {
  token: string;
  phone: string;
  password: string;
}

export interface ActivationContactRequest {
  token: string;
  kind: ContactKind;
  value: string;
}

export interface ActivationCodeRequest {
  token: string;
  code: string;
}

export interface ActivationFinishRequest {
  token: string;
  phone: string;
}

export interface ActivationProgressView {
  invitationId: string;
  kind: ActivationPreview["kind"];
  displayName: string;
  expiresAt: IsoDateTime;
  passwordSet: boolean;
  contactKind?: ContactKind;
  contactMasked?: string;
  contactVerified: boolean;
  twofaEnrolled: boolean;
  completed: boolean;
}

export interface VerifiedContact {
  id: string;
  kind: ContactKind;
  value: string;
  verifiedAt: IsoDateTime;
}

export interface StartContactChangeRequest {
  kind: ContactKind;
  value: string;
  currentPassword: string;
}

export interface ConfirmContactChangeRequest {
  code: string;
}

export interface TwofaStatus {
  enabled: boolean;
  confirmedAt?: IsoDateTime;
  recoveryCodesRemaining: number;
}

export interface TwofaEnrollment {
  secret: string;
  provisioningUri: string;
}

export interface TwofaCodeRequest {
  code: string;
}

export interface CurrentPasswordRequest {
  currentPassword: string;
}

export interface DisableTwofaRequest {
  currentPassword: string;
  code: string;
}

export interface RecoveryCodesResponse {
  recoveryCodes: string[];
}

export const decodeSignInOutcome: Decoder<SignInOutcome> = (value) => {
  const contract = "SignInOutcome";
  const source = record(value, contract);
  const hasTokens = source.tokens !== undefined;
  const hasChallenge = source.twofaChallenge !== undefined;
  if (hasTokens === hasChallenge) {
    throw new ContractDecodeError(contract, "$");
  }
  if (hasTokens) {
    exactKeys(source, ["tokens"], contract);
    return { tokens: decodeSessionTokens(source.tokens) };
  }
  exactKeys(source, ["twofaChallenge", "twofaExpiresAt"], contract);
  return {
    twofaChallenge: opaqueTokenField(source, "twofaChallenge", contract),
    twofaExpiresAt: isoDateField(source, "twofaExpiresAt", contract)!,
  };
};

function contactKindField(
  value: unknown,
  contract: string,
  path: string,
): ContactKind {
  return oneOf(value, CONTACT_KINDS, contract, path);
}

export const decodeActivationProgress: Decoder<ActivationProgressView> = (
  value,
) => {
  const contract = "ActivationProgressView";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "invitationId",
      "kind",
      "displayName",
      "expiresAt",
      "passwordSet",
      "contactKind",
      "contactMasked",
      "contactVerified",
      "twofaEnrolled",
      "completed",
    ],
    contract,
  );
  const view: ActivationProgressView = {
    invitationId: identifierField(source, "invitationId", contract)!,
    kind: oneOf(
      source.kind,
      ["owner_bootstrap", "staff_activation", "student_activation"] as const,
      contract,
      "$.kind",
    ),
    displayName: stringField(source, "displayName", contract)!,
    expiresAt: isoDateField(source, "expiresAt", contract)!,
    passwordSet: booleanField(source, "passwordSet", contract),
    contactVerified: booleanField(source, "contactVerified", contract),
    twofaEnrolled: booleanField(source, "twofaEnrolled", contract),
    completed: booleanField(source, "completed", contract),
  };
  if (source.contactKind !== undefined) {
    view.contactKind = contactKindField(source.contactKind, contract, "$.contactKind");
  }
  const contactMasked = stringField(source, "contactMasked", contract, true);
  if (contactMasked !== undefined) {
    view.contactMasked = contactMasked;
  }
  if (view.contactVerified && view.contactKind === undefined) {
    throw new ContractDecodeError(contract, "$.contactKind");
  }
  return view;
};

function decodeVerifiedContactEntry(
  value: unknown,
  contract: string,
  path: string,
): VerifiedContact {
  const source = record(value, contract, path);
  exactKeys(source, ["id", "kind", "value", "verifiedAt"], contract, path);
  return {
    id: identifierField(source, "id", contract)!,
    kind: contactKindField(source.kind, contract, `${path}.kind`),
    value: stringField(source, "value", contract)!,
    verifiedAt: isoDateField(source, "verifiedAt", contract)!,
  };
}

export const decodeVerifiedContact: Decoder<VerifiedContact> = (value) =>
  decodeVerifiedContactEntry(value, "VerifiedContact", "$");

export const decodeVerifiedContacts: Decoder<VerifiedContact[]> = (value) => {
  const contract = "VerifiedContactList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const contacts = value.map((entry, index) =>
    decodeVerifiedContactEntry(entry, contract, `$[${index}]`),
  );
  unique(
    contacts.map((contact) => contact.kind),
    contract,
    "$[].kind",
  );
  return contacts;
};

export const decodeTwofaStatus: Decoder<TwofaStatus> = (value) => {
  const contract = "TwofaStatus";
  const source = record(value, contract);
  exactKeys(source, ["enabled", "confirmedAt", "recoveryCodesRemaining"], contract);
  const status: TwofaStatus = {
    enabled: booleanField(source, "enabled", contract),
    recoveryCodesRemaining: numberField(source, "recoveryCodesRemaining", contract),
  };
  const confirmedAt = isoDateField(source, "confirmedAt", contract, true);
  if (confirmedAt !== undefined) {
    status.confirmedAt = confirmedAt;
  }
  if (status.enabled && status.confirmedAt === undefined) {
    throw new ContractDecodeError(contract, "$.confirmedAt");
  }
  return status;
};

export const decodeTwofaEnrollment: Decoder<TwofaEnrollment> = (value) => {
  const contract = "TwofaEnrollment";
  const source = record(value, contract);
  exactKeys(source, ["secret", "provisioningUri"], contract);
  const enrollment = {
    secret: stringField(source, "secret", contract)!,
    provisioningUri: stringField(source, "provisioningUri", contract)!,
  };
  if (!enrollment.provisioningUri.startsWith("otpauth://totp/")) {
    throw new ContractDecodeError(contract, "$.provisioningUri");
  }
  return enrollment;
};

const RECOVERY_CODE_PATTERN = /^[A-Z2-9]{4}-[A-Z2-9]{4}-[A-Z2-9]{2}$/;

export const decodeRecoveryCodes: Decoder<string[]> = (value) => {
  const contract = "RecoveryCodesResponse";
  const source = record(value, contract);
  exactKeys(source, ["recoveryCodes"], contract);
  if (!Array.isArray(source.recoveryCodes) || source.recoveryCodes.length !== 10) {
    throw new ContractDecodeError(contract, "$.recoveryCodes");
  }
  const codes = source.recoveryCodes.map((entry, index) => {
    if (typeof entry !== "string" || !RECOVERY_CODE_PATTERN.test(entry)) {
      throw new ContractDecodeError(contract, `$.recoveryCodes[${index}]`);
    }
    return entry;
  });
  unique(codes, contract, "$.recoveryCodes");
  return codes;
};

// ---- P.1 account profile (ACC-01/02) ----

export interface ProfileView {
  accountId: string;
  fullName: string;
  tenantName: string;
  roles: Role[];
  phone: string;
}

export interface UpdateProfileRequest {
  fullName: string;
}

export const decodeProfileView: Decoder<ProfileView> = (value) => {
  const contract = "ProfileView";
  const source = record(value, contract);
  exactKeys(source, ["accountId", "fullName", "tenantName", "roles", "phone"], contract);
  if (!Array.isArray(source.roles)) {
    throw new ContractDecodeError(contract, "$.roles");
  }
  const roles = source.roles.map((entry, index) =>
    oneOf(entry, ROLES, contract, `$.roles[${index}]`),
  );
  unique(roles, contract, "$.roles");
  const phone = stringField(source, "phone", contract, true) ?? "";
  return {
    accountId: identifierField(source, "accountId", contract)!,
    fullName: stringField(source, "fullName", contract)!,
    tenantName: stringField(source, "tenantName", contract)!,
    roles,
    phone,
  };
};

// ---- P.1 policies, privacy and data rights (ACC-10..12, ACC-14..18) ----

export const POLICY_KINDS = [
  "privacy",
  "terms",
  "community",
  "media_consent",
] as const;

export type PolicyKind = (typeof POLICY_KINDS)[number];

export interface PolicyVersion {
  id: string;
  kind: PolicyKind;
  version: string;
  title: string;
  bodyRef: string;
  effectiveFrom: IsoDateTime;
  acceptedAt?: IsoDateTime;
}

export interface AcceptPolicyRequest {
  policyVersionId: string;
}

export const PUSH_PREVIEW_MODES = ["hidden", "title", "full"] as const;

export type PushPreviewMode = (typeof PUSH_PREVIEW_MODES)[number];

export interface PrivacySettings {
  communityProfileVisible: boolean;
  achievementsVisible: boolean;
  staffMessagesAllowed: boolean;
  mentionsAllowed: boolean;
  pushPreview: PushPreviewMode;
  version: number;
}

export const DATA_EXPORT_STATUSES = [
  "requested",
  "processing",
  "ready",
  "expired",
  "cancelled",
] as const;

export type DataExportStatus = (typeof DATA_EXPORT_STATUSES)[number];

export interface DataExportRequest {
  id: string;
  status: DataExportStatus;
  requestedAt: IsoDateTime;
  readyAt?: IsoDateTime;
  expiresAt?: IsoDateTime;
}

export const DELETION_REQUEST_STATUSES = [
  "requested",
  "pending_review",
  "cancelled",
] as const;

export type DeletionRequestStatus = (typeof DELETION_REQUEST_STATUSES)[number];

export interface DeletionRequest {
  id: string;
  status: DeletionRequestStatus;
  requestedAt: IsoDateTime;
  cancelledAt?: IsoDateTime;
}

function decodePolicyVersionEntry(
  value: unknown,
  contract: string,
  path: string,
): PolicyVersion {
  const source = record(value, contract, path);
  exactKeys(
    source,
    ["id", "kind", "version", "title", "bodyRef", "effectiveFrom", "acceptedAt"],
    contract,
    path,
  );
  const policy: PolicyVersion = {
    id: identifierField(source, "id", contract)!,
    kind: oneOf(source.kind, POLICY_KINDS, contract, `${path}.kind`),
    version: stringField(source, "version", contract)!,
    title: stringField(source, "title", contract)!,
    bodyRef: stringField(source, "bodyRef", contract)!,
    effectiveFrom: isoDateField(source, "effectiveFrom", contract)!,
  };
  const acceptedAt = isoDateField(source, "acceptedAt", contract, true);
  if (acceptedAt !== undefined) {
    policy.acceptedAt = acceptedAt;
  }
  return policy;
}

export const decodePolicyVersions: Decoder<PolicyVersion[]> = (value) => {
  const contract = "PolicyVersionList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const policies = value.map((entry, index) =>
    decodePolicyVersionEntry(entry, contract, `$[${index}]`),
  );
  unique(
    policies.map((policy) => policy.id),
    contract,
    "$[].id",
  );
  return policies;
};

export const decodePrivacySettings: Decoder<PrivacySettings> = (value) => {
  const contract = "PrivacySettings";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "communityProfileVisible",
      "achievementsVisible",
      "staffMessagesAllowed",
      "mentionsAllowed",
      "pushPreview",
      "version",
    ],
    contract,
  );
  return {
    communityProfileVisible: booleanField(source, "communityProfileVisible", contract),
    achievementsVisible: booleanField(source, "achievementsVisible", contract),
    staffMessagesAllowed: booleanField(source, "staffMessagesAllowed", contract),
    mentionsAllowed: booleanField(source, "mentionsAllowed", contract),
    pushPreview: oneOf(source.pushPreview, PUSH_PREVIEW_MODES, contract, "$.pushPreview"),
    version: numberField(source, "version", contract, 0),
  };
};

function decodeDataExportEntry(
  value: unknown,
  contract: string,
  path: string,
): DataExportRequest {
  const source = record(value, contract, path);
  exactKeys(
    source,
    ["id", "status", "requestedAt", "readyAt", "expiresAt"],
    contract,
    path,
  );
  const request: DataExportRequest = {
    id: identifierField(source, "id", contract)!,
    status: oneOf(source.status, DATA_EXPORT_STATUSES, contract, `${path}.status`),
    requestedAt: isoDateField(source, "requestedAt", contract)!,
  };
  const readyAt = isoDateField(source, "readyAt", contract, true);
  if (readyAt !== undefined) {
    request.readyAt = readyAt;
  }
  const expiresAt = isoDateField(source, "expiresAt", contract, true);
  if (expiresAt !== undefined) {
    request.expiresAt = expiresAt;
  }
  const requiresReadyAt = request.status === "ready" || request.status === "expired";
  if (requiresReadyAt !== (request.readyAt !== undefined)) {
    throw new ContractDecodeError(contract, `${path}.readyAt`);
  }
  return request;
}

export const decodeDataExport: Decoder<DataExportRequest> = (value) =>
  decodeDataExportEntry(value, "DataExportRequest", "$");

export const decodeDataExports: Decoder<DataExportRequest[]> = (value) => {
  const contract = "DataExportRequestList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  if (value.length > 10) {
    throw new ContractDecodeError(contract, "$");
  }
  const exports = value.map((entry, index) =>
    decodeDataExportEntry(entry, contract, `$[${index}]`),
  );
  unique(
    exports.map((entry) => entry.id),
    contract,
    "$[].id",
  );
  const open = exports.filter(
    (entry) => entry.status === "requested" || entry.status === "processing",
  );
  if (open.length > 1) {
    throw new ContractDecodeError(contract, "$[].status");
  }
  return exports;
};

export const decodeDeletionRequest: Decoder<DeletionRequest> = (value) => {
  const contract = "DeletionRequest";
  const source = record(value, contract);
  exactKeys(source, ["id", "status", "requestedAt", "cancelledAt"], contract);
  const request: DeletionRequest = {
    id: identifierField(source, "id", contract)!,
    status: oneOf(source.status, DELETION_REQUEST_STATUSES, contract, "$.status"),
    requestedAt: isoDateField(source, "requestedAt", contract)!,
  };
  const cancelledAt = isoDateField(source, "cancelledAt", contract, true);
  if (cancelledAt !== undefined) {
    request.cancelledAt = cancelledAt;
  }
  if ((request.status === "cancelled") !== (request.cancelledAt !== undefined)) {
    throw new ContractDecodeError(contract, "$.cancelledAt");
  }
  return request;
};

// ---- L.2 rooms and core lesson series (Pages 24/26/29; DEC-002/004) ----

export interface Room {
  id: string;
  name: string;
  capacity?: number;
  status: "active" | "archived";
  version: number;
}

export interface CreateRoomRequest {
  name: string;
  capacity?: number;
}

export const LESSON_FORMATS = ["individual", "group"] as const;

export type LessonFormat = (typeof LESSON_FORMATS)[number];

export interface CoreLessonSeries {
  id: string;
  format: LessonFormat;
  title: string;
  teacher: LessonTeacher;
  roomId?: string;
  weekday: number;
  startMinutes: number;
  durationMinutes: number;
  effectiveFrom: string;
  effectiveUntil?: string;
  status: "active" | "paused" | "ended";
  version: number;
  students: LessonStudent[];
}

export interface CreateLessonSeriesRequest {
  format: LessonFormat;
  title: string;
  teacherAccountId: string;
  roomId?: string;
  weekday: number;
  startMinutes: number;
  durationMinutes: number;
  effectiveFrom: string;
  effectiveUntil?: string;
  studentIds: string[];
}

export interface GenerateOccurrencesRequest {
  weeks: number;
}

export interface ChangeSeriesStatusRequest {
  status: "active" | "paused" | "ended";
  expectedVersion: number;
}

export interface SeriesGenerationResult {
  seriesId: string;
  createdCount: number;
  occurrenceIds: string[];
}

export const decodeRoom: Decoder<Room> = (value) => {
  const contract = "Room";
  const source = record(value, contract);
  exactKeys(source, ["id", "name", "capacity", "status", "version"], contract);
  const roomView: Room = {
    id: identifierField(source, "id", contract)!,
    name: stringField(source, "name", contract)!,
    status: oneOf(source.status, ["active", "archived"] as const, contract, "$.status"),
    version: numberField(source, "version", contract, 0),
  };
  if (source.capacity !== undefined) {
    const capacity = numberField(source, "capacity", contract, 1);
    roomView.capacity = capacity;
  }
  return roomView;
};

export const decodeRooms: Decoder<Room[]> = (value) => {
  if (!Array.isArray(value)) {
    throw new ContractDecodeError("RoomList", "$");
  }
  const rooms = value.map((entry) => decodeRoom(entry));
  unique(
    rooms.map((roomView) => roomView.id),
    "RoomList",
    "$[].id",
  );
  return rooms;
};

function decodeSeriesEntry(
  value: unknown,
  contract: string,
  path: string,
): CoreLessonSeries {
  const source = record(value, contract, path);
  exactKeys(
    source,
    [
      "id",
      "format",
      "title",
      "teacher",
      "roomId",
      "weekday",
      "startMinutes",
      "durationMinutes",
      "effectiveFrom",
      "effectiveUntil",
      "status",
      "version",
      "students",
    ],
    contract,
    path,
  );
  const teacherSource = record(source.teacher, contract, `${path}.teacher`);
  exactKeys(teacherSource, ["accountId", "fullName"], contract, `${path}.teacher`);
  if (!Array.isArray(source.students)) {
    throw new ContractDecodeError(contract, `${path}.students`);
  }
  const format = oneOf(source.format, LESSON_FORMATS, contract, `${path}.format`);
  const students = source.students.map((entry, index) => {
    const studentSource = record(entry, contract, `${path}.students[${index}]`);
    exactKeys(studentSource, ["studentId", "fullName"], contract, `${path}.students[${index}]`);
    return {
      studentId: identifierField(studentSource, "studentId", contract)!,
      fullName: stringField(studentSource, "fullName", contract)!,
    };
  });
  unique(
    students.map((student) => student.studentId),
    contract,
    `${path}.students[].studentId`,
  );
  if (format === "individual" && students.length !== 1) {
    throw new ContractDecodeError(contract, `${path}.students`);
  }
  if (format === "group" && (students.length < 1 || students.length > 3)) {
    throw new ContractDecodeError(contract, `${path}.students`);
  }
  const weekday = numberField(source, "weekday", contract, 0);
  if (weekday > 6) {
    throw new ContractDecodeError(contract, `${path}.weekday`);
  }
  const series: CoreLessonSeries = {
    id: identifierField(source, "id", contract)!,
    format,
    title: stringField(source, "title", contract)!,
    teacher: {
      accountId: identifierField(teacherSource, "accountId", contract)!,
      fullName: stringField(teacherSource, "fullName", contract)!,
    },
    weekday,
    startMinutes: numberField(source, "startMinutes", contract, 0),
    durationMinutes: numberField(source, "durationMinutes", contract, 1),
    effectiveFrom: stringField(source, "effectiveFrom", contract)!,
    status: oneOf(source.status, ["active", "paused", "ended"] as const, contract, `${path}.status`),
    version: numberField(source, "version", contract, 0),
    students,
  };
  const roomId = identifierField(source, "roomId", contract, true);
  if (roomId !== undefined) {
    series.roomId = roomId;
  }
  const effectiveUntil = stringField(source, "effectiveUntil", contract, true);
  if (effectiveUntil !== undefined) {
    series.effectiveUntil = effectiveUntil;
  }
  return series;
}

export const decodeCoreLessonSeries: Decoder<CoreLessonSeries> = (value) =>
  decodeSeriesEntry(value, "CoreLessonSeries", "$");

export const decodeCoreLessonSeriesList: Decoder<CoreLessonSeries[]> = (value) => {
  const contract = "CoreLessonSeriesList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const seriesList = value.map((entry, index) =>
    decodeSeriesEntry(entry, contract, `$[${index}]`),
  );
  unique(
    seriesList.map((series) => series.id),
    contract,
    "$[].id",
  );
  return seriesList;
};

export const decodeSeriesGenerationResult: Decoder<SeriesGenerationResult> = (
  value,
) => {
  const contract = "SeriesGenerationResult";
  const source = record(value, contract);
  exactKeys(source, ["seriesId", "createdCount", "occurrenceIds"], contract);
  if (!Array.isArray(source.occurrenceIds)) {
    throw new ContractDecodeError(contract, "$.occurrenceIds");
  }
  const occurrenceIds = source.occurrenceIds.map((entry, index) => {
    if (typeof entry !== "string" || entry.length === 0) {
      throw new ContractDecodeError(contract, `$.occurrenceIds[${index}]`);
    }
    return entry;
  });
  unique(occurrenceIds, contract, "$.occurrenceIds");
  const createdCount = numberField(source, "createdCount", contract, 0);
  if (createdCount !== occurrenceIds.length) {
    throw new ContractDecodeError(contract, "$.createdCount");
  }
  return {
    seriesId: identifierField(source, "seriesId", contract)!,
    createdCount,
    occurrenceIds,
  };
};

// ---- L.2 events and RSVP (Page 24 catalog; DEC-001/003/101) ----

export interface EventCategory {
  id: string;
  name: string;
  status: "active" | "archived";
}

export interface CreateEventCategoryRequest {
  name: string;
}

export interface EventSeries {
  id: string;
  categoryId: string;
  title: string;
  description?: string;
  host: LessonTeacher;
  roomId?: string;
  capacity: number;
  weekday: number;
  startMinutes: number;
  durationMinutes: number;
  effectiveFrom: string;
  effectiveUntil?: string;
  status: "active" | "paused" | "ended";
  version: number;
}

export interface CreateEventSeriesRequest {
  categoryId: string;
  title: string;
  description?: string;
  hostAccountId: string;
  roomId?: string;
  capacity: number;
  weekday: number;
  startMinutes: number;
  durationMinutes: number;
  effectiveFrom: string;
  effectiveUntil?: string;
}

export interface EventSpotOffer {
  id: string;
  occurrenceId: string;
  status: "pending" | "confirmed" | "declined" | "expired";
  offeredAt: IsoDateTime;
  expiresAt: IsoDateTime;
}

export interface EventOccurrence {
  id: string;
  seriesId?: string;
  categoryId: string;
  categoryName: string;
  title: string;
  description?: string;
  startsAt: IsoDateTime;
  durationMinutes: number;
  host: LessonTeacher;
  roomId?: string;
  capacity: number;
  confirmedCount: number;
  status: "scheduled" | "completed" | "cancelled";
  version: number;
  myRsvp?: "confirmed" | "cancelled";
  myWaitlistPosition?: number;
  myOffer?: EventSpotOffer;
}

export interface CreateEventRequest {
  categoryId: string;
  title: string;
  description?: string;
  startsAt: IsoDateTime;
  durationMinutes: number;
  hostAccountId: string;
  roomId?: string;
  capacity: number;
}

export interface EventListWindow {
  from: IsoDateTime;
  to: IsoDateTime;
}

export const decodeEventCategory: Decoder<EventCategory> = (value) => {
  const contract = "EventCategory";
  const source = record(value, contract);
  exactKeys(source, ["id", "name", "status"], contract);
  return {
    id: identifierField(source, "id", contract)!,
    name: stringField(source, "name", contract)!,
    status: oneOf(source.status, ["active", "archived"] as const, contract, "$.status"),
  };
};

export const decodeEventCategories: Decoder<EventCategory[]> = (value) => {
  if (!Array.isArray(value)) {
    throw new ContractDecodeError("EventCategoryList", "$");
  }
  const categories = value.map((entry) => decodeEventCategory(entry));
  unique(
    categories.map((category) => category.id),
    "EventCategoryList",
    "$[].id",
  );
  return categories;
};

export const decodeEventSeries: Decoder<EventSeries> = (value) => {
  const contract = "EventSeries";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "id",
      "categoryId",
      "title",
      "description",
      "host",
      "roomId",
      "capacity",
      "weekday",
      "startMinutes",
      "durationMinutes",
      "effectiveFrom",
      "effectiveUntil",
      "status",
      "version",
    ],
    contract,
  );
  const hostSource = record(source.host, contract, "$.host");
  exactKeys(hostSource, ["accountId", "fullName"], contract, "$.host");
  const weekday = numberField(source, "weekday", contract, 0);
  if (weekday > 6) {
    throw new ContractDecodeError(contract, "$.weekday");
  }
  const series: EventSeries = {
    id: identifierField(source, "id", contract)!,
    categoryId: identifierField(source, "categoryId", contract)!,
    title: stringField(source, "title", contract)!,
    host: {
      accountId: identifierField(hostSource, "accountId", contract)!,
      fullName: stringField(hostSource, "fullName", contract)!,
    },
    capacity: numberField(source, "capacity", contract, 1),
    weekday,
    startMinutes: numberField(source, "startMinutes", contract, 0),
    durationMinutes: numberField(source, "durationMinutes", contract, 1),
    effectiveFrom: stringField(source, "effectiveFrom", contract)!,
    status: oneOf(source.status, ["active", "paused", "ended"] as const, contract, "$.status"),
    version: numberField(source, "version", contract, 0),
  };
  const description = stringField(source, "description", contract, true);
  if (description !== undefined) {
    series.description = description;
  }
  const roomId = identifierField(source, "roomId", contract, true);
  if (roomId !== undefined) {
    series.roomId = roomId;
  }
  const effectiveUntil = stringField(source, "effectiveUntil", contract, true);
  if (effectiveUntil !== undefined) {
    series.effectiveUntil = effectiveUntil;
  }
  return series;
};

function decodeEventOccurrenceEntry(
  value: unknown,
  contract: string,
  path: string,
): EventOccurrence {
  const source = record(value, contract, path);
  exactKeys(
    source,
    [
      "id",
      "seriesId",
      "categoryId",
      "categoryName",
      "title",
      "description",
      "startsAt",
      "durationMinutes",
      "host",
      "roomId",
      "capacity",
      "confirmedCount",
      "status",
      "version",
      "myRsvp",
      "myWaitlistPosition",
      "myOffer",
    ],
    contract,
    path,
  );
  const hostSource = record(source.host, contract, `${path}.host`);
  exactKeys(hostSource, ["accountId", "fullName"], contract, `${path}.host`);
  const occurrence: EventOccurrence = {
    id: identifierField(source, "id", contract)!,
    categoryId: identifierField(source, "categoryId", contract)!,
    categoryName: stringField(source, "categoryName", contract)!,
    title: stringField(source, "title", contract)!,
    startsAt: isoDateField(source, "startsAt", contract)!,
    durationMinutes: numberField(source, "durationMinutes", contract, 1),
    host: {
      accountId: identifierField(hostSource, "accountId", contract)!,
      fullName: stringField(hostSource, "fullName", contract)!,
    },
    capacity: numberField(source, "capacity", contract, 1),
    confirmedCount: numberField(source, "confirmedCount", contract, 0),
    status: oneOf(
      source.status,
      ["scheduled", "completed", "cancelled"] as const,
      contract,
      `${path}.status`,
    ),
    version: numberField(source, "version", contract, 0),
  };
  if (occurrence.confirmedCount > occurrence.capacity) {
    throw new ContractDecodeError(contract, `${path}.confirmedCount`);
  }
  const seriesId = identifierField(source, "seriesId", contract, true);
  if (seriesId !== undefined) {
    occurrence.seriesId = seriesId;
  }
  const description = stringField(source, "description", contract, true);
  if (description !== undefined) {
    occurrence.description = description;
  }
  const roomId = identifierField(source, "roomId", contract, true);
  if (roomId !== undefined) {
    occurrence.roomId = roomId;
  }
  if (source.myRsvp !== undefined) {
    occurrence.myRsvp = oneOf(
      source.myRsvp,
      ["confirmed", "cancelled"] as const,
      contract,
      `${path}.myRsvp`,
    );
  }
  if (source.myWaitlistPosition !== undefined) {
    const position = numberField(source, "myWaitlistPosition", contract, 1);
    occurrence.myWaitlistPosition = position;
  }
  if (source.myOffer !== undefined) {
    const offerSource = record(source.myOffer, contract, `${path}.myOffer`);
    exactKeys(
      offerSource,
      ["id", "occurrenceId", "status", "offeredAt", "expiresAt"],
      contract,
      `${path}.myOffer`,
    );
    const offer: EventSpotOffer = {
      id: identifierField(offerSource, "id", contract)!,
      occurrenceId: identifierField(offerSource, "occurrenceId", contract)!,
      status: oneOf(
        offerSource.status,
        ["pending", "confirmed", "declined", "expired"] as const,
        contract,
        `${path}.myOffer.status`,
      ),
      offeredAt: isoDateField(offerSource, "offeredAt", contract)!,
      expiresAt: isoDateField(offerSource, "expiresAt", contract)!,
    };
    if (offer.occurrenceId !== occurrence.id) {
      throw new ContractDecodeError(contract, `${path}.myOffer.occurrenceId`);
    }
    occurrence.myOffer = offer;
  }
  if (occurrence.myOffer !== undefined && occurrence.myWaitlistPosition !== undefined) {
    throw new ContractDecodeError(contract, `${path}.myWaitlistPosition`);
  }
  return occurrence;
}

export const decodeEventOccurrence: Decoder<EventOccurrence> = (value) =>
  decodeEventOccurrenceEntry(value, "EventOccurrence", "$");

export const decodeEventOccurrences: Decoder<EventOccurrence[]> = (value) => {
  const contract = "EventOccurrenceList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const occurrences = value.map((entry, index) =>
    decodeEventOccurrenceEntry(entry, contract, `$[${index}]`),
  );
  unique(
    occurrences.map((occurrence) => occurrence.id),
    contract,
    "$[].id",
  );
  return occurrences;
};

// ---- L.2 reschedule and cancellation requests (flows J/K/L) ----

export interface RescheduleRequest {
  id: string;
  occurrenceId: string;
  kind: "reschedule" | "cancellation";
  proposedStartsAt?: IsoDateTime;
  reason: string;
  status: "pending" | "approved" | "declined" | "withdrawn";
  requestedBy: LessonTeacher;
  decisionNote?: string;
  decidedAt?: IsoDateTime;
  createdAt: IsoDateTime;
  version: number;
}

export interface CreateRescheduleRequestRequest {
  occurrenceId: string;
  kind: "reschedule" | "cancellation";
  proposedStartsAt?: IsoDateTime;
  reason: string;
}

export interface DecideRescheduleRequestRequest {
  approve: boolean;
  decisionNote?: string;
  expectedVersion: number;
}

function decodeRescheduleEntry(
  value: unknown,
  contract: string,
  path: string,
): RescheduleRequest {
  const source = record(value, contract, path);
  exactKeys(
    source,
    [
      "id",
      "occurrenceId",
      "kind",
      "proposedStartsAt",
      "reason",
      "status",
      "requestedBy",
      "decisionNote",
      "decidedAt",
      "createdAt",
      "version",
    ],
    contract,
    path,
  );
  const requesterSource = record(source.requestedBy, contract, `${path}.requestedBy`);
  exactKeys(requesterSource, ["accountId", "fullName"], contract, `${path}.requestedBy`);
  const kind = oneOf(
    source.kind,
    ["reschedule", "cancellation"] as const,
    contract,
    `${path}.kind`,
  );
  const request: RescheduleRequest = {
    id: identifierField(source, "id", contract)!,
    occurrenceId: identifierField(source, "occurrenceId", contract)!,
    kind,
    reason: stringField(source, "reason", contract)!,
    status: oneOf(
      source.status,
      ["pending", "approved", "declined", "withdrawn"] as const,
      contract,
      `${path}.status`,
    ),
    requestedBy: {
      accountId: identifierField(requesterSource, "accountId", contract)!,
      fullName: stringField(requesterSource, "fullName", contract)!,
    },
    createdAt: isoDateField(source, "createdAt", contract)!,
    version: numberField(source, "version", contract, 0),
  };
  const proposed = isoDateField(source, "proposedStartsAt", contract, true);
  if (proposed !== undefined) {
    request.proposedStartsAt = proposed;
  }
  if ((kind === "reschedule") !== (request.proposedStartsAt !== undefined)) {
    throw new ContractDecodeError(contract, `${path}.proposedStartsAt`);
  }
  const decisionNote = stringField(source, "decisionNote", contract, true);
  if (decisionNote !== undefined) {
    request.decisionNote = decisionNote;
  }
  const decidedAt = isoDateField(source, "decidedAt", contract, true);
  if (decidedAt !== undefined) {
    request.decidedAt = decidedAt;
  }
  const decided = request.status === "approved" || request.status === "declined";
  if (decided !== (request.decidedAt !== undefined)) {
    throw new ContractDecodeError(contract, `${path}.decidedAt`);
  }
  return request;
}

export const decodeRescheduleRequest: Decoder<RescheduleRequest> = (value) =>
  decodeRescheduleEntry(value, "RescheduleRequest", "$");

export const decodeRescheduleRequests: Decoder<RescheduleRequest[]> = (value) => {
  const contract = "RescheduleRequestList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const requests = value.map((entry, index) =>
    decodeRescheduleEntry(entry, contract, `$[${index}]`),
  );
  unique(
    requests.map((request) => request.id),
    contract,
    "$[].id",
  );
  return requests;
};

// ---- L.3 lesson journals and progress (DEC-006/007) ----

export interface JournalDraft {
  whatWorked: string;
  currentFocus: string;
  nextStep: string;
}

export interface JournalVersion {
  version: number;
  whatWorked: string;
  currentFocus: string;
  nextStep: string;
  correctionNote?: string;
  publishedAt: IsoDateTime;
}

export interface LessonJournal {
  id: string;
  occurrenceId: string;
  studentId: string;
  teacher: LessonTeacher;
  status: "draft" | "published";
  currentVersion: number;
  draft?: JournalDraft;
  versions: JournalVersion[];
  updatedAt: IsoDateTime;
}

export interface JournalDraftRequest {
  occurrenceId: string;
  studentId: string;
  whatWorked: string;
  currentFocus: string;
  nextStep: string;
}

export interface JournalEvidenceInput {
  area: string;
  note: string;
}

export interface PublishJournalRequest {
  occurrenceId: string;
  studentId: string;
  correctionNote?: string;
  evidence?: JournalEvidenceInput[];
}

export interface ProgressEvidence {
  id: string;
  area: string;
  note: string;
  sourceKind: "lesson_journal" | "practice" | "review";
  sourceId: string;
  recordedAt: IsoDateTime;
}

export const decodeLessonJournal: Decoder<LessonJournal> = (value) => {
  const contract = "LessonJournal";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "id",
      "occurrenceId",
      "studentId",
      "teacher",
      "status",
      "currentVersion",
      "draft",
      "versions",
      "updatedAt",
    ],
    contract,
  );
  const teacherSource = record(source.teacher, contract, "$.teacher");
  exactKeys(teacherSource, ["accountId", "fullName"], contract, "$.teacher");
  if (!Array.isArray(source.versions)) {
    throw new ContractDecodeError(contract, "$.versions");
  }
  const versions = source.versions.map((entry, index) => {
    const path = `$.versions[${index}]`;
    const versionSource = record(entry, contract, path);
    exactKeys(
      versionSource,
      ["version", "whatWorked", "currentFocus", "nextStep", "correctionNote", "publishedAt"],
      contract,
      path,
    );
    const version: JournalVersion = {
      version: numberField(versionSource, "version", contract, 1),
      whatWorked: stringField(versionSource, "whatWorked", contract)!,
      currentFocus: stringField(versionSource, "currentFocus", contract)!,
      nextStep: stringField(versionSource, "nextStep", contract)!,
      publishedAt: isoDateField(versionSource, "publishedAt", contract)!,
    };
    const correctionNote = stringField(versionSource, "correctionNote", contract, true);
    if (correctionNote !== undefined) {
      version.correctionNote = correctionNote;
    }
    if (version.version > 1 && version.correctionNote === undefined) {
      throw new ContractDecodeError(contract, `${path}.correctionNote`);
    }
    return version;
  });
  for (let index = 1; index < versions.length; index += 1) {
    if (versions[index]!.version >= versions[index - 1]!.version) {
      throw new ContractDecodeError(contract, `$.versions[${index}].version`);
    }
  }
  const journal: LessonJournal = {
    id: identifierField(source, "id", contract)!,
    occurrenceId: identifierField(source, "occurrenceId", contract)!,
    studentId: identifierField(source, "studentId", contract)!,
    teacher: {
      accountId: identifierField(teacherSource, "accountId", contract)!,
      fullName: stringField(teacherSource, "fullName", contract)!,
    },
    status: oneOf(source.status, ["draft", "published"] as const, contract, "$.status"),
    currentVersion: numberField(source, "currentVersion", contract, 0),
    versions,
    updatedAt: isoDateField(source, "updatedAt", contract)!,
  };
  if (journal.status === "published" && journal.currentVersion < 1) {
    throw new ContractDecodeError(contract, "$.currentVersion");
  }
  if (source.draft !== undefined) {
    const draftSource = record(source.draft, contract, "$.draft");
    exactKeys(draftSource, ["whatWorked", "currentFocus", "nextStep"], contract, "$.draft");
    journal.draft = {
      whatWorked: stringField(draftSource, "whatWorked", contract)!,
      currentFocus: stringField(draftSource, "currentFocus", contract)!,
      nextStep: stringField(draftSource, "nextStep", contract)!,
    };
  }
  return journal;
};

export const decodeLessonJournals: Decoder<LessonJournal[]> = (value) => {
  const contract = "LessonJournalList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const journals = value.map((entry) => decodeLessonJournal(entry));
  unique(
    journals.map((journal) => journal.id),
    contract,
    "$[].id",
  );
  return journals;
};

export const decodeProgressEvidence: Decoder<ProgressEvidence[]> = (value) => {
  const contract = "ProgressEvidenceList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const entries = value.map((entry, index) => {
    const path = `$[${index}]`;
    const source = record(entry, contract, path);
    exactKeys(source, ["id", "area", "note", "sourceKind", "sourceId", "recordedAt"], contract, path);
    return {
      id: identifierField(source, "id", contract)!,
      area: stringField(source, "area", contract)!,
      note: stringField(source, "note", contract)!,
      sourceKind: oneOf(
        source.sourceKind,
        ["lesson_journal", "practice", "review"] as const,
        contract,
        `${path}.sourceKind`,
      ),
      sourceId: stringField(source, "sourceId", contract)!,
      recordedAt: isoDateField(source, "recordedAt", contract)!,
    };
  });
  unique(
    entries.map((entry) => entry.id),
    contract,
    "$[].id",
  );
  return entries;
};

// ---- L.3 homework, practice and media (domain/homework.md) ----

export const MEDIA_KINDS = ["audio", "video", "image", "pdf"] as const;
export type MediaKind = (typeof MEDIA_KINDS)[number];

export const MEDIA_STATUSES = ["pending", "uploading", "ready", "failed"] as const;
export type MediaStatus = (typeof MEDIA_STATUSES)[number];

export interface MediaObject {
  id: string;
  kind: MediaKind;
  contentType: string;
  byteSize: number;
  uploadedBytes: number;
  status: MediaStatus;
  createdAt: IsoDateTime;
  updatedAt: IsoDateTime;
}

export interface CreateMediaRequest {
  kind: MediaKind;
  contentType: string;
  byteSize: number;
}

export interface MediaAccess {
  url: string;
  expiresAt: IsoDateTime;
}

export const HOMEWORK_STATUSES = [
  "draft",
  "assigned",
  "in_progress",
  "submitted",
  "reviewed",
  "completed",
  "cancelled",
  "expired",
] as const;
export type HomeworkStatus = (typeof HOMEWORK_STATUSES)[number];

export interface HomeworkTask {
  id: string;
  position: number;
  title: string;
  description?: string;
  recommendedMinutes?: number;
  skillArea?: string;
  songTitle?: string;
  status: "pending" | "done";
}

export interface PracticeSubmission {
  id: string;
  attempt: number;
  note?: string;
  media: MediaObject[];
  submittedAt: IsoDateTime;
}

export interface PracticeFeedback {
  id: string;
  submissionId: string;
  teacher: LessonTeacher;
  decision: "needs_revision" | "accepted";
  body: string;
  nextStep?: string;
  evidenceArea?: string;
  evidenceNote?: string;
  createdAt: IsoDateTime;
}

export interface HomeworkAssignment {
  id: string;
  occurrenceId: string;
  studentId: string;
  teacher: LessonTeacher;
  status: HomeworkStatus;
  goal: string;
  readinessCriteria?: string;
  dueAt?: IsoDateTime;
  cancelReason?: string;
  tasks: HomeworkTask[];
  attachments: MediaObject[];
  submissions: PracticeSubmission[];
  feedback: PracticeFeedback[];
  version: number;
  createdAt: IsoDateTime;
  updatedAt: IsoDateTime;
}

export interface HomeworkTaskInput {
  title: string;
  description?: string;
  recommendedMinutes?: number;
  skillArea?: string;
  songTitle?: string;
}

export interface CreateHomeworkRequest {
  occurrenceId: string;
  studentId: string;
  goal: string;
  readinessCriteria?: string;
  dueAt?: IsoDateTime;
  tasks?: HomeworkTaskInput[];
  attachmentMediaIds?: string[];
  assign?: boolean;
}

export interface CancelHomeworkRequest {
  reason: string;
}

export interface MarkHomeworkTaskRequest {
  done: boolean;
}

export interface SubmitHomeworkRequest {
  note?: string;
  mediaIds?: string[];
  expectedVersion: number;
}

export interface ReviewHomeworkRequest {
  decision: "needs_revision" | "accepted";
  body: string;
  nextStep?: string;
  evidenceArea?: string;
  evidenceNote?: string;
  expectedVersion: number;
}

export const decodeMediaObject: Decoder<MediaObject> = (value) => {
  const contract = "MediaObject";
  const source = record(value, contract);
  exactKeys(
    source,
    ["id", "kind", "contentType", "byteSize", "uploadedBytes", "status", "createdAt", "updatedAt"],
    contract,
  );
  const object: MediaObject = {
    id: identifierField(source, "id", contract)!,
    kind: oneOf(source.kind, MEDIA_KINDS, contract, "$.kind"),
    contentType: stringField(source, "contentType", contract)!,
    byteSize: numberField(source, "byteSize", contract, 1),
    uploadedBytes: numberField(source, "uploadedBytes", contract, 0),
    status: oneOf(source.status, MEDIA_STATUSES, contract, "$.status"),
    createdAt: isoDateField(source, "createdAt", contract)!,
    updatedAt: isoDateField(source, "updatedAt", contract)!,
  };
  if (object.uploadedBytes > object.byteSize) {
    throw new ContractDecodeError(contract, "$.uploadedBytes");
  }
  if (object.status === "ready" && object.uploadedBytes !== object.byteSize) {
    throw new ContractDecodeError(contract, "$.status");
  }
  return object;
};

export const decodeMediaAccess: Decoder<MediaAccess> = (value) => {
  const contract = "MediaAccess";
  const source = record(value, contract);
  exactKeys(source, ["url", "expiresAt"], contract);
  const url = stringField(source, "url", contract)!;
  if (!url.includes("/v1/media/") || !url.includes("token=")) {
    throw new ContractDecodeError(contract, "$.url");
  }
  return { url, expiresAt: isoDateField(source, "expiresAt", contract)! };
};

export const decodeHomeworkAssignment: Decoder<HomeworkAssignment> = (value) => {
  const contract = "HomeworkAssignment";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "id",
      "occurrenceId",
      "studentId",
      "teacher",
      "status",
      "goal",
      "readinessCriteria",
      "dueAt",
      "cancelReason",
      "tasks",
      "attachments",
      "submissions",
      "feedback",
      "version",
      "createdAt",
      "updatedAt",
    ],
    contract,
  );
  const teacherSource = record(source.teacher, contract, "$.teacher");
  exactKeys(teacherSource, ["accountId", "fullName"], contract, "$.teacher");
  if (!Array.isArray(source.tasks) || !Array.isArray(source.attachments) ||
      !Array.isArray(source.submissions) || !Array.isArray(source.feedback)) {
    throw new ContractDecodeError(contract, "$.tasks");
  }
  const tasks = source.tasks.map((entry, index) => {
    const path = `$.tasks[${index}]`;
    const taskSource = record(entry, contract, path);
    exactKeys(
      taskSource,
      ["id", "position", "title", "description", "recommendedMinutes", "skillArea", "songTitle", "status"],
      contract,
      path,
    );
    const task: HomeworkTask = {
      id: identifierField(taskSource, "id", contract)!,
      position: numberField(taskSource, "position", contract, 1),
      title: stringField(taskSource, "title", contract)!,
      status: oneOf(taskSource.status, ["pending", "done"] as const, contract, `${path}.status`),
    };
    const description = stringField(taskSource, "description", contract, true);
    if (description !== undefined) task.description = description;
    if (taskSource.recommendedMinutes !== undefined) {
      task.recommendedMinutes = numberField(taskSource, "recommendedMinutes", contract, 1);
    }
    const skillArea = stringField(taskSource, "skillArea", contract, true);
    if (skillArea !== undefined) task.skillArea = skillArea;
    const songTitle = stringField(taskSource, "songTitle", contract, true);
    if (songTitle !== undefined) task.songTitle = songTitle;
    return task;
  });
  for (let index = 1; index < tasks.length; index += 1) {
    if (tasks[index]!.position <= tasks[index - 1]!.position) {
      throw new ContractDecodeError(contract, `$.tasks[${index}].position`);
    }
  }
  const submissions = source.submissions.map((entry, index) => {
    const path = `$.submissions[${index}]`;
    const submissionSource = record(entry, contract, path);
    exactKeys(submissionSource, ["id", "attempt", "note", "media", "submittedAt"], contract, path);
    if (!Array.isArray(submissionSource.media)) {
      throw new ContractDecodeError(contract, `${path}.media`);
    }
    const submission: PracticeSubmission = {
      id: identifierField(submissionSource, "id", contract)!,
      attempt: numberField(submissionSource, "attempt", contract, 1),
      media: submissionSource.media.map((mediaEntry) => decodeMediaObject(mediaEntry)),
      submittedAt: isoDateField(submissionSource, "submittedAt", contract)!,
    };
    const note = stringField(submissionSource, "note", contract, true);
    if (note !== undefined) submission.note = note;
    return submission;
  });
  for (let index = 1; index < submissions.length; index += 1) {
    if (submissions[index]!.attempt >= submissions[index - 1]!.attempt) {
      throw new ContractDecodeError(contract, `$.submissions[${index}].attempt`);
    }
  }
  const feedback = source.feedback.map((entry, index) => {
    const path = `$.feedback[${index}]`;
    const feedbackSource = record(entry, contract, path);
    exactKeys(
      feedbackSource,
      ["id", "submissionId", "teacher", "decision", "body", "nextStep", "evidenceArea", "evidenceNote", "createdAt"],
      contract,
      path,
    );
    const feedbackTeacher = record(feedbackSource.teacher, contract, `${path}.teacher`);
    exactKeys(feedbackTeacher, ["accountId", "fullName"], contract, `${path}.teacher`);
    const item: PracticeFeedback = {
      id: identifierField(feedbackSource, "id", contract)!,
      submissionId: identifierField(feedbackSource, "submissionId", contract)!,
      teacher: {
        accountId: identifierField(feedbackTeacher, "accountId", contract)!,
        fullName: stringField(feedbackTeacher, "fullName", contract)!,
      },
      decision: oneOf(
        feedbackSource.decision,
        ["needs_revision", "accepted"] as const,
        contract,
        `${path}.decision`,
      ),
      body: stringField(feedbackSource, "body", contract)!,
      createdAt: isoDateField(feedbackSource, "createdAt", contract)!,
    };
    const nextStep = stringField(feedbackSource, "nextStep", contract, true);
    if (nextStep !== undefined) item.nextStep = nextStep;
    const evidenceArea = stringField(feedbackSource, "evidenceArea", contract, true);
    if (evidenceArea !== undefined) item.evidenceArea = evidenceArea;
    const evidenceNote = stringField(feedbackSource, "evidenceNote", contract, true);
    if (evidenceNote !== undefined) item.evidenceNote = evidenceNote;
    if ((item.evidenceArea === undefined) !== (item.evidenceNote === undefined)) {
      throw new ContractDecodeError(contract, `${path}.evidenceArea`);
    }
    if (item.decision === "needs_revision" && item.evidenceArea !== undefined) {
      throw new ContractDecodeError(contract, `${path}.evidenceArea`);
    }
    return item;
  });
  const homework: HomeworkAssignment = {
    id: identifierField(source, "id", contract)!,
    occurrenceId: identifierField(source, "occurrenceId", contract)!,
    studentId: identifierField(source, "studentId", contract)!,
    teacher: {
      accountId: identifierField(teacherSource, "accountId", contract)!,
      fullName: stringField(teacherSource, "fullName", contract)!,
    },
    status: oneOf(source.status, HOMEWORK_STATUSES, contract, "$.status"),
    goal: stringField(source, "goal", contract)!,
    tasks,
    attachments: source.attachments.map((entry) => decodeMediaObject(entry)),
    submissions,
    feedback,
    version: numberField(source, "version", contract, 1),
    createdAt: isoDateField(source, "createdAt", contract)!,
    updatedAt: isoDateField(source, "updatedAt", contract)!,
  };
  const readiness = stringField(source, "readinessCriteria", contract, true);
  if (readiness !== undefined) homework.readinessCriteria = readiness;
  const dueAt = isoDateField(source, "dueAt", contract, true);
  if (dueAt !== undefined) homework.dueAt = dueAt;
  const cancelReason = stringField(source, "cancelReason", contract, true);
  if (cancelReason !== undefined) homework.cancelReason = cancelReason;
  if (homework.status === "cancelled" && homework.cancelReason === undefined) {
    throw new ContractDecodeError(contract, "$.cancelReason");
  }
  if (
    homework.status === "completed" &&
    !homework.feedback.some((entry) => entry.decision === "accepted")
  ) {
    throw new ContractDecodeError(contract, "$.feedback");
  }
  return homework;
};

export const decodeHomeworkAssignments: Decoder<HomeworkAssignment[]> = (value) => {
  const contract = "HomeworkAssignmentList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const items = value.map((entry) => decodeHomeworkAssignment(entry));
  unique(
    items.map((item) => item.id),
    contract,
    "$[].id",
  );
  return items;
};

// ---- L.4 lesson attendance (domain/lesson.md) ----

export const ATTENDANCE_STATUSES = ["present", "late", "absent"] as const;
export type AttendanceStatus = (typeof ATTENDANCE_STATUSES)[number];

export interface AttendanceRecord {
  studentId: string;
  studentName: string;
  status: AttendanceStatus;
  lateMinutes?: number;
  note?: string;
  recordedAt: IsoDateTime;
  updatedAt: IsoDateTime;
}

export interface MarkAttendanceRequest {
  status: AttendanceStatus;
  lateMinutes?: number;
  note?: string;
  changeReason?: string;
}

export const decodeAttendanceRecords: Decoder<AttendanceRecord[]> = (value) => {
  const contract = "AttendanceRecordList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const records = value.map((entry, index) => {
    const path = `$[${index}]`;
    const source = record(entry, contract, path);
    exactKeys(
      source,
      ["studentId", "studentName", "status", "lateMinutes", "note", "recordedAt", "updatedAt"],
      contract,
      path,
    );
    const item: AttendanceRecord = {
      studentId: identifierField(source, "studentId", contract)!,
      studentName: stringField(source, "studentName", contract)!,
      status: oneOf(source.status, ATTENDANCE_STATUSES, contract, `${path}.status`),
      recordedAt: isoDateField(source, "recordedAt", contract)!,
      updatedAt: isoDateField(source, "updatedAt", contract)!,
    };
    if (source.lateMinutes !== undefined) {
      item.lateMinutes = numberField(source, "lateMinutes", contract, 1);
    }
    const note = stringField(source, "note", contract, true);
    if (note !== undefined) item.note = note;
    if ((item.status === "late") !== (item.lateMinutes !== undefined)) {
      throw new ContractDecodeError(contract, `${path}.lateMinutes`);
    }
    // An absence without a note is acceptable on read: the Student's
    // own view legitimately hides the teacher note.
    return item;
  });
  unique(
    records.map((entry) => entry.studentId),
    contract,
    "$[].studentId",
  );
  return records;
};

// ---- L.3 student repertoire (aggregate StudentSong) ----

export const SONG_STAGES = [
  "acquaintance",
  "learning",
  "technically_stable",
  "interpretation",
  "stage_ready",
] as const;
export type SongStage = (typeof SONG_STAGES)[number];

export interface SongStageChange {
  fromStage?: SongStage;
  toStage: SongStage;
  note?: string;
  changedAt: IsoDateTime;
}

export interface StudentSong {
  id: string;
  studentId: string;
  title: string;
  artist?: string;
  stage: SongStage;
  stageNote?: string;
  assignedBy: LessonTeacher;
  history: SongStageChange[];
  version: number;
  createdAt: IsoDateTime;
  updatedAt: IsoDateTime;
}

export interface AddStudentSongRequest {
  title: string;
  artist?: string;
  stage?: SongStage;
  stageNote?: string;
}

export interface ChangeSongStageRequest {
  stage: SongStage;
  stageNote?: string;
  expectedVersion: number;
}

export const decodeStudentSong: Decoder<StudentSong> = (value) => {
  const contract = "StudentSong";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "id",
      "studentId",
      "title",
      "artist",
      "stage",
      "stageNote",
      "assignedBy",
      "history",
      "version",
      "createdAt",
      "updatedAt",
    ],
    contract,
  );
  const teacherSource = record(source.assignedBy, contract, "$.assignedBy");
  exactKeys(teacherSource, ["accountId", "fullName"], contract, "$.assignedBy");
  if (!Array.isArray(source.history) || source.history.length === 0) {
    throw new ContractDecodeError(contract, "$.history");
  }
  const history = source.history.map((entry, index) => {
    const path = `$.history[${index}]`;
    const changeSource = record(entry, contract, path);
    exactKeys(changeSource, ["fromStage", "toStage", "note", "changedAt"], contract, path);
    const change: SongStageChange = {
      toStage: oneOf(changeSource.toStage, SONG_STAGES, contract, `${path}.toStage`),
      changedAt: isoDateField(changeSource, "changedAt", contract)!,
    };
    if (changeSource.fromStage !== undefined) {
      change.fromStage = oneOf(changeSource.fromStage, SONG_STAGES, contract, `${path}.fromStage`);
    }
    const note = stringField(changeSource, "note", contract, true);
    if (note !== undefined) change.note = note;
    return change;
  });
  // The oldest entry is the assignment itself and has no fromStage;
  // every later entry records where the song came from.
  const oldest = history[history.length - 1]!;
  if (oldest.fromStage !== undefined) {
    throw new ContractDecodeError(contract, "$.history");
  }
  const song: StudentSong = {
    id: identifierField(source, "id", contract)!,
    studentId: identifierField(source, "studentId", contract)!,
    title: stringField(source, "title", contract)!,
    stage: oneOf(source.stage, SONG_STAGES, contract, "$.stage"),
    assignedBy: {
      accountId: identifierField(teacherSource, "accountId", contract)!,
      fullName: stringField(teacherSource, "fullName", contract)!,
    },
    history,
    version: numberField(source, "version", contract, 1),
    createdAt: isoDateField(source, "createdAt", contract)!,
    updatedAt: isoDateField(source, "updatedAt", contract)!,
  };
  const artist = stringField(source, "artist", contract, true);
  if (artist !== undefined) song.artist = artist;
  const stageNote = stringField(source, "stageNote", contract, true);
  if (stageNote !== undefined) song.stageNote = stageNote;
  if (history[0]!.toStage !== song.stage) {
    throw new ContractDecodeError(contract, "$.stage");
  }
  return song;
};

export const decodeStudentSongs: Decoder<StudentSong[]> = (value) => {
  const contract = "StudentSongList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const songs = value.map((entry) => decodeStudentSong(entry));
  unique(
    songs.map((song) => song.id),
    contract,
    "$[].id",
  );
  return songs;
};

// ---- L.3 goals and achievements ----

export const GOAL_STATUSES = ["active", "completed", "cancelled"] as const;
export type GoalStatus = (typeof GOAL_STATUSES)[number];

export interface StudentGoal {
  id: string;
  studentId: string;
  criterion: string;
  description?: string;
  relatedSongId?: string;
  relatedSkillArea?: string;
  status: GoalStatus;
  completionNote?: string;
  cancelReason?: string;
  replacedByGoalId?: string;
  createdBy: LessonTeacher;
  version: number;
  createdAt: IsoDateTime;
  updatedAt: IsoDateTime;
}

export interface CreateGoalRequest {
  criterion: string;
  description?: string;
  relatedSongId?: string;
  relatedSkillArea?: string;
}

export interface CompleteGoalRequest {
  completionNote: string;
  expectedVersion: number;
}

export interface ReframeGoalRequest {
  reason: string;
  newCriterion?: string;
  newDescription?: string;
  expectedVersion: number;
}

export interface AchievementDefinition {
  id: string;
  name: string;
  description: string;
  category: string;
  evidenceRequirement?: string;
  status: "published" | "retired";
  definitionVersion: number;
  createdAt: IsoDateTime;
  retiredAt?: IsoDateTime;
}

export interface CreateAchievementDefinitionRequest {
  name: string;
  description: string;
  category: string;
  evidenceRequirement?: string;
}

export interface AchievementAward {
  id: string;
  definitionId: string;
  definitionName: string;
  category: string;
  studentId: string;
  evidenceNote: string;
  status: "awarded" | "revoked";
  revokeReason?: string;
  revokedAt?: IsoDateTime;
  awardedBy: LessonTeacher;
  awardedAt: IsoDateTime;
  definitionVersion: number;
}

export interface AwardAchievementRequest {
  definitionId: string;
  evidenceNote: string;
}

export interface RevokeAchievementRequest {
  reason: string;
}

export const decodeStudentGoal: Decoder<StudentGoal> = (value) => {
  const contract = "StudentGoal";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "id",
      "studentId",
      "criterion",
      "description",
      "relatedSongId",
      "relatedSkillArea",
      "status",
      "completionNote",
      "cancelReason",
      "replacedByGoalId",
      "createdBy",
      "version",
      "createdAt",
      "updatedAt",
    ],
    contract,
  );
  const teacherSource = record(source.createdBy, contract, "$.createdBy");
  exactKeys(teacherSource, ["accountId", "fullName"], contract, "$.createdBy");
  const goal: StudentGoal = {
    id: identifierField(source, "id", contract)!,
    studentId: identifierField(source, "studentId", contract)!,
    criterion: stringField(source, "criterion", contract)!,
    status: oneOf(source.status, GOAL_STATUSES, contract, "$.status"),
    createdBy: {
      accountId: identifierField(teacherSource, "accountId", contract)!,
      fullName: stringField(teacherSource, "fullName", contract)!,
    },
    version: numberField(source, "version", contract, 1),
    createdAt: isoDateField(source, "createdAt", contract)!,
    updatedAt: isoDateField(source, "updatedAt", contract)!,
  };
  const description = stringField(source, "description", contract, true);
  if (description !== undefined) goal.description = description;
  const relatedSongId = stringField(source, "relatedSongId", contract, true);
  if (relatedSongId !== undefined) goal.relatedSongId = relatedSongId;
  const relatedSkillArea = stringField(source, "relatedSkillArea", contract, true);
  if (relatedSkillArea !== undefined) goal.relatedSkillArea = relatedSkillArea;
  const completionNote = stringField(source, "completionNote", contract, true);
  if (completionNote !== undefined) goal.completionNote = completionNote;
  const cancelReason = stringField(source, "cancelReason", contract, true);
  if (cancelReason !== undefined) goal.cancelReason = cancelReason;
  const replacedByGoalId = stringField(source, "replacedByGoalId", contract, true);
  if (replacedByGoalId !== undefined) goal.replacedByGoalId = replacedByGoalId;
  if (goal.status === "completed" && goal.completionNote === undefined) {
    throw new ContractDecodeError(contract, "$.completionNote");
  }
  if (goal.status === "cancelled" && goal.cancelReason === undefined) {
    throw new ContractDecodeError(contract, "$.cancelReason");
  }
  if (goal.replacedByGoalId !== undefined && goal.status !== "cancelled") {
    throw new ContractDecodeError(contract, "$.replacedByGoalId");
  }
  return goal;
};

export const decodeStudentGoals: Decoder<StudentGoal[]> = (value) => {
  const contract = "StudentGoalList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const goals = value.map((entry) => decodeStudentGoal(entry));
  unique(
    goals.map((goal) => goal.id),
    contract,
    "$[].id",
  );
  return goals;
};

export const decodeAchievementDefinition: Decoder<AchievementDefinition> = (value) => {
  const contract = "AchievementDefinition";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "id",
      "name",
      "description",
      "category",
      "evidenceRequirement",
      "status",
      "definitionVersion",
      "createdAt",
      "retiredAt",
    ],
    contract,
  );
  const definition: AchievementDefinition = {
    id: identifierField(source, "id", contract)!,
    name: stringField(source, "name", contract)!,
    description: stringField(source, "description", contract)!,
    category: stringField(source, "category", contract)!,
    status: oneOf(source.status, ["published", "retired"] as const, contract, "$.status"),
    definitionVersion: numberField(source, "definitionVersion", contract, 1),
    createdAt: isoDateField(source, "createdAt", contract)!,
  };
  const evidenceRequirement = stringField(source, "evidenceRequirement", contract, true);
  if (evidenceRequirement !== undefined) definition.evidenceRequirement = evidenceRequirement;
  const retiredAt = isoDateField(source, "retiredAt", contract, true);
  if (retiredAt !== undefined) definition.retiredAt = retiredAt;
  if ((definition.status === "retired") !== (definition.retiredAt !== undefined)) {
    throw new ContractDecodeError(contract, "$.retiredAt");
  }
  return definition;
};

export const decodeAchievementDefinitions: Decoder<AchievementDefinition[]> = (value) => {
  const contract = "AchievementDefinitionList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const definitions = value.map((entry) => decodeAchievementDefinition(entry));
  unique(
    definitions.map((definition) => definition.id),
    contract,
    "$[].id",
  );
  return definitions;
};

export const decodeAchievementAward: Decoder<AchievementAward> = (value) => {
  const contract = "AchievementAward";
  const source = record(value, contract);
  exactKeys(
    source,
    [
      "id",
      "definitionId",
      "definitionName",
      "category",
      "studentId",
      "evidenceNote",
      "status",
      "revokeReason",
      "revokedAt",
      "awardedBy",
      "awardedAt",
      "definitionVersion",
    ],
    contract,
  );
  const teacherSource = record(source.awardedBy, contract, "$.awardedBy");
  exactKeys(teacherSource, ["accountId", "fullName"], contract, "$.awardedBy");
  const award: AchievementAward = {
    id: identifierField(source, "id", contract)!,
    definitionId: identifierField(source, "definitionId", contract)!,
    definitionName: stringField(source, "definitionName", contract)!,
    category: stringField(source, "category", contract)!,
    studentId: identifierField(source, "studentId", contract)!,
    evidenceNote: stringField(source, "evidenceNote", contract)!,
    status: oneOf(source.status, ["awarded", "revoked"] as const, contract, "$.status"),
    awardedBy: {
      accountId: identifierField(teacherSource, "accountId", contract)!,
      fullName: stringField(teacherSource, "fullName", contract)!,
    },
    awardedAt: isoDateField(source, "awardedAt", contract)!,
    definitionVersion: numberField(source, "definitionVersion", contract, 1),
  };
  const revokeReason = stringField(source, "revokeReason", contract, true);
  if (revokeReason !== undefined) award.revokeReason = revokeReason;
  const revokedAt = isoDateField(source, "revokedAt", contract, true);
  if (revokedAt !== undefined) award.revokedAt = revokedAt;
  if ((award.status === "revoked") !== (award.revokeReason !== undefined && award.revokedAt !== undefined)) {
    throw new ContractDecodeError(contract, "$.revokeReason");
  }
  return award;
};

export const decodeAchievementAwards: Decoder<AchievementAward[]> = (value) => {
  const contract = "AchievementAwardList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const awards = value.map((entry) => decodeAchievementAward(entry));
  unique(
    awards.map((award) => award.id),
    contract,
    "$[].id",
  );
  return awards;
};

// ---- L.5 activity feed and notification preferences ----

export const NOTIFICATION_CATEGORIES = [
  "important",
  "learning",
  "messages",
  "community",
] as const;
export type NotificationCategory = (typeof NOTIFICATION_CATEGORIES)[number];

export interface ActivityEntry {
  id: string;
  category: NotificationCategory;
  kind: string;
  targetType: string;
  targetId: string;
  payload: Record<string, unknown>;
  occurredAt: IsoDateTime;
  readAt?: IsoDateTime;
}

export interface ActivityFeed {
  unreadCount: number;
  entries: ActivityEntry[];
}

export interface MarkActivityReadRequest {
  upTo: IsoDateTime;
}

export interface MarkActivityReadResult {
  marked: number;
}

export interface NotificationPreference {
  category: NotificationCategory;
  pushEnabled: boolean;
}

export interface UpdateNotificationPreferenceRequest {
  category: NotificationCategory;
  pushEnabled: boolean;
}

export const decodeActivityFeed: Decoder<ActivityFeed> = (value) => {
  const contract = "ActivityFeed";
  const source = record(value, contract);
  exactKeys(source, ["unreadCount", "entries"], contract);
  if (!Array.isArray(source.entries)) {
    throw new ContractDecodeError(contract, "$.entries");
  }
  const entries = source.entries.map((entry, index) => {
    const path = `$.entries[${index}]`;
    const entrySource = record(entry, contract, path);
    exactKeys(
      entrySource,
      ["id", "category", "kind", "targetType", "targetId", "payload", "occurredAt", "readAt"],
      contract,
      path,
    );
    const payload = record(entrySource.payload, contract, `${path}.payload`);
    const item: ActivityEntry = {
      id: identifierField(entrySource, "id", contract)!,
      category: oneOf(entrySource.category, NOTIFICATION_CATEGORIES, contract, `${path}.category`),
      kind: stringField(entrySource, "kind", contract)!,
      targetType: stringField(entrySource, "targetType", contract)!,
      targetId: stringField(entrySource, "targetId", contract)!,
      payload,
      occurredAt: isoDateField(entrySource, "occurredAt", contract)!,
    };
    const readAt = isoDateField(entrySource, "readAt", contract, true);
    if (readAt !== undefined) item.readAt = readAt;
    return item;
  });
  unique(
    entries.map((entry) => entry.id),
    contract,
    "$.entries[].id",
  );
  const unreadCount = numberField(source, "unreadCount", contract, 0);
  const unreadInList = entries.filter((entry) => entry.readAt === undefined).length;
  if (unreadCount < unreadInList && entries.length < 100) {
    throw new ContractDecodeError(contract, "$.unreadCount");
  }
  return { unreadCount, entries };
};

export const decodeMarkActivityReadResult: Decoder<MarkActivityReadResult> = (value) => {
  const contract = "MarkActivityReadResult";
  const source = record(value, contract);
  exactKeys(source, ["marked"], contract);
  return { marked: numberField(source, "marked", contract, 0) };
};

export const decodeNotificationPreferences: Decoder<NotificationPreference[]> = (value) => {
  const contract = "NotificationPreferenceList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const preferences = value.map((entry, index) => {
    const path = `$[${index}]`;
    const source = record(entry, contract, path);
    exactKeys(source, ["category", "pushEnabled"], contract, path);
    if (typeof source.pushEnabled !== "boolean") {
      throw new ContractDecodeError(contract, `${path}.pushEnabled`);
    }
    return {
      category: oneOf(source.category, NOTIFICATION_CATEGORIES, contract, `${path}.category`),
      pushEnabled: source.pushEnabled,
    };
  });
  unique(
    preferences.map((preference) => preference.category),
    contract,
    "$[].category",
  );
  return preferences;
};

// ---- L.5 community and safety (Page 28; COM-SAFE-02/03/05) ----

export const COMMUNITY_POST_KINDS = ["post", "announcement"] as const;
export type CommunityPostKind = (typeof COMMUNITY_POST_KINDS)[number];

export const COMMUNITY_AUDIENCES = ["school", "staff"] as const;
export type CommunityAudience = (typeof COMMUNITY_AUDIENCES)[number];

export const COMMUNITY_CONTENT_STATUSES = ["published", "hidden", "removed"] as const;
export type CommunityContentStatus = (typeof COMMUNITY_CONTENT_STATUSES)[number];

export const COMMUNITY_REPORT_REASONS = ["abuse", "personal_data", "spam", "other"] as const;
export type CommunityReportReason = (typeof COMMUNITY_REPORT_REASONS)[number];

export const COMMUNITY_AUTHOR_ROLES = ["Owner", "Administrator", "Teacher", "Student"] as const;
export type CommunityAuthorRole = (typeof COMMUNITY_AUTHOR_ROLES)[number];

/** Empty fields mean a tombstone: the words and the author leave. */
export interface CommunityAuthor {
  accountId: string;
  fullName: string;
  role: CommunityAuthorRole | "";
}

export interface CommunityComment {
  id: string;
  author: CommunityAuthor;
  body?: string | undefined;
  status: CommunityContentStatus;
  createdAt: IsoDateTime;
}

export interface CommunityPost {
  id: string;
  kind: CommunityPostKind;
  title?: string | undefined;
  body?: string | undefined;
  audience: CommunityAudience;
  commentsEnabled: boolean;
  pinned: boolean;
  status: CommunityContentStatus;
  author: CommunityAuthor;
  commentCount: number;
  comments?: CommunityComment[] | undefined;
  createdAt: IsoDateTime;
}

export interface CreateCommunityPostRequest {
  kind?: CommunityPostKind;
  title?: string;
  body: string;
  audience?: CommunityAudience;
  commentsEnabled: boolean;
  pinned?: boolean;
}

export interface AddCommunityCommentRequest {
  body: string;
}

export type CommunityTargetType = "post" | "comment";

export interface RemoveCommunityContentRequest {
  targetType: CommunityTargetType;
  targetId: string;
}

export interface ReportCommunityContentRequest {
  targetType: CommunityTargetType;
  targetId: string;
  reason: CommunityReportReason;
  note?: string;
}

export interface CommunityReport {
  id: string;
  targetType: CommunityTargetType;
  targetId: string;
  reason: CommunityReportReason;
  note?: string | undefined;
  status: "new" | "reviewed";
  decision?: "hidden" | "kept" | undefined;
  decisionReason?: string | undefined;
  decidedAt?: IsoDateTime | undefined;
  createdAt: IsoDateTime;
  targetExcerpt?: string | undefined;
}

export interface DecideCommunityReportRequest {
  decision: "hidden" | "kept";
  decisionReason: string;
}

export interface BlockCommunityMemberRequest {
  accountId: string;
  blocked: boolean;
}

export interface BlockedMembers {
  blocked: string[];
}

function decodeCommunityAuthor(
  value: unknown,
  contract: string,
  path: string,
): CommunityAuthor {
  const source = record(value, contract, path);
  exactKeys(source, ["accountId", "fullName", "role"], contract, path);
  const { accountId, fullName, role } = source;
  if (
    typeof accountId !== "string" ||
    typeof fullName !== "string" ||
    typeof role !== "string"
  ) {
    throw new ContractDecodeError(contract, path);
  }
  if (accountId === "") {
    // A tombstone author is empty entirely, never partially.
    if (fullName !== "" || role !== "") {
      throw new ContractDecodeError(contract, path);
    }
    return { accountId: "", fullName: "", role: "" };
  }
  if (fullName === "") {
    throw new ContractDecodeError(contract, `${path}.fullName`);
  }
  const authorRole = oneOf(role, COMMUNITY_AUTHOR_ROLES, contract, `${path}.role`);
  return { accountId, fullName, role: authorRole };
}

function decodeCommunityComment(
  value: unknown,
  contract: string,
  path: string,
): CommunityComment {
  const source = record(value, contract, path);
  exactKeys(source, ["id", "author", "body", "status", "createdAt"], contract, path);
  const status = oneOf(source.status, COMMUNITY_CONTENT_STATUSES, contract, `${path}.status`);
  const comment: CommunityComment = {
    id: identifierField(source, "id", contract)!,
    author: decodeCommunityAuthor(source.author, contract, `${path}.author`),
    status,
    createdAt: isoDateField(source, "createdAt", contract)!,
  };
  const body = stringField(source, "body", contract, true);
  if (body !== undefined) comment.body = body;
  // A published reply always shows its words and author; a tombstone
  // for members hides both together (moderators keep the words).
  if (status === "published" && (comment.body === undefined || comment.author.accountId === "")) {
    throw new ContractDecodeError(contract, path);
  }
  if (status !== "published" && comment.body !== undefined && comment.author.accountId === "") {
    throw new ContractDecodeError(contract, path);
  }
  return comment;
}

function decodeCommunityPostShape(
  value: unknown,
  contract: string,
  path: string,
): CommunityPost {
  const source = record(value, contract, path);
  exactKeys(
    source,
    [
      "id",
      "kind",
      "title",
      "body",
      "audience",
      "commentsEnabled",
      "pinned",
      "status",
      "author",
      "commentCount",
      "comments",
      "createdAt",
    ],
    contract,
    path,
  );
  const status = oneOf(source.status, COMMUNITY_CONTENT_STATUSES, contract, `${path}.status`);
  const kind = oneOf(source.kind, COMMUNITY_POST_KINDS, contract, `${path}.kind`);
  const post: CommunityPost = {
    id: identifierField(source, "id", contract)!,
    kind,
    audience: oneOf(source.audience, COMMUNITY_AUDIENCES, contract, `${path}.audience`),
    commentsEnabled: booleanField(source, "commentsEnabled", contract),
    pinned: booleanField(source, "pinned", contract),
    status,
    author: decodeCommunityAuthor(source.author, contract, `${path}.author`),
    commentCount: numberField(source, "commentCount", contract, 0),
    createdAt: isoDateField(source, "createdAt", contract)!,
  };
  const title = stringField(source, "title", contract, true);
  if (title !== undefined) post.title = title;
  const body = stringField(source, "body", contract, true);
  if (body !== undefined) post.body = body;
  if (post.pinned && kind !== "announcement") {
    throw new ContractDecodeError(contract, `${path}.pinned`);
  }
  if (status === "published") {
    if (post.body === undefined || post.author.accountId === "") {
      throw new ContractDecodeError(contract, path);
    }
    if (kind === "announcement" && post.title === undefined) {
      throw new ContractDecodeError(contract, `${path}.title`);
    }
  }
  if (source.comments !== undefined) {
    if (!Array.isArray(source.comments)) {
      throw new ContractDecodeError(contract, `${path}.comments`);
    }
    const comments = source.comments.map((entry, index) =>
      decodeCommunityComment(entry, contract, `${path}.comments[${index}]`),
    );
    unique(
      comments.map((comment) => comment.id),
      contract,
      `${path}.comments[].id`,
    );
    for (let index = 1; index < comments.length; index += 1) {
      if (comments[index]!.createdAt < comments[index - 1]!.createdAt) {
        throw new ContractDecodeError(contract, `${path}.comments[${index}].createdAt`);
      }
    }
    // The count is the published replies the viewer can see.
    const published = comments.filter((comment) => comment.status === "published").length;
    if (post.commentCount !== published) {
      throw new ContractDecodeError(contract, `${path}.commentCount`);
    }
    post.comments = comments;
  }
  return post;
}

export const decodeCommunityPost: Decoder<CommunityPost> = (value) =>
  decodeCommunityPostShape(value, "CommunityPost", "$");

export const decodeCommunityFeed: Decoder<CommunityPost[]> = (value) => {
  const contract = "CommunityFeed";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const posts = value.map((entry, index) =>
    decodeCommunityPostShape(entry, contract, `$[${index}]`),
  );
  unique(
    posts.map((post) => post.id),
    contract,
    "$[].id",
  );
  let unpinnedSeen = false;
  posts.forEach((post, index) => {
    if (post.status !== "published") {
      throw new ContractDecodeError(contract, `$[${index}].status`);
    }
    if (post.pinned && unpinnedSeen) {
      throw new ContractDecodeError(contract, `$[${index}].pinned`);
    }
    if (!post.pinned) unpinnedSeen = true;
  });
  return posts;
};

function decodeCommunityReportShape(
  value: unknown,
  contract: string,
  path: string,
): CommunityReport {
  const source = record(value, contract, path);
  exactKeys(
    source,
    [
      "id",
      "targetType",
      "targetId",
      "reason",
      "note",
      "status",
      "decision",
      "decisionReason",
      "decidedAt",
      "createdAt",
      "targetExcerpt",
    ],
    contract,
    path,
  );
  const status = oneOf(source.status, ["new", "reviewed"] as const, contract, `${path}.status`);
  const reason = oneOf(source.reason, COMMUNITY_REPORT_REASONS, contract, `${path}.reason`);
  const report: CommunityReport = {
    id: identifierField(source, "id", contract)!,
    targetType: oneOf(source.targetType, ["post", "comment"] as const, contract, `${path}.targetType`),
    targetId: identifierField(source, "targetId", contract)!,
    reason,
    status,
    createdAt: isoDateField(source, "createdAt", contract)!,
  };
  const note = stringField(source, "note", contract, true);
  if (note !== undefined) report.note = note;
  const decisionReason = stringField(source, "decisionReason", contract, true);
  if (decisionReason !== undefined) report.decisionReason = decisionReason;
  const decidedAt = isoDateField(source, "decidedAt", contract, true);
  if (decidedAt !== undefined) report.decidedAt = decidedAt;
  const targetExcerpt = stringField(source, "targetExcerpt", contract, true);
  if (targetExcerpt !== undefined) report.targetExcerpt = targetExcerpt;
  if (source.decision !== undefined) {
    report.decision = oneOf(
      source.decision,
      ["hidden", "kept"] as const,
      contract,
      `${path}.decision`,
    );
  }
  if (reason === "other" && report.note === undefined) {
    throw new ContractDecodeError(contract, `${path}.note`);
  }
  // A report is either new or fully decided — never in between.
  const decided =
    report.decision !== undefined &&
    report.decisionReason !== undefined &&
    report.decidedAt !== undefined;
  const undecided =
    report.decision === undefined &&
    report.decisionReason === undefined &&
    report.decidedAt === undefined;
  if (status === "reviewed" ? !decided : !undecided) {
    throw new ContractDecodeError(contract, `${path}.status`);
  }
  return report;
}

export const decodeCommunityReport: Decoder<CommunityReport> = (value) =>
  decodeCommunityReportShape(value, "CommunityReport", "$");

export const decodeCommunityReports: Decoder<CommunityReport[]> = (value) => {
  const contract = "CommunityReportList";
  if (!Array.isArray(value)) {
    throw new ContractDecodeError(contract, "$");
  }
  const reports = value.map((entry, index) =>
    decodeCommunityReportShape(entry, contract, `$[${index}]`),
  );
  unique(
    reports.map((report) => report.id),
    contract,
    "$[].id",
  );
  let reviewedSeen = false;
  reports.forEach((report, index) => {
    if (report.status === "reviewed") {
      reviewedSeen = true;
    } else if (reviewedSeen) {
      // The queue lists what awaits a decision first.
      throw new ContractDecodeError(contract, `$[${index}].status`);
    }
  });
  return reports;
};

export const decodeBlockedMembers: Decoder<BlockedMembers> = (value) => {
  const contract = "BlockedMembers";
  const source = record(value, contract);
  exactKeys(source, ["blocked"], contract);
  if (!Array.isArray(source.blocked)) {
    throw new ContractDecodeError(contract, "$.blocked");
  }
  const blocked = source.blocked.map((entry, index) => {
    if (typeof entry !== "string" || entry.length === 0 || entry.length > 128) {
      throw new ContractDecodeError(contract, `$.blocked[${index}]`);
    }
    return entry;
  });
  unique(blocked, contract, "$.blocked[]");
  return { blocked };
};
