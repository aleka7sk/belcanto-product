import type { BootstrapView, SessionTokens } from "@/api/contracts";

export type SessionPhase =
  | "restoring"
  | "anonymous"
  | "authenticating"
  | "refreshing"
  | "bootstrapping"
  | "authenticated"
  | "blocked";

export type SessionOperation =
  | "restore"
  | "sign_in"
  | "refresh"
  | "bootstrap";

export interface SessionFailure {
  operation: SessionOperation;
  error: unknown;
}

export interface SessionState {
  phase: SessionPhase;
  tokens: SessionTokens | null;
  bootstrap: BootstrapView | null;
  failure: SessionFailure | null;
}

export type SessionEvent =
  | { type: "RESTORE_STARTED" }
  | { type: "SIGN_IN_STARTED" }
  | { type: "REFRESH_STARTED" }
  | { type: "TOKENS_AVAILABLE"; tokens: SessionTokens }
  | { type: "BOOTSTRAP_READY"; bootstrap: BootstrapView }
  | { type: "ANONYMOUS" }
  | { type: "FAILED"; operation: SessionOperation; error: unknown };

export const initialSessionState: SessionState = {
  phase: "restoring",
  tokens: null,
  bootstrap: null,
  failure: null,
};

export function isSessionRestoring(
  state: Pick<SessionState, "phase">,
): boolean {
  return state.phase === "restoring";
}

export function sessionReducer(
  state: SessionState,
  event: SessionEvent,
): SessionState {
  switch (event.type) {
    case "RESTORE_STARTED":
      return initialSessionState;
    case "SIGN_IN_STARTED":
      return {
        phase: "authenticating",
        tokens: null,
        bootstrap: null,
        failure: null,
      };
    case "REFRESH_STARTED":
      if (state.tokens === null) {
        return {
          phase: "anonymous",
          tokens: null,
          bootstrap: null,
          failure: null,
        };
      }
      return { ...state, phase: "refreshing", failure: null };
    case "TOKENS_AVAILABLE":
      return {
        phase: "bootstrapping",
        tokens: event.tokens,
        bootstrap: null,
        failure: null,
      };
    case "BOOTSTRAP_READY":
      if (state.tokens === null) {
        return state;
      }
      return {
        phase: "authenticated",
        tokens: state.tokens,
        bootstrap: event.bootstrap,
        failure: null,
      };
    case "ANONYMOUS":
      return {
        phase: "anonymous",
        tokens: null,
        bootstrap: null,
        failure: null,
      };
    case "FAILED":
      return {
        phase: state.tokens === null ? "anonymous" : "blocked",
        tokens: state.tokens,
        bootstrap: null,
        failure: { operation: event.operation, error: event.error },
      };
  }
}
