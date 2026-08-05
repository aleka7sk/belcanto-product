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

export interface Lesson {
  id: string;
  title: string;
  startsAt: IsoDateTime;
  durationMinutes: number;
  location?: string;
  teacher: LessonTeacher;
  students: LessonStudent[];
  status: "scheduled";
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
    status: oneOf(source.status, ["scheduled"] as const, contract, "$.status"),
    version: numberField(source, "version", contract),
  };
  const location = stringField(source, "location", contract, true);
  if (location !== undefined) lesson.location = location;
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
