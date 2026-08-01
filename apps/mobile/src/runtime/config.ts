import Constants from "expo-constants";

import type { ActivationLinkPolicy } from "@/activation/links";

export interface RuntimeConfig {
  apiBaseUrl: string;
  activationHttpsOrigin: string | null;
  activationLinkPolicy: ActivationLinkPolicy;
  production: boolean;
}

type ManifestExtra = Record<string, unknown>;

function cleanHttpsOrigin(value: string): string {
  let parsed: URL;
  try {
    parsed = new URL(value.trim());
  } catch (error) {
    throw new Error("Activation origin must be an absolute URL", { cause: error });
  }
  if (
    parsed.protocol !== "https:" ||
    parsed.username !== "" ||
    parsed.password !== "" ||
    parsed.pathname !== "/" ||
    parsed.search !== "" ||
    parsed.hash !== "" ||
    parsed.port !== ""
  ) {
    throw new Error("Activation origin must be a clean HTTPS origin without a port");
  }
  return parsed.origin;
}

export function readRuntimeConfig(
  environment: Record<string, string | undefined> = process.env,
  manifestExtra: ManifestExtra = Constants.expoConfig?.extra ?? {},
): RuntimeConfig {
  const manifestApi = manifestExtra.apiBaseUrl;
  const apiBaseUrl =
    (typeof manifestApi === "string" ? manifestApi.trim() : "") ||
    environment.EXPO_PUBLIC_API_BASE_URL?.trim();
  if (!apiBaseUrl) {
    throw new Error("EXPO_PUBLIC_API_BASE_URL is required");
  }
  const production =
    manifestExtra.production === true ||
    environment.EAS_BUILD_PROFILE?.trim() === "production";
  const manifestActivation = manifestExtra.activationHttpsOrigin;
  const configuredActivation =
    (typeof manifestActivation === "string" ? manifestActivation.trim() : "") ||
    environment.EXPO_PUBLIC_ACTIVATION_ORIGIN?.trim();
  if (production && !configuredActivation) {
    throw new Error("EXPO_PUBLIC_ACTIVATION_ORIGIN is required in production");
  }
  const activationHttpsOrigin = configuredActivation
    ? cleanHttpsOrigin(configuredActivation)
    : null;
  const manifestCustom = manifestExtra.allowCustomActivationScheme;
  const allowCustomScheme = production
    ? false
    : typeof manifestCustom === "boolean"
      ? manifestCustom
      : true;
  return {
    apiBaseUrl,
    activationHttpsOrigin,
    activationLinkPolicy: {
      allowedHttpsOrigins:
        activationHttpsOrigin === null ? [] : [activationHttpsOrigin],
      allowCustomScheme,
    },
    production,
  };
}
