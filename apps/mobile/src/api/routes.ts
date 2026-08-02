import {
  decodeActivationPreview,
  decodeBootstrapView,
  decodeDelegationResult,
  decodeFirstMinute,
  decodeInvitationResult,
  decodeLesson,
  decodeLessons,
  decodeReassignPrimaryTeachersResult,
  decodeReplaceLessonTeachersResult,
  decodeSessionTokens,
  decodeStaffMembers,
  decodeStudentDirectory,
  decodeStudentOnboardingItems,
  decodeStudentResult,
  decodeVoid,
  type ActivationPreview,
  type ActivationPreviewRequest,
  type BootstrapView,
  type CompleteActivationRequest,
  type CreateLessonRequest,
  type CreateStudentRequest,
  type Decoder,
  type DelegationResult,
  type FirstMinute,
  type GrantDelegationRequest,
  type InvitationResult,
  type Lesson,
  type LessonListQuery,
  type PublishFirstMinuteRequest,
  type ReassignPrimaryTeachersRequest,
  type ReassignPrimaryTeachersResult,
  type RefreshSessionRequest,
  type ReplaceLessonTeachersRequest,
  type ReplaceLessonTeachersResult,
  type RevokeDelegationRequest,
  type SessionTokens,
  type SignInRequest,
  type StaffMember,
  type StaffRole,
  type StudentDirectoryItem,
  type StudentDirectoryQuery,
  type StudentOnboardingItem,
  type StudentResult,
} from "./contracts";
import { backendIdentifierIssue } from "@/validation/backend";

export type HttpMethod = "GET" | "POST" | "PUT" | "DELETE";
export type AuthMode = "public" | "required";
export type ErrorStatus = 400 | 401 | 403 | 404 | 409 | 422 | 429 | 500;

export interface RouteDescriptor<Response> {
  readonly method: HttpMethod;
  readonly path: string;
  readonly auth: AuthMode;
  readonly successStatus: 200 | 201 | 204;
  readonly errorStatuses: readonly ErrorStatus[];
  readonly decode: Decoder<Response>;
}

function route<Response>(
  method: HttpMethod,
  path: string,
  auth: AuthMode,
  successStatus: 200 | 201 | 204,
  errorStatuses: readonly ErrorStatus[],
  decode: Decoder<Response>,
): RouteDescriptor<Response> {
  return { method, path, auth, successStatus, errorStatuses, decode };
}

function pathPart(value: string, name: string): string {
  const normalized = value.trim();
  if (normalized.length === 0) {
    throw new TypeError(`${name} is required`);
  }
  if (backendIdentifierIssue(normalized) !== null) {
    throw new TypeError(`${name} must be a valid backend identifier`);
  }
  return encodeURIComponent(normalized);
}

function lessonsPath(query: LessonListQuery): string {
  const parameters = [
    `from=${encodeURIComponent(query.from)}`,
    `to=${encodeURIComponent(query.to)}`,
  ];
  if (query.studentId !== undefined) {
    parameters.push(`studentId=${pathPart(query.studentId, "studentId")}`);
  }
  if (query.teacherAccountId !== undefined) {
    parameters.push(
      `teacherAccountId=${pathPart(query.teacherAccountId, "teacherAccountId")}`,
    );
  }
  return `/v1/lessons?${parameters.join("&")}`;
}

function studentDirectoryPath(query: StudentDirectoryQuery): string {
  return query.asOf === undefined
    ? "/v1/students"
    : `/v1/students?asOf=${encodeURIComponent(query.asOf)}`;
}

