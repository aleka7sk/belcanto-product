import type { ConfigContext, ExpoConfig } from "expo/config";

const DEFAULT_DEVELOPMENT_API_ORIGIN = "http://localhost:8080";
const DEFAULT_DEVELOPMENT_IOS_BUNDLE_ID = "com.belcanto.mobile.dev";
const DEFAULT_DEVELOPMENT_ANDROID_PACKAGE = "com.belcanto.mobile.dev";

function requiredProductionValue(
  environment: Record<string, string | undefined>,
  name: string,
): string {
  const value = environment[name]?.trim();
  if (!value) {
    throw new Error(`${name} is required for EAS_BUILD_PROFILE=production`);
  }
  return value;
}

function cleanOrigin(value: string, name: string, allowLocalHttp: boolean): URL {
  let parsed: URL;
  try {
    parsed = new URL(value.trim());
  } catch (error) {
    throw new Error(`${name} must be an absolute URL`, { cause: error });
  }
  const local = parsed.hostname === "localhost" || parsed.hostname === "127.0.0.1";
  if (
    parsed.protocol !== "https:" &&
    !(allowLocalHttp && parsed.protocol === "http:" && local)
  ) {
    throw new Error(`${name} must use HTTPS`);
  }
  if (
    parsed.username !== "" ||
    parsed.password !== "" ||
    parsed.pathname !== "/" ||
    parsed.search !== "" ||
    parsed.hash !== ""
  ) {
    throw new Error(`${name} must be an origin without credentials or path`);
  }
  return parsed;
}

function activationOrigin(value: string): URL {
  const parsed = cleanOrigin(value, "EXPO_PUBLIC_ACTIVATION_ORIGIN", false);
  if (parsed.port !== "") {
    throw new Error(
      "EXPO_PUBLIC_ACTIVATION_ORIGIN must not include a port for verified app links",
    );
  }
  return parsed;
}

function iosBundleIdentifier(value: string): string {
  if (
    value.length > 255 ||
    !/^[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)+$/.test(value)
  ) {
    throw new Error("BELCANTO_IOS_BUNDLE_IDENTIFIER must be a reverse-DNS identifier");
  }
  return value;
}

function androidPackage(value: string): string {
  if (!/^[A-Za-z][A-Za-z0-9_]*(?:\.[A-Za-z][A-Za-z0-9_]*)+$/.test(value)) {
    throw new Error("BELCANTO_ANDROID_PACKAGE must be a reverse-DNS package name");
  }
  return value;
}

export function buildExpoConfig(
  environment: Record<string, string | undefined> = process.env,
): ExpoConfig {
  const production = environment.EAS_BUILD_PROFILE?.trim() === "production";
  const apiValue = production
    ? requiredProductionValue(environment, "EXPO_PUBLIC_API_BASE_URL")
    : environment.EXPO_PUBLIC_API_BASE_URL?.trim() ||
      DEFAULT_DEVELOPMENT_API_ORIGIN;
  const api = cleanOrigin(apiValue, "EXPO_PUBLIC_API_BASE_URL", !production);
  const configuredActivation = environment.EXPO_PUBLIC_ACTIVATION_ORIGIN?.trim();
  const activation = production
    ? activationOrigin(
        requiredProductionValue(environment, "EXPO_PUBLIC_ACTIVATION_ORIGIN"),
      )
    : configuredActivation
      ? activationOrigin(configuredActivation)
      : null;
  const bundleIdentifier = iosBundleIdentifier(
    production
      ? requiredProductionValue(environment, "BELCANTO_IOS_BUNDLE_IDENTIFIER")
      : environment.BELCANTO_IOS_BUNDLE_IDENTIFIER?.trim() ||
        DEFAULT_DEVELOPMENT_IOS_BUNDLE_ID,
  );
  const packageName = androidPackage(
    production
      ? requiredProductionValue(environment, "BELCANTO_ANDROID_PACKAGE")
      : environment.BELCANTO_ANDROID_PACKAGE?.trim() ||
        DEFAULT_DEVELOPMENT_ANDROID_PACKAGE,
  );

  const config: ExpoConfig = {
    name: "Belcanto",
    slug: "belcanto-mobile",
    version: "0.1.0",
    plugins: ["expo-router", "expo-secure-store"],
    extra: {
      apiBaseUrl: api.origin,
      allowCustomActivationScheme: !production,
      production,
      ...(activation ? { activationHttpsOrigin: activation.origin } : {}),
    },
    ios: {
      bundleIdentifier,
      associatedDomains: activation ? [`applinks:${activation.hostname}`] : [],
    },
    android: {
      package: packageName,
      intentFilters: activation
        ? [
            {
              action: "VIEW",
              autoVerify: true,
              data: [
                {
                  scheme: "https",
                  host: activation.hostname,
                  path: "/activate",
                },
              ],
              category: ["BROWSABLE", "DEFAULT"],
            },
          ]
        : [],
    },
  };
  if (!production) {
    config.scheme = "belcanto";
  }
  return config;
}

export function mergeExpoConfig(
  config: Partial<ExpoConfig>,
  environment: Record<string, string | undefined> = process.env,
): ExpoConfig {
  const built = buildExpoConfig(environment);
  const merged: ExpoConfig = { ...config, ...built };
  if (built.scheme === undefined) {
    delete merged.scheme;
  }
  return merged;
}

export default function expoConfig({ config }: ConfigContext): ExpoConfig {
  return mergeExpoConfig(config);
}
