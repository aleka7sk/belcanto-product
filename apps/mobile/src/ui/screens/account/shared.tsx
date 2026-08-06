import { router } from "expo-router";
import { useCallback, useEffect, useRef, useState } from "react";

import { navigationRoleFor, resolveActiveRole, useActiveRole } from "@/access/activeRole";
import type { Role } from "@/api/contracts";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import type { TabKey } from "@/navigation/tabs";

import { RoleBottomNav } from "../../roleNavigation";

/**
 * Shared scaffolding for the Account & Security area (Page 32). The
 * bottom navigation always reflects the active working role; modules that
 * are not built yet stay visible but inert (HOF-03 — no empty slots).
 */

const BUILT_TABS: ReadonlySet<TabKey> = new Set([
  "today",
  "schedule",
  "practice",
  "community",
  "profile",
]);

export function useWorkingRole(): Role | null {
  const { state } = useSession();
  const { preferredRole } = useActiveRole();
  const roles = state.bootstrap?.roles ?? [];
  return resolveActiveRole(roles, preferredRole);
}

export function AccountNav({ active = "profile" }: { active?: TabKey }) {
  const message = useMessage();
  const workingRole = useWorkingRole();
  if (workingRole === null) return null;
  return (
    <RoleBottomNav
      active={active}
      isTabEnabled={(tab) => BUILT_TABS.has(tab.key)}
      label={(key) => message(`nav.${key}`)}
      onSelectTab={(tab) => {
        if (tab.key === "today") router.push("/(protected)");
        if (tab.key === "schedule") router.push("/(protected)/schedule");
        if (tab.key === "practice") router.push("/(protected)/practice");
        if (tab.key === "community") router.push("/(protected)/community");
        if (tab.key === "profile") router.push("/(protected)/account");
      }}
      role={navigationRoleFor(workingRole)}
    />
  );
}

/**
 * Load an authenticated resource on focus with reload support. Stale
 * responses are dropped when a newer load has started.
 */
export function useAccountResource<Value>(
  load: (accessToken: string) => Promise<Value>,
): {
  value: Value | null;
  error: unknown;
  loading: boolean;
  reload: () => Promise<void>;
} {
  const { runAuthenticated } = useSession();
  const [value, setValue] = useState<Value | null>(null);
  const [error, setError] = useState<unknown>(null);
  const [loading, setLoading] = useState(true);
  const generation = useRef(0);
  const loadRef = useRef<typeof load | null>(null);

  useEffect(() => {
    loadRef.current = load;
  });

  const reload = useCallback(async () => {
    const ticket = ++generation.current;
    setLoading(true);
    setError(null);
    try {
      const next = await runAuthenticated(async (accessToken) => {
        const currentLoad = loadRef.current;
        if (currentLoad === null) throw new Error("resource loader not ready");
        return currentLoad(accessToken);
      });
      if (generation.current === ticket) setValue(next);
    } catch (cause) {
      if (generation.current === ticket) setError(cause);
    } finally {
      if (generation.current === ticket) setLoading(false);
    }
  }, [runAuthenticated]);

  useEffect(() => {
    let active = true;
    queueMicrotask(() => {
      if (active) void reload();
    });
    return () => {
      active = false;
    };
  }, [reload]);

  return { value, error, loading, reload };
}

export function initialsOf(fullName: string): string {
  const parts = fullName.trim().split(/\s+/u).filter(Boolean);
  const first = parts[0]?.[0] ?? "";
  const second = parts[1]?.[0] ?? "";
  return `${first}${second}`.toUpperCase() || "•";
}
