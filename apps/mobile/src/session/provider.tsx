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

import { ApiClient, type SignInRequest } from "@/api";
import { interpretBootstrap } from "@/controllers/bootstrap";

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
  const refreshFlightRef = useRef(createKeyedSingleFlight<number, void>());
  const operationEpochRef = useRef(createOperationEpoch());
  const storeMutationsRef = useRef(createSerialExecutor());
  useEffect(() => {
    stateRef.current = state;
  }, [state]);

  const loadBootstrap = useCallback(
    async (
      accessToken: string,
      signal?: AbortSignal,
      epoch = operationEpochRef.current.current(),
    ) => {
      try {
        const bootstrap = await api.bootstrap(accessToken, signal);
        if (!operationEpochRef.current.isCurrent(epoch)) return;
        const interpretation = interpretBootstrap(bootstrap);
        if (!interpretation.ready) {
          throw new Error(`Bootstrap invariant failed: ${interpretation.reason}`);
        }
        dispatch({ type: "BOOTSTRAP_READY", bootstrap: interpretation.view });
      } catch (error) {
        if (operationEpochRef.current.isCurrent(epoch)) {
          dispatch({ type: "FAILED", operation: "bootstrap", error });
        }
      }
    },
    [api],
  );

  const acceptTokens = useCallback(
    async (
      tokens: Awaited<ReturnType<ApiClient["signIn"]>>,
      signal: AbortSignal | undefined,
      epoch: number,
    ) => {
      if (!operationEpochRef.current.isCurrent(epoch)) return;
      await storeMutationsRef.current.run(() => store.save(tokens));
      if (!operationEpochRef.current.isCurrent(epoch)) return;
      dispatch({ type: "TOKENS_AVAILABLE", tokens });
      await loadBootstrap(tokens.accessToken, signal, epoch);
    },
    [loadBootstrap, store],
  );

  const refreshTokens = useCallback(
    (
      tokens: Awaited<ReturnType<ApiClient["signIn"]>>,
      operation: "restore" | "refresh",
      signal?: AbortSignal,
      epoch = operationEpochRef.current.current(),
    ) =>
      refreshFlightRef.current.run(epoch, async () => {
        if (!operationEpochRef.current.isCurrent(epoch)) return;
        dispatch({ type: "REFRESH_STARTED" });
        try {
          const refreshed = await api.refreshSession(
            { refreshToken: tokens.refreshToken },
            signal,
          );
          await acceptTokens(refreshed, signal, epoch);
        } catch (error) {
          if (!operationEpochRef.current.isCurrent(epoch)) return;
          await storeMutationsRef.current.run(() => store.clear());
          if (!operationEpochRef.current.isCurrent(epoch)) return;
          dispatch({ type: "FAILED", operation, error });
          throw error;
        }
      }),
    [acceptTokens, api, store],
  );

  useEffect(() => {
    const abortController = new AbortController();
    const operationEpoch = operationEpochRef.current;
    let active = true;
    const epoch = operationEpoch.current();
    const restore = async () => {
      dispatch({ type: "RESTORE_STARTED" });
      let refreshHandled = false;
      try {
        const tokens = await store.load();
        if (!active || !operationEpoch.isCurrent(epoch)) return;
        if (tokens === null || !isRefreshTokenUsable(tokens, now())) {
          await storeMutationsRef.current.run(() => store.clear());
          if (active && operationEpoch.isCurrent(epoch)) {
            dispatch({ type: "ANONYMOUS" });
          }
          return;
        }
        if (isAccessTokenUsable(tokens, now())) {
          dispatch({ type: "TOKENS_AVAILABLE", tokens });
          await loadBootstrap(tokens.accessToken, abortController.signal, epoch);
          return;
        }
        dispatch({ type: "TOKENS_AVAILABLE", tokens });
        if (active) {
          refreshHandled = true;
          await refreshTokens(
            tokens,
            "restore",
            abortController.signal,
            epoch,
          );
        }
      } catch (error) {
        if (
          refreshHandled ||
          !active ||
          !operationEpoch.isCurrent(epoch)
        ) {
          return;
        }
        await storeMutationsRef.current.run(() => store.clear());
        if (active && operationEpoch.isCurrent(epoch)) {
          dispatch({ type: "FAILED", operation: "restore", error });
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
  }, [loadBootstrap, now, refreshTokens, store]);

  const signIn = useCallback(
    async (request: SignInRequest) => {
      const epoch = operationEpochRef.current.begin();
      dispatch({ type: "SIGN_IN_STARTED" });
      try {
        const tokens = await api.signIn(request);
        await acceptTokens(tokens, undefined, epoch);
      } catch (error) {
        if (!operationEpochRef.current.isCurrent(epoch)) return;
        await storeMutationsRef.current.run(() => store.clear());
        if (!operationEpochRef.current.isCurrent(epoch)) return;
        dispatch({ type: "FAILED", operation: "sign_in", error });
        throw error;
      }
    },
    [acceptTokens, api, store],
  );

  const refresh = useCallback(async () => {
    const tokens = stateRef.current.tokens;
    if (tokens === null || !isRefreshTokenUsable(tokens, now())) {
      operationEpochRef.current.begin();
      await storeMutationsRef.current.run(() => store.clear());
      dispatch({ type: "ANONYMOUS" });
      return;
    }
    const epoch = operationEpochRef.current.current();
    await refreshTokens(tokens, "refresh", undefined, epoch);
  }, [now, refreshTokens, store]);

  const retryBootstrap = useCallback(async () => {
    const accessToken = stateRef.current.tokens?.accessToken;
    if (accessToken === undefined) {
      dispatch({ type: "ANONYMOUS" });
      return;
    }
    const epoch = operationEpochRef.current.current();
    await loadBootstrap(accessToken, undefined, epoch);
  }, [loadBootstrap]);

  const signOut = useCallback(async () => {
    operationEpochRef.current.begin();
    const accessToken = stateRef.current.tokens?.accessToken;
    dispatch({ type: "ANONYMOUS" });
    await storeMutationsRef.current.run(() => store.clear());
    if (accessToken !== undefined) {
      try {
        await api.signOut(accessToken);
      } catch {
        // Local credentials are already removed. Remote expiry remains the fallback.
      }
    }
  }, [api, store]);

  const value = useMemo<SessionContextValue>(
    () => ({ state, signIn, refresh, retryBootstrap, signOut }),
    [refresh, retryBootstrap, signIn, signOut, state],
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
