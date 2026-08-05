import type { IconName } from "../ui/icons";

/**
 * Role-aware Bottom Navigation contract from Figma component
 * «Bottom Navigation · Production» (310:20542, Page 35): four roles,
 * five positional slots each, twenty variants total. Active is positional
 * within the role mapping; role switching replaces the whole navigation
 * context and unavailable modules never leave empty tabs (HOF-03).
 */
export type NavigationRole = "Student" | "Teacher" | "Administrator" | "Owner";

export type TabSlot = 1 | 2 | 3 | 4 | 5;

export type TabKey =
  | "today"
  | "schedule"
  | "practice"
  | "community"
  | "profile"
  | "students"
  | "review"
  | "operations"
  | "people"
  | "more"
  | "overview"
  | "analytics"
  | "team";

export type NavLabelKey = `nav.${TabKey}`;

export type TabDefinition = {
  slot: TabSlot;
  key: TabKey;
  icon: IconName;
  labelKey: NavLabelKey;
};

function tab(slot: TabSlot, key: TabKey, icon: IconName): TabDefinition {
  return { slot, key, icon, labelKey: `nav.${key}` };
}

/**
 * Exact icon assignment per variant set 310:20542 — including the
 * design's deliberate quirks (Teacher/Administrator community uses the
 * mic glyph, Owner analytics uses the trophy glyph).
 */
export const ROLE_TABS: Record<NavigationRole, readonly TabDefinition[]> = {
  Student: [
    tab(1, "today", "home"),
    tab(2, "schedule", "calendar"),
    tab(3, "practice", "mic"),
    tab(4, "community", "users"),
    tab(5, "profile", "trophy"),
  ],
  Teacher: [
    tab(1, "today", "home"),
    tab(2, "schedule", "calendar"),
    tab(3, "students", "users"),
    tab(4, "review", "check"),
    tab(5, "community", "mic"),
  ],
  Administrator: [
    tab(1, "operations", "home"),
    tab(2, "schedule", "calendar"),
    tab(3, "people", "users"),
    tab(4, "community", "mic"),
    tab(5, "more", "more"),
  ],
  Owner: [
    tab(1, "overview", "home"),
    tab(2, "analytics", "trophy"),
    tab(3, "operations", "calendar"),
    tab(4, "team", "users"),
    tab(5, "more", "more"),
  ],
} as const;

export function tabsForRole(role: NavigationRole): readonly TabDefinition[] {
  return ROLE_TABS[role];
}

export function findTab(role: NavigationRole, key: TabKey): TabDefinition | null {
  return ROLE_TABS[role].find((definition) => definition.key === key) ?? null;
}
