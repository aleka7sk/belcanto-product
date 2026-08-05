import {
  decodeActivationPreview,
  decodeActivationProgress,
  decodeRecoveryCodes,
  decodeRevokeOtherSessionsResult,
  decodeSecurityEventsPage,
  decodeSessionDevices,
  decodeSignInOutcome,
  decodeTwofaEnrollment,
  decodeTwofaStatus,
  decodeVerifiedContact,
  decodeVerifiedContacts,
  decodeDataExport,
  decodeDataExports,
  decodeDeletionRequest,
  decodePolicyVersions,
  decodePrivacySettings,
  decodeProfileView,
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
  type ActivationCodeRequest,
  type ActivationContactRequest,
  type ActivationFinishRequest,
  type ActivationPasswordRequest,
  type ActivationProgressView,
  type ActivationTokenRequest,
  type ConfirmContactChangeRequest,
  type CurrentPasswordRequest,
  type AcceptPolicyRequest,
  type DataExportRequest,
  type DeletionRequest,
  type PolicyVersion,
  type PrivacySettings,
  type ProfileView,
  type UpdateProfileRequest,
  type DisableTwofaRequest,
  type RequestPasswordResetRequest,
  type SignInOutcome,
  type StartContactChangeRequest,
  type TwofaCodeRequest,
  type TwofaEnrollment,
  type TwofaSignInRequest,
  type TwofaStatus,
  type VerifiedContact,
  type RevokeOtherSessionsResult,
  type RevokeSessionRequest,
  type CompletePasswordResetRequest,
  type SecurityEventsPage,
  type SecurityEventsQuery,
  type SessionDevice,
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

function securityEventsPath(query: SecurityEventsQuery): string {
  const parameters: string[] = [];
  if (query.cursor !== undefined) {
    const cursor = query.cursor.trim();
    if (cursor.length === 0 || cursor.length > 64) {
      throw new TypeError("cursor must contain 1 to 64 characters");
    }
    parameters.push(`cursor=${encodeURIComponent(cursor)}`);
  }
  if (query.limit !== undefined) {
    if (!Number.isSafeInteger(query.limit) || query.limit < 1 || query.limit > 50) {
      throw new TypeError("limit must be an integer between 1 and 50");
    }
    parameters.push(`limit=${query.limit}`);
  }
  const suffix = parameters.length === 0 ? "" : `?${parameters.join("&")}`;
  return `/v1/me/security-events${suffix}`;
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
  signIn: route<SignInOutcome>(
    "POST",
    "/v1/sessions",
    "public",
    200,
    [401, 422, 429, 500],
    decodeSignInOutcome,
  ),
  signInWithTwofa: route<SessionTokens>(
    "POST",
    "/v1/sessions/twofa",
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
  activationProgress: route<ActivationProgressView>(
    "POST",
    "/v1/activations/progress",
    "public",
    200,
    [400, 422, 429, 500],
    decodeActivationProgress,
  ),
  setActivationPassword: route<void>(
    "POST",
    "/v1/activations/password",
    "public",
    204,
    [400, 422, 429, 500],
    decodeVoid,
  ),
  startActivationContact: route<void>(
    "POST",
    "/v1/activations/contact",
    "public",
    204,
    [400, 409, 422, 429, 500],
    decodeVoid,
  ),
  verifyActivationContact: route<void>(
    "POST",
    "/v1/activations/contact/verify",
    "public",
    204,
    [400, 409, 422, 429, 500],
    decodeVoid,
  ),
  startActivationTwofa: route<TwofaEnrollment>(
    "POST",
    "/v1/activations/twofa",
    "public",
    200,
    [400, 409, 422, 429, 500],
    decodeTwofaEnrollment,
  ),
  confirmActivationTwofa: route<string[]>(
    "POST",
    "/v1/activations/twofa/confirm",
    "public",
    200,
    [400, 409, 422, 429, 500],
    decodeRecoveryCodes,
  ),
  finishActivation: route<void>(
    "POST",
    "/v1/activations/finish",
    "public",
    204,
    [400, 409, 422, 429, 500],
    decodeVoid,
  ),
  listMyContacts: route<VerifiedContact[]>(
    "GET",
    "/v1/me/contacts",
    "required",
    200,
    [401, 422, 500],
    decodeVerifiedContacts,
  ),
  startContactChange: route<void>(
    "POST",
    "/v1/me/contacts/change",
    "required",
    204,
    [401, 409, 422, 429, 500],
    decodeVoid,
  ),
  confirmContactChange: route<VerifiedContact>(
    "POST",
    "/v1/me/contacts/confirm",
    "required",
    200,
    [401, 409, 422, 429, 500],
    decodeVerifiedContact,
  ),
  twofaStatus: route<TwofaStatus>(
    "GET",
    "/v1/me/twofa",
    "required",
    200,
    [401, 422, 500],
    decodeTwofaStatus,
  ),
  startTwofaEnrollment: route<TwofaEnrollment>(
    "POST",
    "/v1/me/twofa/enroll",
    "required",
    200,
    [401, 409, 422, 429, 500],
    decodeTwofaEnrollment,
  ),
  confirmTwofaEnrollment: route<string[]>(
    "POST",
    "/v1/me/twofa/confirm",
    "required",
    200,
    [401, 409, 422, 429, 500],
    decodeRecoveryCodes,
  ),
  disableTwofa: route<void>(
    "POST",
    "/v1/me/twofa/disable",
    "required",
    204,
    [401, 409, 422, 429, 500],
    decodeVoid,
  ),
  requestPasswordReset: route<void>(
    "POST",
    "/v1/password-resets",
    "public",
    204,
    [422, 429, 500],
    decodeVoid,
  ),
  completePasswordReset: route<void>(
    "POST",
    "/v1/password-resets/complete",
    "public",
    204,
    [401, 422, 429, 500],
    decodeVoid,
  ),
  listMySessions: route<SessionDevice[]>(
    "GET",
    "/v1/me/sessions",
    "required",
    200,
    [401, 422, 500],
    decodeSessionDevices,
  ),
  revokeOtherSessions: route<RevokeOtherSessionsResult>(
    "POST",
    "/v1/me/sessions/revoke-others",
    "required",
    200,
    [401, 422, 429, 500],
    decodeRevokeOtherSessionsResult,
  ),
  revokeMySession: (sessionId: string) =>
    route<void>(
      "POST",
      `/v1/me/sessions/${pathPart(sessionId, "sessionId")}/revoke`,
      "required",
      204,
      [401, 404, 409, 422, 429, 500],
      decodeVoid,
    ),
  listSecurityEvents: (query: SecurityEventsQuery = {}) =>
    route<SecurityEventsPage>(
      "GET",
      securityEventsPath(query),
      "required",
      200,
      [401, 422, 500],
      decodeSecurityEventsPage,
    ),
  myProfile: route<ProfileView>(
    "GET",
    "/v1/me/profile",
    "required",
    200,
    [401, 422, 500],
    decodeProfileView,
  ),
  updateMyProfile: route<ProfileView>(
    "PUT",
    "/v1/me/profile",
    "required",
    200,
    [401, 422, 500],
    decodeProfileView,
  ),
  listPolicies: route<PolicyVersion[]>(
    "GET",
    "/v1/policies",
    "required",
    200,
    [401, 422, 500],
    decodePolicyVersions,
  ),
  acceptPolicy: route<void>(
    "POST",
    "/v1/me/policy-acceptances",
    "required",
    204,
    [401, 404, 422, 500],
    decodeVoid,
  ),
  privacySettings: route<PrivacySettings>(
    "GET",
    "/v1/me/privacy",
    "required",
    200,
    [401, 422, 500],
    decodePrivacySettings,
  ),
  updatePrivacySettings: route<PrivacySettings>(
    "PUT",
    "/v1/me/privacy",
    "required",
    200,
    [401, 409, 422, 500],
    decodePrivacySettings,
  ),
  listDataExports: route<DataExportRequest[]>(
    "GET",
    "/v1/me/data-exports",
    "required",
    200,
    [401, 422, 500],
    decodeDataExports,
  ),
  createDataExport: route<DataExportRequest>(
    "POST",
    "/v1/me/data-exports",
    "required",
    201,
    [401, 409, 422, 429, 500],
    decodeDataExport,
  ),
  deletionRequest: route<DeletionRequest>(
    "GET",
    "/v1/me/deletion-request",
    "required",
    200,
    [401, 404, 422, 500],
    decodeDeletionRequest,
  ),
  createDeletionRequest: route<DeletionRequest>(
    "POST",
    "/v1/me/deletion-request",
    "required",
    201,
    [401, 409, 422, 429, 500],
    decodeDeletionRequest,
  ),
  cancelDeletionRequest: route<DeletionRequest>(
    "POST",
    "/v1/me/deletion-request/cancel",
    "required",
    200,
    [401, 409, 422, 500],
    decodeDeletionRequest,
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
  requestPasswordReset: RequestPasswordResetRequest;
  signInWithTwofa: TwofaSignInRequest;
  activationProgress: ActivationTokenRequest;
  setActivationPassword: ActivationPasswordRequest;
  startActivationContact: ActivationContactRequest;
  verifyActivationContact: ActivationCodeRequest;
  startActivationTwofa: ActivationTokenRequest;
  confirmActivationTwofa: ActivationCodeRequest;
  finishActivation: ActivationFinishRequest;
  startContactChange: StartContactChangeRequest;
  confirmContactChange: ConfirmContactChangeRequest;
  startTwofaEnrollment: CurrentPasswordRequest;
  confirmTwofaEnrollment: TwofaCodeRequest;
  disableTwofa: DisableTwofaRequest;
  updateMyProfile: UpdateProfileRequest;
  acceptPolicy: AcceptPolicyRequest;
  updatePrivacySettings: PrivacySettings;
  createDataExport: CurrentPasswordRequest;
  createDeletionRequest: CurrentPasswordRequest;
  completePasswordReset: CompletePasswordResetRequest;
  revokeOtherSessions: RevokeSessionRequest;
  revokeMySession: RevokeSessionRequest;
  grantDelegation: GrantDelegationRequest;
  revokeDelegation: RevokeDelegationRequest;
  createStudent: CreateStudentRequest;
  publishFirstMinute: PublishFirstMinuteRequest;
  createLesson: CreateLessonRequest;
  reassignPrimaryTeachers: ReassignPrimaryTeachersRequest;
  replaceLessonTeachers: ReplaceLessonTeachersRequest;
};
