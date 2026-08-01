import type { BootstrapView, IsoDateTime, SessionTokens } from "@/api/contracts";

import { initialSessionState, sessionReducer } from "./machine";
import {
  createKeyedSingleFlight,
  createOperationEpoch,
  createSerialExecutor,
  createSingleFlight,
} from "./singleFlight";

const tokens: SessionTokens = {
  accessToken: "A".repeat(43),
  refreshToken: "R".repeat(43),
  accessExpiresAt: "2026-08-01T10:00:00Z" as IsoDateTime,
  refreshExpiresAt: "2026-09-01T10:00:00Z" as IsoDateTime,
};

const bootstrap: BootstrapView = {
  accountId: "account_1",
  roles: ["Owner"],
  accessProfiles: [],
  permissions: [
    "students.create",
    "student_onboarding.read",
    "student_invitations.issue",
    "student_invitations.reissue",
    "student_invitations.revoke",
    "student_onboarding.delegate",
  ],
};

describe("session state machine", () => {
  it("does not authenticate until bootstrap is available", () => {
    const withTokens = sessionReducer(initialSessionState, {
      type: "TOKENS_AVAILABLE",
      tokens,
    });
    expect(withTokens.phase).toBe("bootstrapping");
    expect(
      sessionReducer(withTokens, { type: "BOOTSTRAP_READY", bootstrap }),
    ).toMatchObject({ phase: "authenticated", tokens, bootstrap });
  });

  it("blocks rather than exposing protected state after bootstrap failure", () => {
    const withTokens = sessionReducer(initialSessionState, {
      type: "TOKENS_AVAILABLE",
      tokens,
    });
    expect(
      sessionReducer(withTokens, {
        type: "FAILED",
        operation: "bootstrap",
        error: new Error("offline"),
      }).phase,
    ).toBe("blocked");
  });

  it("coalesces simultaneous refresh operations", async () => {
    const flight = createSingleFlight<number>();
    let calls = 0;
    let release: ((value: number) => void) | undefined;
    const operation = () => {
      calls += 1;
      return new Promise<number>((resolve) => {
        release = resolve;
      });
    };
    const first = flight.run(operation);
    const second = flight.run(operation);
    expect(first).toBe(second);
    expect(calls).toBe(0);
    await Promise.resolve();
    expect(calls).toBe(1);
    release?.(7);
    await expect(Promise.all([first, second])).resolves.toEqual([7, 7]);
    expect(flight.isRunning()).toBe(false);
  });

  it("queues a new refresh epoch behind a stale in-flight epoch", async () => {
    const flight = createKeyedSingleFlight<number, number>();
    const releases: ((value: number) => void)[] = [];
    const calls: number[] = [];
    const operation = (epoch: number) => () => {
      calls.push(epoch);
      return new Promise<number>((resolve) => releases.push(resolve));
    };
    const stale = flight.run(1, operation(1));
    const current = flight.run(2, operation(2));
    await Promise.resolve();
    expect(calls).toEqual([1]);
    releases[0]?.(1);
    await stale;
    await Promise.resolve();
    expect(calls).toEqual([1, 2]);
    releases[1]?.(2);
    await expect(current).resolves.toBe(2);
  });

  it("cleans up a rejected flight and permits a retry", async () => {
    const flight = createSingleFlight<number>();
    await expect(
      flight.run(async () => {
        throw new Error("refresh failed");
      }),
    ).rejects.toThrow("refresh failed");
    expect(flight.isRunning()).toBe(false);
    await expect(flight.run(async () => 2)).resolves.toBe(2);
  });

  it("invalidates stale async session operations", () => {
    const epoch = createOperationEpoch();
    const restore = epoch.current();
    const signIn = epoch.begin();
    expect(epoch.isCurrent(restore)).toBe(false);
    expect(epoch.isCurrent(signIn)).toBe(true);
    const signOut = epoch.begin();
    expect(epoch.isCurrent(signIn)).toBe(false);
    expect(epoch.isCurrent(signOut)).toBe(true);
  });

  it("serializes secure-store mutations in call order", async () => {
    const executor = createSerialExecutor();
    const events: string[] = [];
    let releaseSave: (() => void) | undefined;
    const save = executor.run(
      () =>
        new Promise<void>((resolve) => {
          events.push("save:start");
          releaseSave = resolve;
        }),
    );
    const clear = executor.run(async () => {
      events.push("clear");
    });
    await Promise.resolve();
    expect(events).toEqual(["save:start"]);
    releaseSave?.();
    await Promise.all([save, clear]);
    expect(events).toEqual(["save:start", "clear"]);
  });

  it("keeps sign-out clear ordered after an already-started token save", async () => {
    const executor = createSerialExecutor();
    const epoch = createOperationEpoch();
    const refreshEpoch = epoch.current();
    let persisted: string | null = null;
    let releaseSave: (() => void) | undefined;
    const save = executor.run(
      () =>
        new Promise<void>((resolve) => {
          persisted = "refreshed";
          releaseSave = resolve;
        }),
    );
    epoch.begin();
    const clear = executor.run(async () => {
      persisted = null;
    });
    await Promise.resolve();
    releaseSave?.();
    await Promise.all([save, clear]);
    expect(epoch.isCurrent(refreshEpoch)).toBe(false);
    expect(persisted).toBeNull();
  });
});
