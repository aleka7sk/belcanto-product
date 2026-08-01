import { buildExpoConfig, mergeExpoConfig } from "../../app.config";

import { readRuntimeConfig } from "./config";

describe("production application config", () => {
  const productionEnvironment = {
    EAS_BUILD_PROFILE: "production",
    EXPO_PUBLIC_API_BASE_URL: "https://api.example",
    EXPO_PUBLIC_ACTIVATION_ORIGIN: "https://app.example",
    BELCANTO_IOS_BUNDLE_IDENTIFIER: "com.belcanto.mobile",
    BELCANTO_ANDROID_PACKAGE: "com.belcanto.mobile",
  };

  it("fails closed when a production identity or origin is missing", () => {
    for (const missing of [
      "EXPO_PUBLIC_API_BASE_URL",
      "EXPO_PUBLIC_ACTIVATION_ORIGIN",
      "BELCANTO_IOS_BUNDLE_IDENTIFIER",
      "BELCANTO_ANDROID_PACKAGE",
    ] as const) {
      expect(() =>
        buildExpoConfig({ ...productionEnvironment, [missing]: undefined }),
      ).toThrow(`${missing} is required`);
    }
    expect(() =>
      buildExpoConfig({
        ...productionEnvironment,
        EXPO_PUBLIC_ACTIVATION_ORIGIN: "https://app.example:8443",
      }),
    ).toThrow("must not include a port");
  });

  it("emits verified HTTPS associations and no production custom scheme", () => {
    const config = buildExpoConfig(productionEnvironment);
    expect(config.scheme).toBeUndefined();
    expect(config.ios).toMatchObject({
      bundleIdentifier: "com.belcanto.mobile",
      associatedDomains: ["applinks:app.example"],
    });
    expect(config.android).toMatchObject({
      package: "com.belcanto.mobile",
      intentFilters: [
        expect.objectContaining({
          autoVerify: true,
          data: [
            { scheme: "https", host: "app.example", path: "/activate" },
          ],
        }),
      ],
    });
  });

  it("removes an inherited custom scheme from the merged production config", () => {
    const config = mergeExpoConfig(
      {
        name: "Inherited",
        slug: "inherited",
        scheme: "belcanto",
      },
      productionEnvironment,
    );
    expect(config.scheme).toBeUndefined();
  });

  it("derives the exact runtime trust policy from manifest extra", () => {
    const config = readRuntimeConfig(
      {},
      {
        apiBaseUrl: "https://api.example",
        activationHttpsOrigin: "https://app.example",
        allowCustomActivationScheme: false,
        production: true,
      },
    );
    expect(config.activationLinkPolicy).toEqual({
      allowedHttpsOrigins: ["https://app.example"],
      allowCustomScheme: false,
    });
  });
});