export const routes = {
  previewActivation: route<ActivationPreview>(
    "POST",
    "/v1/activations/preview",
    "public",
    200,
    [400, 422, 429, 500],
    decodeActivationPreview,
  ),
  completeActivation: route<void>(
    "POST",
    "/v1/activations/complete",
    "public",
    204,
    [400, 409, 422, 429, 500],
    decodeVoid,
  ),
  signIn: route<SessionTokens>(
    "POST",
    "/v1/sessions",
    "public",
    200,
    [401, 422, 429, 500],
    decodeSessionTokens,
  ),
  refreshSession: route<SessionTokens>(
    "POST",
    "/v1/sessions/refresh",
    "public",
    200,
    [401, 422, 429, 500],
    decodeSessionTokens,
  ),
  signOut: route<void>(
    "DELETE",
    "/v1/sessions/current",
    "required",
    204,
    [401, 500],
    decodeVoid,
  ),
  bootstrap: route<BootstrapView>(
    "GET",
    "/v1/me/bootstrap",
    "required",
    200,
    [401, 500],
    decodeBootstrapView,
  ),
  listStaff: (role: StaffRole) =>
    route<StaffMember[]>(
      "GET",
      `/v1/staff?role=${encodeURIComponent(role)}`,
      "required",
      200,
      [401, 403, 422, 500],
      decodeStaffMembers,
    ),
  listStudents: (query: StudentDirectoryQuery = {}) =>
    route<StudentDirectoryItem[]>(
      "GET",
      studentDirectoryPath(query),
      "required",
      200,
      [401, 403, 422, 500],
      decodeStudentDirectory,
    ),
  listLessons: (query: LessonListQuery) =>
    route<Lesson[]>(
      "GET",
      lessonsPath(query),
      "required",
      200,
      [401, 403, 422, 500],
      decodeLessons,
    ),
  getLesson: (lessonId: string) =>
    route<Lesson>(
      "GET",
      `/v1/lessons/${pathPart(lessonId, "lessonId")}`,
      "required",
      200,
      [401, 403, 404, 500],
      decodeLesson,
    ),
  createLesson: route<Lesson>(
    "POST",
    "/v1/lessons",
    "required",
    201,
    [401, 403, 409, 422, 500],
    decodeLesson,
  ),
  reassignPrimaryTeachers: route<ReassignPrimaryTeachersResult>(
    "POST",
    "/v1/students/primary-teacher-reassignments",
    "required",
    201,
    [401, 403, 404, 409, 422, 500],
    decodeReassignPrimaryTeachersResult,
  ),
  replaceLessonTeachers: route<ReplaceLessonTeachersResult>(
    "POST",
    "/v1/lessons/teacher-replacements",
    "required",
    200,
    [401, 403, 404, 409, 422, 500],
    decodeReplaceLessonTeachersResult,
  ),
  listStudentOnboarding: route<StudentOnboardingItem[]>(
    "GET",
    "/v1/student-onboarding",
    "required",
    200,
    [401, 403, 422, 500],
    decodeStudentOnboardingItems,
  ),
  grantDelegation: route<DelegationResult>(
    "POST",
    "/v1/access/delegations",
    "required",
    201,
    [401, 403, 409, 422, 429, 500],
    decodeDelegationResult,
  ),
  revokeDelegation: (delegationId: string) =>
    route<void>(
      "POST",
      `/v1/access/delegations/${pathPart(delegationId, "delegationId")}/revoke`,
      "required",
      204,
      [401, 403, 404, 409, 422, 429, 500],
      decodeVoid,
    ),
  createStudent: route<StudentResult>(
    "POST",
    "/v1/students",
    "required",
    201,
    [401, 403, 409, 422, 500],
    decodeStudentResult,
  ),
  publishFirstMinute: (studentId: string) =>
    route<FirstMinute>(
      "PUT",
      `/v1/students/${pathPart(studentId, "studentId")}/first-minute`,
      "required",
      200,
      [401, 403, 404, 409, 422, 500],
      decodeFirstMinute,
    ),
  issueInvitation: (studentId: string) =>
    route<InvitationResult>(
      "POST",
      `/v1/students/${pathPart(studentId, "studentId")}/activation-invitations`,
      "required",
      201,
      [401, 403, 404, 409, 422, 500],
      decodeInvitationResult,
    ),
  reissueInvitation: (studentId: string) =>
    route<InvitationResult>(
      "POST",
      `/v1/students/${pathPart(studentId, "studentId")}/activation-invitations/reissue`,
      "required",
      201,
      [401, 403, 404, 409, 422, 500],
      decodeInvitationResult,
    ),
  revokeInvitation: (invitationId: string) =>
    route<void>(
      "POST",
      `/v1/activation-invitations/${pathPart(invitationId, "invitationId")}/revoke`,
      "required",
      204,
      [401, 403, 404, 409, 422, 500],
      decodeVoid,
    ),
} as const;

export type RouteRequestBodies = {
  previewActivation: ActivationPreviewRequest;
  completeActivation: CompleteActivationRequest;
  signIn: SignInRequest;
  refreshSession: RefreshSessionRequest;
  grantDelegation: GrantDelegationRequest;
  revokeDelegation: RevokeDelegationRequest;
  createStudent: CreateStudentRequest;
  publishFirstMinute: PublishFirstMinuteRequest;
  createLesson: CreateLessonRequest;
  reassignPrimaryTeachers: ReassignPrimaryTeachersRequest;
  replaceLessonTeachers: ReplaceLessonTeachersRequest;
};
