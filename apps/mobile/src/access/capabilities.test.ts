import type { BootstrapView, IsoDateTime, SessionTokens } from "@/api/contracts";
import type { SessionState } from "@/session/machine";

import {
  canCreateStudents,
  canDelegateStudentOnboarding,
  canOpenStudentOnboardingQueue,
  type EffectiveAccess,
} from "./capabilities";
import { evaluateRouteGuard } from "./routeGuards";

const ownerWithoutPermission: EffectiveAccess = {
  roles: ["Owner"],
  permissions: [],
  accessProfiles: [],
};

const administratorWithProfileOnly: EffectiveAccess = {
  roles: ["Administrator"],
  permissions: [],
  accessProfiles: ["StudentOnboardingManager.v1"],
};

describe("effective access", () => {
  it("never infers permissions from a role or display profile", () => {
    expect(canCreateStudents(ownerWithoutPermission)).toBe(false);
    expect(canCreateStudents(administratorWithProfileOnly)).toBe(false);
    expect(canDelegateStudentOnboarding(ownerWithoutPermission)).toBe(false);
  });

  it("opens the onboarding queue for assigned Teachers or explicit readers", () => {
    expect(
      canOpenStudentOnboardingQueue({
        roles: ["Teacher"],
        permissions: [],
        accessProfiles: [],
      }),
    ).toBe(true);
    expect(
      canOpenStudentOnboardingQueue({
        roles: ["Administrator"],
        permissions: ["student_onboarding.read"],
        accessProfiles: [],
      }),
    ).toBe(true);
    expect(canOpenStudentOnboardingQueue(administratorWithProfileOnly)).toBe(
      false,
    );
  });

  it("gates routes by exact bootstrap permission literals", () => {
    const tokens: SessionTokens = {
      accessToken: "A".repeat(43),
      refreshToken: "R".repeat(43),
      accessExpiresAt: "2026-08-01T10:00:00Z" as IsoDateTime,
      refreshExpiresAt: "2026-09-01T10:00:00Z" as IsoDateTime,
    };
    const bootstrap: BootstrapView = {
      accountId: "account_1",
      roles: ["Administrator"],
      permissions: ["students.create"],
      accessProfiles: [],
    };
    const state: SessionState = {
      phase: "authenticated",
      tokens,
      bootstrap,
      failure: null,
    };
    expect(
      evaluateRouteGuard(state, {
        kind: "permission",
        permission: "students.create",
      }),
    ).toBe("allow");
    expect(
      evaluateRouteGuard(state, {
        kind: "permission",
        permission: "student_invitations.issue",
      }),
    ).toBe("forbidden");
  });
});
