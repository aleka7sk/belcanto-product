import { ApiError } from "@/api/client";

import {
  runAuthenticatedOperation,
  type AuthenticationRecovery,
} from "./authenticated";

function recovery(
  overrides: Partial<AuthenticationRecovery> = {},
): AuthenticationRecovery {
  return {
    currentAccessToken: () => "access-1",
    ensureAccessToken: async () => "access-1",
    refreshAccessToken: async () => "access-2",
    refreshBootstrap: async () => undefined,
    invalidate: async () => undefined,
    ...overrides,
  };
}

describe("authenticated operation recovery", () => {
  it("refreshes once and retries the same operation after a 401", async () => {
    const operation = jest
      .fn<Promise<string>, [string]>()
      .mockRejectedValueOnce(new ApiError(401, "UNAUTHENTICATED", "expired"))
      .mockResolvedValueOnce("done");
    const refreshAccessToken = jest.fn(async () => "access-2");

    await expect(
      runAuthenticatedOperation(
        operation,
        recovery({ refreshAccessToken }),
      ),
    ).resolves.toBe("done");

    expect(operation.mock.calls).toEqual([["access-1"], ["access-2"]]);
    expect(refreshAccessToken).toHaveBeenCalledTimes(1);
  });

  it("uses a token refreshed by another request without a second refresh", async () => {
    const operation = jest
      .fn<Promise<string>, [string]>()
      .mockRejectedValueOnce(new ApiError(401, "UNAUTHENTICATED", "expired"))
      .mockResolvedValueOnce("done");
    const refreshAccessToken = jest.fn(async () => "unexpected");

    await expect(
      runAuthenticatedOperation(
        operation,
        recovery({
          currentAccessToken: () => "access-concurrent",
          refreshAccessToken,
        }),
      ),
    ).resolves.toBe("done");

    expect(operation.mock.calls).toEqual([
      ["access-1"],
      ["access-concurrent"],
    ]);
    expect(refreshAccessToken).not.toHaveBeenCalled();
  });

  it("invalidates the session after the single retry also returns 401", async () => {
    const error = new ApiError(401, "UNAUTHENTICATED", "revoked");
    const operation = jest.fn(async () => {
      throw error;
    });
    const invalidate = jest.fn(async () => undefined);

    await expect(
      runAuthenticatedOperation(operation, recovery({ invalidate })),
    ).rejects.toBe(error);

    expect(operation).toHaveBeenCalledTimes(2);
    expect(invalidate).toHaveBeenCalledTimes(1);
  });

  it("refreshes permissions but does not retry a forbidden operation", async () => {
    const error = new ApiError(403, "FORBIDDEN", "permission revoked");
    const operation = jest.fn(async () => {
      throw error;
    });
    const refreshBootstrap = jest.fn(async () => undefined);

    await expect(
      runAuthenticatedOperation(
        operation,
        recovery({ refreshBootstrap }),
      ),
    ).rejects.toBe(error);

    expect(operation).toHaveBeenCalledTimes(1);
    expect(refreshBootstrap).toHaveBeenCalledTimes(1);
  });
});
