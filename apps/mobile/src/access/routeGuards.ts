import type { Permission, Role } from "@/api/contracts";
import type { SessionState } from "@/session/machine";

import { hasPermission, hasRole, type EffectiveAccess } from "./capabilities";

export type RouteRequirement =
  | { kind: "public" }
  | { kind: "authenticated" }
  | { kind: "any_role"; roles: readonly Role[] }
  | { kind: "permission"; permission: Permission }
  | { kind: "all_permissions"; permissions: readonly Permission[] };

export type RouteGuardDecision =
  | "allow"
  | "pending"
  | "sign_in_required"
  | "forbidden";

const PENDING_PHASES = new Set([
  "restoring",
  "authenticating",
  "refreshing",
  "bootstrapping",
]);

export function evaluateRouteGuard(
  state: SessionState,
  requirement: RouteRequirement,
): RouteGuardDecision {
  if (requirement.kind === "public") return "allow";
  if (PENDING_PHASES.has(state.phase)) return "pending";
  if (state.phase === "anonymous") return "sign_in_required";
  if (state.phase !== "authenticated" || state.bootstrap === null) return "forbidden";

  const access: EffectiveAccess = {
    roles: state.bootstrap.roles,
    permissions: state.bootstrap.permissions,
    accessProfiles: state.bootstrap.accessProfiles,
  };
  switch (requirement.kind) {
    case "authenticated":
      return "allow";
    case "any_role":
      return requirement.roles.some((role) => hasRole(access, role))
        ? "allow"
        : "forbidden";
    case "permission":
      return hasPermission(access, requirement.permission) ? "allow" : "forbidden";
    case "all_permissions":
      return requirement.permissions.every((permission) =>
        hasPermission(access, permission),
      )
        ? "allow"
        : "forbidden";
  }
}
