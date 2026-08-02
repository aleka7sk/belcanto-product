import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
  useRef,
  type PropsWithChildren,
} from "react";
import { AppState } from "react-native";

import {
  ApiClient,
  ApiError,
  type SessionTokens,
  type SignInRequest,
} from "@/api";
import { interpretBootstrap } from "@/controllers/bootstrap";

import { runAuthenticatedOperation } from "./authenticated";
import { initialSessionState, sessionReducer, type SessionState } from "./machine";
import {
  createKeyedSingleFlight,
  createOperationEpoch,
  createSerialExecutor,
} from "./singleFlight";
import {
  createSessionStore,
  isAccessTokenUsable,
  isRefreshTokenUsable,
  type SessionStore,
} from "./store";

export interface SessionContextValue {
  state: SessionState;
  signIn(request: SignInRequest): Promise<void>;
  refresh(): Promise<void>;
  retryBootstrap(): Promise<void>;
  runAuthenticated<Value>(
    operation: (accessToken: string) => Promise<Value>,
  ): Promise<Value>;
  signOut(): Promise<void>;
}

const SessionContext = createContext<SessionContextValue | null>(null);
const systemNow = () => new Date();

export interface SessionProviderProps extends PropsWithChildren {
  api: ApiClient;
  store?: SessionStore;
  now?: () => Date;
}

