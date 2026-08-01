import {
  decodeActivationPreview,
  decodeBootstrapView,
  decodeDelegationResult,
  decodeFirstMinute,
  decodeInvitationResult,
  decodeSessionTokens,
  decodeStaffMembers,
  decodeStudentOnboardingItems,
  decodeStudentResult,
  decodeVoid,
  type ActivationPreview,
  type ActivationPreviewRequest,
  type BootstrapView,
  type CompleteActivationRequest,
  type CreateStudentRequest,
  type Decoder,
  type DelegationResult,
  type FirstMinute,
  type GrantDelegationRequest,
  type InvitationResult,
  type PublishFirstMinuteRequest,
  type RefreshSessionRequest,
  type RevokeDelegationRequest,
  type SessionTokens,
  type SignInRequest,
  type StaffMember,
  type StaffRole,
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
};
