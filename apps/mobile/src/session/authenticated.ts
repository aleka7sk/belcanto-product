import { ApiError } from "@/api/client";

export interface AuthenticationRecovery {
  currentAccessToken(): string | null;
  ensureAccessToken(): Promise<string>;
  refreshAccessToken(): Promise<string>;
  refreshBootstrap(): Promise<void>;
  invalidate(): Promise<void>;
}

async function refreshBootstrapBestEffort(
  recovery: AuthenticationRecovery,
): Promise<void> {
  try {
    await recovery.refreshBootstrap();
  } catch {
    // The original authorization error is the actionable result for the caller.
  }
}

async function invalidateBestEffort(
  recovery: AuthenticationRecovery,
): Promise<void> {
  try {
    await recovery.invalidate();
  } catch {
    // Session state is invalidated before persistent cleanup is attempted.
  }
}

/**
 * Runs a protected API operation with one authentication recovery attempt.
 *
 * The same operation closure is invoked on retry so callers can capture and
 * reuse their original idempotency key. A 403 is never retried: permissions
 * are refreshed for the next render while the original error is preserved.
 */
export async function runAuthenticatedOperation<Value>(
  operation: (accessToken: string) => Promise<Value>,
  recovery: AuthenticationRecovery,
): Promise<Value> {
  const attemptedToken = await recovery.ensureAccessToken();
  try {
    return await operation(attemptedToken);
  } catch (error) {
    if (!(error instanceof ApiError)) throw error;
    if (error.status === 403) {
      await refreshBootstrapBestEffort(recovery);
      throw error;
    }
    if (error.status !== 401) throw error;
  }

  const currentToken = recovery.currentAccessToken();
  const retryToken =
    currentToken !== null && currentToken !== attemptedToken
      ? currentToken
      : await recovery.refreshAccessToken();

  try {
    return await operation(retryToken);
  } catch (error) {
    if (error instanceof ApiError && error.status === 403) {
      await refreshBootstrapBestEffort(recovery);
    }
    if (error instanceof ApiError && error.status === 401) {
      await invalidateBestEffort(recovery);
    }
    throw error;
  }
}
