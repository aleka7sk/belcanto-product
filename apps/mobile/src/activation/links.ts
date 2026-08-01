import * as Linking from "expo-linking";

const OPAQUE_TOKEN_PATTERN = /^[A-Za-z0-9_-]{43}$/;
const CANONICAL_ACTIVATION_LINK_PATTERN =
  /^(?:belcanto:\/\/activate|https:\/\/[^/?#@]+\/activate)#token=[A-Za-z0-9_-]{43}$/;

export type ActivationLinkFailure =
  | "invalid_url"
  | "unsupported_protocol"
  | "untrusted_origin"
  | "custom_scheme_disabled"
  | "wrong_route"
  | "missing_token"
  | "multiple_tokens"
  | "invalid_token";

export type ActivationLinkResult =
  | { ok: true; token: string }
  | { ok: false; reason: ActivationLinkFailure };

export interface LinkingSubscription {
  remove(): void;
}

export interface LinkingAdapter {
  getInitialURL(): Promise<string | null>;
  addEventListener(
    event: "url",
    listener: (event: { url: string }) => void,
  ): LinkingSubscription;
}

export interface ActivationLinkPolicy {
  /**
   * Exact HTTPS origins accepted by the running environment. Production must
   * provide its associated-domain origin; the OS association remains the
   * first boundary, and this check prevents cross-origin link confusion.
   * HTTPS links fail closed when this list is absent or empty.
   */
  allowedHttpsOrigins?: readonly string[];
  /** Custom schemes are an explicit development-only fallback. */
  allowCustomScheme?: boolean;
}

export const expoLinkingAdapter: LinkingAdapter = {
  getInitialURL: () => Linking.getInitialURL(),
  addEventListener: (event, listener) => Linking.addEventListener(event, listener),
};

export function parseOpaqueActivationToken(value: string): string | null {
  const normalized = value.trim();
  return OPAQUE_TOKEN_PATTERN.test(normalized) ? normalized : null;
}

function isActivationLocation(url: URL): boolean {
  if (url.username !== "" || url.password !== "") return false;
  if (url.protocol === "belcanto:") {
    return (
      url.hostname === "activate" &&
      url.port === "" &&
      (url.pathname === "" || url.pathname === "/")
    );
  }
  return url.protocol === "https:" && url.pathname === "/activate";
}

function tokenValues(url: URL): string[] {
  const values = url.searchParams.getAll("token");
  const fragment = url.hash.startsWith("#") ? url.hash.slice(1) : url.hash;
  if (fragment.length > 0) {
    const fragmentParameters = new URLSearchParams(
      fragment.startsWith("?") ? fragment.slice(1) : fragment,
    );
    values.push(...fragmentParameters.getAll("token"));
  }
  return values;
}

export function parseActivationLink(
  value: string,
  policy: ActivationLinkPolicy = {},
): ActivationLinkResult {
  let url: URL;
  try {
    url = new URL(value);
  } catch {
    return { ok: false, reason: "invalid_url" };
  }
  if (url.protocol !== "belcanto:" && url.protocol !== "https:") {
    return { ok: false, reason: "unsupported_protocol" };
  }
  if (url.protocol === "belcanto:" && policy.allowCustomScheme !== true) {
    return { ok: false, reason: "custom_scheme_disabled" };
  }
  if (
    url.protocol === "https:" &&
    !policy.allowedHttpsOrigins?.includes(url.origin)
  ) {
    return { ok: false, reason: "untrusted_origin" };
  }
  if (!isActivationLocation(url)) {
    return { ok: false, reason: "wrong_route" };
  }
  const values = tokenValues(url);
  if (values.length === 0) {
    return { ok: false, reason: "missing_token" };
  }
  if (values.length !== 1) {
    return { ok: false, reason: "multiple_tokens" };
  }
  const token = parseOpaqueActivationToken(values[0]!);
  return token === null
    ? { ok: false, reason: "invalid_token" }
    : { ok: true, token };
}

export function parseCanonicalActivationLink(
  value: string,
  policy: ActivationLinkPolicy = {},
): ActivationLinkResult {
  if (!CANONICAL_ACTIVATION_LINK_PATTERN.test(value)) {
    return { ok: false, reason: "invalid_url" };
  }
  const result = parseActivationLink(value, policy);
  if (!result.ok) return result;
  const url = new URL(value);
  if (url.search !== "" || url.hash !== `#token=${result.token}`) {
    return { ok: false, reason: "wrong_route" };
  }
  return result;
}

export interface ActivationLinkObserver {
  readInitial(): Promise<ActivationLinkResult | null>;
  subscribe(listener: (result: ActivationLinkResult) => void): LinkingSubscription;
}

export function createActivationLinkObserver(
  linking: LinkingAdapter = expoLinkingAdapter,
  policy: ActivationLinkPolicy = {},
): ActivationLinkObserver {
  return {
    async readInitial() {
      const url = await linking.getInitialURL();
      return url === null ? null : parseActivationLink(url, policy);
    },
    subscribe(listener) {
      return linking.addEventListener("url", ({ url }) =>
        listener(parseActivationLink(url, policy)),
      );
    },
  };
}
