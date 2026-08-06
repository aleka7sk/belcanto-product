import { router, type Href } from "expo-router";
import { useCallback, useEffect, useRef, useState } from "react";

import { navigationRoleFor, resolveActiveRole, useActiveRole } from "@/access/activeRole";
import type { Role } from "@/api/contracts";
import { useMessage } from "@/i18n";
import { useSession } from "@/session";
import type { TabKey } from "@/navigation/tabs";
import { resolveActiveTab, tabTarget } from "@/navigation/routes";

import { RoleBottomNav } from "../../roleNavigation";

/**
 * The one bottom-navigation host of the application. Every screen that
 * shows the tab bar renders this component inside the Screen scaffold's
 * navigation slot; where a tab leads is decided once, in
 * navigation/routes.ts. Tab switches replace the current entry — the
 * stack never grows from moving between tabs.
 */

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
  const navigationRole = navigationRoleFor(workingRole);
  return (
    <RoleBottomNav
      active={resolveActiveTab(navigationRole, active)}
      label={(key) => message(`nav.${key}`)}
      onSelectTab={(tab) => {
        const target = tabTarget(navigationRole, tab.key);
        if (target !== null) router.replace(target as Href);
      }}
      role={navigationRole}
    />
  );
}

/** Readable name outside the account area; the very same host. */
export const RoleNav = AccountNav;

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
  /** A reload while previous data is still shown — pull-to-refresh state. */
  refreshing: boolean;
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

  return { value, error, loading, refreshing: loading && value !== null, reload };
}

export function initialsOf(fullName: string): string {
  const parts = fullName.trim().split(/\s+/u).filter(Boolean);
  const first = parts[0]?.[0] ?? "";
  const second = parts[1]?.[0] ?? "";
  return `${first}${second}`.toUpperCase() || "•";
}
