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

export const canReadLessons = (access: EffectiveAccess) =>
  hasPermission(access, "lessons.read");

export const canCreateLessons = (access: EffectiveAccess) =>
  hasPermission(access, "lessons.create");

export const canReplaceLessonTeachers = (access: EffectiveAccess) =>
  hasPermission(access, "lesson_teachers.replace");

export const canReassignPrimaryTeachers = (access: EffectiveAccess) =>
  hasPermission(access, "student_primary_teachers.reassign");

export const canOpenOperationalWorkspace = (access: EffectiveAccess) =>
  hasRole(access, "Owner") ||
  hasRole(access, "Administrator") ||
  hasRole(access, "Teacher") ||
  canCreateLessons(access) ||
  canReplaceLessonTeachers(access) ||
  canReassignPrimaryTeachers(access);
