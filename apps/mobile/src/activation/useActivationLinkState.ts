import {
  createContext,
  createElement,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type PropsWithChildren,
} from "react";

import {
  createActivationLinkObserver,
  type ActivationLinkFailure,
  type ActivationLinkPolicy,
  type LinkingAdapter,
} from "./links";

export type ActivationLinkState =
  | { status: "loading" }
  | { status: "absent" }
  | { status: "ready"; token: string }
  | { status: "invalid"; reason: ActivationLinkFailure };

const ActivationLinkContext = createContext<ActivationLinkState | null>(null);

export interface ActivationLinkProviderProps extends PropsWithChildren {
  policy: ActivationLinkPolicy;
  linking?: LinkingAdapter;
}

export function useActivationLinkState(
  policy: ActivationLinkPolicy,
  linking?: LinkingAdapter,
): ActivationLinkState {
  const [state, setState] = useState<ActivationLinkState>({ status: "loading" });
  const eventVersion = useRef(0);
  const originsKey = policy.allowedHttpsOrigins?.join("\u0000") ?? "";
  const stablePolicy = useMemo<ActivationLinkPolicy>(
    () => ({
      allowedHttpsOrigins: originsKey === "" ? [] : originsKey.split("\u0000"),
      allowCustomScheme: policy.allowCustomScheme === true,
    }),
    [originsKey, policy.allowCustomScheme],
  );

  useEffect(() => {
    let active = true;
    const observer = createActivationLinkObserver(linking, stablePolicy);
    const subscription = observer.subscribe((result) => {
      eventVersion.current += 1;
      setState(
        result.ok
          ? { status: "ready", token: result.token }
          : { status: "invalid", reason: result.reason },
      );
    });
    const initialVersion = eventVersion.current;
    void observer
      .readInitial()
      .then((result) => {
        if (!active || eventVersion.current !== initialVersion) return;
        if (result === null) {
          setState({ status: "absent" });
        } else {
          setState(
            result.ok
              ? { status: "ready", token: result.token }
              : { status: "invalid", reason: result.reason },
          );
        }
      })
      .catch(() => {
        if (active && eventVersion.current === initialVersion) {
          setState({ status: "invalid", reason: "invalid_url" });
        }
      });
    return () => {
      active = false;
      subscription.remove();
    };
  }, [linking, stablePolicy]);

  return state;
}

export function ActivationLinkProvider({
  children,
  linking,
  policy,
}: ActivationLinkProviderProps) {
  const state = useActivationLinkState(policy, linking);
  return createElement(ActivationLinkContext.Provider, { value: state }, children);
}

export function useActivationLink(): ActivationLinkState {
  const state = useContext(ActivationLinkContext);
  if (state === null) {
    throw new Error("useActivationLink must be used within ActivationLinkProvider");
  }
  return state;
}
