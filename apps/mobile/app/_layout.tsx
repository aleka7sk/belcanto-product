import "react-native-gesture-handler";
import "react-native-reanimated";

import { Slot } from "expo-router";
import { useMemo } from "react";

import { ApiClient } from "@/api";
import { ActivationLinkProvider } from "@/activation/useActivationLinkState";
import { readRuntimeConfig } from "@/runtime/config";
import { SessionProvider } from "@/session";

export default function RootLayout() {
  const config = useMemo(() => readRuntimeConfig(), []);
  const api = useMemo(() => {
    return new ApiClient({
      baseUrl: config.apiBaseUrl,
      activationLinkPolicy: config.activationLinkPolicy,
    });
  }, [config]);

  return (
    <ActivationLinkProvider policy={config.activationLinkPolicy}>
      <SessionProvider api={api}>
        <Slot />
      </SessionProvider>
    </ActivationLinkProvider>
  );
}
