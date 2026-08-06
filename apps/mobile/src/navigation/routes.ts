import type { NavigationRole, TabKey } from "./tabs";
import { tabsForRole } from "./tabs";

/**
 * The one tab→route map of the application. Every screen hosts the same
 * role-aware bottom navigation; where a tab leads must be decided once,
 * here — not re-implemented per screen. Targets follow the production
 * information architecture (Pages 21–30):
 *
 * - the Student schedule is the personal day view, the staff schedule is
 *   the internal lessons workspace — the same tab key routes by role, so
 *   no role ever lands on a guard it cannot leave;
 * - the Teacher «today» is the Page-26 cockpit, not the generic staff
 *   hub;
 * - «people» opens the staff workspace explicitly (workspace=staff), so
 *   a dual-role account is not bounced to the student home.
 */
export interface TabTarget {
  pathname: string;
  params?: Record<string, string>;
}

const STUDENT_TARGETS: Partial<Record<TabKey, TabTarget>> = {
  today: { pathname: "/(protected)", params: { workspace: "student" } },
  schedule: { pathname: "/(protected)/schedule" },
  practice: { pathname: "/(protected)/practice" },
  community: { pathname: "/(protected)/community" },
  profile: { pathname: "/(protected)/account" },
};

const TEACHER_TARGETS: Partial<Record<TabKey, TabTarget>> = {
  today: { pathname: "/(protected)/teacher" },
  schedule: { pathname: "/(protected)/lessons" },
  students: { pathname: "/(protected)/teacher/students" },
  review: { pathname: "/(protected)/teacher/students", params: { segment: "review" } },
  community: { pathname: "/(protected)/community" },
};

const ADMINISTRATOR_TARGETS: Partial<Record<TabKey, TabTarget>> = {
  operations: { pathname: "/(protected)/operations" },
  schedule: { pathname: "/(protected)/lessons" },
  people: { pathname: "/(protected)", params: { workspace: "staff" } },
  community: { pathname: "/(protected)/community" },
  more: { pathname: "/(protected)/account" },
};

const OWNER_TARGETS: Partial<Record<TabKey, TabTarget>> = {
  overview: { pathname: "/(protected)/overview" },
  analytics: { pathname: "/(protected)/overview", params: { segment: "analytics" } },
  operations: { pathname: "/(protected)/operations" },
  team: { pathname: "/(protected)/access" },
  more: { pathname: "/(protected)/account" },
};

const TARGETS: Record<NavigationRole, Partial<Record<TabKey, TabTarget>>> = {
  Student: STUDENT_TARGETS,
  Teacher: TEACHER_TARGETS,
  Administrator: ADMINISTRATOR_TARGETS,
  Owner: OWNER_TARGETS,
};

export function tabTarget(role: NavigationRole, key: TabKey): TabTarget | null {
  return TARGETS[role][key] ?? null;
}

/**
 * Resolve which tab a screen should highlight for the active role.
 * The account/profile area is «profile» for the Student and «more» for
 * the Administrator and Owner; a role without either key highlights
 * nothing rather than lying (the Teacher has no account tab).
 */
export function resolveActiveTab(role: NavigationRole, requested: TabKey): TabKey {
  const keys = new Set(tabsForRole(role).map((tab) => tab.key));
  if (keys.has(requested)) return requested;
  if (requested === "profile" && keys.has("more")) return "more";
  if (requested === "more" && keys.has("profile")) return "profile";
  return requested;
}
