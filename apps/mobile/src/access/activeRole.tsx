import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";

import type { Role } from "@/api/contracts";
import type { NavigationRole } from "@/navigation/tabs";

/**
 * Active working role (HOF-03, Figma ACC-04/19/20). The account may hold
 * several roles; exactly one drives navigation and workspace context at a
 * time. Switching replaces the whole navigation context. When the chosen
 * role disappears from the bootstrap role set (access revoked), consumers
 * fall back to `resolveActiveRole` and ACC-20 explains the change.
 */

export const ROLE_PRIORITY: readonly Role[] = [
  "Student",
  "Teacher",
  "Administrator",
  "Owner",
];

export function resolveActiveRole(
  roles: readonly Role[],
  preferred: Role | null,
): Role | null {
  if (preferred !== null && roles.includes(preferred)) return preferred;
  for (const role of ROLE_PRIORITY) {
    if (roles.includes(role)) return role;
  }
  return null;
}

/** Roles map 1:1 onto navigation contexts (same identifiers by design). */
export function navigationRoleFor(role: Role): NavigationRole {
  return role;
}

interface ActiveRoleContextValue {
  preferredRole: Role | null;
  setPreferredRole: (role: Role) => void;
  /** True when the previously chosen role is gone from the bootstrap set. */
  roleRevoked: (roles: readonly Role[]) => boolean;
}

const ActiveRoleContext = createContext<ActiveRoleContextValue | null>(null);

export function ActiveRoleProvider({ children }: PropsWithChildren) {
  const [preferredRole, setPreferredRoleState] = useState<Role | null>(null);
  const [everChosen, setEverChosen] = useState(false);
  const setPreferredRole = useCallback((role: Role) => {
    setPreferredRoleState(role);
    setEverChosen(true);
  }, []);
  const roleRevoked = useCallback(
    (roles: readonly Role[]) =>
      everChosen && preferredRole !== null && !roles.includes(preferredRole),
    [everChosen, preferredRole],
  );
  const value = useMemo(
    () => ({ preferredRole, setPreferredRole, roleRevoked }),
    [preferredRole, setPreferredRole, roleRevoked],
  );
  return (
    <ActiveRoleContext.Provider value={value}>
      {children}
    </ActiveRoleContext.Provider>
  );
}

export function useActiveRole(): ActiveRoleContextValue {
  const value = useContext(ActiveRoleContext);
  if (value === null) {
    throw new Error("useActiveRole must be used within ActiveRoleProvider");
  }
  return value;
}
