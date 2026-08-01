import type {
  CapabilityBundle,
  Permission,
  Role,
} from "@/api/contracts";

export interface EffectiveAccess {
  roles: readonly Role[];
  permissions: readonly Permission[];
  accessProfiles: readonly CapabilityBundle[];
}

export function hasRole(access: EffectiveAccess, role: Role): boolean {
  return access.roles.includes(role);
}

export function hasPermission(
  access: EffectiveAccess,
  permission: Permission,
): boolean {
  return access.permissions.includes(permission);
}

export const canCreateStudents = (access: EffectiveAccess) =>
  hasPermission(access, "students.create");

export const canReadStudentOnboarding = (access: EffectiveAccess) =>
  hasPermission(access, "student_onboarding.read");

export const canOpenStudentOnboardingQueue = (access: EffectiveAccess) =>
  hasRole(access, "Teacher") || canReadStudentOnboarding(access);

export const canIssueStudentInvitations = (access: EffectiveAccess) =>
  hasPermission(access, "student_invitations.issue");

export const canReissueStudentInvitations = (access: EffectiveAccess) =>
  hasPermission(access, "student_invitations.reissue");

export const canRevokeStudentInvitations = (access: EffectiveAccess) =>
  hasPermission(access, "student_invitations.revoke");

export const canDelegateStudentOnboarding = (access: EffectiveAccess) =>
  hasPermission(access, "student_onboarding.delegate");
