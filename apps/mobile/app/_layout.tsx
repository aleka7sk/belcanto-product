import "react-native-gesture-handler";
import "react-native-reanimated";

import { Onest_400Regular } from "@expo-google-fonts/onest/400Regular";
import { Onest_500Medium } from "@expo-google-fonts/onest/500Medium";
import { Onest_600SemiBold } from "@expo-google-fonts/onest/600SemiBold";
import { Onest_700Bold } from "@expo-google-fonts/onest/700Bold";
import { Onest_800ExtraBold } from "@expo-google-fonts/onest/800ExtraBold";
import { useFonts } from "expo-font";
import { Slot } from "expo-router";
import * as SplashScreen from "expo-splash-screen";
import { StatusBar } from "expo-status-bar";
import { useEffect, useMemo } from "react";
import { SafeAreaProvider } from "react-native-safe-area-context";

import { ActiveRoleProvider } from "@/access/activeRole";
import { ApiClient, ApiClientProvider } from "@/api";
import { ActivationLinkProvider } from "@/activation/useActivationLinkState";
import { LocaleProvider } from "@/i18n";
import { readRuntimeConfig } from "@/runtime/config";
import { SessionProvider } from "@/session";

void SplashScreen.preventAutoHideAsync().catch(() => undefined);

export default function RootLayout() {
  const [fontsLoaded, fontError] = useFonts({
    Onest_400Regular,
    Onest_500Medium,
    Onest_600SemiBold,
    Onest_700Bold,
    Onest_800ExtraBold,
  });
  const config = useMemo(() => readRuntimeConfig(), []);
  const api = useMemo(() => {
    return new ApiClient({
      baseUrl: config.apiBaseUrl,
      activationLinkPolicy: config.activationLinkPolicy,
      allowInsecureDevelopmentOrigin: !config.production,
    });
  }, [config]);
  useEffect(() => {
    if (fontsLoaded || fontError) {
      void SplashScreen.hideAsync();
    }
  }, [fontError, fontsLoaded]);

  if (!fontsLoaded && !fontError) return null;

  return (
    <SafeAreaProvider>
      <StatusBar style="light" />
      <LocaleProvider>
        <ApiClientProvider client={api}>
          <ActivationLinkProvider policy={config.activationLinkPolicy}>
            <SessionProvider api={api}>
              <ActiveRoleProvider>
                <Slot />
              </ActiveRoleProvider>
            </SessionProvider>
          </ActivationLinkProvider>
        </ApiClientProvider>
      </LocaleProvider>
    </SafeAreaProvider>
  );
}