export function SessionProvider({
  api,
  children,
  store: suppliedStore,
  now = systemNow,
}: SessionProviderProps) {
  const store = useMemo(() => suppliedStore ?? createSessionStore(), [suppliedStore]);
  const [state, dispatch] = useReducer(sessionReducer, initialSessionState);
  const stateRef = useRef(state);
  const tokensRef = useRef<SessionTokens | null>(null);
  const bootstrapFlightRef = useRef(createKeyedSingleFlight<string, void>());
  const refreshFlightRef = useRef(
    createKeyedSingleFlight<number, SessionTokens>(),
  );
  const operationEpochRef = useRef(createOperationEpoch());
  const storeMutationsRef = useRef(createSerialExecutor());

  useEffect(() => {
    stateRef.current = state;
  }, [state]);

  const clearStoredSession = useCallback(async () => {
    try {
      await storeMutationsRef.current.run(() => store.clear());
    } catch {
      // In-memory state is authoritative. Persistent cleanup remains best effort.
    }
  }, [store]);

  const invalidateSession = useCallback(async () => {
    operationEpochRef.current.begin();
    tokensRef.current = null;
    dispatch({ type: "ANONYMOUS" });
    await clearStoredSession();
  }, [clearStoredSession]);

  const loadBootstrap = useCallback(
    async (
      accessToken: string,
      signal?: AbortSignal,
      epoch = operationEpochRef.current.current(),
    ): Promise<void> => {
      if (tokensRef.current?.accessToken !== accessToken) return;
      await bootstrapFlightRef.current.run(`${epoch}:${accessToken}`, async () => {
        const stale =
          !operationEpochRef.current.isCurrent(epoch) ||
          tokensRef.current?.accessToken !== accessToken;
        if (stale) return;
        let bootstrap: Awaited<ReturnType<ApiClient["bootstrap"]>>;
        try {
          bootstrap = await api.bootstrap(accessToken, signal);
        } catch (error) {
          const failedRequestIsStale =
            !operationEpochRef.current.isCurrent(epoch) ||
            tokensRef.current?.accessToken !== accessToken;
          if (failedRequestIsStale) return;
          throw error;
        }
        const completedRequestIsStale =
          !operationEpochRef.current.isCurrent(epoch) ||
          tokensRef.current?.accessToken !== accessToken;
        if (completedRequestIsStale) return;
        const interpretation = interpretBootstrap(bootstrap);
        if (!interpretation.ready) {
          throw new Error(`Bootstrap invariant failed: ${interpretation.reason}`);
        }
        dispatch({ type: "BOOTSTRAP_READY", bootstrap: interpretation.view });
      });
    },
    [api],
  );

  const persistTokens = useCallback(
    async (tokens: SessionTokens, epoch: number): Promise<boolean> => {
      if (!operationEpochRef.current.isCurrent(epoch)) return false;
      await storeMutationsRef.current.run(() => store.save(tokens));
      if (!operationEpochRef.current.isCurrent(epoch)) return false;
      tokensRef.current = tokens;
      dispatch({ type: "TOKENS_AVAILABLE", tokens });
      return true;
    },
    [store],
  );

  const refreshTokens = useCallback(
    (
      tokens: SessionTokens,
      operation: "restore" | "refresh",
      signal?: AbortSignal,
      epoch = operationEpochRef.current.current(),
    ): Promise<SessionTokens> =>
      refreshFlightRef.current.run(epoch, async () => {
        if (!operationEpochRef.current.isCurrent(epoch)) {
          throw new Error("Session operation was superseded");
        }
        dispatch({ type: "REFRESH_STARTED" });

        let refreshed: SessionTokens;
        try {
          refreshed = await api.refreshSession(
            { refreshToken: tokens.refreshToken },
            signal,
          );
        } catch (error) {
          if (!operationEpochRef.current.isCurrent(epoch)) throw error;
          if (error instanceof ApiError && error.status === 401) {
            tokensRef.current = null;
            dispatch({ type: "ANONYMOUS" });
            await clearStoredSession();
          } else {
            dispatch({ type: "FAILED", operation, error });
          }
          throw error;
        }

        try {
          const persisted = await persistTokens(refreshed, epoch);
          if (!persisted) {
            throw new Error("Session operation was superseded");
          }
        } catch (error) {
          if (operationEpochRef.current.isCurrent(epoch)) {
            tokensRef.current = null;
            dispatch({ type: "ANONYMOUS" });
            dispatch({ type: "FAILED", operation, error });
            await clearStoredSession();
          }
          throw error;
        }

        try {
          await loadBootstrap(refreshed.accessToken, signal, epoch);
        } catch (error) {
          if (operationEpochRef.current.isCurrent(epoch)) {
            dispatch({ type: "FAILED", operation: "bootstrap", error });
          }
          throw error;
        }
        return refreshed;
      }),
    [api, clearStoredSession, loadBootstrap, persistTokens],
  );

  useEffect(() => {
    const abortController = new AbortController();
    const operationEpoch = operationEpochRef.current;
    let active = true;
    const epoch = operationEpoch.current();

    const restore = async () => {
      dispatch({ type: "RESTORE_STARTED" });
      let tokens: SessionTokens | null;
      try {
        tokens = await store.load();
      } catch (error) {
        if (!active || !operationEpoch.isCurrent(epoch)) return;
        tokensRef.current = null;
        dispatch({ type: "FAILED", operation: "restore", error });
        await clearStoredSession();
        return;
      }

      if (!active || !operationEpoch.isCurrent(epoch)) return;
      if (tokens === null || !isRefreshTokenUsable(tokens, now())) {
        tokensRef.current = null;
        dispatch({ type: "ANONYMOUS" });
        await clearStoredSession();
        return;
      }

      tokensRef.current = tokens;
      dispatch({ type: "TOKENS_AVAILABLE", tokens });
      if (!isAccessTokenUsable(tokens, now())) {
        try {
          await refreshTokens(tokens, "restore", abortController.signal, epoch);
        } catch {
          // refreshTokens has already moved the state to its safe failure phase.
        }
        return;
      }

      try {
        await loadBootstrap(tokens.accessToken, abortController.signal, epoch);
      } catch (error) {
        if (active && operationEpoch.isCurrent(epoch)) {
          dispatch({ type: "FAILED", operation: "bootstrap", error });
        }
      }
    };

    void restore();
    return () => {
      active = false;
      abortController.abort();
      if (operationEpoch.isCurrent(epoch)) {
        operationEpoch.begin();
      }
    };
  }, [clearStoredSession, loadBootstrap, now, refreshTokens, store]);

  const signIn = useCallback(
    async (request: SignInRequest) => {
      const epoch = operationEpochRef.current.begin();
      tokensRef.current = null;
      dispatch({ type: "SIGN_IN_STARTED" });

      let tokens: SessionTokens;
      try {
        tokens = await api.signIn(request);
        const persisted = await persistTokens(tokens, epoch);
        if (!persisted) return;
      } catch (error) {
        if (!operationEpochRef.current.isCurrent(epoch)) return;
        tokensRef.current = null;
        dispatch({ type: "ANONYMOUS" });
        dispatch({ type: "FAILED", operation: "sign_in", error });
        await clearStoredSession();
        throw error;
      }

      try {
        await loadBootstrap(tokens.accessToken, undefined, epoch);
      } catch (error) {
        if (!operationEpochRef.current.isCurrent(epoch)) return;
        dispatch({ type: "FAILED", operation: "bootstrap", error });
        throw error;
      }
    },
    [api, clearStoredSession, loadBootstrap, persistTokens],
  );

  const ensureAccessToken = useCallback(async (): Promise<string> => {
    const tokens = tokensRef.current;
    if (tokens === null || !isRefreshTokenUsable(tokens, now())) {
      await invalidateSession();
      throw new ApiError(401, "UNAUTHENTICATED", "Session is unavailable");
    }
    if (isAccessTokenUsable(tokens, now())) return tokens.accessToken;
    const epoch = operationEpochRef.current.current();
    const refreshed = await refreshTokens(tokens, "refresh", undefined, epoch);
    return refreshed.accessToken;
  }, [invalidateSession, now, refreshTokens]);

  const refreshAccessToken = useCallback(async (): Promise<string> => {
    const tokens = tokensRef.current;
    if (tokens === null || !isRefreshTokenUsable(tokens, now())) {
      await invalidateSession();
      throw new ApiError(401, "UNAUTHENTICATED", "Session is unavailable");
    }
    const epoch = operationEpochRef.current.current();
    const refreshed = await refreshTokens(tokens, "refresh", undefined, epoch);
    return refreshed.accessToken;
  }, [invalidateSession, now, refreshTokens]);

  const refreshBootstrap = useCallback(async (): Promise<void> => {
    const tokens = tokensRef.current;
    if (tokens === null || !isRefreshTokenUsable(tokens, now())) {
      await invalidateSession();
      throw new ApiError(401, "UNAUTHENTICATED", "Session is unavailable");
    }
    const epoch = operationEpochRef.current.current();
    if (!isAccessTokenUsable(tokens, now())) {
      await refreshTokens(tokens, "refresh", undefined, epoch);
      return;
    }
    try {
      await loadBootstrap(tokens.accessToken, undefined, epoch);
    } catch (error) {
      if (operationEpochRef.current.isCurrent(epoch)) {
        dispatch({ type: "FAILED", operation: "bootstrap", error });
      }
      throw error;
    }
  }, [invalidateSession, loadBootstrap, now, refreshTokens]);

  const runAuthenticated = useCallback(
    <Value,>(
      operation: (accessToken: string) => Promise<Value>,
    ): Promise<Value> =>
      runAuthenticatedOperation(operation, {
        currentAccessToken: () => tokensRef.current?.accessToken ?? null,
        ensureAccessToken,
        refreshAccessToken,
        refreshBootstrap,
        invalidate: invalidateSession,
      }),
    [ensureAccessToken, invalidateSession, refreshAccessToken, refreshBootstrap],
  );

  const refresh = useCallback(async () => {
    await refreshAccessToken();
  }, [refreshAccessToken]);

  const retryBootstrap = useCallback(async () => {
    await refreshBootstrap();
  }, [refreshBootstrap]);

  useEffect(() => {
    let previousState = AppState.currentState;
    const synchronize = async () => {
      const pending = [
        "restoring",
        "authenticating",
        "refreshing",
        "bootstrapping",
      ].includes(stateRef.current.phase);
      if (pending || tokensRef.current === null) return;
      try {
        await refreshBootstrap();
      } catch {
        // The refresh/bootstrap path has already established a safe state.
      }
    };
    const subscription = AppState.addEventListener("change", (nextState) => {
      const becameActive = previousState !== "active" && nextState === "active";
      previousState = nextState;
      if (becameActive) void synchronize();
    });
    return () => subscription.remove();
  }, [refreshBootstrap]);

  const signOut = useCallback(async () => {
    const accessToken = tokensRef.current?.accessToken;
    operationEpochRef.current.begin();
    tokensRef.current = null;
    dispatch({ type: "ANONYMOUS" });
    await clearStoredSession();
    if (accessToken !== undefined) {
      try {
        await api.signOut(accessToken);
      } catch {
        // Local credentials are already removed. Remote expiry remains the fallback.
      }
    }
  }, [api, clearStoredSession]);

  const value = useMemo<SessionContextValue>(
    () => ({
      state,
      signIn,
      refresh,
      retryBootstrap,
      runAuthenticated,
      signOut,
    }),
    [refresh, retryBootstrap, runAuthenticated, signIn, signOut, state],
  );

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession(): SessionContextValue {
  const context = useContext(SessionContext);
  if (context === null) {
    throw new Error("useSession must be used within SessionProvider");
  }
  return context;
}
